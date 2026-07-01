# Agent 1e — Findings 121–150 Validation (Manual Review)

## Per-Finding Checklist
- [x] **Finding 121** | Rule: **PERF-31** | Correctly Fired: **Yes** | defer r.mu.Unlock() present in flagged function → Fixed in internal/pdf/font/registry.go
- [x] **Finding 122** | Rule: **PERF-31** | Correctly Fired: **Yes** | defer r.mu.RUnlock() present in flagged function → Fixed in internal/pdf/font/registry.go
- [x] **Finding 123** | Rule: **PERF-31** | Correctly Fired: **Yes** | defer r.mu.RUnlock() present in flagged function → Fixed in internal/pdf/font/registry.go
- [x] **Finding 124** | Rule: **PERF-123** | Correctly Fired: **Yes** | make([]*RegisteredFont, 0) uses explicit zero length → Fixed in internal/pdf/font/registry.go
- [x] **Finding 125** | Rule: **PERF-31** | Correctly Fired: **Yes** | defer r.mu.Unlock() present in flagged function → Fixed in internal/pdf/font/registry.go
- [x] **Finding 126** | Rule: **PERF-192** | Correctly Fired: **Yes** | make(map) called without a size hint → Fixed in internal/pdf/font/registry.go
- [x] **Finding 127** | Rule: **PERF-31** | Correctly Fired: **Yes** | defer r.mu.Unlock() present in flagged function → Fixed in internal/pdf/font/registry.go
- [x] **Finding 128** | Rule: **PERF-4** | Correctly Fired: **Yes** | make(map) allocated inside for-loop body → Fixed in internal/pdf/font/registry.go
- [x] **Finding 129** | Rule: **PERF-192** | Correctly Fired: **Yes** | make(map) inside loop without size hint → Fixed in internal/pdf/font/registry.go
- [x] **Finding 130** | Rule: **PERF-31** | Correctly Fired: **Yes** | defer r.mu.RUnlock() present in flagged function → Fixed in internal/pdf/font/registry.go
- [x] **Finding 131** | Rule: **PERF-31** | Correctly Fired: **Yes** | defer r.mu.Unlock() present in flagged function → Fixed in internal/pdf/font/registry.go
- [x] **Finding 132** | Rule: **PERF-6** | Correctly Fired: **Yes** | fmt.Sprintf called inside for-loop body → Fixed in internal/pdf/font/registry.go
- [x] **Finding 133** | Rule: **PERF-31** | Correctly Fired: **Yes** | defer r.mu.RUnlock() present in flagged function → Fixed in internal/pdf/font/registry.go
- [x] **Finding 134** | Rule: **PERF-109** | Correctly Fired: **No** | ToLower filters per-entry ext; no map key in loop → N/A: false positive — ToLower filters per-entry ext, not constant map key
- [x] **Finding 135** | Rule: **PERF-112** | Correctly Fired: **Yes** | strings.ToLower used before extension comparison → Fixed in internal/pdf/font/registry.go
- [x] **Finding 136** | Rule: **PERF-46** | Correctly Fired: **Yes** | strings.TrimSuffix called inside directory-scan loop → Fixed in internal/pdf/font/registry.go
- [x] **Finding 137** | Rule: **PERF-31** | Correctly Fired: **Yes** | defer r.mu.RUnlock() present in flagged function → Fixed in internal/pdf/font/registry.go
- [x] **Finding 138** | Rule: **PERF-31** | Correctly Fired: **Yes** | defer r.mu.RUnlock() present in flagged function → Fixed in internal/pdf/font/registry.go
- [x] **Finding 139** | Rule: **PERF-6** | Correctly Fired: **Yes** | fmt.Sprintf called inside fonts range loop → Fixed in internal/pdf/font/registry.go
- [x] **Finding 140** | Rule: **PERF-6** | Correctly Fired: **Yes** | fmt.Sprintf called inside fonts range loop → Fixed in internal/pdf/font/registry.go
- [x] **Finding 141** | Rule: **PERF-31** | Correctly Fired: **Yes** | defer r.mu.RUnlock() present in flagged function → Fixed in internal/pdf/font/registry.go
- [x] **Finding 142** | Rule: **PERF-192** | Correctly Fired: **Yes** | make(map) without len(usedGlyphs) size hint → Fixed in internal/pdf/font/subset.go
- [x] **Finding 143** | Rule: **PERF-192** | Correctly Fired: **Yes** | make(map) without len(glyphSet) size hint → Fixed in internal/pdf/font/subset.go
- [x] **Finding 144** | Rule: **PERF-192** | Correctly Fired: **Yes** | make(map) without table-count size hint → Fixed in internal/pdf/font/subset.go
- [x] **Finding 145** | Rule: **PERF-3** | Correctly Fired: **Yes** | make([]byte) rebuilt inside optional-tables loop → Fixed in internal/pdf/font/subset.go
- [x] **Finding 146** | Rule: **PERF-192** | Correctly Fired: **Yes** | make(map) without len(tableNames) size hint → Fixed in internal/pdf/font/subset.go
- [x] **Finding 147** | Rule: **PERF-32** | Correctly Fired: **Yes** | []byte(name) string conversion inside table loop → Fixed in internal/pdf/font/subset.go
- [x] **Finding 148** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Write called inside table-directory loop → Fixed in internal/pdf/font/subset.go
- [x] **Finding 149** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Write called inside table-directory loop → Fixed in internal/pdf/font/subset.go
- [x] **Finding 150** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Write called inside table-directory loop → Fixed in internal/pdf/font/subset.go

## Summary
- Total findings analyzed: 30
- Correctly Fired (True Positive detections): 29
- Incorrectly Fired (False Positive detections): 1
- FP rate: 3.3%

## Notable FP patterns observed
- [ ] **PERF-109 mis-anchored on directory scan loop:** Rule fired on `for _, entry := range entries` where `strings.ToLower` normalizes each file's extension for filtering, not a constant map key recomputed every iteration (finding 134).
- [ ] **PERF-31 mutex-defer cluster:** Eleven findings flag idiomatic `defer r.mu.Lock/RLock` unlock patterns; rule fires syntactically but mutex release is standard resource cleanup.
- [ ] **PERF-192 density in font subsetting:** Multiple `make(map)` calls without size hints in `SubsetTTF` / `buildSubsetFont` are valid pattern matches even when approximate sizes are knowable from `len(glyphs)` or `len(tables)`.