// Package main provides CGO exports for the Python bindings.
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
	"encoding/json"
	"errors"
	"math"
	"unsafe"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/merge"
	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

// maxCGOPayloadBytes caps a single input buffer accepted from C callers
// (512 MiB, matching the MergePDFs per-part limit).
const maxCGOPayloadBytes = 512 << 20

// errResult builds an error ByteResult.
func errResult(err error) C.ByteResult {
	var result C.ByteResult
	result.error = C.CString(err.Error())
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

// checkByteLen rejects lengths that cannot be represented as C.int.
func checkByteLen(n int) error {
	if n < 0 || int64(n) > math.MaxInt32 {
		return errors.New("length exceeds maximum representable length")
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
		return nil, errors.New("nil data pointer with non-zero length")
	}
	return C.GoBytes(unsafe.Pointer(p), n), nil
}

// cString safely converts a C string pointer, rejecting nil pointers.
func cString(p *C.char) (string, error) {
	if p == nil {
		return "", errors.New("nil string pointer")
	}
	return C.GoString(p), nil
}

// GeneratePDF generates a PDF from a JSON template.
// The caller must free the result using FreeBytesResult.
//
//export GeneratePDF
func GeneratePDF(jsonTemplate *C.char) C.ByteResult {
	goTemplate, err := cString(jsonTemplate)
	if err != nil {
		return errResult(err)
	}
	var template gopdflib.PDFTemplate
	if err := json.Unmarshal([]byte(goTemplate), &template); err != nil {
		return errResult(err)
	}

	pdfBytes, err := gopdflib.GeneratePDF(template)
	if err != nil {
		return errResult(err)
	}

	return bytesResult(pdfBytes)
}

// MergePDFs merges multiple PDF files into one.
// The caller must free the result using FreeBytesResult.
//
//export MergePDFs
func MergePDFs(pdfData **C.char, pdfLengths *C.int, count C.int) C.ByteResult {
	if pdfData == nil || pdfLengths == nil {
		return errResult(errors.New("nil array pointer"))
	}
	if count <= 0 || int64(count) >= 1<<16 {
		return errResult(errors.New("invalid PDF count (must be 0 < count < 65536)"))
	}

	// Convert C arrays to Go slices
	dataSlice := unsafe.Slice(pdfData, int(count))
	lengthSlice := unsafe.Slice(pdfLengths, int(count))

	files := make([][]byte, int(count))
	for i := 0; i < int(count); i++ {
		length := int(lengthSlice[i])
		if length < 0 || length > maxCGOPayloadBytes {
			return errResult(errors.New("PDF part length out of range"))
		}
		if length == 0 {
			files[i] = nil
			continue
		}
		if dataSlice[i] == nil {
			return errResult(errors.New("nil PDF data pointer"))
		}
		files[i] = C.GoBytes(unsafe.Pointer(dataSlice[i]), C.int(length))
	}

	merged, err := gopdflib.MergePDFs(files)
	if err != nil {
		return errResult(err)
	}

	return bytesResult(merged)
}

// SplitPDF splits a PDF according to the given specification.
// The caller must free the result using FreeBytesArrayResult.
//
//export SplitPDF
func SplitPDF(pdfData *C.char, pdfLength C.int, specJSON *C.char) C.ByteArrayResult {
	var result C.ByteArrayResult

	file, err := cBytes(pdfData, pdfLength)
	if err != nil {
		result.error = C.CString(err.Error())
		return result
	}
	specStr, err := cString(specJSON)
	if err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	var spec merge.SplitSpec
	if err := json.Unmarshal([]byte(specStr), &spec); err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	parts, err := gopdflib.SplitPDF(file, spec)
	if err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	if len(parts) == 0 {
		result.error = C.CString("no output parts generated")
		return result
	}

	// Allocate arrays for data and lengths
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
			result.error = C.CString(err.Error())
			return result
		}
		lengthSlice[i] = C.int(len(part))
		dataSlice[i] = (*C.char)(C.malloc(C.size_t(len(part))))
		C.memcpy(unsafe.Pointer(dataSlice[i]), unsafe.Pointer(&part[0]), C.size_t(len(part)))
	}

	return result
}

// ParsePageSpec parses a page specification string.
// The caller must free the result using FreeIntArrayResult.
//
//export ParsePageSpec
func ParsePageSpec(spec *C.char, totalPages C.int) C.ByteResult {
	specStr, err := cString(spec)
	if err != nil {
		return errResult(err)
	}
	pages, err := gopdflib.ParsePageSpec(specStr, int(totalPages))
	if err != nil {
		return errResult(err)
	}

	// Return pages as JSON array
	pagesJSON, err := json.Marshal(pages)
	if err != nil {
		return errResult(err)
	}

	return bytesResult(pagesJSON)
}

// FillPDFWithXFDF fills a PDF form with XFDF data.
// The caller must free the result using FreeBytesResult.
//
//export FillPDFWithXFDF
func FillPDFWithXFDF(pdfData *C.char, pdfLen C.int, xfdfData *C.char, xfdfLen C.int) C.ByteResult {
	pdfBytes, err := cBytes(pdfData, pdfLen)
	if err != nil {
		return errResult(err)
	}
	xfdfBytes, err := cBytes(xfdfData, xfdfLen)
	if err != nil {
		return errResult(err)
	}

	filled, err := gopdflib.FillPDFWithXFDF(pdfBytes, xfdfBytes)
	if err != nil {
		return errResult(err)
	}

	return bytesResult(filled)
}

