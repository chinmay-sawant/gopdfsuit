# Agent 1b — Findings 031–060 Validation (Manual Review)

## Per-Finding Table
| Finding | Rule | Correctly Fired? | Reason |
|---------|------|------------------|--------|
| 031 | BP-1 | Yes | `_ = pdfFile.Close()` explicitly discards Close error |
| 032 | PERF-35 | Yes | fmt.Sprintf in handler loop boxes args via interface{} |
| 033 | PERF-6 | Yes | fmt.Sprintf called inside for-loop body |
| 034 | BP-1 | Yes | `_ = zw.Close()` discards error on zip-create failure |
| 035 | BP-1 | Yes | `_ = zw.Close()` discards error on zip-write failure |
| 036 | BP-1 | Yes | `_ = zw.Close()` discards error after successful zip build |
| 037 | PERF-41 | Yes | log.Printf used at start of gin request handler |
| 038 | PERF-109 | No | ToLower key varies per item, not recomputed constant |
| 039 | PERF-46 | Yes | strings.TrimSpace inside loop allocates per iteration |
| 040 | BP-1 | Yes | defer uses `_ = f.Close()` discarding returned error |
| 041 | BP-5 | Yes | Close() return value ignored via blank identifier |
| 042 | BP-1 | Yes | defer `_ = f.Close()` explicitly discards error |
| 043 | BP-5 | Yes | Close() return ignored in deferred cleanup |
| 044 | BP-1 | Yes | defer `_ = f.Close()` discards Close error |
| 045 | BP-5 | Yes | Close() return value ignored in defer |
| 046 | PERF-32 | Yes | `[]byte(blocksJSON)` copies string for json.Unmarshal |
| 047 | PERF-32 | Yes | `[]byte(textSearchJSON)` string-to-byte copy present |
| 048 | PERF-32 | Yes | `[]byte(textSearchJSON)` conversion copies underlying data |
| 049 | PERF-32 | Yes | `[]byte(ocrJSON)` conversion in request handler path |
| 050 | PERF-32 | Yes | `[]byte(redactionsJSON)` string-to-byte copy present |
| 051 | BP-1 | Yes | defer `_ = f.Close()` discards returned error |
| 052 | BP-5 | Yes | Close() ignored with blank identifier in defer |
| 053 | PERF-32 | Yes | `[]byte(textsJSON)` copies string for Unmarshal |
| 054 | BP-1 | Yes | defer `_ = f.Close()` explicitly discards error |
| 055 | BP-5 | Yes | Close() return ignored in deferred file cleanup |
| 056 | PERF-30 | No | context.Background used synchronously; no goroutine shown |
| 057 | BP-13 | Yes | context.Background() used instead of caller/request context |
| 058 | BP-13 | Yes | context.Background() instead of propagated request context |
| 059 | PERF-192 | Yes | make(map) called without a size hint |
| 060 | PERF-201 | Yes | Custom OPTIONS branch implements manual preflight handling |

## Summary
- Total findings analyzed: 30
- Correctly Fired (True Positive detections): 28
- Incorrectly Fired (False Positive detections): 2
- FP rate: 6.7%

## Notable FP patterns observed
- **PERF-109 on per-item keys:** Rule flagged `strings.ToLower(term)` inside a dedup loop where each map key is derived from a distinct input term, not a constant key recomputed every iteration (finding 038).
- **PERF-30 without goroutine:** Rule title requires context.Background in a goroutine spawned from a request; snippet shows only synchronous middleware usage with no `go` statement (finding 056).
- **Duplicate BP-1/BP-5 pairs:** Multiple findings on the same `defer func() { _ = f.Close() }()` line fire both discarded-error and ignored-Close rules; each rule matches its stated pattern independently (findings 040–045, 051–052, 054–055).