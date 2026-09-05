//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

func registerWasmHTMLBindings() {
	js.Global().Set("goHtmlToPDF", js.FuncOf(htmlToPDF))
	js.Global().Set("goHtmlToImage", js.FuncOf(htmlToImage))
}

// htmlSource extracts the inline HTML string (first arg) plus an optional
// options object (second arg). URL conversions bypass it: when the options
// object carries a non-empty `url` string field, the caller runs in URL mode
// (page HTML pre-fetched via browser fetch, req.HTML plus req.URL set
// together so the engine renders inline HTML with the page URL as base).
func htmlSource(args []js.Value, op string) (string, error) {
	if len(args) < 1 || args[0].Type() != js.TypeString {
		return "", invalidInput(op, "expected HTML content as a string")
	}
	html := args[0].String()
	if html == "" {
		return "", invalidInput(op, "empty HTML content")
	}
	return html, nil
}

// htmlOptionsURL returns the non-empty `url` string field from the options
// object arg (second arg), or "" when absent. It accepts the same object or
// JSON-string shapes as decodeJS.
func htmlOptionsURL(args []js.Value) string {
	if len(args) < 2 {
		return ""
	}
	v := args[1]
	if v.IsNull() || v.IsUndefined() {
		return ""
	}
	switch v.Type() {
	case js.TypeObject:
		if u := v.Get("url"); u.Type() == js.TypeString {
			return u.String()
		}
		return ""
	case js.TypeString:
		var m map[string]any
		if err := json.Unmarshal([]byte(v.String()), &m); err != nil {
			return ""
		}
		if s, ok := m["url"].(string); ok {
			return s
		}
		return ""
	default:
		return ""
	}
}

// promiseReasonMessage extracts a readable message from a fetch rejection
// reason (usually a TypeError DOMException carrying `message`).
func promiseReasonMessage(op string, v js.Value) string {
	if v.Type() == js.TypeObject {
		if m := v.Get("message"); m.Type() == js.TypeString && m.String() != "" {
			return m.String()
		}
		if s := v.Get("toString"); s.Type() == js.TypeFunction {
			if out := v.Call("toString"); out.Type() == js.TypeString && out.String() != "" {
				return out.String()
			}
		}
	}
	if v.Type() == js.TypeString && v.String() != "" {
		return v.String()
	}
	return op + ": fetch rejected the request"
}

// awaitPromise blocks the calling goroutine until a JS Promise settles. It
// must run off the synchronous JS-to-Go call stack (inside a goroutine
// spawned by a Promise executor, as below): blocking the sync call stack
// would deadlock the browser event loop the settlement callbacks need.
func awaitPromise(p js.Value, op string) (js.Value, error) {
	resCh := make(chan js.Value, 1)
	errCh := make(chan js.Value, 1)
	onFulfilled := js.FuncOf(func(_ js.Value, args []js.Value) any {
		resCh <- args[0]
		return nil
	})
	defer onFulfilled.Release()
	onRejected := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 0 {
			errCh <- args[0]
		} else {
			errCh <- js.Undefined()
		}
		return nil
	})
	defer onRejected.Release()
	p.Call("then", onFulfilled, onRejected)
	select {
	case v := <-resCh:
		return v, nil
	case e := <-errCh:
		return js.Undefined(), fmt.Errorf("%s: %s", op, promiseReasonMessage(op, e))
	}
}

// fetchPageHTML retrieves rawURL through the browser fetch API and returns
// the page markup as a string. The engine loader cannot fetch under
// js/wasm (it dials raw sockets for DNS/HTTP, which browsers do not allow),
// so URL mode always goes through here first. Rejection reasons are folded
// into a user-readable upstream error naming the likely cause (CORS block,
// DNS, offline) instead of leaking engine dial errors.
func fetchPageHTML(rawURL, op string) (string, error) {
	lower := strings.ToLower(strings.TrimSpace(rawURL))
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		return "", fmt.Errorf("%w: %s: URL must start with http:// or https://", gopdflib.ErrInvalidInput, op)
	}
	fetchFn := js.Global().Get("fetch")
	if fetchFn.Type() != js.TypeFunction {
		return "", fmt.Errorf("%w: %s: browser fetch is unavailable in this context", gopdflib.ErrUpstream, op)
	}
	respVal, err := awaitPromise(fetchFn.Invoke(rawURL), op)
	if err != nil {
		return "", fmt.Errorf("%w: %s: could not load %s in the browser (the site may block cross-origin fetch via CORS, or there may be no network connection): %s",
			gopdflib.ErrUpstream, op, rawURL, strings.TrimPrefix(err.Error(), op+": "))
	}
	if respVal.Type() != js.TypeObject {
		return "", fmt.Errorf("%w: %s: could not load %s in the browser (unexpected fetch response)", gopdflib.ErrUpstream, op, rawURL)
	}
	if ok := respVal.Get("ok"); ok.Type() == js.TypeBoolean && !ok.Bool() {
		status := respVal.Get("status")
		return "", fmt.Errorf("%w: %s: could not load %s in the browser (site returned HTTP status %s)",
			gopdflib.ErrUpstream, op, rawURL, status.String())
	}
	textVal, err := awaitPromise(respVal.Call("text"), op)
	if err != nil {
		return "", fmt.Errorf("%w: %s: could not read %s in the browser: %s",
			gopdflib.ErrUpstream, op, rawURL, strings.TrimPrefix(err.Error(), op+": "))
	}
	html := textVal.String()
	if strings.TrimSpace(html) == "" {
		return "", fmt.Errorf("%w: %s: %s returned an empty page", gopdflib.ErrUpstream, op, rawURL)
	}
	return html, nil
}

