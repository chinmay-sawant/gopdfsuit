//go:build js

package pdf

import (
	"errors"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
)

// ErrUpstream marks headless-Chrome HTML conversion as unavailable in the
// js/wasm build. The conversion backend (chinmay-sawant/gochromedp) stays
// server-side behind a //go:build !js gate on pdf.go so the browser closure
// never pulls it in. Callers that need HTML rendering must use /api/v1/*.
var ErrUpstream = errors.New("pdf: upstream failure: headless-Chrome HTML conversion is server-side only, unsupported in WASM")

// ConvertHTMLToPDF is a WASM stub. It keeps the server signature so shared
// callers compile, but always fails: HTML rendering needs headless Chrome.
func ConvertHTMLToPDF(req models.HTMLToPDFRequest) ([]byte, error) {
	_ = req
	return nil, ErrUpstream
}

// ConvertHTMLToImage is a WASM stub. It keeps the server signature so shared
// callers compile, but always fails: HTML rendering needs headless Chrome.
func ConvertHTMLToImage(req models.HTMLToImageRequest) ([]byte, error) {
	_ = req
	return nil, ErrUpstream
}
