//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf/font"
)

// pdfaStandardFaces mirrors font.LiberationFontMapping keys: the standard
// names the engine resolves before embedding. JS registers each TTF under
// its standard name so the generator picks up subsets with no download.
var pdfaStandardFaces = []string{
	"Helvetica", "Helvetica-Bold", "Helvetica-Oblique", "Helvetica-BoldOblique",
	"Times-Roman", "Times-Bold", "Times-Italic", "Times-BoldItalic",
	"Courier", "Courier-Bold", "Courier-Oblique", "Courier-BoldOblique",
}

func registerWasmFontBindings() {
	js.Global().Set("goRegisterFont", js.FuncOf(registerFont))
	js.Global().Set("goEnsurePDFAFonts", js.FuncOf(ensurePDFAFonts))
}

// registerFont(name, bytes) parses TTF/OTF bytes and registers them under a
// standard PDF name (e.g. "Helvetica"). Returns {registered:name} or the
// shared {code,message,error} envelope.
func registerFont(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				result = errResult(err)
			} else {
				result = errResult(invalidInput("RegisterFont", "registration failed"))
			}
		}
	}()
	const op = "RegisterFont"
	if len(args) < 2 {
		return errResult(invalidInput(op, "expected (name, bytes)"))
	}
	if args[0].Type() != js.TypeString {
		return errResult(invalidInput(op, "expected font name as a string"))
	}
	name := args[0].String()
	if name == "" {
		return errResult(invalidInput(op, "empty font name"))
	}
	data, err := copyBytes(args[1], "font bytes", op)
	if err != nil {
		return errResult(err)
	}
	if err := font.GetFontRegistry().RegisterFontFromData(name, data); err != nil {
		return errResult(err)
	}
	return map[string]any{"registered": name}
}

// ensurePDFAFonts() reports which of the 12 standard faces are registered
// for compliant generation: {registered:[...], missing:[...]}.
func ensurePDFAFonts(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				result = errResult(err)
			} else {
				result = errResult(invalidInput("EnsurePDFAFonts", "status check failed"))
			}
		}
	}()
	registry := font.GetFontRegistry()
	var registered, missing []string
	for _, name := range pdfaStandardFaces {
		if registry.HasFont(name) {
			registered = append(registered, name)
		} else {
			missing = append(missing, name)
		}
	}
	if registered == nil {
		registered = []string{}
	}
	if missing == nil {
		missing = []string{}
	}
	return valueToJS(map[string]any{"registered": registered, "missing": missing})
}
