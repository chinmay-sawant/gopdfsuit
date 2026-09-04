package gopdflib

import (
	"fmt"
	"strconv"
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

func toInternalTemplate(t PDFTemplate) (models.PDFTemplate, error) {
	in, err := toInternal[PDFTemplate, models.PDFTemplate](t)
	if err != nil {
		return in, err
	}
	// The font hint is not part of the JSON shape, so it does not survive
	// translation above: carry it over explicitly.
	in.SetPrecomputedStandardFonts(t.precomputedStandardFonts...)
	return in, nil
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

// compressLevelByNumber maps the 1|2|3 tier numbers shared with the frontend
// (compressLevels.js COMPRESS_LEVELS) and the WASM entry point. Numbers
// outside 1-3 fall back to Medium, matching the frontend levelByValue policy.
func compressLevelByNumber(n int) CompressLevel {
	switch n {
	case 1:
		return CompressLight
	case 3:
		return CompressHeavy
	default:
		return CompressMedium
	}
}

// ToServerLevel normalizes a flexible compress level to its canonical server
// string (light|medium|heavy), mirroring frontend toServerLevel: nil and ""
// select Medium, ints and numeric strings map 1|2|3 (other numbers fall back
// to Medium), and unknown non-numeric strings return an error. Use
// ParseCompressLevel instead when the engine-compatible silent default
// (unknown selects Medium) is wanted.
func ToServerLevel(level any) (CompressLevel, error) {
	switch v := level.(type) {
	case nil:
		return CompressMedium, nil
	case CompressLevel:
		return ToServerLevel(string(v))
	case string:
		key := strings.ToLower(strings.TrimSpace(v))
		if key == "" {
			return CompressMedium, nil
		}
		switch CompressLevel(key) {
		case CompressLight, CompressMedium, CompressHeavy:
			return CompressLevel(key), nil
		}
		if n, err := strconv.Atoi(key); err == nil {
			return compressLevelByNumber(n), nil
		}
		return "", fmt.Errorf("gopdflib: invalid compression level: %q (use 1|2|3 or light|medium|heavy)", v)
	case int:
		return compressLevelByNumber(v), nil
	default:
		return "", fmt.Errorf("gopdflib: invalid compression level type %T (use 1|2|3 or light|medium|heavy)", level)
	}
}

// ToWasmLevel normalizes a flexible compress level to its WASM tier number
// (1|2|3), mirroring frontend toWasmLevel with the same defaults and the
// same invalid-input error contract as ToServerLevel.
func ToWasmLevel(level any) (int, error) {
	normalized, err := ToServerLevel(level)
	if err != nil {
		return 0, err
	}
	switch normalized {
	case CompressLight:
		return 1, nil
	case CompressHeavy:
		return 3, nil
	default:
		return 2, nil
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
