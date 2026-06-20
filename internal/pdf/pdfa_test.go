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
