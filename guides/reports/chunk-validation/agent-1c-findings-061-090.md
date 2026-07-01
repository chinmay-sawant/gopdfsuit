# Agent 1c — Findings 061–090 Validation (Manual Review)

## Per-Finding Table
| Finding | Rule | Correctly Fired? | Reason |
|---------|------|------------------|--------|
| 061 | PERF-119 | No | strconv.AppendInt separates the two append calls |
| 062 | PERF-128 | Yes | Three consecutive append calls on same buffer |
| 063 | PERF-119 | Yes | Consecutive append(dst,'.') and append(dst,'0') calls |
| 064 | PERF-128 | No | appendFmtNum calls break three-or-more append sequence |
| 065 | BP-1 | No | parseHexColor returns bool valid, not an error |
| 066 | BP-1 | No | parseHexColor returns bool valid, not an error |
| 067 | BP-1 | No | parseHexColor returns bool valid, not an error |
| 068 | PERF-15 | Yes | strconv.Itoa inside nested table cell loop |
| 069 | BP-1 | No | parseHexColor returns bool valid, not an error |
| 070 | PERF-15 | Yes | strconv.Itoa inside nested table cell loop |
| 071 | BP-1 | No | parseHexColor returns bool valid, not an error |
| 072 | BP-1 | No | parseHexColor returns bool valid, not an error |
| 073 | PERF-15 | Yes | strconv.Itoa inside nested table row loop |
| 074 | BP-1 | No | parseHexColor returns bool valid, not an error |
| 075 | PERF-15 | Yes | strconv.Itoa inside nested table row loop |
| 076 | PERF-35 | Yes | fmt.Sprintf boxes args during per-page rendering |
| 077 | PERF-42 | No | One-time constructor validation, not hot path |
| 078 | PERF-32 | No | One-time encryption setup, not hot path |
| 079 | CWE-916 | Yes | MD5 password hash with insufficient computational effort |
| 080 | CWE-328 | Yes | MD5 used for password digest derivation |
| 081 | PERF-3 | Yes | make([]byte) rebuilt inside owner-hash loop |
| 082 | PERF-3 | Yes | make([]byte) rebuilt inside user-hash loop |
| 083 | PERF-35 | No | One-time encrypt dictionary build, not hot path |
| 084 | PERF-110 | No | Pool New returns *zlib.Writer pointer, not value |
| 085 | BP-1 | Yes | zlib.NewWriterLevel error discarded into blank identifier |
| 086 | PERF-109 | No | Slice range loop, no map key recomputation |
| 087 | PERF-15 | Yes | strconv.Itoa inside metrics.Widths loop |
| 088 | PERF-15 | Yes | strconv.Itoa inside metrics.Widths loop |
| 089 | PERF-35 | No | One-time font resource string build, not hot path |
| 090 | PERF-192 | Yes | make(map) without known object-count size hint |

## Summary
- Total findings analyzed: 30
- Correctly Fired (True Positive detections): 15
- Incorrectly Fired (False Positive detections): 15
- FP rate: 50.0%

## Notable FP patterns observed
- **BP-1 on parseHexColor multi-return:** Rule flagged `r, g, b, _, valid := parseHexColor(...)` eight times, treating the blank identifier as a discarded error; `parseHexColor` returns `(float64, float64, float64, float64, bool)` with no error (findings 065, 066, 067, 069, 071, 072, 074).
- **PERF-119/128 with intervening strconv/appendFmtNum:** Append-merge rules fired where `strconv.AppendInt` or `appendFmtNum` separates append calls, so they are not consecutive independent appends (findings 061, 064).
- **Hot-path rules on cold paths:** PERF-42, PERF-32, and PERF-35 flagged one-time encryption or font-dictionary setup code outside tight loops (findings 077, 078, 083, 089).
- **PERF-110 incomplete match:** Finding anchored on `sync.Pool` declaration where `New` already returns `*zlib.Writer`, not a value type subject to boxing on Put (finding 084).
- **PERF-109 mis-anchoring:** Rule fired on a plain `for range metrics.Widths` slice iteration with no map key construction (finding 086).