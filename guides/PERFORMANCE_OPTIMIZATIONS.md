# Performance Optimizations

This document describes performance optimizations made based on pprof profiling analysis.

## Profiling Summary

The pprof report identified these hotspots:

| Function | CPU% | Issue |
|----------|------|-------|
| runtime.memclrNoHeapPointers | 13% | Memory allocations |
| runtime.memmove | 8.7% | Memory copying |
| image/png.readImagePass | 21.7% | PNG decoding |
| compress/flate.* | ~20% | Compression overhead |
| convertToRGBWithAlpha | 2.2% | Alpha blending math |

## Optimizations Implemented

### 1. Zlib Writer Pooling (~20% compression overhead reduction)

**Problem**: Each image and content stream compression created a new `zlib.Writer`, which allocates ~256KB for compression tables.

**Solution**: Added `sync.Pool` for zlib writers that are reset and reused.

```go
var zlibWriterPool = sync.Pool{
    New: func() interface{} {
        w, _ := zlib.NewWriterLevel(io.Discard, zlib.BestSpeed)
        return w
    },
}
```

**Files Changed**: [image.go](../internal/pdf/image.go), [generator.go](../internal/pdf/generator.go)

### 2. Compression Buffer Pooling

**Problem**: Each compression operation allocated a new `bytes.Buffer` for output.

**Solution**: Added `sync.Pool` for compression output buffers.

```go
var compressBufPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}
```

**Files Changed**: [image.go](../internal/pdf/image.go)

### 3. Fast Alpha Blending Math

**Problem**: Alpha blending used integer division (`/ 255`) for each pixel component.

**Solution**: Replaced division with fast multiplication approximation:

```go
// Old: (r*a + 255*invA) / 255
// New: ((r*a + white) * 0x8081) >> 23
```

The magic number `0x8081` with right shift by 23 approximates division by 255 with high accuracy.

**Performance Impact**: Eliminates expensive integer division per RGB component per pixel.

**Files Changed**: [image.go](../internal/pdf/image.go)

### 4. Image Deduplication Cache

**Problem**: Same image appearing multiple times (e.g., company logo in every row) was decoded and compressed repeatedly.

**Solution**: Added FNV-1a hash-based cache for decoded images.

```go
type imageCache struct {
    mu    sync.RWMutex
    cache map[uint64]*ImageObject
}
```

**Benefits**:
- Skips PNG decoding for duplicate images (21.7% hotspot)
- Skips compression for duplicate images
- Thread-safe with read-write mutex

**Files Changed**: [image.go](../internal/pdf/image.go)

## Expected Improvements

| Optimization | Target Hotspot | Est. Improvement |
|-------------|----------------|------------------|
| Zlib pooling | compress/flate (~20%) | 10-15% faster |
| Buffer pooling | runtime.memclr (13%) | 5-10% less GC |
| Fast division | convertToRGBWithAlpha | 2-3x faster |
| Image cache | PNG decoding (21.7%) | Varies by duplication |

## API Changes

### New Public Function

```go
// ResetImageCache clears the image cache
// Call between unrelated PDF generations if memory is a concern
func ResetImageCache()
```

## Backward Compatibility

All changes are internal optimizations. No API breaking changes.

## Testing

Run the existing test suite to verify functionality:

```bash
go test ./internal/pdf/...
```

Profile again after changes to verify improvements:

```bash
go tool pprof -http=:8081 "http://localhost:8080/debug/pprof/profile?seconds=30"
```
