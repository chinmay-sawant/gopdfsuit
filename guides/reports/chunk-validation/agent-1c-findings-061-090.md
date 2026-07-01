# Agent 1c — Findings 061–090 Validation (Manual Review)

## Per-Finding Checklist
- [ ] **Finding 061** | Rule: **PERF-119** | Correctly Fired: **No** | strconv.AppendInt separates the two append calls
- [ ] **Finding 062** | Rule: **PERF-128** | Correctly Fired: **Yes** | Three consecutive append calls on same buffer
- [ ] **Finding 063** | Rule: **PERF-119** | Correctly Fired: **Yes** | Consecutive append(dst,'.') and append(dst,'0') calls
- [ ] **Finding 064** | Rule: **PERF-128** | Correctly Fired: **No** | appendFmtNum calls break three-or-more append sequence
- [ ] **Finding 065** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor returns bool valid, not an error
- [ ] **Finding 066** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor returns bool valid, not an error
- [ ] **Finding 067** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor returns bool valid, not an error
- [ ] **Finding 068** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa inside nested table cell loop
- [ ] **Finding 069** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor returns bool valid, not an error
- [ ] **Finding 070** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa inside nested table cell loop
- [ ] **Finding 071** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor returns bool valid, not an error
- [ ] **Finding 072** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor returns bool valid, not an error
- [ ] **Finding 073** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa inside nested table row loop
- [ ] **Finding 074** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor returns bool valid, not an error
- [ ] **Finding 075** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa inside nested table row loop
- [ ] **Finding 076** | Rule: **PERF-35** | Correctly Fired: **Yes** | fmt.Sprintf boxes args during per-page rendering
- [ ] **Finding 077** | Rule: **PERF-42** | Correctly Fired: **No** | One-time constructor validation, not hot path
- [ ] **Finding 078** | Rule: **PERF-32** | Correctly Fired: **No** | One-time encryption setup, not hot path
- [ ] **Finding 079** | Rule: **CWE-916** | Correctly Fired: **Yes** | MD5 password hash with insufficient computational effort
- [ ] **Finding 080** | Rule: **CWE-328** | Correctly Fired: **Yes** | MD5 used for password digest derivation
- [ ] **Finding 081** | Rule: **PERF-3** | Correctly Fired: **Yes** | make([]byte) rebuilt inside owner-hash loop
- [ ] **Finding 082** | Rule: **PERF-3** | Correctly Fired: **Yes** | make([]byte) rebuilt inside user-hash loop
- [ ] **Finding 083** | Rule: **PERF-35** | Correctly Fired: **No** | One-time encrypt dictionary build, not hot path
- [ ] **Finding 084** | Rule: **PERF-110** | Correctly Fired: **No** | Pool New returns *zlib.Writer pointer, not value
- [ ] **Finding 085** | Rule: **BP-1** | Correctly Fired: **Yes** | zlib.NewWriterLevel error discarded into blank identifier
- [ ] **Finding 086** | Rule: **PERF-109** | Correctly Fired: **No** | Slice range loop, no map key recomputation
- [ ] **Finding 087** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa inside metrics.Widths loop
- [ ] **Finding 088** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa inside metrics.Widths loop
- [ ] **Finding 089** | Rule: **PERF-35** | Correctly Fired: **No** | One-time font resource string build, not hot path
- [ ] **Finding 090** | Rule: **PERF-192** | Correctly Fired: **Yes** | make(map) without known object-count size hint

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