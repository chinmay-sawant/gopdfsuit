//go:build js && wasm

// Package main exposes the pure-Go gopdflib surface to JavaScript via WebAssembly.
//
// Bound ops: goGeneratePDF, goMergePDFs (alias goMergePDF), goSplitPDF,
// goFillPDF, goCompressPDF, the text-path redact ops goRedactGetPageInfo,
// goRedactExtractText, goRedactFindText (alias goRedactSearch), goRedactApply,
// goRedactAdvanced, plus the font ops goRegisterFont and goEnsurePDFAFonts
// (fonts.go) for compliant generation. Chrome-backed HTML conversion, the
// CGO bindings, and the OCR subprocess stay server-side and are never
// referenced here.
//
// All validation and size caps are shared with the Go, HTTP, and CGO entry
// points through pkg/gopdflib; this shim only translates between JS values
// and Go values. Bytes cross the boundary with js.CopyBytesToGo/CopyBytesToJS.
// Failures return the shared {code,message} envelope (plus a legacy `error`
// alias equal to message, which existing callers read) via recover() plus
// gopdflib.EnvelopeOf, mirroring cmd/wasmcompress/main.go.
//
// Multi-file split results come back as a JS Array of Uint8Array; zipping is
// left to JS on purpose so archive/zip never enters the WASM closure.
package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
)

func main() {
	js.Global().Set("goGeneratePDF", js.FuncOf(generatePDF))
	js.Global().Set("goMergePDFs", js.FuncOf(mergePDFs))
	js.Global().Set("goMergePDF", js.FuncOf(mergePDFs))
	js.Global().Set("goSplitPDF", js.FuncOf(splitPDF))
	js.Global().Set("goFillPDF", js.FuncOf(fillPDF))
	js.Global().Set("goCompressPDF", js.FuncOf(compressPDF))
	js.Global().Set("goRedactGetPageInfo", js.FuncOf(redactGetPageInfo))
	js.Global().Set("goRedactExtractText", js.FuncOf(redactExtractText))
	js.Global().Set("goRedactFindText", js.FuncOf(redactFindText))
	js.Global().Set("goRedactSearch", js.FuncOf(redactFindText))
	js.Global().Set("goRedactApply", js.FuncOf(redactApply))
	js.Global().Set("goRedactAdvanced", js.FuncOf(redactAdvanced))
	registerWasmFontBindings()
	registerWasmHTMLBindings()
	select {}
}

// errResult folds err into the shared envelope map.
func errResult(err error) map[string]any {
	env := gopdflib.EnvelopeOf(err)
	return map[string]any{"code": string(env.Code), "message": env.Message, "error": env.Message}
}

func invalidInput(op, msg string) error {
	return fmt.Errorf("%w: %s: %s", gopdflib.ErrInvalidInput, op, msg)
}

// copyBytes copies a JS Uint8Array into Go memory.
func copyBytes(v js.Value, what string, op string) ([]byte, error) {
	if v.Type() != js.TypeObject {
		return nil, invalidInput(op, "expected a Uint8Array of "+what)
	}
	sizeVal := v.Get("byteLength")
	if sizeVal.Type() != js.TypeNumber {
		return nil, invalidInput(op, "expected a Uint8Array of "+what)
	}
	n := sizeVal.Int()
	if n <= 0 {
		return nil, invalidInput(op, "empty "+what)
	}
	buf := make([]byte, n)
	copied := js.CopyBytesToGo(buf, v)
	if copied == 0 {
		return nil, invalidInput(op, "expected a Uint8Array of "+what)
	}
	return buf[:copied], nil
}

func jsByteLength(v js.Value, what, op string) (int, error) {
	if v.Type() != js.TypeObject {
		return 0, invalidInput(op, "expected a Uint8Array of "+what)
	}
	sizeVal := v.Get("byteLength")
	if sizeVal.Type() != js.TypeNumber {
		return 0, invalidInput(op, "expected a Uint8Array of "+what)
	}
	n := sizeVal.Int()
	if n <= 0 {
		return 0, invalidInput(op, "empty "+what)
	}
	return n, nil
}

func copyMergeBytes(v js.Value, what, op string) ([]byte, error) {
	n, err := jsByteLength(v, what, op)
	if err != nil {
		return nil, err
	}
	if n > gopdflib.MaxMergeInputBytes {
		return nil, fmt.Errorf("%w: %s: %s exceeds %d bytes", gopdflib.ErrLimitExceeded, op, what, gopdflib.MaxMergeInputBytes)
	}
	return copyBytes(v, what, op)
}

