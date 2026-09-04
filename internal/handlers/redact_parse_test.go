package handlers

import (
	"testing"
)

// TestParseRedactApply exercises the pure redact-apply parser: block JSON,
// plain-string textSearch compat, legacy redactions/text aliases, and OCR.
func TestParseRedactApply(t *testing.T) {
	opts, err := parseRedactApply(redactApplyForm{
		mode:       " visual_allowed ",
		blocks:     `[{"pageNum":1,"x":1,"y":2,"width":10,"height":5}]`,
		textSearch: `[" secret ", "secret", "other"]`,
		ocr:        `{"enabled":true,"language":"eng"}`,
	})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if opts.Mode != "visual_allowed" {
		t.Fatalf("mode = %q", opts.Mode)
	}
	if len(opts.Blocks) != 1 || opts.Blocks[0].PageNum != 1 {
		t.Fatalf("blocks = %+v", opts.Blocks)
	}
	if len(opts.TextSearch) != 2 || opts.TextSearch[0].Text != "secret" {
		t.Fatalf("textSearch = %+v", opts.TextSearch)
	}
	if opts.OCR == nil || !opts.OCR.Enabled || opts.OCR.Language != "eng" {
		t.Fatalf("ocr = %+v", opts.OCR)
	}

	legacy, err := parseRedactApply(redactApplyForm{
		redactions: `[{"pageNum":2,"x":0,"y":0,"width":1,"height":1}]`,
		text:       "alpha, beta",
	})
	if err != nil {
		t.Fatalf("legacy parse failed: %v", err)
	}
	if len(legacy.Blocks) != 1 || legacy.Blocks[0].PageNum != 2 {
		t.Fatalf("legacy blocks = %+v", legacy.Blocks)
	}
	if len(legacy.TextSearch) != 2 {
		t.Fatalf("legacy textSearch = %+v", legacy.TextSearch)
	}

	for _, bad := range []redactApplyForm{
		{blocks: "{bad"},
		{textSearch: "{bad"},
		{ocr: "{bad"},
		{redactions: "{bad"},
	} {
		if _, err := parseRedactApply(bad); err == nil {
			t.Fatalf("expected error for %+v", bad)
		}
	}
}
