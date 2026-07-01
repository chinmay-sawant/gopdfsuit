# Agent 1c — Findings 061–090 Validation (Manual Review)

## Per-Finding Checklist
- [x] **Finding 061** | Rule: **PERF-119** | Correctly Fired: **No** | strconv.AppendInt separates the two append calls → N/A: false positive — strconv.AppendInt separates append calls
- [x] **Finding 062** | Rule: **PERF-128** | Correctly Fired: **Yes** | Three consecutive append calls on same buffer → Fixed in internal/pdf/bookmarks.go
- [x] **Finding 063** | Rule: **PERF-119** | Correctly Fired: **Yes** | Consecutive append(dst,'.') and append(dst,'0') calls → Fixed in internal/pdf/draw.go
- [x] **Finding 064** | Rule: **PERF-128** | Correctly Fired: **No** | appendFmtNum calls break three-or-more append sequence → N/A: false positive — appendFmtNum breaks three-or-more append sequence
- [x] **Finding 065** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor returns bool valid, not an error → N/A: false positive — parseHexColor returns bool, not error
- [x] **Finding 066** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor returns bool valid, not an error → N/A: false positive — parseHexColor returns bool, not error
- [x] **Finding 067** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor returns bool valid, not an error → N/A: false positive — parseHexColor returns bool, not error
- [x] **Finding 068** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa inside nested table cell loop → Fixed in internal/pdf/draw.go
- [x] **Finding 069** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor returns bool valid, not an error → N/A: false positive — parseHexColor returns bool, not error
- [x] **Finding 070** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa inside nested table cell loop → Fixed in internal/pdf/draw.go
- [x] **Finding 071** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor returns bool valid, not an error → N/A: false positive — parseHexColor returns bool, not error
- [x] **Finding 072** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor returns bool valid, not an error → N/A: false positive — parseHexColor returns bool, not error
- [x] **Finding 073** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa inside nested table row loop → Fixed in internal/pdf/draw.go
- [x] **Finding 074** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor returns bool valid, not an error → N/A: false positive — parseHexColor returns bool, not error
- [x] **Finding 075** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa inside nested table row loop → Fixed in internal/pdf/draw.go
- [x] **Finding 076** | Rule: **PERF-35** | Correctly Fired: **Yes** | fmt.Sprintf boxes args during per-page rendering → Fixed in internal/pdf/draw.go
- [x] **Finding 077** | Rule: **PERF-42** | Correctly Fired: **No** | One-time constructor validation, not hot path → N/A: false positive — one-time constructor validation, not hot path
- [x] **Finding 078** | Rule: **PERF-32** | Correctly Fired: **No** | One-time encryption setup, not hot path → N/A: false positive — one-time encryption setup, not hot path
- [x] **Finding 079** | Rule: **CWE-916** | Correctly Fired: **Yes** | MD5 password hash with insufficient computational effort → Cannot fix: PDF Standard Security Handler requires MD5 per ISO 32000
- [x] **Finding 080** | Rule: **CWE-328** | Correctly Fired: **Yes** | MD5 used for password digest derivation → Cannot fix: PDF Standard Security Handler requires MD5 per ISO 32000
- [x] **Finding 081** | Rule: **PERF-3** | Correctly Fired: **Yes** | make([]byte) rebuilt inside owner-hash loop → Fixed in internal/pdf/encryption/encrypt.go
- [x] **Finding 082** | Rule: **PERF-3** | Correctly Fired: **Yes** | make([]byte) rebuilt inside user-hash loop → Fixed in internal/pdf/encryption/encrypt.go
- [x] **Finding 083** | Rule: **PERF-35** | Correctly Fired: **No** | One-time encrypt dictionary build, not hot path → N/A: false positive — one-time encrypt dictionary build, not hot path
- [x] **Finding 084** | Rule: **PERF-110** | Correctly Fired: **No** | Pool New returns *zlib.Writer pointer, not value → N/A: false positive — Pool New returns *zlib.Writer pointer
- [x] **Finding 085** | Rule: **BP-1** | Correctly Fired: **Yes** | zlib.NewWriterLevel error discarded into blank identifier → Fixed in internal/pdf/font/compression.go
- [x] **Finding 086** | Rule: **PERF-109** | Correctly Fired: **No** | Slice range loop, no map key recomputation → N/A: false positive — slice range loop, no map key recomputation
- [x] **Finding 087** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa inside metrics.Widths loop → Fixed in internal/pdf/font/metrics.go
- [x] **Finding 088** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa inside metrics.Widths loop → Fixed in internal/pdf/font/metrics.go
- [x] **Finding 089** | Rule: **PERF-35** | Correctly Fired: **No** | One-time font resource string build, not hot path → N/A: false positive — one-time font resource string build, not hot path
- [x] **Finding 090** | Rule: **PERF-192** | Correctly Fired: **Yes** | make(map) without known object-count size hint → Fixed in internal/pdf/font/ttf.go

## Summary
- Total findings analyzed: 30
- Correctly Fired (True Positive detections): 15
- Incorrectly Fired (False Positive detections): 15
- FP rate: 50.0%

## Notable FP patterns observed
- [ ] **BP-1 on parseHexColor multi-return:** Rule flagged `r, g, b, _, valid := parseHexColor(...)` eight times, treating the blank identifier as a discarded error; `parseHexColor` returns `(float64, float64, float64, float64, bool)` with no error (findings 065, 066, 067, 069, 071, 072, 074).
- [ ] **PERF-119/128 with intervening strconv/appendFmtNum:** Append-merge rules fired where `strconv.AppendInt` or `appendFmtNum` separates append calls, so they are not consecutive independent appends (findings 061, 064).
- [ ] **Hot-path rules on cold paths:** PERF-42, PERF-32, and PERF-35 flagged one-time encryption or font-dictionary setup code outside tight loops (findings 077, 078, 083, 089).
- [ ] **PERF-110 incomplete match:** Finding anchored on `sync.Pool` declaration where `New` already returns `*zlib.Writer`, not a value type subject to boxing on Put (finding 084).
- [ ] **PERF-109 mis-anchoring:** Rule fired on a plain `for range metrics.Widths` slice iteration with no map key construction (finding 086).