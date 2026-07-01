# Agent 1b — Findings 031–060 Validation (Manual Review)

## Per-Finding Checklist
- [ ] **Finding 031** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = pdfFile.Close()` explicitly discards Close error
- [ ] **Finding 032** | Rule: **PERF-35** | Correctly Fired: **Yes** | fmt.Sprintf in handler loop boxes args via interface{}
- [ ] **Finding 033** | Rule: **PERF-6** | Correctly Fired: **Yes** | fmt.Sprintf called inside for-loop body
- [ ] **Finding 034** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = zw.Close()` discards error on zip-create failure
- [ ] **Finding 035** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = zw.Close()` discards error on zip-write failure
- [ ] **Finding 036** | Rule: **BP-1** | Correctly Fired: **Yes** | `_ = zw.Close()` discards error after successful zip build
- [ ] **Finding 037** | Rule: **PERF-41** | Correctly Fired: **Yes** | log.Printf used at start of gin request handler
- [ ] **Finding 038** | Rule: **PERF-109** | Correctly Fired: **No** | ToLower key varies per item, not recomputed constant
- [ ] **Finding 039** | Rule: **PERF-46** | Correctly Fired: **Yes** | strings.TrimSpace inside loop allocates per iteration
- [ ] **Finding 040** | Rule: **BP-1** | Correctly Fired: **Yes** | defer uses `_ = f.Close()` discarding returned error
- [ ] **Finding 041** | Rule: **BP-5** | Correctly Fired: **Yes** | Close() return value ignored via blank identifier
- [ ] **Finding 042** | Rule: **BP-1** | Correctly Fired: **Yes** | defer `_ = f.Close()` explicitly discards error
- [ ] **Finding 043** | Rule: **BP-5** | Correctly Fired: **Yes** | Close() return ignored in deferred cleanup
- [ ] **Finding 044** | Rule: **BP-1** | Correctly Fired: **Yes** | defer `_ = f.Close()` discards Close error
- [ ] **Finding 045** | Rule: **BP-5** | Correctly Fired: **Yes** | Close() return value ignored in defer
- [ ] **Finding 046** | Rule: **PERF-32** | Correctly Fired: **Yes** | `[]byte(blocksJSON)` copies string for json.Unmarshal
- [ ] **Finding 047** | Rule: **PERF-32** | Correctly Fired: **Yes** | `[]byte(textSearchJSON)` string-to-byte copy present
- [ ] **Finding 048** | Rule: **PERF-32** | Correctly Fired: **Yes** | `[]byte(textSearchJSON)` conversion copies underlying data
- [ ] **Finding 049** | Rule: **PERF-32** | Correctly Fired: **Yes** | `[]byte(ocrJSON)` conversion in request handler path
- [ ] **Finding 050** | Rule: **PERF-32** | Correctly Fired: **Yes** | `[]byte(redactionsJSON)` string-to-byte copy present
- [ ] **Finding 051** | Rule: **BP-1** | Correctly Fired: **Yes** | defer `_ = f.Close()` discards returned error
- [ ] **Finding 052** | Rule: **BP-5** | Correctly Fired: **Yes** | Close() ignored with blank identifier in defer
- [ ] **Finding 053** | Rule: **PERF-32** | Correctly Fired: **Yes** | `[]byte(textsJSON)` copies string for Unmarshal
- [ ] **Finding 054** | Rule: **BP-1** | Correctly Fired: **Yes** | defer `_ = f.Close()` explicitly discards error
- [ ] **Finding 055** | Rule: **BP-5** | Correctly Fired: **Yes** | Close() return ignored in deferred file cleanup
- [ ] **Finding 056** | Rule: **PERF-30** | Correctly Fired: **No** | context.Background used synchronously; no goroutine shown
- [ ] **Finding 057** | Rule: **BP-13** | Correctly Fired: **Yes** | context.Background() used instead of caller/request context
- [ ] **Finding 058** | Rule: **BP-13** | Correctly Fired: **Yes** | context.Background() instead of propagated request context
- [ ] **Finding 059** | Rule: **PERF-192** | Correctly Fired: **Yes** | make(map) called without a size hint
- [ ] **Finding 060** | Rule: **PERF-201** | Correctly Fired: **Yes** | Custom OPTIONS branch implements manual preflight handling

## Summary
- Total findings analyzed: 30
- Correctly Fired (True Positive detections): 28
- Incorrectly Fired (False Positive detections): 2
- FP rate: 6.7%

## Notable FP patterns observed
- [ ] **PERF-109 on per-item keys:** Rule flagged `strings.ToLower(term)` inside a dedup loop where each map key is derived from a distinct input term, not a constant key recomputed every iteration (finding 038).
- [ ] **PERF-30 without goroutine:** Rule title requires context.Background in a goroutine spawned from a request; snippet shows only synchronous middleware usage with no `go` statement (finding 056).
- [ ] **Duplicate BP-1/BP-5 pairs:** Multiple findings on the same `defer func() { _ = f.Close() }()` line fire both discarded-error and ignored-Close rules; each rule matches its stated pattern independently (findings 040–045, 051–052, 054–055).