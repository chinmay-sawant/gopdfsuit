package pdf

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/models"
)

// Pure-Go golden test: inline HTML strings render with no browser and no
// network, so these tests never skip. Outputs are committed under
// sampledata/htmltopdf|htmltoimg as the post-Chrome baselines.

const pureGoGoldenHTML = `<html><head><style>h1{color:#154360}table{border-collapse:collapse}td{border:1px solid #999;padding:6px}</style></head><body><h1>Invoice #42</h1><table><tr><td>Widget</td><td>$10.00</td></tr><tr><td>Gadget</td><td>$20.00</td></tr></table></body></html>`

func TestHTMLToPDFPureGoGolden(t *testing.T) {
	out, err := ConvertHTMLToPDF(models.HTMLToPDFRequest{HTML: pureGoGoldenHTML, PageSize: "A4", Orientation: "Portrait"})
	if err != nil {
		t.Fatalf("ConvertHTMLToPDF failed: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF-")) {
		t.Error("output does not start with PDF header")
	}
	if !bytes.HasSuffix(out, []byte("%%EOF\n")) {
		t.Error("output does not end with EOF marker")
	}
	if len(out) == 0 {
		t.Error("empty PDF output")
	}
	path := filepath.Join("..", "..", "sampledata", "htmltopdf", "purego_invoice.pdf")
	if err := os.WriteFile(path, out, 0644); err != nil {
		t.Fatalf("failed to write golden %s: %v", path, err)
	}
	t.Logf("purego_invoice.pdf: %d bytes", len(out))
}

func TestHTMLToImagePureGoGolden(t *testing.T) {
	out, err := ConvertHTMLToImage(models.HTMLToImageRequest{HTML: pureGoGoldenHTML, Format: "png", Width: 800, Height: 600})
	if err != nil {
		t.Fatalf("ConvertHTMLToImage failed: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("\x89PNG")) {
		t.Error("output does not start with PNG header")
	}
	if len(out) == 0 {
		t.Error("empty PNG output")
	}
	path := filepath.Join("..", "..", "sampledata", "htmltoimg", "purego_invoice.png")
	if err := os.WriteFile(path, out, 0644); err != nil {
		t.Fatalf("failed to write golden %s: %v", path, err)
	}
	t.Logf("purego_invoice.png: %d bytes", len(out))
}
