package form

import (
	"strings"
	"testing"
	"time"
)

// TestParseXFDFRejectsEntities proves a billion-laughs-style payload is
// rejected fast instead of being expanded by encoding/xml.
func TestParseXFDFRejectsEntities(t *testing.T) {
	laughs := `<?xml version="1.0"?>
<!DOCTYPE xfdf [<!ENTITY lol "lollollollollollollollol"><!ENTITY lol2 "&lol;&lol;&lol;&lol;&lol;&lol;&lol;">]>
<xfdf xmlns="http://ns.adobe.com/xfdf/"><fields><field name="a"><value>&lol2;</value></field></fields></xfdf>`
	start := time.Now()
	if _, err := ParseXFDF([]byte(laughs)); err == nil {
		t.Fatalf("expected rejection of ENTITY payload")
	}
	if time.Since(start) > 10*time.Second {
		t.Fatalf("entity rejection took too long")
	}
}

// TestParseXFDFRejectsOversized proves inputs over the cap are rejected.
func TestParseXFDFRejectsOversized(t *testing.T) {
	big := make([]byte, MaxXFDFBytes+1)
	big[0] = '<'
	if _, err := ParseXFDF(big); err == nil {
		t.Fatalf("expected rejection of oversized XFDF")
	}
}

// TestParseXFDFValidStillWorks guards the cap against false positives.
func TestParseXFDFValidStillWorks(t *testing.T) {
	xfdf := `<xfdf xmlns="http://ns.adobe.com/xfdf/"><fields><field name="a"><value>b</value></field></fields></xfdf>`
	m, err := ParseXFDF([]byte(xfdf))
	if err != nil {
		t.Fatalf("valid XFDF rejected: %v", err)
	}
	if m["a"] != "b" {
		t.Fatalf("unexpected parse result: %v", m)
	}
}

// TestRadioValueRewriteAnchored proves the /V rewrite only touches a real
// /V value and leaves lookalikes like /Version intact.
func TestRadioValueRewriteAnchored(t *testing.T) {
	dict := []byte("<< /Version 1 /V /Yes /Kids [3 0 R] >>")
	newV := []byte("/V /Off")
	got := reVBroad.ReplaceAll(dict, newV)
	if !strings.Contains(string(got), "/Version 1") {
		t.Fatalf("unrelated /Version entry was clobbered: %q", got)
	}
	if !strings.Contains(string(got), "/V /Off") {
		t.Fatalf("/V value was not rewritten: %q", got)
	}
	paren := []byte("<< /T (choice) /V (Old) >>")
	gotParen := reVBroad.ReplaceAll(paren, []byte("/V (New)"))
	if string(gotParen) != "<< /T (choice) /V (New) >>" {
		t.Fatalf("paren value rewrite broke: %q", gotParen)
	}
}

// TestFormTrailerHasEncryptIgnoresStreamText mirrors the merge/redact
// guarantee inside the form package.
func TestFormTrailerHasEncryptIgnoresStreamText(t *testing.T) {
	plain := []byte("%PDF-1.4\n1 0 obj\n<< /Length 20 >>\nstream\nBT (/Encrypt) Tj ET\nendstream\nendobj\ntrailer\n<< /Size 2 /Root 1 0 R >>")
	if trailerHasEncrypt(plain) {
		t.Fatalf("stream /Encrypt text falsely detected as encryption")
	}
	encrypted := []byte("trailer\n<< /Size 2 /Root 1 0 R /Encrypt 9 0 R >>")
	if !trailerHasEncrypt(encrypted) {
		t.Fatalf("real /Encrypt trailer entry was not detected")
	}
}
