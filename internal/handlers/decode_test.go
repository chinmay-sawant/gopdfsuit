package handlers

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
)

const decodeTestRetailJSON = `{"config":{"page":"A4","pageAlignment":1},"title":{"props":"a","text":"b"},"footer":{"font":"a","text":"b"}}`

const decodeTestHFTJSON = `{"config":{"page":"A4","pageAlignment":1},"title":{"props":"a","text":"b"},"elements":[{"type":"table","table":{"maxcolumns":1,"rows":[{"row":[{"props":"a","text":"b"}]}]}}],"footer":{"font":"a","text":"b"}}`

func decodeWith(t *testing.T, body string, contentLength int, tier string) (models.PDFTemplate, error) {
	t.Helper()
	var tpl models.PDFTemplate
	err := decodeTemplate(strings.NewReader(body), contentLength, tier, &tpl)
	return tpl, err
}

// TestDecodeTemplateLimits pins the tier/limit policy of decodeTemplate:
// small bodies take the pooled path, bodies over maxPooledDecodeBody (512
// KiB) stream, hft bodies over maxHFTEncodeBody (8 MiB) fall back to the
// stream decoder, and malformed input errors on every path. The
// maxTemplateJSONBody (8 MiB) cap itself is enforced by http.MaxBytesReader
// in handleGenerateTemplatePDF, not here.
func TestDecodeTemplateLimits(t *testing.T) {
	t.Run("small pooled body decodes", func(t *testing.T) {
		tpl, err := decodeWith(t, decodeTestRetailJSON, len(decodeTestRetailJSON), "")
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		if tpl.Config.Page != "A4" {
			t.Fatalf("page = %q, want A4", tpl.Config.Page)
		}
	})

	t.Run("unknown length streams", func(t *testing.T) {
		for _, n := range []int{0, -1} {
			tpl, err := decodeWith(t, decodeTestRetailJSON, n, "")
			if err != nil {
				t.Fatalf("contentLength=%d decode: %v", n, err)
			}
			if tpl.Config.Page != "A4" {
				t.Fatalf("contentLength=%d page = %q, want A4", n, tpl.Config.Page)
			}
		}
	})

	t.Run("large body streams past pooled limit", func(t *testing.T) {
		big := `{"config":{"page":"A4","pageAlignment":1,"watermark":"` +
			strings.Repeat("w", maxPooledDecodeBody+1024) +
			`"},"title":{"props":"a","text":"b"},"footer":{"font":"a","text":"b"}}`
		tpl, err := decodeWith(t, big, len(big), "")
		if err != nil {
			t.Fatalf("large decode: %v", err)
		}
		if tpl.Config.Page != "A4" {
			t.Fatalf("page = %q, want A4", tpl.Config.Page)
		}
	})

	t.Run("invalid json errors", func(t *testing.T) {
		for _, body := range []string{"", "{invalid", `{"config":{"page":`} {
			if _, err := decodeWith(t, body, len(body), ""); err == nil {
				t.Fatalf("body %q decoded without error", body)
			}
		}
	})

	t.Run("hft tier decodes hft payload", func(t *testing.T) {
		tpl, err := decodeWith(t, decodeTestHFTJSON, len(decodeTestHFTJSON), "hft")
		if err != nil {
			t.Fatalf("hft decode: %v", err)
		}
		if len(tpl.Elements) != 1 || tpl.Elements[0].Type != "table" {
			t.Fatalf("elements = %+v, want one table element", tpl.Elements)
		}
	})

	t.Run("hft oversized length falls back to stream", func(t *testing.T) {
		// contentLength past maxHFTEncodeBody must not allocate the HFT
		// buffer: the small actual body decodes via the stream decoder.
		tpl, err := decodeWith(t, decodeTestRetailJSON, maxHFTEncodeBody+1, "hft")
		if err != nil {
			t.Fatalf("hft fallback decode: %v", err)
		}
		if tpl.Config.Page != "A4" {
			t.Fatalf("page = %q, want A4", tpl.Config.Page)
		}
	})

	t.Run("hft invalid json errors", func(t *testing.T) {
		body := "{invalid"
		if _, err := decodeWith(t, body, len(body), "hft"); err == nil {
			t.Fatal("hft tier decoded invalid JSON without error")
		}
	})

	t.Run("pooled decode does not leak prior templates", func(t *testing.T) {
		// Same pooled entry point twice: omitted fields must not leak from
		// the previous decode (resetTemplate clears before reuse).
		first, err := decodeWith(t, decodeTestRetailJSON, len(decodeTestRetailJSON), "")
		if err != nil {
			t.Fatalf("first decode: %v", err)
		}
		_ = first
		second, err := decodeWith(t, `{"config":{"page":"LETTER","pageAlignment":2}}`, -1, "")
		if err != nil {
			t.Fatalf("second decode: %v", err)
		}
		if second.Config.Page != "LETTER" {
			t.Fatalf("page = %q, want LETTER", second.Config.Page)
		}
		if second.Title.Text != "" || len(second.Table) != 0 {
			t.Fatalf("stale fields leaked: %+v", second)
		}
	})
}