// bytesToJS copies Go bytes out as a fresh Uint8Array.
func bytesToJS(b []byte) js.Value {
	dst := js.Global().Get("Uint8Array").New(len(b))
	js.CopyBytesToJS(dst, b)
	return dst
}

// decodeJS parses a JS object (or JSON string) into dst via encoding/json so
// the shim never depends on bytedance/sonic inside the WASM closure.
func decodeJS(v js.Value, dst any, op, what string) error {
	var raw string
	switch v.Type() {
	case js.TypeString:
		raw = v.String()
	case js.TypeObject:
		raw = js.Global().Get("JSON").Call("stringify", v).String()
	default:
		return invalidInput(op, "expected "+what+" as an object or JSON string")
	}
	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		return fmt.Errorf("%w: %s: invalid %s: %w", gopdflib.ErrInvalidInput, op, what, err)
	}
	return nil
}

// valueToJS converts a Go value into a plain JS object graph via a
// JSON round-trip. Never use for PDF bytes (they travel via bytesToJS).
func valueToJS(v any) js.Value {
	raw, err := json.Marshal(v)
	if err != nil {
		return js.Null()
	}
	return js.Global().Get("JSON").Call("parse", string(raw))
}

// generatePDF(templateObj|string) returns a Uint8Array on success or
// {code,message,error} on failure. Large binary assets should stay base64
// imagedata/fontData strings, matching the server template contract.
func generatePDF(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if rec := recover(); rec != nil {
			result = map[string]any{"code": string(gopdflib.CodeInternal), "message": "generation failed", "error": "generation failed"}
		}
	}()
	const op = "gopdflib: GeneratePDF"
	if len(args) < 1 || (args[0].Type() != js.TypeObject && args[0].Type() != js.TypeString) {
		return errResult(invalidInput(op, "expected a template object or JSON string"))
	}
	var template gopdflib.PDFTemplate
	if err := decodeJS(args[0], &template, op, "template"); err != nil {
		return errResult(err)
	}
	out, err := gopdflib.GeneratePDF(template)
	if err != nil {
		return errResult(err)
	}
	return bytesToJS(out)
}

// mergePDFs(pdfArray) takes a JS array of Uint8Array and returns one
// Uint8Array. Also registered as goMergePDF for the frontend wasmLoader.
func mergePDFs(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if rec := recover(); rec != nil {
			result = map[string]any{"code": string(gopdflib.CodeInternal), "message": "merge failed", "error": "merge failed"}
		}
	}()
	const op = "gopdflib: MergePDFs"
	if len(args) < 1 || args[0].Type() != js.TypeObject {
		return errResult(invalidInput(op, "expected an array of Uint8Array"))
	}
	isArray := js.Global().Get("Array").Call("isArray", args[0])
	if !isArray.Bool() {
		return errResult(invalidInput(op, "expected an array of Uint8Array"))
	}
	n := args[0].Get("length").Int()
	if n == 0 {
		return errResult(invalidInput(op, "needs at least 1 PDF file"))
	}
	if n >= gopdflib.MaxMergeFileCount {
		return errResult(invalidInput(op, "too many PDF files"))
	}
	var totalBytes uint64
	for i := 0; i < n; i++ {
		length, err := jsByteLength(args[0].Index(i), fmt.Sprintf("PDF bytes at index %d", i), op)
		if err != nil {
			return errResult(err)
		}
		if length > gopdflib.MaxMergeInputBytes {
			return errResult(fmt.Errorf("%w: %s: PDF bytes at index %d exceeds %d bytes", gopdflib.ErrLimitExceeded, op, i, gopdflib.MaxMergeInputBytes))
		}
		totalBytes += uint64(length)
		if totalBytes > gopdflib.MaxMergeTotalInputBytes {
			return errResult(fmt.Errorf("%w: %s: combined input exceeds %d bytes", gopdflib.ErrLimitExceeded, op, gopdflib.MaxMergeTotalInputBytes))
		}
	}
	files := make([][]byte, 0, n)
	for i := 0; i < n; i++ {
		b, err := copyMergeBytes(args[0].Index(i), fmt.Sprintf("PDF bytes at index %d", i), op)
		if err != nil {
			return errResult(err)
		}
		files = append(files, b)
	}
	out, err := gopdflib.MergePDFs(files)
	if err != nil {
		return errResult(err)
	}
	return bytesToJS(out)
}

