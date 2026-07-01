# Agent 1a — Findings 001–030 Validation (Manual Review)

## Per-Finding Checklist
- [ ] **Finding 1** | Rule: **PERF-151** | Correctly Fired: **No** | main is startup code, not a per-request hot-path handler
- [ ] **Finding 2** | Rule: **PERF-41** | Correctly Fired: **No** | flagged log.Printf is in startup profiling setup, not request path
- [ ] **Finding 3** | Rule: **PERF-43** | Correctly Fired: **Yes** | defer-recover runs in per-request Gin middleware
- [ ] **Finding 4** | Rule: **PERF-68** | Correctly Fired: **No** | gin.Logger registered only when gin.DebugMode, not production
- [ ] **Finding 5** | Rule: **CWE-497** | Correctly Fired: **No** | highlighted line is a comment, not a diagnostics endpoint
- [ ] **Finding 6** | Rule: **PERF-148** | Correctly Fired: **No** | channel is buffered and middleware receives in same function
- [ ] **Finding 7** | Rule: **PERF-195** | Correctly Fired: **Yes** | log.Fatalf is called inside a goroutine
- [ ] **Finding 8** | Rule: **BP-2** | Correctly Fired: **Yes** | bare `return err` without wrapping
- [ ] **Finding 9** | Rule: **PERF-40** | Correctly Fired: **Yes** | multiple time.Now calls in same function body
- [ ] **Finding 10** | Rule: **BP-2** | Correctly Fired: **Yes** | bare `return err` without wrapping
- [ ] **Finding 11** | Rule: **BP-1** | Correctly Fired: **No** | discarded value is runtime.Caller ok bool, not an error
- [ ] **Finding 12** | Rule: **PERF-35** | Correctly Fired: **No** | fmt.Sprintf in one-off benchmark header, not hot path
- [ ] **Finding 13** | Rule: **PERF-151** | Correctly Fired: **No** | utility helper, not a request handler on hot path
- [ ] **Finding 14** | Rule: **PERF-61** | Correctly Fired: **Yes** | router.Static with no cache headers shown in snippet
- [ ] **Finding 15** | Rule: **PERF-88** | Correctly Fired: **No** | Echo static-cache rule applied to Gin router.Static
- [ ] **Finding 16** | Rule: **PERF-200** | Correctly Fired: **No** | CORS is registered before Auth; ordering already correct
- [ ] **Finding 17** | Rule: **CWE-22** | Correctly Fired: **No** | filepath.Base confines user input before join with project root
- [ ] **Finding 18** | Rule: **PERF-22** | Correctly Fired: **Yes** | os.ReadFile invoked inside HTTP handler
- [ ] **Finding 19** | Rule: **PERF-112** | Correctly Fired: **Yes** | strings.ToLower used before string extension comparison
- [ ] **Finding 20** | Rule: **BP-1** | Correctly Fired: **Yes** | Close error discarded via `_ =` without handling comment
- [ ] **Finding 21** | Rule: **PERF-57** | Correctly Fired: **No** | io.ReadAll is in route handler, not Gin middleware
- [ ] **Finding 22** | Rule: **PERF-46** | Correctly Fired: **Yes** | strings.TrimSuffix allocates on request handler path
- [ ] **Finding 23** | Rule: **BP-1** | Correctly Fired: **Yes** | FormFile error return discarded with blank identifier
- [ ] **Finding 24** | Rule: **BP-1** | Correctly Fired: **Yes** | Close error discarded via `_ =` without handling comment
- [ ] **Finding 25** | Rule: **BP-1** | Correctly Fired: **Yes** | FormFile error return discarded with blank identifier
- [ ] **Finding 26** | Rule: **BP-1** | Correctly Fired: **Yes** | Close error discarded via `_ =` without handling comment
- [ ] **Finding 27** | Rule: **PERF-32** | Correctly Fired: **Yes** | `[]byte(b)` string-to-byte conversion in handler
- [ ] **Finding 28** | Rule: **PERF-32** | Correctly Fired: **Yes** | `[]byte(b)` string-to-byte conversion in handler
- [ ] **Finding 29** | Rule: **PERF-56** | Correctly Fired: **Yes** | c.JSON called inside for-loop body on error path
- [ ] **Finding 30** | Rule: **BP-1** | Correctly Fired: **Yes** | Close error discarded via `_ =` without handling comment

## Summary
- Total findings analyzed: 30
- Correctly Fired (True Positives): 18
- Incorrectly Fired (False Positives): 12
- FP rate: 40.0%

## Notable FP patterns observed
- [ ] **Framework mismatch:** PERF-88 (Echo static-cache rule) fired on Gin `router.Static` (finding 15).
- [ ] **Comment / anchor mis-association:** CWE-497 flagged a commented `runtime.NumCPU` line, not an exposed endpoint (finding 5).
- [ ] **Hot-path rules on cold/startup code:** PERF-151 on `main`, PERF-41 on startup profiling logs, PERF-35 on benchmark header helper (findings 1, 2, 12).
- [ ] **Conditional production paths ignored:** PERF-68 flagged `gin.Logger()` though it is gated behind `gin.DebugMode` (finding 4).
- [ ] **Inverted middleware-order logic:** PERF-200 reported Auth-before-CORS while snippet shows CORS before Auth (finding 16).
- [ ] **Channel semantics mismatch:** PERF-148 claimed unbuffered channel with no receiver on a buffered semaphore with defer receive (finding 6).
- [ ] **Handler vs middleware confusion:** PERF-57 flagged `io.ReadAll` in a route handler as middleware work (finding 21).
- [ ] **Sanitization missed:** CWE-22 flagged path traversal despite `filepath.Base` confinement before `os.ReadFile` (finding 17).
- [ ] **Non-error discard misclassified:** BP-1 flagged `runtime.Caller` `ok` bool discard as discarded error (finding 11).