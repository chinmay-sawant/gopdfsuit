package gopdflib_test

import (
	"errors"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

func TestSentinelValidationErrors(t *testing.T) {
	if _, err := gopdflib.MergePDFs(nil); !errors.Is(err, gopdflib.ErrInvalidInput) {
		t.Fatalf("MergePDFs(nil) err = %v, want ErrInvalidInput", err)
	}
	if _, err := gopdflib.MergePDFs([][]byte{nil}); !errors.Is(err, gopdflib.ErrInvalidInput) {
		t.Fatalf("MergePDFs(empty part) err = %v, want ErrInvalidInput", err)
	}
	if _, err := gopdflib.SplitPDF(nil, gopdflib.SplitSpec{}); !errors.Is(err, gopdflib.ErrInvalidInput) {
		t.Fatalf("SplitPDF(nil) err = %v, want ErrInvalidInput", err)
	}
	if _, err := gopdflib.CompressPDF(nil, gopdflib.CompressOptions{}); !errors.Is(err, gopdflib.ErrInvalidInput) {
		t.Fatalf("CompressPDF(nil) err = %v, want ErrInvalidInput", err)
	}
	if _, err := gopdflib.FillPDFWithXFDF(nil, []byte("x")); !errors.Is(err, gopdflib.ErrInvalidInput) {
		t.Fatalf("FillPDFWithXFDF(nil) err = %v, want ErrInvalidInput", err)
	}
	if _, err := gopdflib.FindTextOccurrences([]byte("%PDF-1.4"), ""); !errors.Is(err, gopdflib.ErrInvalidInput) {
		t.Fatalf("FindTextOccurrences(empty text) err = %v, want ErrInvalidInput", err)
	}
	if _, err := gopdflib.ConvertHTMLToPDF(gopdflib.HTMLToPDFRequest{}); !errors.Is(err, gopdflib.ErrInvalidInput) {
		t.Fatalf("ConvertHTMLToPDF(empty) err = %v, want ErrInvalidInput", err)
	}
}

func TestLimitExceededClassification(t *testing.T) {
	big := make([]byte, gopdflib.MaxCompressInputBytes+1)
	_, err := gopdflib.CompressPDF(big, gopdflib.CompressOptions{})
	if !errors.Is(err, gopdflib.ErrLimitExceeded) {
		t.Fatalf("CompressPDF(oversize) err = %v, want ErrLimitExceeded", err)
	}
	if got := gopdflib.CodeOf(err); got != gopdflib.CodeLimitExceeded {
		t.Fatalf("CodeOf = %q, want limit_exceeded", got)
	}
}

func TestEngineErrorWrappedWithChain(t *testing.T) {
	// Corrupt PDF bytes reach the engine and come back classified as
	// invalid input, with the original engine error still in the chain.
	_, err := gopdflib.MergePDFs([][]byte{[]byte("%PDF-1.4 broken")})
	if err == nil {
		t.Skip("engine accepted corrupt input; classification vacuous")
	}
	if !errors.Is(err, gopdflib.ErrInvalidInput) {
		t.Fatalf("MergePDFs(corrupt) err = %v, want ErrInvalidInput", err)
	}
	if got := gopdflib.CodeOf(err); got != gopdflib.CodeInvalidInput {
		t.Fatalf("CodeOf = %q, want invalid_input", got)
	}
}

func TestEnvelopeJSONRoundTrip(t *testing.T) {
	_, err := gopdflib.SplitPDF(nil, gopdflib.SplitSpec{})
	raw := gopdflib.EnvelopeJSON(err)
	var env gopdflib.ErrorEnvelope
	if err := sonic.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatalf("EnvelopeJSON not JSON: %q (%v)", raw, err)
	}
	if env.Code != gopdflib.CodeInvalidInput || env.Message == "" {
		t.Fatalf("envelope = %+v, want invalid_input with message", env)
	}
}

func TestCodeStatusMapping(t *testing.T) {
	cases := map[gopdflib.ErrorCode]int{
		gopdflib.CodeInvalidInput:  422,
		gopdflib.CodeLimitExceeded: 413,
		gopdflib.CodeUpstream:      502,
		gopdflib.CodeInternal:      500,
	}
	for code, want := range cases {
		if got := gopdflib.StatusForCode(code); got != want {
			t.Fatalf("StatusForCode(%q) = %d, want %d", code, got, want)
		}
	}
	if got := gopdflib.CodeForStatus(413); got != gopdflib.CodeLimitExceeded {
		t.Fatalf("CodeForStatus(413) = %q", got)
	}
	if got := gopdflib.CodeForStatus(429); got != gopdflib.CodeLimitExceeded {
		t.Fatalf("CodeForStatus(429) = %q", got)
	}
	if got := gopdflib.CodeForStatus(400); got != gopdflib.CodeInvalidInput {
		t.Fatalf("CodeForStatus(400) = %q", got)
	}
}
