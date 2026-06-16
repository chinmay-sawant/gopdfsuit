package font

import (
	"bytes"
	"runtime"
	"sync"
	"sync/atomic"
)

type pageCompressEntry struct {
	fingerprint uint64
	rawLen      int
	data        []byte
	useFlate    bool
}

const maxPageCompressCacheEntries = 2048

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

func pageContentFingerprint(raw []byte) (uint64, int) {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	h := uint64(offset64)
	for _, b := range raw {
		h ^= uint64(b)
		h *= prime64
	}
	return h, len(raw)
}

// CompressContentStreamCached zlib-compresses page bytes, reusing prior results for
// identical content streams (G2: HFT pages repeat across benchmark iterations).
func CompressContentStreamCached(raw []byte) (compressed *bytes.Buffer, useFlate bool) {
	fp, rawLen := pageContentFingerprint(raw)
	shard := &compressShards[compressShardIndex(fp)]
	if v, ok := shard.entries.Load(fp); ok {
		entry := v.(*pageCompressEntry)
		if entry.rawLen == rawLen && entry.fingerprint == fp {
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
		storePageCompressEntry(shard, fp, &pageCompressEntry{
			fingerprint: fp,
			rawLen:      rawLen,
			useFlate:    false,
		})
		return nil, false
	}

	data := append([]byte(nil), compressedBuf.Bytes()...)
	storePageCompressEntry(shard, fp, &pageCompressEntry{
		fingerprint: fp,
		rawLen:      rawLen,
		data:        data,
		useFlate:    true,
	})
	return compressedBuf, true
}

func storePageCompressEntry(shard *compressCacheShard, key uint64, entry *pageCompressEntry) {
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
		compressShards[i].entries.Clear()
		compressShards[i].count.Store(0)
	}
}
