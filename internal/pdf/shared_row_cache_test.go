package pdf

import (
	"bytes"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
)

func TestSharedRowRenderCacheBoundsEntries(t *testing.T) {
	cache := &sharedRowRenderCacheStore{}
	rows := make([]models.Row, sharedRowRenderCacheMaxEntries+1)
	for i := 0; i <= sharedRowRenderCacheMaxEntries; i++ {
		key := sharedRowRenderCacheKey{row: &rows[i]}
		cache.Store(key, []byte{byte(i)})
	}

	cache.mu.RLock()
	entries := len(cache.entries)
	retainedBytes := cache.bytes
	cache.mu.RUnlock()

	if entries >= sharedRowRenderCacheMaxEntries {
		t.Fatalf("cache entries = %d, want below cap after overflow clear", entries)
	}
	if retainedBytes >= sharedRowRenderCacheMaxEntries {
		t.Fatalf("cache bytes = %d, want below cap after overflow clear", retainedBytes)
	}
}

func TestSharedRowRenderCacheSkipsOversizedValues(t *testing.T) {
	cache := &sharedRowRenderCacheStore{}
	oversized := bytes.Repeat([]byte("x"), sharedRowRenderCacheMaxValue+1)
	row := &models.Row{}
	cache.Store(sharedRowRenderCacheKey{row: row}, oversized)

	if _, ok := cache.Load(sharedRowRenderCacheKey{row: row}); ok {
		t.Fatal("expected oversized render to skip cache storage")
	}
}
