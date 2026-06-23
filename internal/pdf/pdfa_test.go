package pdf

import (
	"bytes"
	"testing"
)

func TestICCProfileCacheMatchesBuilders(t *testing.T) {
	grayBuilt := compressICCProfileBytes(buildGrayICCProfile())
	if !bytes.Equal(grayBuilt, grayICCProfileCompressed) {
		t.Fatalf("gray ICC cache drift: built %d bytes, cached %d bytes", len(grayBuilt), len(grayICCProfileCompressed))
	}

	srgbBuilt := compressICCProfileBytes(buildSRGBICCProfile())
	if !bytes.Equal(srgbBuilt, srgbICCProfileCompressed) {
		t.Fatalf("sRGB ICC cache drift: built %d bytes, cached %d bytes", len(srgbBuilt), len(srgbICCProfileCompressed))
	}
}

func TestGetSRGBICCProfileReturnsDefensiveCopy(t *testing.T) {
	first := GetSRGBICCProfile()
	if len(first) == 0 {
		t.Fatal("expected cached sRGB ICC profile bytes")
	}
	original := first[0]
	first[0] ^= 0xff

	second := GetSRGBICCProfile()
	if second[0] != original {
		t.Fatal("expected cached sRGB ICC profile to be isolated from caller mutation")
	}
	if !bytes.Equal(second, srgbICCProfileRaw) {
		t.Fatal("expected public accessor to match cached raw profile bytes")
	}
}
