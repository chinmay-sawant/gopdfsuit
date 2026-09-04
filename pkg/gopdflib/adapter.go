package gopdflib

import (
	"strings"

	"github.com/bytedance/sonic"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/compress"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/merge"
)

// This file is the only place that converts between owned public types and
// internal engine types. Conversion goes through JSON so field renames on
// either side surface as translation behavior here instead of silent API
// drift. Opaque engine handles (BorrowedPDF, CustomFontRegistry) are not
// translated: they are borrowed, never serialized.

func toInternal[Pub any, In any](pub Pub) (In, error) {
	var zero In
	raw, err := sonic.Marshal(pub)
	if err != nil {
		return zero, err
	}
	if err := sonic.Unmarshal(raw, &zero); err != nil {
		return zero, err
	}
	return zero, nil
}

func fromInternal[In any, Pub any](in In) (Pub, error) {
	var zero Pub
	raw, err := sonic.Marshal(in)
	if err != nil {
		return zero, err
	}
	if err := sonic.Unmarshal(raw, &zero); err != nil {
		return zero, err
	}
	return zero, nil
}

func mustToInternal[Pub any, In any](pub Pub) In {
	in, err := toInternal[Pub, In](pub)
	if err != nil {
		panic("gopdflib: public-to-internal translation failed: " + err.Error())
	}
	return in
}

func mustFromInternal[In any, Pub any](in In) Pub {
	pub, err := fromInternal[In, Pub](in)
	if err != nil {
		panic("gopdflib: internal-to-public translation failed: " + err.Error())
	}
	return pub
}

// Template translation.
func toInternalTemplate(t PDFTemplate) models.PDFTemplate {
	return mustToInternal[PDFTemplate, models.PDFTemplate](t)
}

// ParseCompressLevel normalizes a level string the same way the engine does:
// case-insensitive, empty or unknown selects Medium.
func ParseCompressLevel(s string) CompressLevel {
	switch CompressLevel(strings.ToLower(strings.TrimSpace(s))) {
	case CompressLight:
		return CompressLight
	case CompressHeavy:
		return CompressHeavy
	default:
		return CompressMedium
	}
}

// normalizeCompressOptions applies engine-compatible defaults and caps so
// every entry point (Go, CGO, WASM) shares one policy. The engine re-applies
// its own defaults idempotently, so this never changes output bytes.
func normalizeCompressOptions(o CompressOptions) CompressOptions {
	o.Level = ParseCompressLevel(string(o.Level))
	if o.JPEGQuality <= 0 {
		switch o.Level {
		case CompressLight:
			o.JPEGQuality = 92
		case CompressHeavy:
			o.JPEGQuality = 50
		default:
			o.JPEGQuality = 75
		}
	}
	if o.JPEGQuality > 100 {
		o.JPEGQuality = 100
	}
	if o.MaxImageDim <= 0 {
		switch o.Level {
		case CompressLight:
			o.MaxImageDim = 1920
		case CompressHeavy:
			o.MaxImageDim = 612
		default:
			o.MaxImageDim = 1275
		}
	}
	if o.MaxImageDim > 4096 {
		o.MaxImageDim = 4096
	}
	return o
}

func toInternalCompressOptions(o CompressOptions) compress.Options {
	o = normalizeCompressOptions(o)
	return compress.Options{
		Level:       compress.Level(o.Level),
		JPEGQuality: o.JPEGQuality,
		MaxImageDim: o.MaxImageDim,
	}
}

func toInternalSplitSpec(s SplitSpec) merge.SplitSpec {
	return merge.SplitSpec{
		Pages:      s.Pages,
		Ranges:     s.Ranges,
		MaxPerFile: s.MaxPerFile,
	}
}
