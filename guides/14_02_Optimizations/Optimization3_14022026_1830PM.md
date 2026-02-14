# PDF Generation Optimization Round 3 — Results

## Benchmark Comparison

| Metric        | Before      | After       | Δ          |
| ------------- | ----------- | ----------- | ---------- |
| **ns/op**     | ~99,576,000 | ~93,959,000 | **-5.6%**  |
| **B/op**      | 32,319,000  | 26,760,000  | **-17.2%** |
| **allocs/op** | 631,740     | 400,379     | **-36.6%** |

> [!IMPORTANT]
> 231,000 fewer allocations per PDF generation means significantly less GC pressure under load.

## Changes Made

### 1. Zero-alloc `appendFmtNum` — [draw.go](file:///home/chinmay/ChinmayPersonalProjects/gopdfsuit/internal/pdf/draw.go)

- Added [appendFmtNum](file:///home/chinmay/ChinmayPersonalProjects/gopdfsuit/internal/pdf/draw.go#L19-L22) that writes directly to caller's `[]byte`
- Converted **151 call sites** from `append(buf, fmtNum(x)...)` → `appendFmtNum(buf, x)`

### 2. Fast hex decode — [utils.go](file:///home/chinmay/ChinmayPersonalProjects/gopdfsuit/internal/pdf/utils.go)

- Replaced 3-4 `strconv.ParseInt` calls with [inline hex nibble lookup table](file:///home/chinmay/ChinmayPersonalProjects/gopdfsuit/internal/pdf/utils.go#L12-L28)

### 3. Cached `parseProps` — [utils.go](file:///home/chinmay/ChinmayPersonalProjects/gopdfsuit/internal/pdf/utils.go#L68-L71)

- `sync.Map` memoization eliminates repeated parsing of identical prop strings

### 4. Direct writes in `BeginMarkedContent` — [structure.go](file:///home/chinmay/ChinmayPersonalProjects/gopdfsuit/internal/pdf/structure.go#L131-L143)

- Eliminated `make([]byte, 0, 64)` per call, writes directly to `strings.Builder`

### 5-6. `BeginMarkedContentBuf` / `EndMarkedContentBuf` — [structure.go](file:///home/chinmay/ChinmayPersonalProjects/gopdfsuit/internal/pdf/structure.go#L168-L228)

- New methods write directly to `*bytes.Buffer`, bypassing `strings.Builder` intermediary
- 7 hot-loop call sites converted (every table cell, every image, every title)

## Verification

- ✅ `go build ./...` — clean
- ✅ `TestGenerateTemplatePDF` — PASS
- ✅ Benchmark — 5 runs consistent
