package font

import (
	"bytes"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/cachettl"
)

type pageCompressEntry struct {
	data      []byte
	useFlate  bool
	expiresAt time.Time
}

type pageCompressKey struct {
	fingerprint uint64 // FNV-1a 64 over the FULL content plus length
	rawLen      int
}

const maxPageCompressCacheEntries = 2048
const maxFingerprintCachedContentLen = 32 * 1024

func maxEntriesPerShard() int64 {
	per := int64(maxPageCompressCacheEntries / compressShardCount)
	if per < 64 {
		return 64
	}
	return per
}

type compressCacheShard struct {
	entries sync.Map
	count   atomic.Int64
	// mu makes per-shard clear-all atomic with overflow stores.
	mu sync.Mutex
}

var (
	compressShardCount = maxCompressShards()
	compressShards     = make([]compressCacheShard, compressShardCount)
)

func maxCompressShards() int {
	n := runtime.NumCPU()
	if n < 4 {
		return 4
	}
	if n > 64 {
		return 64
	}
	return n
}

func compressShardIndex(fp uint64) int {
	return int(fp % uint64(compressShardCount))
}

func pageContentFingerprint(raw []byte) pageCompressKey {
	n := len(raw)
	if n == 0 {
		return pageCompressKey{}
	}
	// FNV-1a over the full content: sampling only len/first/mid/last let
	// distinct pages share a fingerprint and reuse each other's streams.
	const (
		offset64 = 1469598103934665603
		prime64  = 1099511628211
	)
	h := uint64(offset64) ^ uint64(n)*0x9e3779b97f4a7c15
	for _, b := range raw {
		h ^= uint64(b)
		h *= prime64
	}
	return pageCompressKey{fingerprint: h, rawLen: n}
}

// CompressContentStreamCached zlib-compresses page bytes, reusing prior results for
// identical content streams (G2: HFT pages repeat across benchmark iterations).
func CompressContentStreamCached(raw []byte) (compressed *bytes.Buffer, useFlate bool) {
	if len(raw) > maxFingerprintCachedContentLen {
		return CompressContentStream(raw)
	}
	key := pageContentFingerprint(raw)
	shard := &compressShards[compressShardIndex(key.fingerprint)]
	if v, ok := shard.entries.Load(key); ok {
		entry := v.(*pageCompressEntry)
		if cachettl.Expired(entry.expiresAt, time.Now()) {
			// Drop and fall through to the miss path below, which
			// recompresses and stores a fresh entry with a new expiry.
			shard.entries.Delete(key)
		} else {
			if !entry.useFlate {
				return nil, false
			}
			buf := GetCompressBuffer()
			buf.Write(entry.data)
			return buf, true
		}
	}

	compressedBuf, ok := CompressContentStream(raw)
	if !ok {
		storePageCompressEntry(shard, key, &pageCompressEntry{
			useFlate:  false,
			expiresAt: cachettl.ExpiresAt(time.Now()),
		})
		return nil, false
	}

	data := append([]byte(nil), compressedBuf.Bytes()...)
	storePageCompressEntry(shard, key, &pageCompressEntry{
		data:      data,
		useFlate:  true,
		expiresAt: cachettl.ExpiresAt(time.Now()),
	})
	return compressedBuf, true
}

func storePageCompressEntry(shard *compressCacheShard, key pageCompressKey, entry *pageCompressEntry) {
	shard.mu.Lock()
	defer shard.mu.Unlock()
	if _, exists := shard.entries.Load(key); exists {
		shard.entries.Store(key, entry)
		return
	}
	if shard.count.Load() >= maxEntriesPerShard() {
		shard.entries.Clear()
		shard.count.Store(0)
	}
	if _, loaded := shard.entries.LoadOrStore(key, entry); !loaded {
		shard.count.Add(1)
	}
}

// ClearPageCompressCache drops all shard entries (tests / memory pressure).
func ClearPageCompressCache() {
	for i := range compressShards {
		compressShards[i].mu.Lock()
		compressShards[i].entries.Clear()
		compressShards[i].count.Store(0)
		compressShards[i].mu.Unlock()
	}
}
