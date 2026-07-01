# Agent 1b — Findings 031–060 Validation (Manual Review)

## Per-Finding Checklist
- [x] **Finding 031** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = pdfFile.Close()` explicitly discards Close error → Fixed in internal/handlers/handlers.go
- [x] **Finding 032** | Rule: **PERF-35** | Correctly Fired: **Yes** | fmt.Sprintf in handler loop boxes args via interface{} → Fixed in internal/handlers/handlers.go
- [x] **Finding 033** | Rule: **PERF-6** | Correctly Fired: **Yes** | fmt.Sprintf called inside for-loop body → Fixed in internal/handlers/handlers.go
- [x] **Finding 034** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = zw.Close()` discards error on zip-create failure → Fixed in internal/handlers/handlers.go
- [x] **Finding 035** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = zw.Close()` discards error on zip-write failure → Fixed in internal/handlers/handlers.go
- [x] **Finding 036** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = zw.Close()` discards error after successful zip build → Fixed in internal/handlers/handlers.go
- [x] **Finding 037** | Rule: **PERF-41** | Correctly Fired: **Yes** | log.Printf used at start of gin request handler → Fixed in internal/handlers/handlers.go
- [x] **Finding 038** | Rule: **PERF-109** | Correctly Fired: **No** | ToLower key varies per item, not recomputed constant → N/A: false positive — ToLower key varies per dedup term, not constant map key
- [x] **Finding 039** | Rule: **PERF-46** | Correctly Fired: **Yes** | strings.TrimSpace inside loop allocates per iteration → Fixed in internal/handlers/redact.go
- [x] **Finding 040** | Rule: **BP-1** | Correctly Fired: **Yes** | defer uses `_ = f.Close()` discarding returned error → Fixed in internal/handlers/redact.go
- [x] **Finding 041** | Rule: **BP-5** | Correctly Fired: **Yes** | Close() return value ignored via blank identifier → Fixed in internal/handlers/redact.go
- [x] **Finding 042** | Rule: **BP-1** | Correctly Fired: **Yes** | defer `_ = f.Close()` explicitly discards error → Fixed in internal/handlers/redact.go
- [x] **Finding 043** | Rule: **BP-5** | Correctly Fired: **Yes** | Close() return ignored in deferred cleanup → Fixed in internal/handlers/redact.go
- [x] **Finding 044** | Rule: **BP-1** | Correctly Fired: **Yes** | defer `_ = f.Close()` discards Close error → Fixed in internal/handlers/redact.go
- [x] **Finding 045** | Rule: **BP-5** | Correctly Fired: **Yes** | Close() return value ignored in defer → Fixed in internal/handlers/redact.go
- [x] **Finding 046** | Rule: **PERF-32** | Correctly Fired: **Yes** | `[]byte(blocksJSON)` copies string for json.Unmarshal → Fixed in internal/handlers/redact.go
- [x] **Finding 047** | Rule: **PERF-32** | Correctly Fired: **Yes** | `[]byte(textSearchJSON)` string-to-byte copy present → Fixed in internal/handlers/redact.go
- [x] **Finding 048** | Rule: **PERF-32** | Correctly Fired: **Yes** | `[]byte(textSearchJSON)` conversion copies underlying data → Fixed in internal/handlers/redact.go
- [x] **Finding 049** | Rule: **PERF-32** | Correctly Fired: **Yes** | `[]byte(ocrJSON)` conversion in request handler path → Fixed in internal/handlers/redact.go
- [x] **Finding 050** | Rule: **PERF-32** | Correctly Fired: **Yes** | `[]byte(redactionsJSON)` string-to-byte copy present → Fixed in internal/handlers/redact.go
- [x] **Finding 051** | Rule: **BP-1** | Correctly Fired: **Yes** | defer `_ = f.Close()` discards returned error → Fixed in internal/handlers/redact.go
- [x] **Finding 052** | Rule: **BP-5** | Correctly Fired: **Yes** | Close() ignored with blank identifier in defer → Fixed in internal/handlers/redact.go
- [x] **Finding 053** | Rule: **PERF-32** | Correctly Fired: **Yes** | `[]byte(textsJSON)` copies string for Unmarshal → Fixed in internal/handlers/redact.go
- [x] **Finding 054** | Rule: **BP-1** | Correctly Fired: **Yes** | defer `_ = f.Close()` explicitly discards error → Fixed in internal/handlers/redact.go
- [x] **Finding 055** | Rule: **BP-5** | Correctly Fired: **Yes** | Close() return ignored in deferred file cleanup → Fixed in internal/handlers/redact.go
- [x] **Finding 056** | Rule: **PERF-30** | Correctly Fired: **No** | context.Background used synchronously; no goroutine shown → N/A: false positive — synchronous middleware, no goroutine
- [x] **Finding 057** | Rule: **BP-13** | Correctly Fired: **Yes** | context.Background() used instead of caller/request context → Fixed in internal/middleware/auth.go
- [x] **Finding 058** | Rule: **BP-13** | Correctly Fired: **Yes** | context.Background() instead of propagated request context → Fixed in internal/middleware/auth.go
- [x] **Finding 059** | Rule: **PERF-192** | Correctly Fired: **Yes** | make(map) called without a size hint → Fixed in internal/middleware/auth.go
- [x] **Finding 060** | Rule: **PERF-201** | Correctly Fired: **Yes** | Custom OPTIONS branch implements manual preflight handling → Fixed in internal/middleware/cors.go

## Summary
- Total findings analyzed: 30
- Correctly Fired (True Positive detections): 28
- Incorrectly Fired (False Positive detections): 2
- FP rate: 6.7%

## Notable FP patterns observed
- [ ] **PERF-109 on per-item keys:** Rule flagged `strings.ToLower(term)` inside a dedup loop where each map key is derived from a distinct input term, not a constant key recomputed every iteration (finding 038).
- [ ] **PERF-30 without goroutine:** Rule title requires context.Background in a goroutine spawned from a request; snippet shows only synchronous middleware usage with no `go` statement (finding 056).
- [ ] **Duplicate BP-1/BP-5 pairs:** Multiple findings on the same `defer func() { _ = f.Close() }()` line fire both discarded-error and ignored-Close rules; each rule matches its stated pattern independently (findings 040–045, 051–052, 054–055).