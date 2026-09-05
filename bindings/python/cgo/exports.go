// Package main provides CGO exports for the Python bindings.
//
// Validation ownership (do not blur these lines):
//
//   - Python (bindings/python/pypdfsuit/*.py): type and shape checks on
//     caller-owned values (non-empty bytes, required fields, level names).
//     Python raises ValueError before crossing the ABI.
//   - CGO (this file): transport-only guards (nil pointers, representable
//     lengths, payload caps, malformed JSON). All semantic validation lives
//     in pkg/gopdflib, the single validating interface shared by Go,
//     Python, HTTP, and WASM callers. Do not add content checks here.
//   - gopdflib: semantics (empty PDFs, page specs, format policy, count
//     caps such as MaxMergeFileCount). CGO passes inputs through so
//     gopdflib owns the resulting errors.
package main

/*
#include <stdlib.h>
#include <string.h>

typedef struct {
    char* data;
    int length;
    char* error;
} ByteResult;

typedef struct {
    char** data;
    int* lengths;
    int count;
    char* error;
} ByteArrayResult;
*/
import "C"
import (
	"errors"
	"fmt"
	"math"
	"unsafe"

	"github.com/bytedance/sonic"
	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

// maxCGOPayloadBytes caps a generic input buffer accepted from C callers.
const maxCGOPayloadBytes = 512 << 20

// errResult builds an error ByteResult. The ABI is unchanged (the `error`
// C string), but the payload is now the shared {code,message} JSON envelope
// so Python can raise a typed subclass without parsing message text.
func errResult(err error) C.ByteResult {
	var result C.ByteResult
	result.error = C.CString(gopdflib.EnvelopeJSON(err))
	return result
}

// bytesResult copies b into C memory, returning an empty result (nil data,
// 0 length) when b is empty so callers never index b[0] out of range.
func bytesResult(b []byte) C.ByteResult {
	var result C.ByteResult
	if len(b) == 0 {
		return result
	}
	if len(b) > math.MaxInt32 {
		result.error = C.CString("output exceeds maximum representable length")
		return result
	}
	result.length = C.int(len(b))
	result.data = (*C.char)(C.malloc(C.size_t(len(b))))
	C.memcpy(unsafe.Pointer(result.data), unsafe.Pointer(&b[0]), C.size_t(len(b)))
	return result
}

// invalidArg wraps a transport-guard failure so the envelope classifies it
// as invalid_input, matching the HTTP 400 for the same malformed calls.
func invalidArg(err error) error {
	return fmt.Errorf("%w: %w", gopdflib.ErrInvalidInput, err)
}

// limitArg wraps an over-cap transport rejection so the envelope carries
// limit_exceeded, matching the HTTP 413 for oversized uploads.
func limitArg(err error) error {
	return fmt.Errorf("%w: %w", gopdflib.ErrLimitExceeded, err)
}

// checkByteLen rejects lengths that cannot be represented as C.int.
func checkByteLen(n int) error {
	if n < 0 || int64(n) > math.MaxInt32 {
		return invalidArg(errors.New("length exceeds maximum representable length"))
	}
	return nil
}

// cBytes safely copies n bytes from a C pointer, rejecting nil pointers.
func cBytes(p *C.char, n C.int) ([]byte, error) {
	if err := checkByteLen(int(n)); err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	if p == nil {
		return nil, invalidArg(errors.New("nil data pointer with non-zero length"))
	}
	return C.GoBytes(unsafe.Pointer(p), n), nil
}

// cString safely converts a C string pointer, rejecting nil pointers.
func cString(p *C.char) (string, error) {
	if p == nil {
		return "", invalidArg(errors.New("nil string pointer"))
	}
	return C.GoString(p), nil
}

// pdfBytes is the transport guard for a PDF input buffer: representable
// length within the payload cap. Empty content passes through so gopdflib
// owns the "non-empty PDF" error.
func pdfBytes(p *C.char, n C.int) ([]byte, error) {
	if err := checkByteLen(int(n)); err != nil {
		return nil, err
	}
	if int64(n) > maxCGOPayloadBytes {
		return nil, limitArg(errors.New("PDF length out of range"))
	}
	return cBytes(p, n)
}

// jsonArg parses a C JSON string argument into T via sonic (pooled decoder
// path for small args). Malformed JSON is a transport error; semantic
// validation of the decoded value belongs to gopdflib.
func jsonArg[T any](p *C.char) (T, error) {
	var v T
	s, err := cString(p)
	if err != nil {
		return v, err
	}
	if err := sonic.Unmarshal([]byte(s), &v); err != nil {
		return v, invalidArg(err)
	}
	return v, nil
}

// bytesOp is the generic adapter for byte-or-JSON-out operations: run the op
// and map its result to a ByteResult.
func bytesOp(op func() ([]byte, error)) C.ByteResult {
	out, err := op()
	if err != nil {
		return errResult(err)
	}
	return bytesResult(out)
}

// jsonOp is the generic adapter for value-out operations: run the op and
// return its result marshaled as JSON.
func jsonOp(op func() (any, error)) C.ByteResult {
	return bytesOp(func() ([]byte, error) {
		v, err := op()
		if err != nil {
			return nil, err
		}
		return sonic.Marshal(v)
	})
}

// bytesArrayOp is the generic adapter for multi-part byte-out operations
// (SplitPDF): run the op and map its parts to a ByteArrayResult.
func bytesArrayOp(op func() ([][]byte, error)) C.ByteArrayResult {
	var result C.ByteArrayResult
	fail := func(err error) C.ByteArrayResult {
		result.error = C.CString(gopdflib.EnvelopeJSON(err))
		return result
	}

	parts, err := op()
	if err != nil {
		return fail(err)
	}
	if len(parts) == 0 {
		return fail(errors.New("no output parts generated"))
	}

	result.count = C.int(len(parts))
	result.data = (**C.char)(C.malloc(C.size_t(len(parts)) * C.size_t(unsafe.Sizeof((*C.char)(nil)))))
	result.lengths = (*C.int)(C.malloc(C.size_t(len(parts)) * C.size_t(unsafe.Sizeof(C.int(0)))))

	dataSlice := unsafe.Slice(result.data, len(parts))
	lengthSlice := unsafe.Slice(result.lengths, len(parts))

	for i, part := range parts {
		if len(part) == 0 {
			continue
		}
		if err := checkByteLen(len(part)); err != nil {
			return fail(err)
		}
		lengthSlice[i] = C.int(len(part))
		dataSlice[i] = (*C.char)(C.malloc(C.size_t(len(part))))
		C.memcpy(unsafe.Pointer(dataSlice[i]), unsafe.Pointer(&part[0]), C.size_t(len(part)))
	}

	return result
}

// GeneratePDF generates a PDF from a JSON template.
// The caller must free the result using FreeBytesResult.
//
//export GeneratePDF
func GeneratePDF(jsonTemplate *C.char) C.ByteResult {
	return bytesOp(func() ([]byte, error) {
		template, err := jsonArg[gopdflib.PDFTemplate](jsonTemplate)
		if err != nil {
			return nil, err
		}
		return gopdflib.GeneratePDF(template)
	})
}

// MergePDFs merges multiple PDF files into one.
// The caller must free the result using FreeBytesResult.
//
//export MergePDFs
func MergePDFs(pdfData **C.char, pdfLengths *C.int, count C.int) C.ByteResult {
	return bytesOp(func() ([]byte, error) {
		if pdfData == nil || pdfLengths == nil {
			return nil, invalidArg(errors.New("nil array pointer"))
		}
		if count <= 0 || int64(count) >= gopdflib.MaxMergeFileCount {
			return nil, invalidArg(errors.New("invalid PDF count"))
		}

		dataSlice := unsafe.Slice(pdfData, int(count))
		lengthSlice := unsafe.Slice(pdfLengths, int(count))

		var totalBytes uint64
		for i := 0; i < int(count); i++ {
			length := int64(lengthSlice[i])
			if length < 0 || length > gopdflib.MaxMergeInputBytes {
				return nil, limitArg(errors.New("PDF part length out of range"))
			}
			totalBytes += uint64(length)
			if totalBytes > gopdflib.MaxMergeTotalInputBytes {
				return nil, limitArg(errors.New("combined PDF input length out of range"))
			}
		}

		files := make([][]byte, int(count))
		for i := 0; i < int(count); i++ {
			length := int(lengthSlice[i])
			if length == 0 {
				files[i] = nil
				continue
			}
			if dataSlice[i] == nil {
				return nil, invalidArg(errors.New("nil PDF data pointer"))
			}
			files[i] = C.GoBytes(unsafe.Pointer(dataSlice[i]), C.int(length))
		}

		return gopdflib.MergePDFs(files)
	})
}

// SplitPDF splits a PDF according to the given specification.
// The caller must free the result using FreeBytesArrayResult.
//
//export SplitPDF
func SplitPDF(pdfData *C.char, pdfLength C.int, specJSON *C.char) C.ByteArrayResult {
	return bytesArrayOp(func() ([][]byte, error) {
		file, err := pdfBytes(pdfData, pdfLength)
		if err != nil {
			return nil, err
		}
		spec, err := jsonArg[gopdflib.SplitSpec](specJSON)
		if err != nil {
			return nil, err
		}
		return gopdflib.SplitPDF(file, spec)
	})
}

// ParsePageSpec parses a page specification string.
// The caller must free the result using FreeIntArrayResult.
//
//export ParsePageSpec
func ParsePageSpec(spec *C.char, totalPages C.int) C.ByteResult {
	return bytesOp(func() ([]byte, error) {
		specStr, err := cString(spec)
		if err != nil {
			return nil, err
		}
		pages, err := gopdflib.ParsePageSpec(specStr, int(totalPages))
		if err != nil {
			return nil, err
		}
		return sonic.Marshal(pages)
	})
}

// FillPDFWithXFDF fills a PDF form with XFDF data.
// The caller must free the result using FreeBytesResult.
//
//export FillPDFWithXFDF
func FillPDFWithXFDF(pdfData *C.char, pdfLen C.int, xfdfData *C.char, xfdfLen C.int) C.ByteResult {
	return bytesOp(func() ([]byte, error) {
		pdfBytes, err := pdfBytes(pdfData, pdfLen)
		if err != nil {
			return nil, err
		}
		if int64(xfdfLen) > maxCGOPayloadBytes {
			return nil, limitArg(errors.New("XFDF length out of range"))
		}
		xfdfBytes, err := cBytes(xfdfData, xfdfLen)
		if err != nil {
			return nil, err
		}
		return gopdflib.FillPDFWithXFDF(pdfBytes, xfdfBytes)
	})
}

// CompressPDF compresses a PDF at the tier selected by the JSON options
// ({"level":"light|medium|heavy"}; empty selects Medium).
// The caller must free the result using FreeBytesResult.
//
//export CompressPDF
func CompressPDF(pdfData *C.char, pdfLen C.int, optsJSON *C.char) C.ByteResult {
	return bytesOp(func() ([]byte, error) {
		file, err := pdfBytes(pdfData, pdfLen)
		if err != nil {
			return nil, err
		}
		opts, err := jsonArg[gopdflib.CompressOptions](optsJSON)
		if err != nil {
			return nil, err
		}
		return gopdflib.CompressPDF(file, opts)
	})
}

// ConvertHTMLToPDF converts HTML to PDF.
// The caller must free the result using FreeBytesResult.
//
//export ConvertHTMLToPDF
func ConvertHTMLToPDF(requestJSON *C.char) C.ByteResult {
	return bytesOp(func() ([]byte, error) {
		req, err := jsonArg[gopdflib.HTMLToPDFRequest](requestJSON)
		if err != nil {
			return nil, err
		}
		return gopdflib.ConvertHTMLToPDF(req)
	})
}

// ConvertHTMLToImage converts HTML to an image.
// The caller must free the result using FreeBytesResult.
//
//export ConvertHTMLToImage
func ConvertHTMLToImage(requestJSON *C.char) C.ByteResult {
	return bytesOp(func() ([]byte, error) {
		req, err := jsonArg[gopdflib.HTMLToImageRequest](requestJSON)
		if err != nil {
			return nil, err
		}
		return gopdflib.ConvertHTMLToImage(req)
	})
}

// GetAvailableFonts returns the list of available fonts as JSON.
// The caller must free the result using FreeBytesResult.
//
//export GetAvailableFonts
func GetAvailableFonts() C.ByteResult {
	return jsonOp(func() (any, error) {
		return gopdflib.GetAvailableFonts(), nil
	})
}

// GetPageInfo returns metadata about PDF pages.
// The caller must free the result using FreeBytesResult.
//
//export GetPageInfo
func GetPageInfo(pdfData *C.char, pdfLen C.int) C.ByteResult {
	return jsonOp(func() (any, error) {
		pdfBytes, err := pdfBytes(pdfData, pdfLen)
		if err != nil {
			return nil, err
		}
		return gopdflib.GetPageInfo(pdfBytes)
	})
}

// ExtractTextPositions extracts text coordinates from a specific page.
// The caller must free the result using FreeBytesResult.
//
//export ExtractTextPositions
func ExtractTextPositions(pdfData *C.char, pdfLen C.int, pageNum C.int) C.ByteResult {
	return jsonOp(func() (any, error) {
		pdfBytes, err := pdfBytes(pdfData, pdfLen)
		if err != nil {
			return nil, err
		}
		return gopdflib.ExtractTextPositions(pdfBytes, int(pageNum))
	})
}

// FindTextOccurrences searches for text and returns redaction candidate rectangles.
// The caller must free the result using FreeBytesResult.
//
//export FindTextOccurrences
func FindTextOccurrences(pdfData *C.char, pdfLen C.int, searchText *C.char) C.ByteResult {
	return jsonOp(func() (any, error) {
		pdfBytes, err := pdfBytes(pdfData, pdfLen)
		if err != nil {
			return nil, err
		}
		text, err := cString(searchText)
		if err != nil {
			return nil, err
		}
		return gopdflib.FindTextOccurrences(pdfBytes, text)
	})
}

// ApplyRedactions applies redaction rectangles to the PDF.
// The caller must free the result using FreeBytesResult.
//
//export ApplyRedactions
func ApplyRedactions(pdfData *C.char, pdfLen C.int, redactionsJSON *C.char) C.ByteResult {
	return bytesOp(func() ([]byte, error) {
		pdfBytes, err := pdfBytes(pdfData, pdfLen)
		if err != nil {
			return nil, err
		}
		redactions, err := jsonArg[[]gopdflib.RedactionRect](redactionsJSON)
		if err != nil {
			return nil, err
		}
		return gopdflib.ApplyRedactions(pdfBytes, redactions)
	})
}

// ApplyRedactionsAdvanced applies a unified redaction request to the PDF.
// The caller must free the result using FreeBytesResult.
//
//export ApplyRedactionsAdvanced
func ApplyRedactionsAdvanced(pdfData *C.char, pdfLen C.int, optionsJSON *C.char) C.ByteResult {
	return bytesOp(func() ([]byte, error) {
		pdfBytes, err := pdfBytes(pdfData, pdfLen)
		if err != nil {
			return nil, err
		}
		options, err := jsonArg[gopdflib.ApplyRedactionOptions](optionsJSON)
		if err != nil {
			return nil, err
		}
		return gopdflib.ApplyRedactionsAdvanced(pdfBytes, options)
	})
}

// FreeBytesResult frees memory allocated by functions returning ByteResult.
//
//export FreeBytesResult
func FreeBytesResult(result C.ByteResult) {
	if result.data != nil {
		C.free(unsafe.Pointer(result.data))
	}
	if result.error != nil {
		C.free(unsafe.Pointer(result.error))
	}
}

// FreeBytesArrayResult frees memory allocated by functions returning ByteArrayResult.
//
//export FreeBytesArrayResult
func FreeBytesArrayResult(result C.ByteArrayResult) {
	if result.data != nil {
		dataSlice := unsafe.Slice(result.data, int(result.count))
		for i := 0; i < int(result.count); i++ {
			if dataSlice[i] != nil {
				C.free(unsafe.Pointer(dataSlice[i]))
			}
		}
		C.free(unsafe.Pointer(result.data))
	}
	if result.lengths != nil {
		C.free(unsafe.Pointer(result.lengths))
	}
	if result.error != nil {
		C.free(unsafe.Pointer(result.error))
	}
}

// Required for building as a shared library
func main() {}
