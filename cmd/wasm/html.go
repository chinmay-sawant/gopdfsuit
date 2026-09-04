//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

func registerWasmHTMLBindings() {
	js.Global().Set("goHtmlToPDF", js.FuncOf(htmlToPDF))
	js.Global().Set("goHtmlToImage", js.FuncOf(htmlToImage))
}

// htmlSource extracts the inline HTML string (first arg) plus an optional
// options object (second arg). URL sources stay server-side.
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

// htmlToPDF(htmlString[, optionsObj]) renders inline HTML to a PDF
// Uint8Array. Options: {pageSize, orientation, marginTop/Right/Bottom/Left,
// grayscale}. URL input is server-side only.
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
		req.URL = ""
	}
	out, err := gopdflib.ConvertHTMLToPDF(req)
	if err != nil {
		return errResult(err)
	}
	return bytesToJS(out)
}

// htmlToImage(htmlString[, optionsObj]) renders inline HTML to a png/jpg
// Uint8Array. Options: {format, width, height, quality, zoom}. URL input is
// server-side only.
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
		req.URL = ""
	}
	out, err := gopdflib.ConvertHTMLToImage(req)
	if err != nil {
		return errResult(err)
	}
	return bytesToJS(out)
}