// htmlURLPromise runs fetch-then-convert off the sync JS call stack and
// returns a JS Promise for it. Returning a Promise (instead of blocking)
// keeps the browser event loop free so fetch settlement callbacks can run;
// the promise resolves to a Uint8Array on success or to the shared
// {code,message,error} envelope on failure, matching the sync bindings, so
// the frontend normalizes both shapes through one path. No server calls:
// the fetched markup renders through the in-browser engine with the page
// URL as base (relative subresources still resolve engine-side and stay
// subject to the same browser network limits).
func htmlURLPromise(op, pageURL string, convert func(html string) (any, error)) any {
	executor := js.FuncOf(func(_ js.Value, args []js.Value) any {
		resolve := args[0]
		go func() {
			defer func() {
				if r := recover(); r != nil {
					if err, ok := r.(error); ok {
						resolve.Invoke(errResult(err))
					} else {
						resolve.Invoke(errResult(fmt.Errorf("%w: %s: conversion failed", gopdflib.ErrInternal, op)))
					}
				}
			}()
			html, err := fetchPageHTML(pageURL, op)
			if err != nil {
				resolve.Invoke(errResult(err))
				return
			}
			out, err := convert(html)
			if err != nil {
				resolve.Invoke(errResult(err))
				return
			}
			resolve.Invoke(out)
		}()
		return nil
	})
	defer executor.Release()
	return js.Global().Get("Promise").New(executor)
}

// htmlToPDF(source[, optionsObj]) renders to a PDF Uint8Array, or to a JS
// Promise of one in URL mode (see htmlURLPromise). URL mode: when
// optionsObj carries a non-empty `url` string field, source is ignored and
// the page is fetched via browser fetch, then converted as HTML with the
// page URL as base. HTML mode: source is the inline HTML string. Options:
// {url, pageSize, orientation, marginTop/Right/Bottom/Left, grayscale}.
func htmlToPDF(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				result = errResult(err)
			} else {
				result = errResult(invalidInput("HtmlToPDF", "conversion failed"))
			}
		}
	}()
	const op = "gopdflib: HtmlToPDF"
	if u := htmlOptionsURL(args); u != "" {
		req := gopdflib.HTMLToPDFRequest{URL: u}
		if len(args) >= 2 && !args[1].IsNull() && !args[1].IsUndefined() {
			if err := decodeJS(args[1], &req, op, "PDF options"); err != nil {
				return errResult(err)
			}
			req.URL = u
			req.HTML = ""
		}
		return htmlURLPromise(op, u, func(html string) (any, error) {
			req.HTML = html
			out, err := gopdflib.ConvertHTMLToPDF(req)
			if err != nil {
				return nil, err
			}
			return bytesToJS(out), nil
		})
	}
	html, err := htmlSource(args, op)
	if err != nil {
		return errResult(err)
	}
	req := gopdflib.HTMLToPDFRequest{HTML: html}
	if len(args) >= 2 && !args[1].IsNull() && !args[1].IsUndefined() {
		if err := decodeJS(args[1], &req, op, "PDF options"); err != nil {
			return errResult(err)
		}
		req.HTML = html
	}
	out, err := gopdflib.ConvertHTMLToPDF(req)
	if err != nil {
		return errResult(err)
	}
	return bytesToJS(out)
}

// htmlToImage(source[, optionsObj]) renders to a png/jpg Uint8Array, or to
// a JS Promise of one in URL mode (see htmlURLPromise). URL mode: when
// optionsObj carries a non-empty `url` string field, source is ignored and
// the page is fetched via browser fetch, then converted as HTML with the
// page URL as base. HTML mode: source is the inline HTML string. Options:
// {url, format, width, height, quality, zoom}.
func htmlToImage(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				result = errResult(err)
			} else {
				result = errResult(invalidInput("HtmlToImage", "conversion failed"))
			}
		}
	}()
	const op = "gopdflib: HtmlToImage"
	if u := htmlOptionsURL(args); u != "" {
		req := gopdflib.HTMLToImageRequest{URL: u}
		if len(args) >= 2 && !args[1].IsNull() && !args[1].IsUndefined() {
			if err := decodeJS(args[1], &req, op, "image options"); err != nil {
				return errResult(err)
			}
			req.URL = u
			req.HTML = ""
		}
		return htmlURLPromise(op, u, func(html string) (any, error) {
			req.HTML = html
			out, err := gopdflib.ConvertHTMLToImage(req)
			if err != nil {
				return nil, err
			}
			return bytesToJS(out), nil
		})
	}
	html, err := htmlSource(args, op)
	if err != nil {
		return errResult(err)
	}
	req := gopdflib.HTMLToImageRequest{HTML: html}
	if len(args) >= 2 && !args[1].IsNull() && !args[1].IsUndefined() {
		if err := decodeJS(args[1], &req, op, "image options"); err != nil {
			return errResult(err)
		}
		req.HTML = html
	}
	out, err := gopdflib.ConvertHTMLToImage(req)
	if err != nil {
		return errResult(err)
	}
	return bytesToJS(out)
}
