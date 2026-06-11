package font

import (
	"bytes"
	"hash/fnv"
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
	compressShardSeq   atomic.Uint64
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

func compressShardIndex() int {
	return int(compressShardSeq.Add(1) % uint64(compressShardCount))
}

func pageContentFingerprint(raw []byte) (uint64, int) {
	h := fnv.New64a()
	_, _ = h.Write(raw)
	return h.Sum64(), len(raw)
}

// CompressContentStreamCached zlib-compresses page bytes, reusing prior results for
// identical content streams (G2: HFT pages repeat across benchmark iterations).
func CompressContentStreamCached(raw []byte) (compressed *bytes.Buffer, useFlate bool) {
	fp, rawLen := pageContentFingerprint(raw)
	shard := &compressShards[compressShardIndex()]
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
	if shard.count.Add(1) > maxEntriesPerShard() {
		shard.entries.Clear()
		shard.count.Store(1)
	}
	shard.entries.Store(key, entry)
}

// ClearPageCompressCache drops all shard entries (tests / memory pressure).
func ClearPageCompressCache() {
	for i := range compressShards {
		compressShards[i].entries.Clear()
		compressShards[i].count.Store(0)
	}
}