// splitSpecFromArgs accepts either a {pages,ranges,maxPerFile} object, a
// "1-3,5" page-spec string, or the frontend (pagesString, maxPerFileString)
// pair, and normalizes to gopdflib.SplitSpec.
func splitSpecFromArgs(args []js.Value, op string) (gopdflib.SplitSpec, error) {
	var spec gopdflib.SplitSpec
	if len(args) == 0 {
		return spec, nil
	}
	if args[0].Type() == js.TypeString {
		pages, err := gopdflib.ParsePageSpec(strings.TrimSpace(args[0].String()), 0)
		if err != nil {
			return spec, err
		}
		spec.Pages = pages
		if len(args) >= 2 && args[1].Type() == js.TypeString {
			if s := strings.TrimSpace(args[1].String()); s != "" {
				max, err := strconv.Atoi(s)
				if err != nil {
					return spec, invalidInput(op, "invalid maxPerFile: "+args[1].String())
				}
				spec.MaxPerFile = max
			}
		}
		return spec, nil
	}
	if err := decodeJS(args[0], &spec, op, "split spec"); err != nil {
		return spec, err
	}
	return spec, nil
}

// splitPDF(bytes, specObj|string[, maxPerFileString]) returns a JS Array of
// Uint8Array, one entry per output part. Never a zip: archive/zip stays out
// of the WASM closure and JS does any packaging.
func splitPDF(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if rec := recover(); rec != nil {
			result = map[string]any{"code": string(gopdflib.CodeInternal), "message": "split failed", "error": "split failed"}
		}
	}()
	const op = "gopdflib: SplitPDF"
	if len(args) < 1 {
		return errResult(invalidInput(op, "expected PDF bytes and a split spec"))
	}
	in, err := copyBytes(args[0], "PDF bytes", op)
	if err != nil {
		return errResult(err)
	}
	spec, err := splitSpecFromArgs(args[1:], op)
	if err != nil {
		return errResult(err)
	}
	parts, err := gopdflib.SplitPDF(in, spec)
	if err != nil {
		return errResult(err)
	}
	arr := js.Global().Get("Array").New(len(parts))
	for i, part := range parts {
		arr.SetIndex(i, bytesToJS(part))
	}
	return arr
}

// fillPDF(pdfBytes, xfdfBytes) returns the filled PDF as a Uint8Array.
// Note: like the server path, fill relies on /NeedAppearances and does not
// rewrite compressed object streams.
func fillPDF(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if rec := recover(); rec != nil {
			result = map[string]any{"code": string(gopdflib.CodeInternal), "message": "fill failed", "error": "fill failed"}
		}
	}()
	const op = "gopdflib: FillPDFWithXFDF"
	if len(args) < 2 {
		return errResult(invalidInput(op, "expected PDF bytes and XFDF bytes"))
	}
	pdfBytes, err := copyBytes(args[0], "PDF bytes", op)
	if err != nil {
		return errResult(err)
	}
	xfdfBytes, err := copyBytes(args[1], "XFDF bytes", op)
	if err != nil {
		return errResult(err)
	}
	out, err := gopdflib.FillPDFWithXFDF(pdfBytes, xfdfBytes)
	if err != nil {
		return errResult(err)
	}
	return bytesToJS(out)
}

// compressPDF(bytes, level) returns a Uint8Array on success or
// {code,message,error} on failure. Level accepts the light|medium|heavy
// strings, the 1|2|3 tier numbers (the frontend passes ints), and numeric
// strings; unknown values are rejected instead of silently defaulting.
func compressPDF(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if rec := recover(); rec != nil {
			result = map[string]any{"code": string(gopdflib.CodeInternal), "message": "compression failed", "error": "compression failed"}
		}
	}()
	const op = "gopdflib: CompressPDF"
	if len(args) < 1 {
		return errResult(invalidInput(op, "expected a Uint8Array of PDF bytes"))
	}
	n := 0
	if args[0].Type() == js.TypeObject {
		if sizeVal := args[0].Get("byteLength"); sizeVal.Type() == js.TypeNumber {
			n = sizeVal.Int()
		}
	}
	if n <= 0 {
		return errResult(invalidInput(op, "expected a Uint8Array of PDF bytes"))
	}
	if n > gopdflib.MaxCompressInputBytes {
		return errResult(fmt.Errorf("%w: %s: PDF exceeds maximum size (%d bytes)", gopdflib.ErrLimitExceeded, op, gopdflib.MaxCompressInputBytes))
	}
	in, err := copyBytes(args[0], "PDF bytes", op)
	if err != nil {
		return errResult(err)
	}

	var lvl any
	if len(args) >= 2 {
		switch args[1].Type() {
		case js.TypeString:
			lvl = args[1].String()
		case js.TypeNumber:
			lvl = args[1].Int()
		}
	}
	level, err := gopdflib.ToServerLevel(lvl)
	if err != nil {
		return errResult(fmt.Errorf("%w: %s: %s", gopdflib.ErrInvalidInput, op, err.Error()))
	}
	out, err := gopdflib.CompressPDF(in, gopdflib.CompressOptions{Level: level})
	if err != nil {
		return errResult(err)
	}
	return bytesToJS(out)
}

