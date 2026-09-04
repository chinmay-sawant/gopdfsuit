//go:build js && wasm

// Package main exposes gopdflib.CompressPDF to JavaScript via WebAssembly.
// All validation, level parsing, and size caps are shared with the Go, HTTP,
// and CGO entry points through pkg/gopdflib; this shim only translates
// between JS values and Go values.
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

// compressPDF(bytes, level) returns a Uint8Array on success or {error: string}.
func compressPDF(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if rec := recover(); rec != nil {
			result = map[string]any{"error": "compression failed"}
		}
	}()

	if len(args) < 1 || args[0].Type() != js.TypeObject {
		return map[string]any{"error": "expected a Uint8Array of PDF bytes"}
	}

	n := args[0].Get("byteLength").Int()
	if n <= 0 {
		return map[string]any{"error": "empty PDF"}
	}
	if n > gopdflib.MaxCompressInputBytes {
		return map[string]any{"error": fmt.Sprintf("PDF exceeds maximum size (%d bytes)", gopdflib.MaxCompressInputBytes)}
	}

	in := make([]byte, n)
	copied := js.CopyBytesToGo(in, args[0])
	if copied == 0 {
		return map[string]any{"error": "expected a Uint8Array of PDF bytes"}
	}
	in = in[:copied]

	opts := gopdflib.CompressOptions{Level: gopdflib.CompressMedium}
	if len(args) >= 2 && args[1].Type() == js.TypeString {
		opts.Level = gopdflib.ParseCompressLevel(args[1].String())
	}

	out, err := gopdflib.CompressPDF(in, opts)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}

	dst := js.Global().Get("Uint8Array").New(len(out))
	js.CopyBytesToJS(dst, out)
	return dst
}
