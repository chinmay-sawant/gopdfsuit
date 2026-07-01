# Agent 1a — Findings 001–030 Validation (Manual Review)

## Per-Finding Checklist
- [x] **Finding 1** | Rule: **PERF-151** | Correctly Fired: **No** | main is startup code, not a per-request hot-path handler → N/A: false positive — `main()` is one-time startup, not a request handler
- [x] **Finding 2** | Rule: **PERF-41** | Correctly Fired: **No** | flagged log.Printf is in startup profiling setup, not request path → N/A: false positive — profiling branch runs only when `ENABLE_PROFILING=1`
- [x] **Finding 3** | Rule: **PERF-43** | Correctly Fired: **Yes** | defer-recover runs in per-request Gin middleware → Cannot fix: per-request recovery is required in Gin; custom handler is intentional lighter alternative to `gin.Recovery()`
- [x] **Finding 4** | Rule: **PERF-68** | Correctly Fired: **No** | gin.Logger registered only when gin.DebugMode, not production → N/A: false positive — logger gated behind `gin.DebugMode`
- [x] **Finding 5** | Rule: **CWE-497** | Correctly Fired: **No** | highlighted line is a comment, not a diagnostics endpoint → N/A: false positive — anchor is a commented line, not exposed diagnostics
- [x] **Finding 6** | Rule: **PERF-148** | Correctly Fired: **No** | channel is buffered and middleware receives in same function → N/A: false positive — semaphore channel is buffered with defer receive
- [x] **Finding 7** | Rule: **PERF-195** | Correctly Fired: **Yes** | log.Fatalf is called inside a goroutine → Fixed in `cmd/gopdfsuit/main.go` — listen errors sent via channel; `os.Exit` called from main goroutine
- [x] **Finding 8** | Rule: **BP-2** | Correctly Fired: **Yes** | bare `return err` without wrapping → Fixed in `internal/benchmarktemplates/runner.go` — wrapped with `fmt.Errorf`
- [x] **Finding 9** | Rule: **PERF-40** | Correctly Fired: **Yes** | multiple time.Now calls in same function body → Cannot fix: intentional per-iteration benchmark timing requires separate `time.Now()` per run
- [x] **Finding 10** | Rule: **BP-2** | Correctly Fired: **Yes** | bare `return err` without wrapping → Fixed in `internal/benchmarktemplates/runner.go` — wrapped with `fmt.Errorf`
- [x] **Finding 11** | Rule: **BP-1** | Correctly Fired: **No** | discarded value is runtime.Caller ok bool, not an error → N/A: false positive — blank identifier binds `ok` bool, not an error return
- [x] **Finding 12** | Rule: **PERF-35** | Correctly Fired: **No** | fmt.Sprintf in one-off benchmark header, not hot path → N/A: false positive — one-off benchmark header helper, not request hot path
- [x] **Finding 13** | Rule: **PERF-151** | Correctly Fired: **No** | utility helper, not a request handler on hot path → N/A: false positive — `getProjectRoot()` is a cold-path utility
- [x] **Finding 14** | Rule: **PERF-61** | Correctly Fired: **Yes** | router.Static with no cache headers shown in snippet → Fixed in `internal/handlers/handlers.go` — custom static handler sets `Cache-Control: public, max-age=31536000, immutable`
- [x] **Finding 15** | Rule: **PERF-88** | Correctly Fired: **No** | Echo static-cache rule applied to Gin router.Static → N/A: false positive — Echo-specific rule misfired on Gin
- [x] **Finding 16** | Rule: **PERF-200** | Correctly Fired: **No** | CORS is registered before Auth; ordering already correct → N/A: false positive — snippet shows CORS before Auth as intended
- [x] **Finding 17** | Rule: **CWE-22** | Correctly Fired: **No** | filepath.Base confines user input before join with project root → N/A: false positive — `filepath.Base` prevents traversal
- [x] **Finding 18** | Rule: **PERF-22** | Correctly Fired: **Yes** | os.ReadFile invoked inside HTTP handler → Fixed in `internal/handlers/handlers.go` — `sync.Map` template cache avoids per-request disk I/O
- [x] **Finding 19** | Rule: **PERF-112** | Correctly Fired: **Yes** | strings.ToLower used before string extension comparison → Fixed in `internal/handlers/handlers.go` — `strings.EqualFold` on extension
- [x] **Finding 20** | Rule: **BP-1** | Correctly Fired: **Yes** | Close error discarded via `_ =` without handling comment → Fixed in `internal/handlers/handlers.go` — close errors logged in defer
- [x] **Finding 21** | Rule: **PERF-57** | Correctly Fired: **No** | io.ReadAll is in route handler, not Gin middleware → N/A: false positive — `io.ReadAll` is in route handler, not middleware
- [x] **Finding 22** | Rule: **PERF-46** | Correctly Fired: **Yes** | strings.TrimSuffix allocates on request handler path → Fixed in `internal/handlers/handlers.go` — `strings.CutSuffix` instead of `TrimSuffix`
- [x] **Finding 23** | Rule: **BP-1** | Correctly Fired: **Yes** | FormFile error return discarded with blank identifier → Fixed in `internal/handlers/handlers.go` — FormFile errors checked (allows `ErrMissingFile`)
- [x] **Finding 24** | Rule: **BP-1** | Correctly Fired: **Yes** | Close error discarded via `_ =` without handling comment → Fixed in `internal/handlers/handlers.go` — close errors logged in defer
- [x] **Finding 25** | Rule: **BP-1** | Correctly Fired: **Yes** | FormFile error return discarded with blank identifier → Fixed in `internal/handlers/handlers.go` — FormFile errors checked (allows `ErrMissingFile`)
- [x] **Finding 26** | Rule: **BP-1** | Correctly Fired: **Yes** | Close error discarded via `_ =` without handling comment → Fixed in `internal/handlers/handlers.go` — close errors logged in defer
- [x] **Finding 27** | Rule: **PERF-32** | Correctly Fired: **Yes** | `[]byte(b)` string-to-byte conversion in handler → Cannot fix: `[]byte(string)` required for binary form-field fallback; no allocation-free alternative without `unsafe`
- [x] **Finding 28** | Rule: **PERF-32** | Correctly Fired: **Yes** | `[]byte(b)` string-to-byte conversion in handler → Cannot fix: same as finding 27 — binary post-form data requires byte slice
- [x] **Finding 29** | Rule: **PERF-56** | Correctly Fired: **Yes** | c.JSON called inside for-loop body on error path → Fixed in `internal/handlers/handlers.go` — error responses use formatted messages; close handled outside JSON branch
- [x] **Finding 30** | Rule: **BP-1** | Correctly Fired: **Yes** | Close error discarded via `_ =` without handling comment → Fixed in `internal/handlers/handlers.go` — upload close errors logged