// ConvertHTMLToPDF converts HTML to PDF.
// The caller must free the result using FreeBytesResult.
//
//export ConvertHTMLToPDF
func ConvertHTMLToPDF(requestJSON *C.char) C.ByteResult {
	reqStr, err := cString(requestJSON)
	if err != nil {
		return errResult(err)
	}
	var req gopdflib.HTMLToPDFRequest
	if err := json.Unmarshal([]byte(reqStr), &req); err != nil {
		return errResult(err)
	}

	pdfBytes, err := gopdflib.ConvertHTMLToPDF(req)
	if err != nil {
		return errResult(err)
	}

	return bytesResult(pdfBytes)
}

// ConvertHTMLToImage converts HTML to an image.
// The caller must free the result using FreeBytesResult.
//
//export ConvertHTMLToImage
func ConvertHTMLToImage(requestJSON *C.char) C.ByteResult {
	reqStr, err := cString(requestJSON)
	if err != nil {
		return errResult(err)
	}
	var req gopdflib.HTMLToImageRequest
	if err := json.Unmarshal([]byte(reqStr), &req); err != nil {
		return errResult(err)
	}

	imgBytes, err := gopdflib.ConvertHTMLToImage(req)
	if err != nil {
		return errResult(err)
	}

	return bytesResult(imgBytes)
}

// GetAvailableFonts returns the list of available fonts as JSON.
// The caller must free the result using FreeBytesResult.
//
//export GetAvailableFonts
func GetAvailableFonts() C.ByteResult {
	fonts := gopdflib.GetAvailableFonts()
	fontsJSON, err := json.Marshal(fonts)
	if err != nil {
		return errResult(err)
	}

	return bytesResult(fontsJSON)
}

// GetPageInfo returns metadata about PDF pages.
// The caller must free the result using FreeBytesResult.
//
//export GetPageInfo
func GetPageInfo(pdfData *C.char, pdfLen C.int) C.ByteResult {
	pdfBytes, err := cBytes(pdfData, pdfLen)
	if err != nil {
		return errResult(err)
	}
	info, err := gopdflib.GetPageInfo(pdfBytes)
	if err != nil {
		return errResult(err)
	}

	infoJSON, err := json.Marshal(info)
	if err != nil {
		return errResult(err)
	}

	return bytesResult(infoJSON)
}

// ExtractTextPositions extracts text coordinates from a specific page.
// The caller must free the result using FreeBytesResult.
//
//export ExtractTextPositions
func ExtractTextPositions(pdfData *C.char, pdfLen C.int, pageNum C.int) C.ByteResult {
	pdfBytes, err := cBytes(pdfData, pdfLen)
	if err != nil {
		return errResult(err)
	}
	positions, err := gopdflib.ExtractTextPositions(pdfBytes, int(pageNum))
	if err != nil {
		return errResult(err)
	}

	posJSON, err := json.Marshal(positions)
	if err != nil {
		return errResult(err)
	}

	return bytesResult(posJSON)
}

// FindTextOccurrences searches for text and returns redaction candidate rectangles.
// The caller must free the result using FreeBytesResult.
//
//export FindTextOccurrences
func FindTextOccurrences(pdfData *C.char, pdfLen C.int, searchText *C.char) C.ByteResult {
	pdfBytes, err := cBytes(pdfData, pdfLen)
	if err != nil {
		return errResult(err)
	}
	text, err := cString(searchText)
	if err != nil {
		return errResult(err)
	}
	rects, err := gopdflib.FindTextOccurrences(pdfBytes, text)
	if err != nil {
		return errResult(err)
	}

	rectsJSON, err := json.Marshal(rects)
	if err != nil {
		return errResult(err)
	}

	return bytesResult(rectsJSON)
}

// ApplyRedactions applies redaction rectangles to the PDF.
// The caller must free the result using FreeBytesResult.
//
//export ApplyRedactions
func ApplyRedactions(pdfData *C.char, pdfLen C.int, redactionsJSON *C.char) C.ByteResult {
	pdfBytes, err := cBytes(pdfData, pdfLen)
	if err != nil {
		return errResult(err)
	}
	redactionsStr, err := cString(redactionsJSON)
	if err != nil {
		return errResult(err)
	}

	var redactions []gopdflib.RedactionRect
	if err := json.Unmarshal([]byte(redactionsStr), &redactions); err != nil {
		return errResult(err)
	}

	out, err := gopdflib.ApplyRedactions(pdfBytes, redactions)
	if err != nil {
		return errResult(err)
	}

	return bytesResult(out)
}

// ApplyRedactionsAdvanced applies a unified redaction request to the PDF.
// The caller must free the result using FreeBytesResult.
//
//export ApplyRedactionsAdvanced
func ApplyRedactionsAdvanced(pdfData *C.char, pdfLen C.int, optionsJSON *C.char) C.ByteResult {
	pdfBytes, err := cBytes(pdfData, pdfLen)
	if err != nil {
		return errResult(err)
	}
	optionsStr, err := cString(optionsJSON)
	if err != nil {
		return errResult(err)
	}

	var options gopdflib.ApplyRedactionOptions
	if err := json.Unmarshal([]byte(optionsStr), &options); err != nil {
		return errResult(err)
	}

	out, err := gopdflib.ApplyRedactionsAdvanced(pdfBytes, options)
	if err != nil {
		return errResult(err)
	}

	return bytesResult(out)
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
