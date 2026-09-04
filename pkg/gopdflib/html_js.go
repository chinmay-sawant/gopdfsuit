//go:build js

package gopdflib

import "fmt"

// ConvertHTMLToPDF is a WASM stub. HTML rendering needs headless Chrome via
// chinmay-sawant/gochromedp, which stays server-side behind //go:build !js,
// so browser callers always get ErrUpstream. Use POST /api/v1/htmltopdf.
// Empty requests still report ErrInvalidInput, matching the server contract.
func ConvertHTMLToPDF(req HTMLToPDFRequest) ([]byte, error) {
	const op = "gopdflib: ConvertHTMLToPDF"
	if req.HTML == "" && req.URL == "" {
		return nil, invalidInputError(op, "needs HTML content or a URL")
	}
	return nil, fmt.Errorf("%w: %s: headless-Chrome conversion is server-side only", ErrUpstream, op)
}

// ConvertHTMLToImage is a WASM stub. HTML rendering needs headless Chrome via
// chinmay-sawant/gochromedp, which stays server-side behind //go:build !js,
// so browser callers always get ErrUpstream. Use POST /api/v1/htmltoimage.
// Empty requests still report ErrInvalidInput, matching the server contract.
func ConvertHTMLToImage(req HTMLToImageRequest) ([]byte, error) {
	const op = "gopdflib: ConvertHTMLToImage"
	if req.HTML == "" && req.URL == "" {
		return nil, invalidInputError(op, "needs HTML content or a URL")
	}
	return nil, fmt.Errorf("%w: %s: headless-Chrome conversion is server-side only", ErrUpstream, op)
}
