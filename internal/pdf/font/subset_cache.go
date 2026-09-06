package font

import (
	"crypto/sha256"
	"encoding/binary"
	"maps"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/cachettl"
)

type cachedSubset struct {
	data      []byte
	oldToNew  map[uint16]uint16
	expiresAt time.Time
}

const maxSubsetCacheEntries = 1024

var (
	subsetCache      sync.Map // [32]byte fingerprint -> *cachedSubset
	subsetCacheCount atomic.Int64
	// subsetCacheMu makes clear-all atomic with overflow stores: both the
	// count reset and the map clear happen under one mutex so concurrent
	// stores cannot interleave a Store between Clear and count reset.
	subsetCacheMu sync.Mutex
)

// ClearSubsetCache drops all cached font subsets (tests / memory pressure).
func ClearSubsetCache() {
	subsetCacheMu.Lock()
	defer subsetCacheMu.Unlock()
	subsetCache.Clear()
	subsetCacheCount.Store(0)
}

func glyphSubsetFingerprint(font *TTFFont, usedGlyphs []uint16) [sha256.Size]byte {
	h := sha256.New()
	_, _ = h.Write(font.contentID[:])
	glyphs := append([]uint16(nil), usedGlyphs...)
	slices.Sort(glyphs)
	for _, g := range glyphs {
		var b [2]byte
		binary.BigEndian.PutUint16(b[:], g)
		_, _ = h.Write(b[:])
	}
	var fingerprint [sha256.Size]byte
	copy(fingerprint[:], h.Sum(nil))
	return fingerprint
}

func lookupCachedSubset(font *TTFFont, usedGlyphs []uint16) (*cachedSubset, bool) {
	if font == nil || len(usedGlyphs) == 0 {
		return nil, false
	}
	key := glyphSubsetFingerprint(font, usedGlyphs)
	if v, ok := subsetCache.Load(key); ok {
		if cs, ok := v.(*cachedSubset); ok && cs != nil {
			if cachettl.Expired(cs.expiresAt, time.Now()) {
				subsetCache.Delete(key)
				// Count left untouched: it is an overflow-trip approximation
				// that resets on clear-all. Decrementing here could race the
				// reset and delay eviction; upward drift only triggers an
				// earlier (safe) clear.
				return nil, false
			}
			return cs, true
		}
	}
	return nil, false
}

func storeCachedSubset(font *TTFFont, usedGlyphs []uint16, data []byte, oldToNew map[uint16]uint16) {
	if font == nil || len(data) == 0 || len(usedGlyphs) == 0 {
		return
	}
	key := glyphSubsetFingerprint(font, usedGlyphs)
	oldCopy := make(map[uint16]uint16, len(oldToNew))
	maps.Copy(oldCopy, oldToNew)
	subsetCacheMu.Lock()
	defer subsetCacheMu.Unlock()
	if subsetCacheCount.Add(1) > maxSubsetCacheEntries {
		subsetCache.Clear()
		subsetCacheCount.Store(1)
	}
	subsetCache.Store(key, &cachedSubset{
		data:      append([]byte(nil), data...),
		oldToNew:  oldCopy,
		expiresAt: cachettl.ExpiresAt(time.Now()),
	})
}
