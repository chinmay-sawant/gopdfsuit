# Agent 1e — Findings 121–150 Validation (Manual Review)

## Per-Finding Table
| Finding | Rule | Correctly Fired? | Reason |
|---------|------|------------------|--------|
| 121 | PERF-31 | Yes | defer r.mu.Unlock() present in flagged function |
| 122 | PERF-31 | Yes | defer r.mu.RUnlock() present in flagged function |
| 123 | PERF-31 | Yes | defer r.mu.RUnlock() present in flagged function |
| 124 | PERF-123 | Yes | make([]*RegisteredFont, 0) uses explicit zero length |
| 125 | PERF-31 | Yes | defer r.mu.Unlock() present in flagged function |
| 126 | PERF-192 | Yes | make(map) called without a size hint |
| 127 | PERF-31 | Yes | defer r.mu.Unlock() present in flagged function |
| 128 | PERF-4 | Yes | make(map) allocated inside for-loop body |
| 129 | PERF-192 | Yes | make(map) inside loop without size hint |
| 130 | PERF-31 | Yes | defer r.mu.RUnlock() present in flagged function |
| 131 | PERF-31 | Yes | defer r.mu.Unlock() present in flagged function |
| 132 | PERF-6 | Yes | fmt.Sprintf called inside for-loop body |
| 133 | PERF-31 | Yes | defer r.mu.RUnlock() present in flagged function |
| 134 | PERF-109 | No | ToLower filters per-entry ext; no map key in loop |
| 135 | PERF-112 | Yes | strings.ToLower used before extension comparison |
| 136 | PERF-46 | Yes | strings.TrimSuffix called inside directory-scan loop |
| 137 | PERF-31 | Yes | defer r.mu.RUnlock() present in flagged function |
| 138 | PERF-31 | Yes | defer r.mu.RUnlock() present in flagged function |
| 139 | PERF-6 | Yes | fmt.Sprintf called inside fonts range loop |
| 140 | PERF-6 | Yes | fmt.Sprintf called inside fonts range loop |
| 141 | PERF-31 | Yes | defer r.mu.RUnlock() present in flagged function |
| 142 | PERF-192 | Yes | make(map) without len(usedGlyphs) size hint |
| 143 | PERF-192 | Yes | make(map) without len(glyphSet) size hint |
| 144 | PERF-192 | Yes | make(map) without table-count size hint |
| 145 | PERF-3 | Yes | make([]byte) rebuilt inside optional-tables loop |
| 146 | PERF-192 | Yes | make(map) without len(tableNames) size hint |
| 147 | PERF-32 | Yes | []byte(name) string conversion inside table loop |
| 148 | PERF-107 | Yes | binary.Write called inside table-directory loop |
| 149 | PERF-107 | Yes | binary.Write called inside table-directory loop |
| 150 | PERF-107 | Yes | binary.Write called inside table-directory loop |

## Summary
- Total findings analyzed: 30
- Correctly Fired (True Positive detections): 29
- Incorrectly Fired (False Positive detections): 1
- FP rate: 3.3%

## Notable FP patterns observed
- **PERF-109 mis-anchored on directory scan loop:** Rule fired on `for _, entry := range entries` where `strings.ToLower` normalizes each file's extension for filtering, not a constant map key recomputed every iteration (finding 134).
- **PERF-31 mutex-defer cluster:** Eleven findings flag idiomatic `defer r.mu.Lock/RLock` unlock patterns; rule fires syntactically but mutex release is standard resource cleanup.
- **PERF-192 density in font subsetting:** Multiple `make(map)` calls without size hints in `SubsetTTF` / `buildSubsetFont` are valid pattern matches even when approximate sizes are knowable from `len(glyphs)` or `len(tables)`.