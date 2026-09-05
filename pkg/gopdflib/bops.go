package gopdflib

import (
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/cachettl"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/font"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/signature"
)

// ClearBOPSCaches drops every cross-request content cache so the next
// GeneratePDF call pays full rebuild cost. BOPS means bypass-cache
// operations per second: cold path with no subset, page compress, font
// object, props, image, or signer reuse.
//
// Pools (sync.Pool buffers, zlib writers) are intentionally kept: they
// recycle memory without reusing content across documents. Only content
// caches are cleared.
func ClearBOPSCaches() {
	font.ClearSubsetCache()
	font.ClearPageCompressCache()
	font.ClearFontObjectCaches()
	pdf.ClearPropsCache()
	pdf.ResetImageCache()
	signature.ClearSignerCaches()
}

// DefaultCacheTTL is the out-of-box lifetime for content caches (font
// subsets, page compress output, font objects, prop parses, decoded
// images, signers, template-data files). Entries expire lazily on lookup.
const DefaultCacheTTL = cachettl.DefaultTTL

// SetCacheTTL overrides the shared content cache lifetime. A value <= 0
// disables time-based expiry. It can also be set at startup with the
// GOPDFSUIT_CACHE_TTL environment variable (e.g. "2m").
func SetCacheTTL(d time.Duration) {
	cachettl.SetCacheTTL(d)
}

// CacheTTL reports the current content cache lifetime.
func CacheTTL() time.Duration {
	return cachettl.Get()
}
