package pdf

import (
	"encoding/base64"
	"testing"
)

// tinyPNG1x1 is a valid 1x1 transparent PNG.
const tinyPNG1x1 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

func TestImageCacheReusesDecodedImage(t *testing.T) {
	ResetImageCache()

	img1, err := DecodeImageData(tinyPNG1x1)
	if err != nil {
		t.Fatalf("first decode failed: %v", err)
	}

	img2, err := DecodeImageData(tinyPNG1x1)
	if err != nil {
		t.Fatalf("second decode failed: %v", err)
	}

	if img1.CacheKey != img2.CacheKey {
		t.Fatalf("expected cache keys to match")
	}
	if img1.Width != img2.Width || img1.Height != img2.Height {
		t.Fatalf("expected cached dimensions to match")
	}
}

func TestImageCacheBoundsEntries(t *testing.T) {
	ResetImageCache()

	// Generate many tiny unique PNGs to overflow the cache.
	for i := 0; i < maxImageCacheEntries+10; i++ {
		unique := base64.StdEncoding.EncodeToString([]byte{byte(i), byte(i >> 8), 0x89, 0x50, 0x4e, 0x47})
		// Not all will decode, but each call exercises the store path.
		_, _ = DecodeImageData(unique)
	}

	// The cache should have been cleared at least once and should not exceed the bound.
	imgCache.mu.RLock()
	size := len(imgCache.cache)
	imgCache.mu.RUnlock()

	if size > maxImageCacheEntries {
		t.Fatalf("image cache size = %d, want <= %d", size, maxImageCacheEntries)
	}
}

func TestImageCacheClear(t *testing.T) {
	ResetImageCache()

	_, err := DecodeImageData(tinyPNG1x1)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if imgCache.lastObj == nil {
		t.Fatal("expected cached image before clear")
	}

	ResetImageCache()

	imgCache.mu.RLock()
	size := len(imgCache.cache)
	last := imgCache.lastObj
	imgCache.mu.RUnlock()

	if size != 0 || last != nil {
		t.Fatalf("expected empty cache after reset, got size=%d last=%v", size, last)
	}
}
