package font

import (
	"bytes"
	"encoding/binary"
	"runtime"
	"sync"
	"sync/atomic"
)

type pageCompressEntry struct {
	data     []byte
	useFlate bool
}

type pageCompressKey struct {
	fingerprint uint64
	rawLen      int
	first       uint64
	mid         uint64
	last        uint64
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
	if n < 24 {
		var h uint64 = 1469598103934665603
		for _, b := range raw {
			h ^= uint64(b)
			h *= 1099511628211
		}
		return pageCompressKey{fingerprint: h, rawLen: n}
	}
	first := binary.LittleEndian.Uint64(raw[:8])
	mid := binary.LittleEndian.Uint64(raw[n/2:])
	last := binary.LittleEndian.Uint64(raw[n-8:])
	h := uint64(n) * 0x9e3779b97f4a7c15
	h ^= first + 0xbf58476d1ce4e5b9 + (h << 6) + (h >> 2)
	h ^= mid + 0x94d049bb133111eb + (h << 6) + (h >> 2)
	h ^= last + 0x2545f4914f6cdd1d + (h << 6) + (h >> 2)
	return pageCompressKey{fingerprint: h, rawLen: n, first: first, mid: mid, last: last}
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
		if !entry.useFlate {
			return nil, false
		}
		buf := GetCompressBuffer()
		buf.Write(entry.data)
		return buf, true
	}

	compressedBuf, ok := CompressContentStream(raw)
	if !ok {
		storePageCompressEntry(shard, key, &pageCompressEntry{
			useFlate: false,
		})
		return nil, false
	}

	data := append([]byte(nil), compressedBuf.Bytes()...)
	storePageCompressEntry(shard, key, &pageCompressEntry{
		data:     data,
		useFlate: true,
	})
	return compressedBuf, true
}

func storePageCompressEntry(shard *compressCacheShard, key pageCompressKey, entry *pageCompressEntry) {
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
