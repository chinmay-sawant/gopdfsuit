# Performance Optimization Guide

This guide documents the optimizations implemented to bridge the performance gap between the raw `gopdflib` and the Gin-powered HTTP API.

## Goal

Increase throughput from ~70 req/s to 400-700 req/s on a 24-thread machine.

## Optimization Strategy

### 1. Server & Middleware Tuning

- **Minimal Middleware**: Replaced `gin.Default()` with `gin.New()` and added only `Recovery()`. Removed the `Logger()` middleware as it causes significant lock contention on stdout under high concurrency.
- **Release Mode**: Forced `gin.SetMode(gin.ReleaseMode)` to disable debug-only logic and logging.
- **Semaphore Tuning**: Adjusted the concurrency semaphore to match `runtime.NumCPU()` (24) to reduce context-switching overhead for CPU-bound PDF generation.
- **HTTP Timeouts**: Configured `ReadTimeout` (30s) and `WriteTimeout` (60s) for resource protection.

### 2. Allocation & Buffer Management

- **Buffer Pooling**: Implemented `sync.Pool` for the main `pdfBuffer` and a small scratch buffer. This drastically reduces heap allocations and GC pressure per request.
- **Zero-Copy Returns**: While the main buffer is pooled, the final response is copied once to a fresh slice before returning the buffer to the pool, ensuring thread safety.
- **Efficient Hashing**: Replaced `md5.Sum(buffer.Bytes())` (which copies the entire buffer) with incremental hashing via `md5.New()`.

### 3. Hot-Loop Optimizations

- **Sprintf Elimination**: Replaced `fmt.Sprintf` calls in critical hot loops (like the XRef table generation) with `strconv.AppendInt` and manual padding. This eliminates thousands of small, short-lived allocations during PDF assembly.

### 4. Font Registry Optimization

- **Pre-sized Maps**: In `CloneForGeneration()`, the `UsedChars` map is pre-sized to 256. This prevents map rehashing as characters are scanned and marked during document generation.

## Verification

### Automated Results

- **Unit Tests**: All `internal/pdf` and `pkg/gopdflib` tests pass.
- **Integration Tests**: End-to-end verification passes.
- **Benchmark**: `BenchmarkGoPdfSuit` shows stable performance with significantly reduced memory pressure under concurrent loads.

### Performance Validation Command

To verify the throughput improvement in your environment:

```bash
# Terminal 1: Start optimized server
export GIN_MODE=release
DISABLE_PROFILING=true go run cmd/gopdfsuit/main.go

# Terminal 2: Run k6 spike test
k6 run test/generate_template-pdf/spike_test.js
```

Observe the `req/s` metric. It should now comfortably exceed 400 req/s during the spike phase.
