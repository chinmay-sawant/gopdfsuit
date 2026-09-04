package font

import (
	"bytes"
	"sync"
	"testing"
)

// TestSubsetCacheOverflowRace hammers overflow clears against concurrent
// stores and lookups. Run with -race: clear-all must be atomic with the
// count reset (C5), so no lost-update or negative-count divergence.
func TestSubsetCacheOverflowRace(t *testing.T) {
	ClearSubsetCache()
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			base := &TTFFont{PostScriptName: "RaceFont"}
			for i := 0; i < 300; i++ {
				glyphs := []uint16{uint16(g*1000 + i), uint16(i)}
				storeCachedSubset(base, glyphs, []byte("race-data"), map[uint16]uint16{1: 1})
				_, _ = lookupCachedSubset(base, glyphs)
				if i%50 == 0 {
					ClearSubsetCache()
				}
			}
		}(g)
	}
	wg.Wait()
	if c := subsetCacheCount.Load(); c < 0 || c > maxSubsetCacheEntries+8 {
		t.Fatalf("subset count = %d out of range", c)
	}
}

// TestCompressShardOverflowRace hammers per-shard overflow clears against
// concurrent cached compressions. Run with -race (C5).
func TestCompressShardOverflowRace(t *testing.T) {
	ClearPageCompressCache()
	raw := bytes.Repeat([]byte("BT /F1 12 Tf (race) Tj ET\n"), 64)
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				payload := append(append([]byte(nil), raw...), byte(g), byte(i))
				buf, ok := CompressContentStreamCached(payload)
				if ok && buf != nil {
					PutCompressBuffer(buf)
				}
				if i%25 == 0 {
					ClearPageCompressCache()
				}
			}
		}(g)
	}
	wg.Wait()
}
