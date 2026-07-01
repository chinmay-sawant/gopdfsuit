# Agent 1 — Findings 1–150 Validation (Manual Review)

> Merged from five parallel subagents: 1a (1–30), 1b (31–60), 1c (61–90), 1d (91–120), 1e (121–150)

## Summary

| Sub-agent | Range | TP | FP | FP rate |
|-----------|-------|---:|---:|--------:|
| 1a | 1–30 | 18 | 12 | 40.0% |
| 1b | 31–60 | 28 | 2 | 6.7% |
| 1c | 61–90 | 15 | 15 | 50.0% |
| 1d | 91–120 | 22 | 8 | 26.7% |
| 1e | 121–150 | 29 | 1 | 3.3% |
| **Total** | **1–150** | **112** | **38** | **25.3%** |

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
| 31 | BP-1 | Yes | `_ = pdfFile.Close()` explicitly discards Close error |
| 32 | PERF-35 | Yes | fmt.Sprintf in handler loop boxes args via interface{} |
| 33 | PERF-6 | Yes | fmt.Sprintf called inside for-loop body |
| 34 | BP-1 | Yes | `_ = zw.Close()` discards error on zip-create failure |
| 35 | BP-1 | Yes | `_ = zw.Close()` discards error on zip-write failure |
| 36 | BP-1 | Yes | `_ = zw.Close()` discards error after successful zip build |
| 37 | PERF-41 | Yes | log.Printf used at start of gin request handler |
| 38 | PERF-109 | No | ToLower key varies per item, not recomputed constant |
| 39 | PERF-46 | Yes | strings.TrimSpace inside loop allocates per iteration |
| 40 | BP-1 | Yes | defer uses `_ = f.Close()` discarding returned error |
| 41 | BP-5 | Yes | Close() return value ignored via blank identifier |
| 42 | BP-1 | Yes | defer `_ = f.Close()` explicitly discards error |
| 43 | BP-5 | Yes | Close() return ignored in deferred cleanup |
| 44 | BP-1 | Yes | defer `_ = f.Close()` discards Close error |
| 45 | BP-5 | Yes | Close() return value ignored in defer |
| 46 | PERF-32 | Yes | `[]byte(blocksJSON)` copies string for json.Unmarshal |
| 47 | PERF-32 | Yes | `[]byte(textSearchJSON)` string-to-byte copy present |
| 48 | PERF-32 | Yes | `[]byte(textSearchJSON)` conversion copies underlying data |
| 49 | PERF-32 | Yes | `[]byte(ocrJSON)` conversion in request handler path |
| 50 | PERF-32 | Yes | `[]byte(redactionsJSON)` string-to-byte copy present |
| 51 | BP-1 | Yes | defer `_ = f.Close()` discards returned error |
| 52 | BP-5 | Yes | Close() ignored with blank identifier in defer |
| 53 | PERF-32 | Yes | `[]byte(textsJSON)` copies string for Unmarshal |
| 54 | BP-1 | Yes | defer `_ = f.Close()` explicitly discards error |
| 55 | BP-5 | Yes | Close() return ignored in deferred file cleanup |
| 56 | PERF-30 | No | context.Background used synchronously; no goroutine shown |
| 57 | BP-13 | Yes | context.Background() used instead of caller/request context |
| 58 | BP-13 | Yes | context.Background() instead of propagated request context |
| 59 | PERF-192 | Yes | make(map) called without a size hint |
| 60 | PERF-201 | Yes | Custom OPTIONS branch implements manual preflight handling |
| 61 | PERF-119 | No | strconv.AppendInt separates the two append calls |
| 62 | PERF-128 | Yes | Three consecutive append calls on same buffer |
| 63 | PERF-119 | Yes | Consecutive append(dst,'.') and append(dst,'0') calls |
| 64 | PERF-128 | No | appendFmtNum calls break three-or-more append sequence |
| 65 | BP-1 | No | parseHexColor returns bool valid, not an error |
| 66 | BP-1 | No | parseHexColor returns bool valid, not an error |
| 67 | BP-1 | No | parseHexColor returns bool valid, not an error |
| 68 | PERF-15 | Yes | strconv.Itoa inside nested table cell loop |
| 69 | BP-1 | No | parseHexColor returns bool valid, not an error |
| 70 | PERF-15 | Yes | strconv.Itoa inside nested table cell loop |
| 71 | BP-1 | No | parseHexColor returns bool valid, not an error |
| 72 | BP-1 | No | parseHexColor returns bool valid, not an error |
| 73 | PERF-15 | Yes | strconv.Itoa inside nested table row loop |
| 74 | BP-1 | No | parseHexColor returns bool valid, not an error |
| 75 | PERF-15 | Yes | strconv.Itoa inside nested table row loop |
| 76 | PERF-35 | Yes | fmt.Sprintf boxes args during per-page rendering |
| 77 | PERF-42 | No | One-time constructor validation, not hot path |
| 78 | PERF-32 | No | One-time encryption setup, not hot path |
| 79 | CWE-916 | Yes | MD5 password hash with insufficient computational effort |
| 80 | CWE-328 | Yes | MD5 used for password digest derivation |
| 81 | PERF-3 | Yes | make([]byte) rebuilt inside owner-hash loop |
| 82 | PERF-3 | Yes | make([]byte) rebuilt inside user-hash loop |
| 83 | PERF-35 | No | One-time encrypt dictionary build, not hot path |
| 84 | PERF-110 | No | Pool New returns *zlib.Writer pointer, not value |
| 85 | BP-1 | Yes | zlib.NewWriterLevel error discarded into blank identifier |
| 86 | PERF-109 | No | Slice range loop, no map key recomputation |
| 87 | PERF-15 | Yes | strconv.Itoa inside metrics.Widths loop |
| 88 | PERF-15 | Yes | strconv.Itoa inside metrics.Widths loop |
| 89 | PERF-35 | No | One-time font resource string build, not hot path |
| 90 | PERF-192 | Yes | make(map) without known object-count size hint |
| 91 | BP-1 | Yes | `_ = zlibWriter.Close()` discards Close error without comment |
| 92 | BP-1 | Yes | `_ = zlibWriter.Close()` discards Close error without comment |
| 93 | PERF-15 | Yes | `strconv.Itoa` called inside outer width-range loop |
| 94 | PERF-15 | Yes | `strconv.Itoa` called inside inner width-write loop |
| 95 | BP-1 | Yes | `_ = zlibWriter.Close()` discards Close error on failure path |
| 96 | BP-1 | Yes | `_ = zlibWriter.Close()` discards Close error after success |
| 97 | PERF-15 | Yes | `strconv.Itoa` called inside CMap chunking loop |
| 98 | BP-1 | Yes | `_ = zlibWriter.Close()` discards Close error on failure path |
| 99 | BP-1 | Yes | `_ = zlibWriter.Close()` discards Close error after success |
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

## Notable FP Patterns (Findings 1–150)

1. **Startup/main misclassified as hot path** — PERF-151, PERF-41 on `main()` and profiling setup (1, 2, 13)
2. **Framework/order mismatches** — PERF-88 Echo rule on Gin (15), PERF-200 inverted CORS check (16)
3. **PERF-31 split verdict** — 4 FPs in 116–120 (registry helpers), 11 TPs in 121–141 (syntactic defer match; rule over-broad)
4. **BP-1 on parseHexColor** — 8 FPs discarding alpha `float64`, not error (65–67, 69, 71–72, 74)
5. **Hot-path rules on cold paths** — PERF-35/42/32 on one-time setup (77–78, 83, 89, 102–103, 115)
6. **PERF-109 on slice/dedup loops** — not constant map key recomputation (38, 86, 134)

## Partial Reports

- [`agent-1a-findings-001-030.md`](agent-1a-findings-001-030.md)
- [`agent-1b-findings-031-060.md`](agent-1b-findings-031-060.md)
- [`agent-1c-findings-061-090.md`](agent-1c-findings-061-090.md)
- [`agent-1d-findings-091-120.md`](agent-1d-findings-091-120.md)
- [`agent-1e-findings-121-150.md`](agent-1e-findings-121-150.md)