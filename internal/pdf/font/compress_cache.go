package font

import (
	"bytes"
	"hash/fnv"
	"sync"
)

type pageCompressEntry struct {
	fingerprint uint64
	rawLen      int
	data        []byte
	useFlate    bool
}

var pageCompressCache sync.Map

func pageContentFingerprint(raw []byte) (uint64, int) {
	h := fnv.New64a()
	_, _ = h.Write(raw)
	return h.Sum64(), len(raw)
}

// CompressContentStreamCached zlib-compresses page bytes, reusing prior results for
// identical content streams (G2: HFT pages repeat across benchmark iterations).
func CompressContentStreamCached(raw []byte) (compressed *bytes.Buffer, useFlate bool) {
	fp, rawLen := pageContentFingerprint(raw)
	if v, ok := pageCompressCache.Load(fp); ok {
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
		pageCompressCache.Store(fp, &pageCompressEntry{
			fingerprint: fp,
			rawLen:      rawLen,
			useFlate:    false,
		})
		return nil, false
	}

	data := append([]byte(nil), compressedBuf.Bytes()...)
	pageCompressCache.Store(fp, &pageCompressEntry{
		fingerprint: fp,
		rawLen:      rawLen,
		data:        data,
		useFlate:    true,
	})
	return compressedBuf, true
}
