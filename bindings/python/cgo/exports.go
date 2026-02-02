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
	"unsafe"

	"github.com/chinmay-sawant/gopdfsuit/v4/internal/pdf/merge"
	"github.com/chinmay-sawant/gopdfsuit/v4/pkg/gopdflib"
)

// GeneratePDF generates a PDF from a JSON template.
// The caller must free the result using FreeBytesResult.
//
//export GeneratePDF
func GeneratePDF(jsonTemplate *C.char) C.ByteResult {
	var result C.ByteResult

	goTemplate := C.GoString(jsonTemplate)
	var template gopdflib.PDFTemplate
	if err := json.Unmarshal([]byte(goTemplate), &template); err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	pdfBytes, err := gopdflib.GeneratePDF(template)
	if err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	// Allocate C memory and copy data
	result.length = C.int(len(pdfBytes))
	result.data = (*C.char)(C.malloc(C.size_t(len(pdfBytes))))
	C.memcpy(unsafe.Pointer(result.data), unsafe.Pointer(&pdfBytes[0]), C.size_t(len(pdfBytes)))

	return result
}

// MergePDFs merges multiple PDF files into one.
// The caller must free the result using FreeBytesResult.
//
//export MergePDFs
func MergePDFs(pdfData **C.char, pdfLengths *C.int, count C.int) C.ByteResult {
	var result C.ByteResult

	if count <= 0 {
		result.error = C.CString("no PDF files provided")
		return result
	}

	// Convert C arrays to Go slices
	dataSlice := unsafe.Slice(pdfData, int(count))
	lengthSlice := unsafe.Slice(pdfLengths, int(count))

	files := make([][]byte, int(count))
	for i := 0; i < int(count); i++ {
		length := int(lengthSlice[i])
		files[i] = C.GoBytes(unsafe.Pointer(dataSlice[i]), C.int(length))
	}

	merged, err := gopdflib.MergePDFs(files)
	if err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	result.length = C.int(len(merged))
	result.data = (*C.char)(C.malloc(C.size_t(len(merged))))
	C.memcpy(unsafe.Pointer(result.data), unsafe.Pointer(&merged[0]), C.size_t(len(merged)))

	return result
}

// SplitPDF splits a PDF according to the given specification.
// The caller must free the result using FreeBytesArrayResult.
//
//export SplitPDF
func SplitPDF(pdfData *C.char, pdfLength C.int, specJSON *C.char) C.ByteArrayResult {
	var result C.ByteArrayResult

	file := C.GoBytes(unsafe.Pointer(pdfData), pdfLength)
	specStr := C.GoString(specJSON)

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
	var result C.ByteResult

	specStr := C.GoString(spec)
	pages, err := gopdflib.ParsePageSpec(specStr, int(totalPages))
	if err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	// Return pages as JSON array
	pagesJSON, err := json.Marshal(pages)
	if err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	result.length = C.int(len(pagesJSON))
	result.data = (*C.char)(C.malloc(C.size_t(len(pagesJSON))))
	C.memcpy(unsafe.Pointer(result.data), unsafe.Pointer(&pagesJSON[0]), C.size_t(len(pagesJSON)))

	return result
}

// FillPDFWithXFDF fills a PDF form with XFDF data.
// The caller must free the result using FreeBytesResult.
//
//export FillPDFWithXFDF
func FillPDFWithXFDF(pdfData *C.char, pdfLen C.int, xfdfData *C.char, xfdfLen C.int) C.ByteResult {
	var result C.ByteResult

	pdfBytes := C.GoBytes(unsafe.Pointer(pdfData), pdfLen)
	xfdfBytes := C.GoBytes(unsafe.Pointer(xfdfData), xfdfLen)

	filled, err := gopdflib.FillPDFWithXFDF(pdfBytes, xfdfBytes)
	if err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	result.length = C.int(len(filled))
	result.data = (*C.char)(C.malloc(C.size_t(len(filled))))
	C.memcpy(unsafe.Pointer(result.data), unsafe.Pointer(&filled[0]), C.size_t(len(filled)))

	return result
}

// ConvertHTMLToPDF converts HTML to PDF.
// The caller must free the result using FreeBytesResult.
//
//export ConvertHTMLToPDF
func ConvertHTMLToPDF(requestJSON *C.char) C.ByteResult {
	var result C.ByteResult

	reqStr := C.GoString(requestJSON)
	var req gopdflib.HtmlToPDFRequest
	if err := json.Unmarshal([]byte(reqStr), &req); err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	pdfBytes, err := gopdflib.ConvertHTMLToPDF(req)
	if err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	result.length = C.int(len(pdfBytes))
	result.data = (*C.char)(C.malloc(C.size_t(len(pdfBytes))))
	C.memcpy(unsafe.Pointer(result.data), unsafe.Pointer(&pdfBytes[0]), C.size_t(len(pdfBytes)))

	return result
}

// ConvertHTMLToImage converts HTML to an image.
// The caller must free the result using FreeBytesResult.
//
//export ConvertHTMLToImage
func ConvertHTMLToImage(requestJSON *C.char) C.ByteResult {
	var result C.ByteResult

	reqStr := C.GoString(requestJSON)
	var req gopdflib.HtmlToImageRequest
	if err := json.Unmarshal([]byte(reqStr), &req); err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	imgBytes, err := gopdflib.ConvertHTMLToImage(req)
	if err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	result.length = C.int(len(imgBytes))
	result.data = (*C.char)(C.malloc(C.size_t(len(imgBytes))))
	C.memcpy(unsafe.Pointer(result.data), unsafe.Pointer(&imgBytes[0]), C.size_t(len(imgBytes)))

	return result
}

// GetAvailableFonts returns the list of available fonts as JSON.
// The caller must free the result using FreeBytesResult.
//
//export GetAvailableFonts
func GetAvailableFonts() C.ByteResult {
	var result C.ByteResult

	fonts := gopdflib.GetAvailableFonts()
	fontsJSON, err := json.Marshal(fonts)
	if err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	result.length = C.int(len(fontsJSON))
	result.data = (*C.char)(C.malloc(C.size_t(len(fontsJSON))))
	C.memcpy(unsafe.Pointer(result.data), unsafe.Pointer(&fontsJSON[0]), C.size_t(len(fontsJSON)))

	return result
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
