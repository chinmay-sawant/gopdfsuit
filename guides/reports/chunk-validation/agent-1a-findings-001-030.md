# Agent 1a — Findings 001–030 Validation (Manual Review)

## Per-Finding Table
| Finding | Rule | Correctly Fired? | Reason |
|---------|------|------------------|--------|
| 1 | PERF-151 | No | main is startup code, not a per-request hot-path handler |
| 2 | PERF-41 | No | flagged log.Printf is in startup profiling setup, not request path |
| 3 | PERF-43 | Yes | defer-recover runs in per-request Gin middleware |
| 4 | PERF-68 | No | gin.Logger registered only when gin.DebugMode, not production |
| 5 | CWE-497 | No | highlighted line is a comment, not a diagnostics endpoint |
| 6 | PERF-148 | No | channel is buffered and middleware receives in same function |
| 7 | PERF-195 | Yes | log.Fatalf is called inside a goroutine |
| 8 | BP-2 | Yes | bare `return err` without wrapping |
| 9 | PERF-40 | Yes | multiple time.Now calls in same function body |
| 10 | BP-2 | Yes | bare `return err` without wrapping |
| 11 | BP-1 | No | discarded value is runtime.Caller ok bool, not an error |
| 12 | PERF-35 | No | fmt.Sprintf in one-off benchmark header, not hot path |
| 13 | PERF-151 | No | utility helper, not a request handler on hot path |
| 14 | PERF-61 | Yes | router.Static with no cache headers shown in snippet |
| 15 | PERF-88 | No | Echo static-cache rule applied to Gin router.Static |
| 16 | PERF-200 | No | CORS is registered before Auth; ordering already correct |
| 17 | CWE-22 | No | filepath.Base confines user input before join with project root |
| 18 | PERF-22 | Yes | os.ReadFile invoked inside HTTP handler |
| 19 | PERF-112 | Yes | strings.ToLower used before string extension comparison |
| 20 | BP-1 | Yes | Close error discarded via `_ =` without handling comment |
| 21 | PERF-57 | No | io.ReadAll is in route handler, not Gin middleware |
| 22 | PERF-46 | Yes | strings.TrimSuffix allocates on request handler path |
| 23 | BP-1 | Yes | FormFile error return discarded with blank identifier |
| 24 | BP-1 | Yes | Close error discarded via `_ =` without handling comment |
| 25 | BP-1 | Yes | FormFile error return discarded with blank identifier |
| 26 | BP-1 | Yes | Close error discarded via `_ =` without handling comment |
| 27 | PERF-32 | Yes | `[]byte(b)` string-to-byte conversion in handler |
| 28 | PERF-32 | Yes | `[]byte(b)` string-to-byte conversion in handler |
| 29 | PERF-56 | Yes | c.JSON called inside for-loop body on error path |
| 30 | BP-1 | Yes | Close error discarded via `_ =` without handling comment |

## Summary
- Total findings analyzed: 30
- Correctly Fired (True Positives): 18
- Incorrectly Fired (False Positives): 12
- FP rate: 40.0%

## Notable FP patterns observed
- **Framework mismatch:** PERF-88 (Echo static-cache rule) fired on Gin `router.Static` (finding 15).
- **Comment / anchor mis-association:** CWE-497 flagged a commented `runtime.NumCPU` line, not an exposed endpoint (finding 5).
- **Hot-path rules on cold/startup code:** PERF-151 on `main`, PERF-41 on startup profiling logs, PERF-35 on benchmark header helper (findings 1, 2, 12).
- **Conditional production paths ignored:** PERF-68 flagged `gin.Logger()` though it is gated behind `gin.DebugMode` (finding 4).
- **Inverted middleware-order logic:** PERF-200 reported Auth-before-CORS while snippet shows CORS before Auth (finding 16).
- **Channel semantics mismatch:** PERF-148 claimed unbuffered channel with no receiver on a buffered semaphore with defer receive (finding 6).
- **Handler vs middleware confusion:** PERF-57 flagged `io.ReadAll` in a route handler as middleware work (finding 21).
- **Sanitization missed:** CWE-22 flagged path traversal despite `filepath.Base` confinement before `os.ReadFile` (finding 17).
- **Non-error discard misclassified:** BP-1 flagged `runtime.Caller` `ok` bool discard as discarded error (finding 11).