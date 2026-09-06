package handlers

import (
	"io"
	"sync"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/models"
)

var templatePDFPool = sync.Pool{
	New: func() any {
		return new(models.PDFTemplate)
	},
}

// resetTemplate clears a pooled PDFTemplate before unmarshal (and before Put) so
// omitted JSON fields do not leak from prior requests while still retaining
// hot backing arrays for the next pooled decode.
func resetTemplate(t *models.PDFTemplate) {
	if t == nil {
		return
	}
	t.ResetForReuse()
}

// acquireTemplate checks a cleared PDFTemplate out of the pool.
func acquireTemplate() *models.PDFTemplate {
	t := templatePDFPool.Get().(*models.PDFTemplate)
	resetTemplate(t)
	return t
}

// releaseTemplate clears a PDFTemplate and returns it to the pool.
func releaseTemplate(t *models.PDFTemplate) {
	resetTemplate(t)
	templatePDFPool.Put(t)
}

// decodeTemplate is the single decode entry point for the generate path. It
// owns tier policy (hft fast path vs pooled retail path vs streaming
// fallback), the pooled body buffers, and the HFT shell - all implemented in
// json_decode.go / hft_decode.go. Hot-path allocation behavior is unchanged:
// small bodies reuse bodyBufPool, HFT bodies reuse hftBodyBufPool, large
// bodies stream via StreamDecoder with no intermediate copy.
func decodeTemplate(r io.Reader, contentLength int, tier string, tpl *models.PDFTemplate) error {
	return decodeTemplateJSON(r, contentLength, tier, tpl)
}