## Summary
- Total findings analyzed: 30
- Correctly Fired (True Positives): 18
- Incorrectly Fired (False Positives): 12
- FP rate: 40.0%
- Remediation: 13 fixed, 3 cannot fix, 12 N/A (false positive), 2 cannot fix (PERF-43, PERF-40)

## Notable FP patterns observed
- [x] **Framework mismatch:** PERF-88 (Echo static-cache rule) fired on Gin `router.Static` (finding 15). → Addressed: replaced with custom Gin static handler
- [x] **Comment / anchor mis-association:** CWE-497 flagged a commented `runtime.NumCPU` line, not an exposed endpoint (finding 5).
- [x] **Hot-path rules on cold/startup code:** PERF-151 on `main`, PERF-41 on startup profiling logs, PERF-35 on benchmark header helper (findings 1, 2, 12).
- [x] **Conditional production paths ignored:** PERF-68 flagged `gin.Logger()` though it is gated behind `gin.DebugMode` (finding 4).
- [x] **Inverted middleware-order logic:** PERF-200 reported Auth-before-CORS while snippet shows CORS before Auth (finding 16).
- [x] **Channel semantics mismatch:** PERF-148 claimed unbuffered channel with no receiver on a buffered semaphore with defer receive (finding 6).
- [x] **Handler vs middleware confusion:** PERF-57 flagged `io.ReadAll` in a route handler as middleware work (finding 21).
- [x] **Sanitization missed:** CWE-22 flagged path traversal despite `filepath.Base` confinement before `os.ReadFile` (finding 17).
- [x] **Non-error discard misclassified:** BP-1 flagged `runtime.Caller` `ok` bool discard as discarded error (finding 11).