// redactGetPageInfo(pdfBytes) returns {totalPages, pages:[...]}.
func redactGetPageInfo(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if rec := recover(); rec != nil {
			result = map[string]any{"code": string(gopdflib.CodeInternal), "message": "page info failed", "error": "page info failed"}
		}
	}()
	const op = "gopdflib: GetPageInfo"
	if len(args) < 1 {
		return errResult(invalidInput(op, "expected PDF bytes"))
	}
	in, err := copyBytes(args[0], "PDF bytes", op)
	if err != nil {
		return errResult(err)
	}
	info, err := gopdflib.GetPageInfo(in)
	if err != nil {
		return errResult(err)
	}
	return valueToJS(info)
}

// redactExtractText(pdfBytes, pageNum) returns [{text,x,y,width,height}].
func redactExtractText(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if rec := recover(); rec != nil {
			result = map[string]any{"code": string(gopdflib.CodeInternal), "message": "text extraction failed", "error": "text extraction failed"}
		}
	}()
	const op = "gopdflib: ExtractTextPositions"
	if len(args) < 1 {
		return errResult(invalidInput(op, "expected PDF bytes and a page number"))
	}
	in, err := copyBytes(args[0], "PDF bytes", op)
	if err != nil {
		return errResult(err)
	}
	pageNum := 1
	if len(args) >= 2 {
		switch args[1].Type() {
		case js.TypeNumber:
			pageNum = args[1].Int()
		case js.TypeString:
			p, convErr := strconv.Atoi(strings.TrimSpace(args[1].String()))
			if convErr != nil {
				return errResult(invalidInput(op, "invalid page number"))
			}
			pageNum = p
		default:
			return errResult(invalidInput(op, "expected a page number"))
		}
	}
	positions, err := gopdflib.ExtractTextPositions(in, pageNum)
	if err != nil {
		return errResult(err)
	}
	return valueToJS(positions)
}

// redactTerms extracts search terms from a string, a {text} object, or an
// array mixing both.
func redactTerms(v js.Value, op string) ([]string, error) {
	switch v.Type() {
	case js.TypeString:
		s := strings.TrimSpace(v.String())
		if s == "" {
			return nil, invalidInput(op, "needs non-empty searchText")
		}
		return []string{s}, nil
	case js.TypeObject:
		raw := js.Global().Get("JSON").Call("stringify", v).String()
		var single map[string]any
		if err := json.Unmarshal([]byte(raw), &single); err == nil {
			if t, ok := single["text"].(string); ok && strings.TrimSpace(t) != "" {
				return []string{t}, nil
			}
		}
		var items []any
		if err := json.Unmarshal([]byte(raw), &items); err != nil {
			return nil, invalidInput(op, "expected a search string or array of search strings")
		}
		terms := make([]string, 0, len(items))
		for _, item := range items {
			switch t := item.(type) {
			case string:
				if strings.TrimSpace(t) != "" {
					terms = append(terms, t)
				}
			case map[string]any:
				if s, ok := t["text"].(string); ok && strings.TrimSpace(s) != "" {
					terms = append(terms, s)
				}
			}
		}
		if len(terms) == 0 {
			return nil, invalidInput(op, "needs non-empty searchText")
		}
		return terms, nil
	default:
		return nil, invalidInput(op, "expected a search string or array of search strings")
	}
}

// redactFindText(pdfBytes, searchText|terms) returns [{pageNum,x,y,width,height}].
// Also registered as goRedactSearch for the frontend wasmLoader.
func redactFindText(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if rec := recover(); rec != nil {
			result = map[string]any{"code": string(gopdflib.CodeInternal), "message": "text search failed", "error": "text search failed"}
		}
	}()
	const op = "gopdflib: FindTextOccurrences"
	if len(args) < 2 {
		return errResult(invalidInput(op, "expected PDF bytes and search text"))
	}
	in, err := copyBytes(args[0], "PDF bytes", op)
	if err != nil {
		return errResult(err)
	}
	terms, err := redactTerms(args[1], op)
	if err != nil {
		return errResult(err)
	}
	combined := make([]gopdflib.RedactionRect, 0)
	for _, term := range terms {
		rects, err := gopdflib.FindTextOccurrences(in, term)
		if err != nil {
			return errResult(err)
		}
		combined = append(combined, rects...)
	}
	return valueToJS(combined)
}

