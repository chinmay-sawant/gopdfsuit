//go:build js && wasm

// Package main exposes gopdflib.CompressPDF to JavaScript via WebAssembly.
// All validation, level parsing, and size caps are shared with the Go, HTTP,
// and CGO entry points through pkg/gopdflib; this shim only translates
// between JS values and Go values. Failures return the shared
// {code,message} envelope (plus a legacy `error` alias equal to message,
// which existing callers read).
package main

import (
	"fmt"
	"syscall/js"

	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

func main() {
	js.Global().Set("goCompressPDF", js.FuncOf(compressPDF))
	select {}
}

// errResult folds err into the shared envelope map.
func errResult(err error) map[string]any {
	env := gopdflib.EnvelopeOf(err)
	return map[string]any{"code": string(env.Code), "message": env.Message, "error": env.Message}
}

// compressPDF(bytes, level) returns a Uint8Array on success or
// {code,message,error} on failure.
func compressPDF(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if rec := recover(); rec != nil {
			result = map[string]any{"code": string(gopdflib.CodeInternal), "message": "compression failed", "error": "compression failed"}
		}
	}()

	if len(args) < 1 || args[0].Type() != js.TypeObject {
		return errResult(fmt.Errorf("%w: expected a Uint8Array of PDF bytes", gopdflib.ErrInvalidInput))
	}

	n := args[0].Get("byteLength").Int()
	if n <= 0 {
		return errResult(fmt.Errorf("%w: empty PDF", gopdflib.ErrInvalidInput))
	}
	if n > gopdflib.MaxCompressInputBytes {
		return errResult(fmt.Errorf("%w: PDF exceeds maximum size (%d bytes)", gopdflib.ErrLimitExceeded, gopdflib.MaxCompressInputBytes))
	}

	in := make([]byte, n)
	copied := js.CopyBytesToGo(in, args[0])
	if copied == 0 {
		return errResult(fmt.Errorf("%w: expected a Uint8Array of PDF bytes", gopdflib.ErrInvalidInput))
	}
	in = in[:copied]

	opts := gopdflib.CompressOptions{Level: gopdflib.CompressMedium}
	if len(args) >= 2 && args[1].Type() == js.TypeString {
		opts.Level = gopdflib.ParseCompressLevel(args[1].String())
	}

	out, err := gopdflib.CompressPDF(in, opts)
	if err != nil {
		return errResult(err)
	}

	dst := js.Global().Get("Uint8Array").New(len(out))
	js.CopyBytesToJS(dst, out)
	return dst
}
