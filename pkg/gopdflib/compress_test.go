package gopdflib_test

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"math/rand"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

func TestCompressPDF_NotPDF(t *testing.T) {
	_, err := gopdflib.CompressPDF([]byte("hello"), gopdflib.CompressOptions{})
	if err == nil {
		t.Fatal("expected error for non-PDF input")
	}
}

func TestCompressPDF_ImagePDF(t *testing.T) {
	jpegBytes := noisyJPEG(1200, 800)
	template := gopdflib.PDFTemplate{
		Config: gopdflib.Config{
			Page:          "A4",
			PageAlignment: 1,
		},
		Elements: []gopdflib.Element{
			{
				Type: "image",
				Image: &gopdflib.Image{
					ImageName: "photo.jpg",
					ImageData: base64.StdEncoding.EncodeToString(jpegBytes),
					Width:     400,
					Height:    270,
				},
			},
		},
	}

	src, err := gopdflib.GeneratePDF(template)
	if err != nil {
		t.Fatalf("GeneratePDF: %v", err)
	}
	if !bytes.HasPrefix(src, []byte("%PDF-")) {
		t.Fatal("generated file is not a PDF")
	}

	out, err := gopdflib.CompressPDF(src, gopdflib.CompressOptions{
		Level: gopdflib.CompressHeavy,
	})
	if err != nil {
		t.Fatalf("CompressPDF: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF-")) {
		t.Fatal("compressed file does not start with PDF header")
	}
	if !bytes.Contains(out, []byte("%%EOF")) {
		t.Fatal("compressed file missing EOF marker")
	}
	if len(out) == 0 {
		t.Fatal("compressed output is empty")
	}
	if len(out) > len(src) {
		t.Fatalf("compressed size %d > original %d", len(out), len(src))
	}
	t.Logf("original=%d compressed=%d", len(src), len(out))
}

func TestCompressPDF_StripsMetadata(t *testing.T) {
	template := gopdflib.PDFTemplate{
		Config: gopdflib.Config{
			Page:          "A4",
			PageAlignment: 1,
			PdfTitle:      "Secret Title",
		},
		Title: gopdflib.Title{
			Props: "Helvetica:14:100:center:0:0:0:0",
			Text:  "Hello",
		},
	}
	src, err := gopdflib.GeneratePDF(template)
	if err != nil {
		t.Fatalf("GeneratePDF: %v", err)
	}

	out, err := gopdflib.CompressPDF(src, gopdflib.CompressOptions{Level: gopdflib.CompressMedium})
	if err != nil {
		t.Fatalf("CompressPDF: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF-")) {
		t.Fatal("compressed file is not a PDF")
	}
	trailer := out
	if i := bytes.LastIndex(out, []byte("trailer")); i >= 0 {
		trailer = out[i:]
	}
	if bytes.Contains(trailer, []byte("/Info ")) {
		t.Fatal("expected trailer /Info to be stripped")
	}
	if bytes.Contains(trailer, []byte("/ID ")) {
		t.Fatal("expected trailer /ID to be stripped")
	}
}

func TestCompressPDF_LevelOverride(t *testing.T) {
	src, err := gopdflib.GeneratePDF(gopdflib.PDFTemplate{
		Config: gopdflib.Config{Page: "A4", PageAlignment: 1},
		Title:  gopdflib.Title{Props: "Helvetica:12:000:left:0:0:0:0", Text: "nogs"},
	})
	if err != nil {
		t.Fatalf("GeneratePDF: %v", err)
	}
	out, err := gopdflib.CompressPDF(src, gopdflib.CompressOptions{Level: gopdflib.CompressLight})
	if err != nil {
		t.Fatalf("CompressPDF light: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF-")) {
		t.Fatal("light compress did not return a PDF")
	}
}

func noisyJPEG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	rng := rand.New(rand.NewSource(1))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(rng.Intn(256)),
				G: uint8(rng.Intn(256)),
				B: uint8(rng.Intn(256)),
				A: 255,
			})
		}
	}
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95})
	return buf.Bytes()
}