// redactOptionsFromArgs accepts either a single options object
// {blocks,textSearch,mode,password} or the frontend
// (bytes, blocks, textQueries, mode) tuple.
func redactOptionsFromArgs(args []js.Value, op string) (gopdflib.ApplyRedactionOptions, error) {
	var options gopdflib.ApplyRedactionOptions
	if len(args) == 1 {
		if err := decodeJS(args[0], &options, op, "redaction options"); err != nil {
			return options, err
		}
		return options, nil
	}
	if args[0].Type() != js.TypeObject {
		return options, invalidInput(op, "expected redaction blocks as an array or options object")
	}
	if err := decodeJS(args[0], &options.Blocks, op, "redaction blocks"); err != nil {
		return options, err
	}
	if len(args) >= 2 && args[1].Type() != js.TypeUndefined && args[1].Type() != js.TypeNull {
		var queries []gopdflib.RedactionTextQuery
		raw := ""
		switch args[1].Type() {
		case js.TypeString:
			raw = `["` + args[1].String() + `"]`
		default:
			raw = js.Global().Get("JSON").Call("stringify", args[1]).String()
		}
		var items []any
		if err := json.Unmarshal([]byte(raw), &items); err != nil {
			return options, invalidInput(op, "invalid text queries")
		}
		for _, item := range items {
			switch t := item.(type) {
			case string:
				if strings.TrimSpace(t) != "" {
					queries = append(queries, gopdflib.RedactionTextQuery{Text: t})
				}
			case map[string]any:
				if s, ok := t["text"].(string); ok && strings.TrimSpace(s) != "" {
					queries = append(queries, gopdflib.RedactionTextQuery{Text: s})
				}
			}
		}
		options.TextSearch = queries
	}
	if len(args) >= 3 && args[2].Type() == js.TypeString {
		options.Mode = args[2].String()
	}
	return options, nil
}

// redactApply(pdfBytes, blocks|options[, textQueries, mode]) applies visual
// redactions and returns the redacted PDF as a Uint8Array.
func redactApply(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if rec := recover(); rec != nil {
			result = map[string]any{"code": string(gopdflib.CodeInternal), "message": "redaction failed", "error": "redaction failed"}
		}
	}()
	const op = "gopdflib: ApplyRedactions"
	if len(args) < 2 {
		return errResult(invalidInput(op, "expected PDF bytes and redaction blocks"))
	}
	in, err := copyBytes(args[0], "PDF bytes", op)
	if err != nil {
		return errResult(err)
	}
	options, err := redactOptionsFromArgs(args[1:], op)
	if err != nil {
		return errResult(err)
	}
	if options.OCR != nil && options.OCR.Enabled {
		return errResult(invalidInput(op, "OCR redaction is unsupported in WASM (server-side only)"))
	}
	var out []byte
	if len(options.TextSearch) > 0 || options.Mode != "" || options.Password != "" {
		out, err = gopdflib.ApplyRedactionsAdvanced(in, options)
	} else {
		out, err = gopdflib.ApplyRedactions(in, options.Blocks)
	}
	if err != nil {
		return errResult(err)
	}
	return bytesToJS(out)
}

// redactAdvanced(pdfBytes, options) applies redactions and returns
// {pdf: Uint8Array, report: {...}}. OCR-enabled options are rejected: the OCR
// subprocess cannot run in the browser, text path only.
func redactAdvanced(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if rec := recover(); rec != nil {
			result = map[string]any{"code": string(gopdflib.CodeInternal), "message": "redaction failed", "error": "redaction failed"}
		}
	}()
	const op = "gopdflib: ApplyRedactionsAdvancedWithReport"
	if len(args) < 2 {
		return errResult(invalidInput(op, "expected PDF bytes and redaction options"))
	}
	in, err := copyBytes(args[0], "PDF bytes", op)
	if err != nil {
		return errResult(err)
	}
	var options gopdflib.ApplyRedactionOptions
	if err := decodeJS(args[1], &options, op, "redaction options"); err != nil {
		return errResult(err)
	}
	if options.OCR != nil && options.OCR.Enabled {
		return errResult(invalidInput(op, "OCR redaction is unsupported in WASM (server-side only)"))
	}
	out, report, err := gopdflib.ApplyRedactionsAdvancedWithReport(in, options)
	if err != nil {
		return errResult(err)
	}
	obj := js.Global().Get("Object").New()
	obj.Set("pdf", bytesToJS(out))
	obj.Set("report", valueToJS(report))
	return obj
}
