package pdf

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"strings"
	"sync"
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
	for i := range maxImageCacheEntries + 10 {
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

func TestImageCacheConcurrentDecode(t *testing.T) {
	ResetImageCache()

	// Prime the cache so goroutines hit the MRU promote path.
	if _, err := DecodeImageData(tinyPNG1x1); err != nil {
		t.Fatalf("prime decode failed: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				img, err := DecodeImageData(tinyPNG1x1)
				if err != nil {
					t.Errorf("concurrent decode failed: %v", err)
					return
				}
				if img.Width != 1 || img.Height != 1 {
					t.Errorf("unexpected dimensions %dx%d", img.Width, img.Height)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func TestImageCacheReturnsCopy(t *testing.T) {
	ResetImageCache()

	img1, err := DecodeImageData(tinyPNG1x1)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	// Mutating the returned object (as the PDF/A path does with ColorSpace)
	// must not corrupt the shared cached entry.
	img1.ColorSpace = "/Mutated"

	img2, err := DecodeImageData(tinyPNG1x1)
	if err != nil {
		t.Fatalf("second decode failed: %v", err)
	}
	if img2.ColorSpace == "/Mutated" {
		t.Fatal("cache entry was corrupted by mutating a returned object")
	}
}

func TestDecodeImageDataRejectsExcessiveDecodedPixels(t *testing.T) {
	raw, err := base64.StdEncoding.DecodeString(tinyPNG1x1)
	if err != nil {
		t.Fatal(err)
	}
	const width, height = 4097, 4097
	binary.BigEndian.PutUint32(raw[16:20], width)
	binary.BigEndian.PutUint32(raw[20:24], height)
	binary.BigEndian.PutUint32(raw[29:33], crc32.ChecksumIEEE(raw[12:29]))

	_, err = DecodeImageData(base64.StdEncoding.EncodeToString(raw))
	if err == nil || !strings.Contains(err.Error(), "pixels") {
		t.Fatalf("DecodeImageData error = %v, want pixel-budget error", err)
	}
}

func TestDecodeImageDataPreservesJPEG(t *testing.T) {
	var encoded bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 3))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	if err := jpeg.Encode(&encoded, img, nil); err != nil {
		t.Fatal(err)
	}

	decoded, err := DecodeImageData(base64.StdEncoding.EncodeToString(encoded.Bytes()))
	if err != nil {
		t.Fatalf("DecodeImageData JPEG failed: %v", err)
	}
	if decoded.Filter != "/DCTDecode" || decoded.Width != 2 || decoded.Height != 3 {
		t.Fatalf("JPEG image = filter %q, dimensions %dx%d", decoded.Filter, decoded.Width, decoded.Height)
	}
}
