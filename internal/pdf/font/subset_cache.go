package font

import (
	"encoding/binary"
	"hash/fnv"
	"sort"
	"sync"
	"sync/atomic"
)

type cachedSubset struct {
	data     []byte
	oldToNew map[uint16]uint16
}

const maxSubsetCacheEntries = 1024

var (
	subsetCache      sync.Map // uint64 fingerprint -> *cachedSubset
	subsetCacheCount atomic.Int64
)

// ClearSubsetCache drops all cached font subsets (tests / memory pressure).
func ClearSubsetCache() {
	subsetCache.Clear()
	subsetCacheCount.Store(0)
}

func glyphSubsetFingerprint(font *TTFFont, usedGlyphs []uint16) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(font.PostScriptName))
	glyphs := append([]uint16(nil), usedGlyphs...)
	sort.Slice(glyphs, func(i, j int) bool { return glyphs[i] < glyphs[j] })
	for _, g := range glyphs {
		var b [2]byte
		binary.BigEndian.PutUint16(b[:], g)
		_, _ = h.Write(b[:])
	}
	return h.Sum64()
}

func lookupCachedSubset(font *TTFFont, usedGlyphs []uint16) (*cachedSubset, bool) {
	if font == nil || len(usedGlyphs) == 0 {
		return nil, false
	}
	key := glyphSubsetFingerprint(font, usedGlyphs)
	if v, ok := subsetCache.Load(key); ok {
		if cs, ok := v.(*cachedSubset); ok && cs != nil {
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
	for k, v := range oldToNew {
		oldCopy[k] = v
	}
	if subsetCacheCount.Add(1) > maxSubsetCacheEntries {
		ClearSubsetCache()
		subsetCacheCount.Store(1)
	}
	subsetCache.Store(key, &cachedSubset{
		data:     append([]byte(nil), data...),
		oldToNew: oldCopy,
	})
}
