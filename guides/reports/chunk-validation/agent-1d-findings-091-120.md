# Agent 1d — Findings 091–120 Validation (Manual Review)

## Per-Finding Checklist
- [ ] **Finding 091** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = zlibWriter.Close()` discards Close error without comment
- [ ] **Finding 092** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = zlibWriter.Close()` discards Close error without comment
- [ ] **Finding 093** | Rule: **PERF-15** | Correctly Fired: **Yes** | `strconv.Itoa` called inside outer width-range loop
- [ ] **Finding 094** | Rule: **PERF-15** | Correctly Fired: **Yes** | `strconv.Itoa` called inside inner width-write loop
- [ ] **Finding 095** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = zlibWriter.Close()` discards Close error on failure path
- [ ] **Finding 096** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = zlibWriter.Close()` discards Close error after success
- [ ] **Finding 097** | Rule: **PERF-15** | Correctly Fired: **Yes** | `strconv.Itoa` called inside CMap chunking loop
- [ ] **Finding 098** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = zlibWriter.Close()` discards Close error on failure path
- [ ] **Finding 099** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = zlibWriter.Close()` discards Close error after success
- [ ] **Finding 100** | Rule: **PERF-192** | Correctly Fired: **Yes** | `make(map[string]*TTFFont)` without a size hint
- [ ] **Finding 101** | Rule: **BP-2** | Correctly Fired: **Yes** | bare `return err` without contextual wrapping
- [ ] **Finding 102** | Rule: **PERF-42** | Correctly Fired: **No** | static `fmt.Errorf` on cold font-setup path, not hot
- [ ] **Finding 103** | Rule: **PERF-35** | Correctly Fired: **No** | `fmt.Errorf` on rare AutoDownload-false branch, not hot
- [ ] **Finding 104** | Rule: **BP-1** | Correctly Fired: **No** | `_ = os.Remove(...)` explicitly ignored with `// Clean up`
- [ ] **Finding 105** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = tmpFile.Close()` discards Close error in defer
- [ ] **Finding 106** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = resp.Body.Close()` discards Close error in defer
- [ ] **Finding 107** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = gzr.Close()` discards Close error in defer
- [ ] **Finding 108** | Rule: **PERF-176** | Correctly Fired: **Yes** | `io.Copy` inside tar-extraction loop allocates per call
- [ ] **Finding 109** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = outFile.Close()` discards Close error on failure path
- [ ] **Finding 110** | Rule: **BP-2** | Correctly Fired: **Yes** | bare `return err` without contextual wrapping
- [ ] **Finding 111** | Rule: **BP-2** | Correctly Fired: **Yes** | bare `return err` without contextual wrapping
- [ ] **Finding 112** | Rule: **BP-2** | Correctly Fired: **Yes** | bare `return err` without contextual wrapping
- [ ] **Finding 113** | Rule: **PERF-192** | Correctly Fired: **Yes** | `make(map[string]*RegisteredFont)` without a size hint
- [ ] **Finding 114** | Rule: **PERF-192** | Correctly Fired: **Yes** | `make(map[string]*RegisteredFont)` without a size hint
- [ ] **Finding 115** | Rule: **PERF-35** | Correctly Fired: **No** | `fmt.Errorf` only on font-load error path, not hot
- [ ] **Finding 116** | Rule: **PERF-31** | Correctly Fired: **No** | `defer r.mu.Unlock()` in registry helper, not HTTP handler
- [ ] **Finding 117** | Rule: **PERF-192** | Correctly Fired: **Yes** | `make(map[rune]bool)` without a size hint
- [ ] **Finding 118** | Rule: **PERF-31** | Correctly Fired: **No** | `defer r.mu.RUnlock()` in registry getter, not HTTP handler
- [ ] **Finding 119** | Rule: **PERF-31** | Correctly Fired: **No** | `defer r.mu.RUnlock()` in registry lookup, not HTTP handler
- [ ] **Finding 120** | Rule: **PERF-31** | Correctly Fired: **No** | `defer r.mu.Unlock()` in registry helper, not HTTP handler

## Summary
- Total findings analyzed: 30
- Correctly Fired (True Positive detections): 22
- Incorrectly Fired (False Positive detections): 8
- FP rate: 26.7%

## Notable FP patterns observed
- [ ] **PERF-31 over-broad hot-path heuristic:** Rule fired on idiomatic `defer r.mu.Unlock()` / `RUnlock()` in PDF font-registry methods because any `func (` receiver satisfies `is_request_path`; these are library helpers, not HTTP handlers (findings 116, 118, 119, 120).
- [ ] **Hot-path rules on cold/error paths:** PERF-35 and PERF-42 flagged `fmt.Errorf` in font-manager initialization and one-time font-load error branches that are not on a per-request or loop hot path (findings 102, 103, 115).
- [ ] **BP-1 with explicit ignore comment:** Rule still fired on `_ = os.Remove(...)` even though the discard is annotated with `// Clean up`, matching the rule's allowed explicit-ignore pattern (finding 104).
- [ ] **Duplicate BP-1 on zlib cleanup:** Multiple findings on the same compression helper fire separately for error-path and success-path `_ = zlibWriter.Close()` calls; each matches the discarded-error pattern independently (findings 091–092, 095–096, 098–099).