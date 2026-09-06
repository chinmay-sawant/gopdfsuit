package pdf

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/cachettl"
)

func TestPropsCacheReusesParsedProps(t *testing.T) {
	ClearPropsCache()

	input := "Helvetica:12:000:left:0:0:0:0"
	first := parseProps(input)
	second := parseProps(input)

	if first != second {
		t.Fatalf("expected cached result to be equal: %+v vs %+v", first, second)
	}
}

func TestPropsCacheBoundsEntries(t *testing.T) {
	ClearPropsCache()

	// Fill cache beyond its limit with unique strings.
	for i := range maxPropsCacheEntries + 10 {
		parseProps("Arial:10:000:left:0:0:0:0:" + string(rune(i)))
	}

	// After overflow, an early entry should be evicted while a recent one remains.
	if _, ok := propsCache.Load("Arial:10:000:left:0:0:0:0:\x00"); ok {
		t.Fatal("expected early entry to be evicted after cache overflow")
	}

	recent := "Arial:10:000:left:0:0:0:0:" + string(rune(maxPropsCacheEntries+9))
	if _, ok := propsCache.Load(recent); !ok {
		t.Fatal("expected recent entry to still be cached")
	}
}

func TestPropsCacheClear(t *testing.T) {
	ClearPropsCache()

	input := "Courier:14:111:right:1:1:1:1"
	parseProps(input)

	if _, ok := propsCache.Load(input); !ok {
		t.Fatal("expected entry before clear")
	}

	ClearPropsCache()

	if _, ok := propsCache.Load(input); ok {
		t.Fatal("expected cache miss after clear")
	}
}

func TestPropsCacheExpiresAfterTTL(t *testing.T) {
	old := cachettl.Get()
	defer cachettl.SetCacheTTL(old)
	cachettl.SetCacheTTL(30 * time.Millisecond)
	ClearPropsCache()

	input := "Helvetica:12:000:left:0:0:0:0"
	parseProps(input)
	if _, ok := propsCache.Load(input); !ok {
		t.Fatal("expected entry before TTL expiry")
	}

	time.Sleep(60 * time.Millisecond)
	// Lookup treats the stale entry as a miss and re-parses on demand.
	got := parseProps(input)
	if got.FontName != "Helvetica" || got.FontSize != 12 {
		t.Fatalf("expected re-parsed props after expiry, got %+v", got)
	}
}
