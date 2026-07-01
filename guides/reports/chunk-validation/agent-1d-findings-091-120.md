# Agent 1d — Findings 091–120 Validation (Manual Review)

## Per-Finding Table
| Finding | Rule | Correctly Fired? | Reason |
|---------|------|------------------|--------|
| 091 | BP-1 | Yes | `_ = zlibWriter.Close()` discards Close error without comment |
| 092 | BP-1 | Yes | `_ = zlibWriter.Close()` discards Close error without comment |
| 093 | PERF-15 | Yes | `strconv.Itoa` called inside outer width-range loop |
| 094 | PERF-15 | Yes | `strconv.Itoa` called inside inner width-write loop |
| 095 | BP-1 | Yes | `_ = zlibWriter.Close()` discards Close error on failure path |
| 096 | BP-1 | Yes | `_ = zlibWriter.Close()` discards Close error after success |
| 097 | PERF-15 | Yes | `strconv.Itoa` called inside CMap chunking loop |
| 098 | BP-1 | Yes | `_ = zlibWriter.Close()` discards Close error on failure path |
| 099 | BP-1 | Yes | `_ = zlibWriter.Close()` discards Close error after success |
| 100 | PERF-192 | Yes | `make(map[string]*TTFFont)` without a size hint |
| 101 | BP-2 | Yes | bare `return err` without contextual wrapping |
| 102 | PERF-42 | No | static `fmt.Errorf` on cold font-setup path, not hot |
| 103 | PERF-35 | No | `fmt.Errorf` on rare AutoDownload-false branch, not hot |
| 104 | BP-1 | No | `_ = os.Remove(...)` explicitly ignored with `// Clean up` |
| 105 | BP-1 | Yes | `_ = tmpFile.Close()` discards Close error in defer |
| 106 | BP-1 | Yes | `_ = resp.Body.Close()` discards Close error in defer |
| 107 | BP-1 | Yes | `_ = gzr.Close()` discards Close error in defer |
| 108 | PERF-176 | Yes | `io.Copy` inside tar-extraction loop allocates per call |
| 109 | BP-1 | Yes | `_ = outFile.Close()` discards Close error on failure path |
| 110 | BP-2 | Yes | bare `return err` without contextual wrapping |
| 111 | BP-2 | Yes | bare `return err` without contextual wrapping |
| 112 | BP-2 | Yes | bare `return err` without contextual wrapping |
| 113 | PERF-192 | Yes | `make(map[string]*RegisteredFont)` without a size hint |
| 114 | PERF-192 | Yes | `make(map[string]*RegisteredFont)` without a size hint |
| 115 | PERF-35 | No | `fmt.Errorf` only on font-load error path, not hot |
| 116 | PERF-31 | No | `defer r.mu.Unlock()` in registry helper, not HTTP handler |
| 117 | PERF-192 | Yes | `make(map[rune]bool)` without a size hint |
| 118 | PERF-31 | No | `defer r.mu.RUnlock()` in registry getter, not HTTP handler |
| 119 | PERF-31 | No | `defer r.mu.RUnlock()` in registry lookup, not HTTP handler |
| 120 | PERF-31 | No | `defer r.mu.Unlock()` in registry helper, not HTTP handler |

## Summary
- Total findings analyzed: 30
- Correctly Fired (True Positive detections): 22
- Incorrectly Fired (False Positive detections): 8
- FP rate: 26.7%

## Notable FP patterns observed
- **PERF-31 over-broad hot-path heuristic:** Rule fired on idiomatic `defer r.mu.Unlock()` / `RUnlock()` in PDF font-registry methods because any `func (` receiver satisfies `is_request_path`; these are library helpers, not HTTP handlers (findings 116, 118, 119, 120).
- **Hot-path rules on cold/error paths:** PERF-35 and PERF-42 flagged `fmt.Errorf` in font-manager initialization and one-time font-load error branches that are not on a per-request or loop hot path (findings 102, 103, 115).
- **BP-1 with explicit ignore comment:** Rule still fired on `_ = os.Remove(...)` even though the discard is annotated with `// Clean up`, matching the rule's allowed explicit-ignore pattern (finding 104).
- **Duplicate BP-1 on zlib cleanup:** Multiple findings on the same compression helper fire separately for error-path and success-path `_ = zlibWriter.Close()` calls; each matches the discarded-error pattern independently (findings 091–092, 095–096, 098–099).