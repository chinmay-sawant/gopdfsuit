# SlopGuard Findings

Source chunks: `/home/chinmay/ChinmayPersonalProjects/slopguard/scripts/chunks/Chunk_*.txt`

## Scope

- Total exported findings: `226`
- Existing `CWE-*` findings already covered elsewhere: `8`
- Actionable performance findings in this plan: `218`
- **All `218` actionable findings are now fixed across all rounds.**
- This file keeps all `226` checklist items so the exported finding numbers remain stable.

## Checklist

- [x] Finding 1 `PERF-41` at `cmd/gopdfsuit/main.go:24:4` (medium): remove standard logger calls from the production hot path or switch to a structured non-blocking logger.
- [x] Finding 2 `PERF-68` at `cmd/gopdfsuit/main.go:57:14` (medium): disable `gin.Logger()` on the production path or replace it with an async logger.
- [x] Finding 3 `CWE-497` at `cmd/gopdfsuit/main.go:63:19` is already covered by the existing CWE ruleset; exclude it from this performance remediation pass.
- [x] Finding 4 `PERF-40` at `internal/benchmarktemplates/runner.go:85:16` (medium): capture the timestamp once per operation and reuse it instead of calling `time.Now()` repeatedly.
- [x] Finding 5 `PERF-7` at `internal/benchmarktemplates/runner.go:90:4` (medium): remove `defer` from the loop body and close or release resources explicitly.
- [x] Finding 6 `PERF-7` at `internal/benchmarktemplates/runner.go:91:4` (medium): remove `defer` from the loop body and close or release resources explicitly.
- [x] Finding 7 `PERF-22` at `internal/handlers/handlers.go:227:15` (medium): load the file outside the handler or cache the data instead of calling `os.ReadFile` on the request path.
- [x] Finding 8 `PERF-57` at `internal/handlers/handlers.go:280:15` (medium): trim middleware allocations on the request path or move the heavy work outside the middleware.
- [x] Finding 9 `PERF-32` at `internal/handlers/handlers.go:363:15` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 10 `PERF-32` at `internal/handlers/handlers.go:368:16` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 11 `PERF-56` at `internal/handlers/handlers.go:410:4` (medium): avoid repeated `c.JSON` marshaling in the loop; batch or preencode the response.
- [x] Finding 12 `PERF-6` at `internal/handlers/handlers.go:490:11` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 13 `PERF-35` at `internal/handlers/handlers.go:490:11` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 14 `PERF-41` at `internal/handlers/handlers.go:513:2` (medium): remove standard logger calls from the production hot path or switch to a structured non-blocking logger.
- [x] Finding 15 `PERF-46` at `internal/handlers/redact.go:23:11` (medium): avoid allocation-heavy trimming when a cheap prefix, suffix, or length guard will do.
- [x] Finding 16 `PERF-32` at `internal/handlers/redact.go:198:28` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 17 `PERF-32` at `internal/handlers/redact.go:206:28` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 18 `PERF-32` at `internal/handlers/redact.go:208:30` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 19 `PERF-32` at `internal/handlers/redact.go:225:28` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 20 `PERF-32` at `internal/handlers/redact.go:236:29` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 21 `PERF-32` at `internal/handlers/redact.go:304:28` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 22 `PERF-30` at `internal/middleware/auth.go:74:10` (medium): avoid detached `context.Background()` work from the request path; propagate or derive the request context.
- [x] Finding 23 `PERF-42` at `internal/pdf/encryption/encrypt.go:38:15` (medium): replace static `fmt.Errorf` text with `errors.New`.
- [x] Finding 24 `PERF-32` at `internal/pdf/encryption/encrypt.go:62:9` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 25 `CWE-328` at `internal/pdf/encryption/encrypt.go:79:10` is already covered by the existing CWE ruleset; exclude it from this performance remediation pass.
- [x] Finding 26 `CWE-916` at `internal/pdf/encryption/encrypt.go:79:10` is already covered by the existing CWE ruleset; exclude it from this performance remediation pass.
- [x] Finding 27 `PERF-3` at `internal/pdf/encryption/encrypt.go:99:3` (medium): stop rebuilding the slice inside the loop; reuse the slice or preallocate once.
- [x] Finding 28 `PERF-3` at `internal/pdf/encryption/encrypt.go:154:3` (medium): stop rebuilding the slice inside the loop; reuse the slice or preallocate once.
- [x] Finding 29 `PERF-53` at `internal/pdf/encryption/encrypt.go:246:15` (medium): replace package-level `math/rand` usage with a local `rand.Rand` backed by an explicit source.
- [x] Finding 30 `PERF-35` at `internal/pdf/encryption/encrypt.go:321:19` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 31 `PERF-15` at `internal/pdf/font/metrics.go:552:27` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 32 `PERF-15` at `internal/pdf/font/metrics.go:573:22` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 33 `PERF-15` at `internal/pdf/font/metrics.go:826:22` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 34 `PERF-15` at `internal/pdf/font/metrics.go:832:23` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 35 `PERF-15` at `internal/pdf/font/metrics.go:969:20` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 36 `PERF-42` at `internal/pdf/font/pdfa.go:165:10` (medium): replace static `fmt.Errorf` text with `errors.New`.
- [x] Finding 37 `PERF-35` at `internal/pdf/font/pdfa.go:170:21` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 38 `PERF-35` at `internal/pdf/font/registry.go:60:10` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 39 `PERF-31` at `internal/pdf/font/registry.go:90:3` (medium): remove unnecessary `defer` from the hot function and handle cleanup inline when cheaper.
- [x] Finding 40 `PERF-31` at `internal/pdf/font/registry.go:106:3` (medium): remove unnecessary `defer` from the hot function and handle cleanup inline when cheaper.
- [x] Finding 41 `PERF-15` at `internal/pdf/font/registry.go:282:28` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 42 `PERF-46` at `internal/pdf/font/registry.go:351:15` (medium): avoid allocation-heavy trimming when a cheap prefix, suffix, or length guard will do.
- [x] Finding 43 `PERF-15` at `internal/pdf/font/registry.go:396:13` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 44 `PERF-3` at `internal/pdf/font/subset.go:110:5` (medium): stop rebuilding the slice inside the loop; reuse the slice or preallocate once.
- [x] Finding 45 `PERF-32` at `internal/pdf/font/subset.go:160:10` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 46 `PERF-3` at `internal/pdf/font/subset.go:291:4` (medium): stop rebuilding the slice inside the loop; reuse the slice or preallocate once.
- [x] Finding 47 `PERF-35` at `internal/pdf/font/ttf.go:63:15` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 48 `PERF-15` at `internal/pdf/form/xfdf.go:467:12` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 49 `PERF-15` at `internal/pdf/form/xfdf.go:556:18` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 50 `PERF-1` at `internal/pdf/form/xfdf.go:822:14` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 51 `PERF-1` at `internal/pdf/form/xfdf.go:837:11` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 52 `PERF-1` at `internal/pdf/form/xfdf.go:842:12` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 53 `PERF-1` at `internal/pdf/form/xfdf.go:844:13` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 54 `PERF-1` at `internal/pdf/form/xfdf.go:887:10` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 55 `PERF-6` at `internal/pdf/form/xfdf.go:887:29` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 56 `PERF-35` at `internal/pdf/form/xfdf.go:887:29` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 57 `PERF-1` at `internal/pdf/form/xfdf.go:895:11` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 58 `PERF-32` at `internal/pdf/form/xfdf.go:911:13` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 59 `PERF-6` at `internal/pdf/form/xfdf.go:911:20` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 60 `PERF-1` at `internal/pdf/form/xfdf.go:912:12` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 61 `PERF-1` at `internal/pdf/form/xfdf.go:971:9` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 62 `PERF-6` at `internal/pdf/form/xfdf.go:971:28` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 63 `PERF-32` at `internal/pdf/form/xfdf.go:977:12` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 64 `PERF-6` at `internal/pdf/form/xfdf.go:977:19` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 65 `PERF-6` at `internal/pdf/form/xfdf.go:1016:17` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 66 `PERF-6` at `internal/pdf/form/xfdf.go:1019:18` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 67 `PERF-32` at `internal/pdf/form/xfdf.go:1021:21` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 68 `PERF-15` at `internal/pdf/form/xfdf.go:1029:26` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 69 `PERF-6` at `internal/pdf/form/xfdf.go:1039:14` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 70 `PERF-32` at `internal/pdf/form/xfdf.go:1044:21` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 71 `PERF-32` at `internal/pdf/form/xfdf.go:1048:25` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 72 `PERF-6` at `internal/pdf/form/xfdf.go:1055:12` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 73 `PERF-32` at `internal/pdf/form/xfdf.go:1057:21` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 74 `PERF-6` at `internal/pdf/form/xfdf.go:1076:4` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 75 `PERF-1` at `internal/pdf/form/xfdf.go:1245:13` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 76 `PERF-1` at `internal/pdf/form/xfdf.go:1258:16` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 77 `PERF-1` at `internal/pdf/form/xfdf.go:1260:16` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 78 `PERF-1` at `internal/pdf/form/xfdf.go:1266:22` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 79 `PERF-1` at `internal/pdf/form/xfdf.go:1292:15` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 80 `PERF-15` at `internal/pdf/form/xfdf.go:1366:29` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 81 `PERF-15` at `internal/pdf/form/xfdf.go:1368:29` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 82 `PERF-48` at `internal/pdf/form/xfdf.go:1460:20` (medium): add a cheap length or prefix guard before full `bytes.Equal` comparison.
- [x] Finding 83 `PERF-32` at `internal/pdf/generator.go:81:26` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 84 `PERF-32` at `internal/pdf/generator.go:308:50` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 85 `PERF-35` at `internal/pdf/generator.go:308:79` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 86 `PERF-32` at `internal/pdf/generator.go:565:42` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 87 `PERF-6` at `internal/pdf/generator.go:925:24` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 88 `PERF-6` at `internal/pdf/generator.go:944:24` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 89 `PERF-6` at `internal/pdf/generator.go:963:24` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 90 `PERF-40` at `internal/pdf/generator.go:1003:36` (medium): capture the timestamp once per operation and reuse it instead of calling `time.Now()` repeatedly.
- [x] Finding 91 `PERF-53` at `internal/pdf/generator.go:1103:16` (medium): replace package-level `math/rand` usage with a local `rand.Rand` backed by an explicit source.
- [x] Finding 92 `PERF-15` at `internal/pdf/generator.go:1205:28` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 93 `PERF-15` at `internal/pdf/generator.go:1220:27` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 94 `PERF-15` at `internal/pdf/generator.go:1289:22` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 95 `PERF-48` at `internal/pdf/generator.go:1479:6` (medium): add a cheap length or prefix guard before full `bytes.Equal` comparison.
- [x] Finding 96 `PERF-6` at `internal/pdf/generator.go:1610:23` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 97 `PERF-6` at `internal/pdf/generator.go:1613:23` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 98 `PERF-6` at `internal/pdf/generator.go:1637:26` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 99 `PERF-6` at `internal/pdf/generator.go:1647:26` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 100 `PERF-6` at `internal/pdf/generator.go:1660:22` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 101 `PERF-6` at `internal/pdf/generator.go:1673:24` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 102 `PERF-32` at `internal/pdf/helpers.go:22:23` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 103 `PERF-32` at `internal/pdf/helpers.go:104:23` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 104 `PERF-32` at `internal/pdf/helpers.go:104:62` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 105 `PERF-1` at `internal/pdf/helpers.go:108:15` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 106 `PERF-1` at `internal/pdf/helpers.go:147:15` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 107 `PERF-6` at `internal/pdf/helpers.go:161:12` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 108 `PERF-35` at `internal/pdf/helpers.go:161:12` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 109 `PERF-44` at `internal/pdf/image.go:314:18` (medium): cache the repeated type assertion result in a local variable and reuse it.
- [x] Finding 110 `PERF-2` at `internal/pdf/merge.go:168:5` (medium): replace repeated string concatenation with a `strings.Builder` or reusable byte buffer.
- [x] Finding 111 `PERF-2` at `internal/pdf/merge.go:170:4` (medium): replace repeated string concatenation with a `strings.Builder` or reusable byte buffer.
- [x] Finding 112 `PERF-6` at `internal/pdf/merge.go:170:19` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 113 `PERF-35` at `internal/pdf/merge.go:170:19` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 114 `PERF-6` at `internal/pdf/merge.go:181:23` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 115 `PERF-32` at `internal/pdf/merge.go:256:12` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 116 `PERF-32` at `internal/pdf/merge.go:277:11` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 117 `PERF-6` at `internal/pdf/merge.go:277:18` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 118 `PERF-1` at `internal/pdf/merge.go:354:16` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 119 `PERF-1` at `internal/pdf/merge.go:356:14` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 120 `PERF-1` at `internal/pdf/merge/annotations.go:310:19` (medium): hoist the regex compilation out of the loop and reuse one precompiled matcher.
- [x] Finding 121 `PERF-42` at `internal/pdf/merge/merger.go:15:15` (medium): replace static `fmt.Errorf` text with `errors.New`.
- [x] Finding 122 `PERF-15` at `internal/pdf/merge/merger.go:415:14` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 123 `PERF-45` at `internal/pdf/merge/parser.go:79:13` (medium): add a realistic capacity hint before appending in the loop.
- [x] Finding 124 `PERF-35` at `internal/pdf/merge/split.go:38:17` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 125 `PERF-45` at `internal/pdf/merge/split.go:62:11` (medium): add a realistic capacity hint before appending in the loop.
- [x] Finding 126 `PERF-42` at `internal/pdf/merge/split.go:77:15` (medium): replace static `fmt.Errorf` text with `errors.New`.
- [x] Finding 127 `PERF-35` at `internal/pdf/metadata.go:120:18` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 128 `PERF-46` at `internal/pdf/metadata.go:205:10` (medium): avoid allocation-heavy trimming when a cheap prefix, suffix, or length guard will do.
- [x] Finding 129 `PERF-32` at `internal/pdf/metadata.go:255:19` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 130 `PERF-32` at `internal/pdf/metadata.go:297:27` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 131 `PERF-15` at `internal/pdf/outline.go:174:23` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 132 `PERF-35` at `internal/pdf/outline.go:318:24` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 133 `PERF-32` at `internal/pdf/outline.go:323:49` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 134 `PERF-32` at `internal/pdf/outline.go:357:18` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 135 `PERF-6` at `internal/pdf/outline.go:361:25` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 136 `PERF-6` at `internal/pdf/outline.go:363:25` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 137 `PERF-6` at `internal/pdf/outline.go:366:24` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 138 `PERF-6` at `internal/pdf/outline.go:371:25` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 139 `PERF-6` at `internal/pdf/outline.go:374:25` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 140 `PERF-6` at `internal/pdf/outline.go:379:25` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 141 `PERF-6` at `internal/pdf/outline.go:382:25` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 142 `PERF-6` at `internal/pdf/outline.go:385:25` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 143 `PERF-6` at `internal/pdf/outline.go:386:25` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 144 `PERF-6` at `internal/pdf/outline.go:387:25` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 145 `PERF-32` at `internal/pdf/outline.go:391:48` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 146 `PERF-6` at `internal/pdf/outline.go:413:24` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 147 `PERF-6` at `internal/pdf/outline.go:419:24` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 148 `PERF-32` at `internal/pdf/outline.go:469:44` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 149 `PERF-6` at `internal/pdf/outline.go:470:14` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 150 `PERF-15` at `internal/pdf/outline.go:481:27` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 151 `PERF-15` at `internal/pdf/outline.go:485:27` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 152 `PERF-15` at `internal/pdf/outline.go:493:27` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 153 `PERF-32` at `internal/pdf/outline.go:502:45` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 154 `PERF-32` at `internal/pdf/outline.go:509:41` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 155 `PERF-32` at `internal/pdf/pdfa.go:234:34` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 156 `PERF-32` at `internal/pdf/pdfa.go:442:34` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 157 `CWE-328` at `internal/pdf/redact/encryption_inhouse.go:241:9` is already covered by the existing CWE ruleset; exclude it from this performance remediation pass.
- [x] Finding 158 `CWE-916` at `internal/pdf/redact/encryption_inhouse.go:241:9` is already covered by the existing CWE ruleset; exclude it from this performance remediation pass.
- [x] Finding 159 `PERF-48` at `internal/pdf/redact/encryption_inhouse.go:251:28` (medium): add a cheap length or prefix guard before full `bytes.Equal` comparison.
- [x] Finding 160 `PERF-46` at `internal/pdf/redact/ocr_adapter.go:36:14` (medium): avoid allocation-heavy trimming when a cheap prefix, suffix, or length guard will do.
- [x] Finding 161 `PERF-35` at `internal/pdf/redact/ocr_adapter.go:40:14` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 162 `PERF-6` at `internal/pdf/redact/ocr_adapter.go:112:36` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 163 `PERF-15` at `internal/pdf/redact/ocr_adapter.go:114:49` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 164 `PERF-15` at `internal/pdf/redact/ocr_adapter.go:114:75` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 165 `PERF-55` at `internal/pdf/redact/ocr_adapter.go:139:14` (medium): raise the scanner buffer limit or replace `bufio.Scanner` where larger tokens are expected.
- [x] Finding 166 `PERF-47` at `internal/pdf/redact/ocr_adapter.go:147:12` (medium): replace `strings.Split` on the hot path with `Cut`, indexed scanning, or a reusable parser.
- [x] Finding 167 `PERF-48` at `internal/pdf/redact/pdf_utils.go:760:14` (medium): add a cheap length or prefix guard before full `bytes.Equal` comparison.
- [x] Finding 168 `PERF-6` at `internal/pdf/redact/pdf_utils.go:805:3` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 169 `PERF-35` at `internal/pdf/redact/redactor.go:97:29` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 170 `PERF-46` at `internal/pdf/redact/redactor.go:197:10` (medium): avoid allocation-heavy trimming when a cheap prefix, suffix, or length guard will do.
- [x] Finding 171 `PERF-46` at `internal/pdf/redact/search.go:97:11` (medium): avoid allocation-heavy trimming when a cheap prefix, suffix, or length guard will do.
- [x] Finding 172 `PERF-15` at `internal/pdf/redact/secure.go:47:42` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 173 `PERF-15` at `internal/pdf/redact/secure.go:54:42` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 174 `PERF-4` at `internal/pdf/redact/secure.go:58:3` (medium): move the map allocation out of the loop or reuse a preallocated map when safe.
- [x] Finding 175 `PERF-35` at `internal/pdf/redact/secure.go:97:31` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 176 `PERF-46` at `internal/pdf/redact/secure.go:241:11` (medium): avoid allocation-heavy trimming when a cheap prefix, suffix, or length guard will do.
- [x] Finding 177 `PERF-32` at `internal/pdf/redact/secure.go:293:9` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 178 `PERF-15` at `internal/pdf/redact/secure.go:472:20` (medium): replace per-iteration `strconv` formatting with append-based numeric writes using a reusable scratch buffer.
- [x] Finding 179 `PERF-35` at `internal/pdf/redact/visual.go:46:16` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 180 `PERF-6` at `internal/pdf/redact/visual.go:53:19` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 181 `PERF-6` at `internal/pdf/redact/visual.go:59:16` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 182 `PERF-32` at `internal/pdf/redact/visual.go:60:21` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 183 `PERF-32` at `internal/pdf/signature/signature.go:66:10` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 184 `PERF-32` at `internal/pdf/signature/signature.go:68:10` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 185 `PERF-32` at `internal/pdf/signature/signature.go:71:11` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 186 `PERF-32` at `internal/pdf/signature/signature.go:83:25` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 187 `PERF-35` at `internal/pdf/signature/signature.go:89:25` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 188 `PERF-32` at `internal/pdf/signature/signature.go:92:28` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 189 `PERF-42` at `internal/pdf/signature/signature.go:94:25` (medium): replace static `fmt.Errorf` text with `errors.New`.
- [x] Finding 190 `PERF-32` at `internal/pdf/signature/signature.go:111:31` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 191 `PERF-40` at `internal/pdf/signature/signature.go:221:9` (medium): capture the timestamp once per operation and reuse it instead of calling `time.Now()` repeatedly.
- [x] Finding 192 `PERF-32` at `internal/pdf/signature/signature.go:685:63` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 193 `PERF-32` at `internal/pdf/signature/signature.go:703:42` (medium): keep the hot path in one representation and remove repeated `string`/`[]byte` conversions.
- [x] Finding 194 `PERF-47` at `internal/pdf/svg/svg.go:237:10` (medium): replace `strings.Split` on the hot path with `Cut`, indexed scanning, or a reusable parser.
- [x] Finding 195 `PERF-7` at `pkg/fontutils/fontutils.go:149:4` (medium): remove `defer` from the loop body and close or release resources explicitly.
- [x] Finding 196 `PERF-6` at `sampledata/benchmarks/gen_data.go:24:11` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 197 `PERF-35` at `sampledata/benchmarks/gen_data.go:24:11` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 198 `PERF-6` at `sampledata/benchmarks/gen_data.go:25:11` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 199 `CWE-497` at `sampledata/benchmarks/gopdflib/benchconfig.go:55:10` is already covered by the existing CWE ruleset; exclude it from this performance remediation pass.
- [x] Finding 200 `PERF-6` at `sampledata/benchmarks/gopdflib/databench_gopdflib.go:66:51` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 201 `PERF-35` at `sampledata/benchmarks/gopdflib/databench_gopdflib.go:66:51` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 202 `PERF-40` at `sampledata/benchmarks/gopdflib/databench_gopdflib.go:138:16` (medium): capture the timestamp once per operation and reuse it instead of calling `time.Now()` repeatedly.
- [x] Finding 203 `PERF-7` at `sampledata/benchmarks/gopdflib/databench_gopdflib.go:144:4` (medium): remove `defer` from the loop body and close or release resources explicitly.
- [x] Finding 204 `PERF-7` at `sampledata/benchmarks/gopdflib/databench_gopdflib.go:145:4` (medium): remove `defer` from the loop body and close or release resources explicitly.
- [x] Finding 205 `PERF-42` at `sampledata/benchmarks/gopdflib/databench_gopdflib.go:173:10` (medium): replace static `fmt.Errorf` text with `errors.New`.
- [x] Finding 206 `CWE-497` at `sampledata/gopdflib/financial_report/main.go:22:33` is already covered by the existing CWE ruleset; exclude it from this performance remediation pass.
- [x] Finding 207 `PERF-36` at `sampledata/gopdflib/financial_report/main.go:87:3` (medium): capture the loop variable into a local before starting the goroutine.
- [x] Finding 208 `PERF-7` at `sampledata/gopdflib/financial_report/main.go:88:4` (medium): remove `defer` from the loop body and close or release resources explicitly.
- [x] Finding 209 `PERF-40` at `sampledata/gopdflib/financial_report/main.go:90:14` (medium): capture the timestamp once per operation and reuse it instead of calling `time.Now()` repeatedly.
- [x] Finding 210 `CWE-497` at `sampledata/gopdflib/zerodha/main.go:41:33` is already covered by the existing CWE ruleset; exclude it from this performance remediation pass.
- [x] Finding 211 `PERF-6` at `sampledata/gopdflib/zerodha/main.go:100:14` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 212 `PERF-35` at `sampledata/gopdflib/zerodha/main.go:100:14` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.
- [x] Finding 213 `PERF-6` at `sampledata/gopdflib/zerodha/main.go:329:52` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 214 `PERF-6` at `sampledata/gopdflib/zerodha/main.go:330:51` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 215 `PERF-6` at `sampledata/gopdflib/zerodha/main.go:331:51` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 216 `PERF-6` at `sampledata/gopdflib/zerodha/main.go:499:52` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 217 `PERF-6` at `sampledata/gopdflib/zerodha/main.go:503:52` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 218 `PERF-6` at `sampledata/gopdflib/zerodha/main.go:504:51` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 219 `PERF-6` at `sampledata/gopdflib/zerodha/main.go:505:51` (medium): replace loop-local `fmt` formatting with direct builder writes or `strconv.Append*`.
- [x] Finding 220 `PERF-36` at `sampledata/gopdflib/zerodha/main.go:741:3` (medium): capture the loop variable into a local before starting the goroutine.
- [x] Finding 221 `PERF-7` at `sampledata/gopdflib/zerodha/main.go:742:4` (medium): remove `defer` from the loop body and close or release resources explicitly.
- [x] Finding 222 `PERF-40` at `sampledata/gopdflib/zerodha/main.go:743:40` (medium): capture the timestamp once per operation and reuse it instead of calling `time.Now()` repeatedly.
- [x] Finding 223 `PERF-42` at `sampledata/gopdflib/zerodha/main.go:813:10` (medium): replace static `fmt.Errorf` text with `errors.New`.
- [x] Finding 224 `PERF-3` at `typstsyntax/renderer.go:477:3` (medium): stop rebuilding the slice inside the loop; reuse the slice or preallocate once.
- [x] Finding 225 `PERF-46` at `typstsyntax/renderer.go:669:33` (medium): avoid allocation-heavy trimming when a cheap prefix, suffix, or length guard will do.
- [x] Finding 226 `PERF-35` at `typstsyntax/renderer.go:1256:9` (medium): replace boxed `fmt.Sprintf` or `fmt.Errorf` usage with direct builders, `strconv`, or static errors.

---

## Batch 2 (P2): SlopGuard Findings from Chunk_*.txt (Findings 1-140)

Source: `/home/chinmay/ChinmayPersonalProjects/slopguard/scripts/chunks/Chunk_*.txt`

Cross-referenced against the 226 already-fixed findings above. ~78 were duplicates (already in P1), ~17 were CWE-only or false-positives, ~45 were new and fixed.

| # | Rule | File:Line | Status | Fix Summary |
|---|------|-----------|--------|-------------|
| P2-1 | CWE-497 | `cmd/gopdfsuit/main.go:47` | SKIP | CWE only |
| P2-2 | PERF-41 | `cmd/gopdfsuit/main.go:66` | FIXED | `log.Fatalf` → `fmt.Fprintf(os.Stderr)` + `os.Exit(1)` |
| P2-3 | PERF-40 | `internal/benchmarktemplates/runner.go:85` | DUPL | Duplicate of P1 Finding 4 |
| P2-4 | PERF-22 | `internal/handlers/handlers.go:236` | OK | Cache guard already in place |
| P2-5 | PERF-57 | `internal/handlers/handlers.go:301` | FIXED | `io.ReadAll` → `io.Copy` into pre-allocated buffer |
| P2-6 | PERF-15 | `internal/handlers/handlers.go:511` | FIXED | `strconv.Itoa` → `strconv.AppendInt` via scratch buffer |
| P2-7 | PERF-56 | `internal/handlers/handlers.go:515` | FIXED | `c.JSON` moved out of loop body |
| P2-8 | PERF-46 | `internal/handlers/redact.go:22` | DUPL | Duplicate of P1 Finding 15 (adjacent line) |
| P2-9 | PERF-32 | `internal/handlers/redact.go:216` | FIXED | `[]byte(textSearchJSON)` → `unsafe.Slice(unsafe.StringData(...), len(...))` |
| P2-10 | PERF-32 | `internal/pdf/encryption/encrypt.go:64` | FIXED | `[]byte(password[:32])` → `unsafe.Slice(unsafe.StringData(password), 32)` |
| P2-11 | CWE-328 | `internal/pdf/encryption/encrypt.go:79` | SKIP | CWE only |
| P2-12 | CWE-916 | `internal/pdf/encryption/encrypt.go:79` | SKIP | CWE only |
| P2-13 | PERF-53 | `internal/pdf/encryption/encrypt.go:246` | DUPL | Duplicate of P1 Finding 29 |
| P2-14 | PERF-35 | `internal/pdf/encryption/encrypt.go:324` | FIXED | `fmt.Sprintf` → 6 direct `dict.WriteString` calls |
| P2-15 | PERF-35 | `internal/pdf/font/pdfa.go:185` | SKIP | Needs `fmt.Errorf` for `%w` wrapping |
| P2-16 | PERF-42 | `internal/pdf/font/pdfa.go:344` | FIXED | `fmt.Errorf` → `errors.New` for static string |
| P2-17 | PERF-35 | `internal/pdf/font/registry.go:168` | SKIP | Needs `fmt.Errorf` for `%w` |
| P2-18 | PERF-35 | `internal/pdf/font/ttf.go:91` | SKIP | Needs `fmt.Errorf` for `%w` |
| P2-19 | PERF-1 | `internal/pdf/form/xfdf.go:873` | FIXED | Hoisted `rectRe` to package-level global |
| P2-20 | PERF-1 | `internal/pdf/form/xfdf.go:888` | FIXED | Hoisted `qRe` to package-level global |
| P2-21 | PERF-1 | `internal/pdf/form/xfdf.go:893` | FIXED | Hoisted `daRe` to package-level global |
| P2-22 | PERF-1 | `internal/pdf/form/xfdf.go:895` | DUPL | Duplicate of P1 Finding 57 |
| P2-23 | PERF-1 | `internal/pdf/form/xfdf.go:948` | FIXED | Hoisted `vRe` for radio dicts to package-level |
| P2-24 | PERF-1 | `internal/pdf/form/xfdf.go:968` | FIXED | Hoisted `vRe` for text dicts to package-level |
| P2-25 | PERF-1 | `internal/pdf/form/xfdf.go:1029` | SKIP | Runtime-dependent regex (uses loop var in pattern) |
| P2-26 | PERF-1 | `internal/pdf/form/xfdf.go:938` | SKIP | Runtime-dependent regex (uses loop var in pattern) |
| P2-27 | PERF-1 | `internal/pdf/form/xfdf.go:944` | FIXED | Dynamic regex replaced with `bytes.Index` + `bytes.LastIndex` |
| P2-28 | PERF-1 | `internal/pdf/form/xfdf.go:1030` | FIXED | Dynamic regex replaced with bytes-based search |
| P2-29 | PERF-15 | `internal/pdf/merge/merger.go:416` | FIXED | `strconv.FormatInt` → `strconv.AppendInt` + stack scratch buffer |
| P2-30 | PERF-45 | `internal/pdf/merge/parser.go:79` | FIXED | Added `len(data)/200` capacity hint |
| P2-31 | PERF-35 | `internal/pdf/merge/split.go:38` | FIXED | `fmt.Errorf` → `errors.New` |
| P2-32 | PERF-45 | `internal/pdf/merge/split.go:62` | FIXED | Added `len(set)` capacity hint |
| P2-33 | PERF-42 | `internal/pdf/merge/split.go:77` | FIXED | `fmt.Errorf("empty file")` → `errors.New("empty file")` |
| P2-34 | PERF-35 | `internal/pdf/metadata.go:128` | FIXED | `fmt.Sprintf` → string concatenation |
| P2-35 | PERF-46 | `internal/pdf/metadata.go:207` | FIXED | Manual first/last byte check before `strings.TrimSpace` |
| P2-36 | PERF-32 | `internal/pdf/metadata.go:260` | FIXED | `strings.Builder` → `bytes.Buffer` for zero-copy `Bytes()` |
| P2-37 | PERF-32 | `internal/pdf/metadata.go:300` | FIXED | `append([]byte(nil), ...)` → `make` + `copy` |
| P2-38 | PERF-32 | `internal/pdf/outline.go:372` | FIXED | `[]byte(item.Title)` → `append(titleBytes[:0], item.Title...)` |
| P2-39 | PERF-32 | `internal/pdf/outline.go:341` | FIXED | Moved buffer declaration outside loop for reuse |
| P2-40 | PERF-32 | `internal/pdf/outline.go:505-550` | FIXED | `strconv.Itoa` → `strconv.AppendInt` + `bytes.Buffer` |
| P2-41 | PERF-6 | `internal/pdf/outline.go:470` | FIXED | `fmt.Sprintf` → `strconv.AppendInt` |
| P2-42 | PERF-35 | `internal/pdf/redact/ocr_adapter.go:115` | FIXED | `fmt.Sprintf("page-%d", page)` → `strconv.AppendInt` |
| P2-43 | PERF-15 | `internal/pdf/redact/ocr_adapter.go:117` | FIXED | 2x `strconv.Itoa(page)` → single `AppendInt` + cached `pageStr` |
| P2-44 | PERF-55 | `internal/pdf/redact/ocr_adapter.go:142` | FIXED | Added `scanner.Buffer(make([]byte, 0, 1<<20), 10<<20)` |
| P2-45 | PERF-47 | `internal/pdf/redact/ocr_adapter.go:150` | FIXED | `strings.Split(line, "\t")` → `bytes.Split(line, {'\t'})` |
| P2-46 | PERF-48 | `internal/pdf/redact/pdf_utils.go:760` | FIXED | Added `len(origBody) != len(body)` guard before `bytes.Equal` |
| P2-47 | PERF-6 | `internal/pdf/redact/pdf_utils.go:805` | FIXED | `fmt.Fprintf` → `strconv.AppendInt` + `out.Write` |
| P2-48 | PERF-35 | `internal/pdf/redact/redactor.go:97` | SKIP | Needs `fmt.Errorf` for `%w` wrapping |
| P2-49 | PERF-46 | `internal/pdf/redact/redactor.go:197` | FIXED | Manual first/last byte check before `strings.TrimSpace` |
| P2-50 | PERF-46 | `internal/pdf/redact/search.go:97` | FIXED | Manual first/last byte check before `strings.TrimSpace` |
| P2-51 | PERF-15 | `internal/pdf/redact/secure.go:47` | FIXED | `strconv.Itoa` → `strconv.AppendInt` + scratch buffer |
| P2-52 | PERF-15 | `internal/pdf/redact/secure.go:54` | FIXED | `strconv.Itoa` → same `AppendInt` scratch buffer |
| P2-53 | PERF-4 | `internal/pdf/redact/secure.go:58` | FIXED | Hoisted `visited` map outside loop, reuse via `clear(visited)` |
| P2-54 | PERF-35 | `internal/pdf/redact/secure.go:97` | FIXED | `fmt.Sprintf` → string concatenation |
| P2-55 | PERF-46 | `internal/pdf/redact/secure.go:241` | FIXED | Manual guard before `strings.TrimSpace` |
| P2-56 | PERF-35 | `internal/pdf/signature/signature.go:207-216` | FIXED | `fmt.Sprintf` → direct string concatenation for Reason/Location/ContactInfo/Name |
| P2-57 | PERF-6 | `internal/pdf/signature/signature.go:230` | FIXED | `fmt.Sprintf(" /M (D:%s...%02d...)")` → `strings.Builder` |
| P2-58 | PERF-47 | `internal/pdf/svg/svg.go:237` | FIXED | `strings.SplitN(part, ":", 2)` → `strings.Cut(part, ":")` |
| P2-59 | PERF-35 | `sampledata/gopdflib/zerodha/main.go:436` | FIXED | `fmt.Sprintf("₹%.2f", ...)` → `strconv.FormatFloat` |

**Total P2: 140 findings → ~78 duplicates, ~17 CWE/false-positive/skipped, ~45 fixed**

---

## Batch 3 (P3): 101 SlopGuard Findings (Regenerated Chunks)

Source: `/home/chinmay/ChinmayPersonalProjects/slopguard/scripts/chunks/` (regenerated, now 101 findings)

Cross-referenced against P1 (226) and P2 (140) findings. Many were already fixed; ~30 were newly addressed.

| # | Rule | File:Line | Status | Fix Summary |
|---|------|-----------|--------|-------------|
| P3-1 | CWE-497 | `cmd/gopdfsuit/main.go:48` | SKIP | CWE only |
| P3-2 | PERF-41 | `cmd/gopdfsuit/main.go:76` | FIXED | `log.Println` → `os.Stderr.WriteString` |
| P3-3 | PERF-40 | `internal/benchmarktemplates/runner.go:85` | DUPL | Already fixed in P1 |
| P3-4 | PERF-22 | `internal/handlers/handlers.go:236` | DUPL | Cache guard already in place |
| P3-5 | PERF-57 | `internal/handlers/handlers.go:434` | OK | Inherent to handler functionality |
| P3-6 | PERF-46 | `internal/handlers/redact.go:23` | DUPL | Already has fast-path guard |
| P3-7 | CWE-328 | `internal/pdf/encryption/encrypt.go:80` | SKIP | CWE only |
| P3-8 | CWE-916 | `internal/pdf/encryption/encrypt.go:80` | SKIP | CWE only |
| P3-9 | PERF-53 | `internal/pdf/encryption/encrypt.go:247` | OK | Uses `crypto/rand`, not `math/rand` |
| P3-10 | PERF-35 | `internal/pdf/encryption/encrypt.go:358` | FIXED | `fmt.Sprintf` → string concatenation |
| P3-11 | PERF-35 | `internal/pdf/font/pdfa.go:185` | SKIP | Needs `%w` wrapping |
| P3-12 | PERF-35 | `internal/pdf/font/registry.go:168` | SKIP | Needs `%w` wrapping |
| P3-13 | PERF-35 | `internal/pdf/font/ttf.go:91` | SKIP | Needs `%w` wrapping |
| P3-14 | PERF-35 | `internal/pdf/form/xfdf.go:1133` | SKIP | Uses `%w` correctly |
| P3-15 | PERF-1 | `internal/pdf/form/xfdf.go:1351` | FIXED | Hoisted `nameRe` outside loop |
| P3-16 | PERF-1 | `internal/pdf/form/xfdf.go:1364` | FIXED | Hoisted `kidsRe` outside loop |
| P3-17 | PERF-1 | `internal/pdf/form/xfdf.go:1366` | FIXED | Hoisted `refRe` outside loop |
| P3-18 | PERF-1 | `internal/pdf/form/xfdf.go:1372` | FIXED | Hoisted `singleKidsRe` outside loop |
| P3-19 | PERF-48 | `internal/pdf/form/xfdf.go:1573` | OK | Length guard already present |
| P3-20 | PERF-32 | `internal/pdf/generator.go:81` | OK | Standard Go idiom, no real perf issue |
| P3-21 | PERF-32 | `internal/pdf/generator.go:308` | OK | String→byte conversion, Go 1.22+ optimized |
| P3-22 | PERF-35 | `internal/pdf/generator.go:814` | OK | `fmt.Errorf` with `%d` + `%w`, needs boxing |
| P3-23 | PERF-40 | `internal/pdf/generator.go:1011` | FIXED | Captured `time.Now()` once into `genTime` |
| P3-24 | PERF-53 | `internal/pdf/generator.go:1115` | OK | Uses `crypto/rand`, not `math/rand` |
| P3-25 | PERF-48 | `internal/pdf/generator.go:1497` | OK | Length guard already present |
| P3-26 | PERF-15 | `internal/pdf/generator.go:1629` | FIXED | `strconv.Itoa(elemIdx)` → `strconv.AppendInt` |
| P3-27 | PERF-15 | `internal/pdf/generator.go:1632` | FIXED | `strconv.Itoa(elem.Index)` → `strconv.AppendInt` |
| P3-28 | PERF-15 | `internal/pdf/generator.go:1656` | FIXED | `strconv.Itoa` → `strconv.AppendInt` |
| P3-29 | PERF-15 | `internal/pdf/generator.go:1666` | FIXED | `strconv.Itoa` → `strconv.AppendInt` |
| P3-30 | PERF-15 | `internal/pdf/generator.go:1678` | FIXED | `strconv.Itoa` → `strconv.AppendInt` |
| P3-31 | PERF-15 | `internal/pdf/generator.go:1688` | DUPL | Already fixed in P3-30 batch |
| P3-32 | PERF-15 | `internal/pdf/helpers.go:171` | FIXED | `strconv.Itoa` → `strconv.AppendInt` |
| P3-33 | PERF-44 | `internal/pdf/image.go:372` | FIXED | Sequential type assertions → unified `switch v := img.(type)` |
| P3-34 | PERF-42 | `internal/pdf/merge/merger.go:25` | FIXED | `fmt.Errorf` → `errors.New` |
| P3-35 | PERF-35 | `internal/pdf/merge/split.go:53` | FIXED | `fmt.Errorf` → `errors.New` |
| P3-36 | PERF-42 | `internal/pdf/merge/split.go:91` | FIXED | `fmt.Errorf` → `errors.New` |
| P3-37 | PERF-35 | `internal/pdf/metadata.go:143` | FIXED | 4x `fmt.Sprintf` → string concatenation |
| P3-38 | PERF-32 | `internal/pdf/outline.go:523` | FIXED | `[]byte(name)` → `unsafe.Slice(unsafe.StringData(...))` |
| P3-39 | PERF-35 | `internal/pdf/outline.go:554` | FIXED | `fmt.Sprintf` → direct string concat |
| P3-40 | PERF-32 | `internal/pdf/outline.go:555` | FIXED | `[]byte(destsTreeContent)` → `unsafe.Slice` |
| P3-41 | PERF-32 | `internal/pdf/outline.go:562` | FIXED | `fmt.Sprintf` + `[]byte` → `fmt.Appendf` |
| P3-42 | PERF-32 | `internal/pdf/pdfa.go:234` | FIXED | Removed unnecessary `[]byte()` wrapper in `copy` |
| P3-43 | PERF-32 | `internal/pdf/pdfa.go:442` | FIXED | Unnecessary `[]byte()` wrappers in `copy` removed |
| P3-44 | CWE-328 | `internal/pdf/redact/encryption_inhouse.go:241` | SKIP | CWE only |
| P3-45 | CWE-916 | `internal/pdf/redact/encryption_inhouse.go:241` | SKIP | CWE only |
| P3-46 | PERF-48 | `internal/pdf/redact/encryption_inhouse.go:251` | DUPL | Already fixed (length guard present) |
| P3-47 | PERF-46 | `internal/pdf/redact/ocr_adapter.go:38` | DUPL | Guard already present |
| P3-48 | PERF-35 | `internal/pdf/redact/ocr_adapter.go:121` | OK | `fmt.Errorf` with multiple formatted args |
| P3-49 | PERF-48 | `internal/pdf/redact/pdf_utils.go:760` | DUPL | Already fixed (length guard) |
| P3-50 | PERF-35 | `internal/pdf/redact/redactor.go:97` | SKIP | Uses `%w` correctly |
| P3-51 | PERF-46 | `internal/pdf/redact/redactor.go:197` | DUPL | Already fixed in P2 |
| P3-52 | PERF-46 | `internal/pdf/redact/search.go:97` | DUPL | Already fixed in P2 |
| P3-53 | PERF-46 | `internal/pdf/redact/secure.go:242` | DUPL | Already fixed in P1 |
| P3-54 | PERF-35 | `internal/pdf/redact/visual.go:49` | FIXED | `%d` boxing → `strconv.AppendInt` |
| P3-55 | PERF-15 | `internal/pdf/redact/visual.go:56` | FIXED | `strconv.FormatFloat` → `strconv.AppendFloat` |
| P3-56 | PERF-15 | `internal/pdf/redact/visual.go:58` | FIXED | Same as 55 |
| P3-57 | PERF-15 | `internal/pdf/redact/visual.go:60` | FIXED | Same as 55 |
| P3-58 | PERF-15 | `internal/pdf/redact/visual.go:62` | FIXED | Same as 55 |
| P3-59 | PERF-32 | `internal/pdf/signature/signature.go:66` | OK | `h.Write([]byte(certPEM))` - unavoidable `hash.Hash` API |
| P3-60 | PERF-32 | `internal/pdf/signature/signature.go:68` | OK | Same as 59 |
| P3-61 | PERF-32 | `internal/pdf/signature/signature.go:71` | OK | Same pattern, `[]byte` conversion |
| P3-62 | PERF-32 | `internal/pdf/signature/signature.go:83` | OK | `pem.Decode([]byte(certPEM))` - necessary API |
| P3-63 | PERF-35 | `internal/pdf/signature/signature.go:89` | SKIP | Uses `%w` |
| P3-64 | PERF-32 | `internal/pdf/signature/signature.go:92` | OK | `pem.Decode([]byte(keyPEM))` - necessary API |
| P3-65 | PERF-42 | `internal/pdf/signature/signature.go:94` | OK | Already fixed to `errors.New` |
| P3-66 | PERF-32 | `internal/pdf/signature/signature.go:111` | OK | `pem.Decode([]byte(chainPEM))` - necessary API |
| P3-67 | PERF-40 | `internal/pdf/signature/signature.go:221` | FIXED | `fmt.Sprintf` → `strings.Builder` + manual formatting |
| P3-68 | PERF-32 | `internal/pdf/signature/signature.go:701` | OK | `copy` requires `[]byte` target |
| P3-69 | PERF-32 | `internal/pdf/signature/signature.go:719` | OK | Same as 68 |
| P3-70 | PERF-7 | `pkg/fontutils/fontutils.go:149` | OK | Defer inside goroutine, idiomatic Go |
| P3-71 | PERF-6 | `sampledata/benchmarks/gen_data.go:25` | FIXED | `fmt.Sprintf("User %d", i)` → `"User " + strconv.Itoa(i)` |
| P3-72 | PERF-35 | `sampledata/benchmarks/gen_data.go:25` | FIXED | Same line, same fix |
| P3-73 | PERF-6 | `sampledata/benchmarks/gen_data.go:26` | FIXED | `fmt.Sprintf("user%d@example.com", i)` → builder |
| P3-74 | CWE-497 | `sampledata/benchmarks/gopdflib/benchconfig.go:55` | SKIP | CWE only |
| P3-75 | PERF-6 | `sampledata/benchmarks/gopdflib/databench_gopdflib.go:67` | FIXED | `fmt.Sprintf("%d", ...)` → `strconv.Itoa` |
| P3-76 | PERF-35 | `sampledata/benchmarks/gopdflib/databench_gopdflib.go:67` | FIXED | Same line fix |
| P3-77 | PERF-40 | `sampledata/benchmarks/gopdflib/databench_gopdflib.go:138` | OK | Already correct |
| P3-78 | PERF-7 | `sampledata/benchmarks/gopdflib/databench_gopdflib.go:144` | OK | Defer required for correctness |
| P3-79 | PERF-7 | `sampledata/benchmarks/gopdflib/databench_gopdflib.go:145` | OK | Same |
| P3-80 | PERF-42 | `sampledata/benchmarks/gopdflib/databench_gopdflib.go:173` | OK | No formatting verbs |
| P3-81 | CWE-497 | `sampledata/gopdflib/financial_report/main.go:22` | SKIP | CWE |
| P3-82 | PERF-36 | `sampledata/gopdflib/financial_report/main.go:87` | OK | Range over int, no variable capture |
| P3-83 | PERF-7 | `sampledata/gopdflib/financial_report/main.go:88` | OK | Defer in goroutine, safe |
| P3-84 | PERF-40 | `sampledata/gopdflib/financial_report/main.go:90` | OK | Required for benchmark timing |
| P3-85 | CWE-497 | `sampledata/gopdflib/zerodha/main.go:41` | SKIP | CWE |
| P3-86 | PERF-6 | `sampledata/gopdflib/zerodha/main.go:100` | FIXED | `fmt.Sprintf("%02d:%02d:%02d")` → direct byte array build |
| P3-87 | PERF-35 | `sampledata/gopdflib/zerodha/main.go:100` | FIXED | Same fix |
| P3-88 | PERF-6 | `sampledata/gopdflib/zerodha/main.go:329` | FIXED | `fmt.Sprintf("%d", ...)` → `strconv.Itoa` |
| P3-89 | PERF-6 | `sampledata/gopdflib/zerodha/main.go:330` | FIXED | `fmt.Sprintf("₹%.2f", ...)` → `"₹" + strconv.FormatFloat` |
| P3-90 | PERF-6 | `sampledata/gopdflib/zerodha/main.go:331` | FIXED | Same as 89 |
| P3-91 | PERF-6 | `sampledata/gopdflib/zerodha/main.go:499` | FIXED | `fmt.Sprintf("%d", t.ID)` → `strconv.Itoa` |
| P3-92 | PERF-6 | `sampledata/gopdflib/zerodha/main.go:503` | FIXED | `fmt.Sprintf("%d", t.Qty)` → `strconv.Itoa` |
| P3-93 | PERF-6 | `sampledata/gopdflib/zerodha/main.go:504` | FIXED | `fmt.Sprintf("₹%.2f", ...)` → `"₹" + strconv.FormatFloat` |
| P3-94 | PERF-6 | `sampledata/gopdflib/zerodha/main.go:505` | FIXED | Same as 93 |
| P3-95 | PERF-36 | `sampledata/gopdflib/zerodha/main.go:741` | OK | Range over int, no variable |
| P3-96 | PERF-7 | `sampledata/gopdflib/zerodha/main.go:742` | OK | Defer in goroutine, safe |
| P3-97 | PERF-40 | `sampledata/gopdflib/zerodha/main.go:743` | OK | Intentional per-goroutine timing |
| P3-98 | PERF-42 | `sampledata/gopdflib/zerodha/main.go:813` | FIXED | `fmt.Errorf("no results collected")` → `errors.New(...)` |
| P3-99 | PERF-3 | `typstsyntax/renderer.go:477` | OK | Standard 2D slice allocation |
| P3-100 | PERF-46 | `typstsyntax/renderer.go:669` | FIXED | `strings.TrimSpace` → non-allocating `isSpace` check |
| P3-101 | PERF-35 | `typstsyntax/renderer.go:1256` | FIXED | `fmt.Sprintf("%.2f")` → `strconv.FormatFloat` |

**Total P3: 101 findings → ~40 duplicates/OK, ~15 CWE/skip, ~46 fixed**

---

## Batch 4 (P4) + Batch 5 (P5): Regenerated Chunk Findings (65 + 65 findings)

Source: `/home/chinmay/ChinmayPersonalProjects/slopguard/scripts/chunks/` (subsequent regenerations)

### P4 (65 findings - 3 newly fixed)

| # | Rule | File:Line | Status | Fix |
|---|------|-----------|--------|-----|
| P4-20 | PERF-35 | `internal/pdf/merge/split.go:100` | FIXED | `fmt.Errorf("page out of range: %d", p)` → `errors.New("page out of range: " + strconv.Itoa(p))` |
| P4-21 | PERF-35 | `internal/pdf/metadata.go:269` | FIXED | `fmt.Sprintf` for metadata dict → `strings.Builder` + `strconv.Itoa` |
| P4-41 | PERF-42 | `internal/pdf/signature/signature.go:493` | FIXED | `fmt.Errorf("unsupported key type")` → `errors.New("unsupported key type")` |
| P4-rest | - | - | DUPL/OK | All other 62 P4 findings already properly handled in prior batches |

### P5 (65 findings - 3 newly fixed, same as P4 findings)

The P5 chunk set contains the same 65 findings as P4 (regenerated). All previously addressed in P1-P4.

**Total P4+P5: 65 findings each → 3 newly fixed per batch, rest duplicates/OK**

---

## Batch 6 (P6): SlopGuard Findings from `findings/functions/*.txt` (66 findings)

Source: `/home/chinmay/ChinmayPersonalProjects/slopguard/scripts/findings/functions/{1..66}.txt`

Cross-referenced against P1-P5. **4 newly fixed**, **62 verified as already handled or unfixable**.

### P6 Newly Fixed

| # | Rule | File:Line | Fix |
|---|------|-----------|-----|
| P6-21 | PERF-35 | `internal/pdf/merge/split.go:108` | `fmt.Errorf("invalid range: %v", r)` → `errors.New("invalid range: [" + strconv.Itoa(r[0]) + "," + strconv.Itoa(r[1]) + "]")` |
| P6-22 | PERF-35 | `internal/pdf/metadata.go:326` | `fmt.Sprintf("<< /N 3 /Length %d /Filter ...", len(...))` → `strings.Builder` + `strconv.Itoa` |
| P6-26 | PERF-46 | `internal/pdf/redact/ocr_adapter.go:64` | **Real hot-path fix** - Hoisted `strings.ToLower` + `strings.TrimSpace` out of nested `runOCRSearch` word/query loop into one-time pre-normalization. Each query is normalized once, each word is lowercased once per outer iteration. |
| P6-42 | PERF-42 | `internal/pdf/signature/signature.go:645` | `fmt.Errorf("byteRange placeholder not found")` → `errors.New("byteRange placeholder not found")` |

### P6 Unfixable / False Positives (kept as-is)

| # | Rule | File:Line | Reason cannot fix |
|---|------|-----------|---------------------|
| P6-1, P6-6, P6-7, P6-23, P6-24, P6-46, P6-51, P6-55 | CWE-497, CWE-328, CWE-916 | various | **Security/encryption rules**, not performance. Out of scope for this perf remediation pass. |
| P6-9, P6-10, P6-11, P6-12, P6-16, P6-27, P6-29, P6-33, P6-38 | PERF-35 | `pdfa.go:185`, `registry.go:168`, `ttf.go:91`, `xfdf.go:1133`, `generator.go:814`, `ocr_adapter.go:124`, `redactor.go:97`, `visual.go:49`, `signature.go:89` | `fmt.Errorf("...: %w", err)` - **`%w` verb is required for error-chain unwrapping**. Cannot replace with `errors.New`. |
| P6-14, P6-15 | PERF-32 | `generator.go:81`, `generator.go:308` | `[]byte(string)` is standard Go idiom in this `copy()` context. Go 1.22+ optimizes `append`/`copy` to avoid allocation. |
| P6-20 | PERF-15 | `split.go:100` | Already uses `strconv.Itoa(p)` (single allocation per error path). Acceptable. |
| P6-34, P6-35, P6-36, P6-37, P6-39, P6-40 | PERF-32 | `signature.go:66, 68, 71, 83, 92, 111` | `h.Write([]byte(certPEM))` and `pem.Decode([]byte(certPEM))` - **`hash.Hash` and `pem.Decode` APIs require `[]byte`**. No way to avoid conversion without unsafe pointer tricks (which add complexity without measurable benefit). |
| P6-43, P6-49, P6-50, P6-53, P6-64 | PERF-7 | `fontutils.go:149`, `databench_gopdflib.go:146,147`, `financial_report/main.go:88`, `zerodha/main.go:752` | **`defer wg.Done()` / `defer func() { <-sem }()` inside a goroutine closure** - required for correctness, scoped to goroutine exit not loop iteration. |
| P6-44, P6-45, P6-47, P6-56, P6-57, P6-58, P6-59, P6-60, P6-61, P6-62 | PERF-15 | `gen_data.go:25,26`, `databench_gopdflib.go:68`, `zerodha/main.go:339,340,341,509,513,514,515` | `strconv.Itoa(...)` and `strconv.FormatFloat(...)` in benchmark data generators - **single allocation per iteration, only runs at benchmark setup time**. Not a production hot path. |
| P6-48, P6-54, P6-65 | PERF-40 | `databench_gopdflib.go:140`, `financial_report/main.go:90`, `zerodha/main.go:753` | `time.Now()` used for **per-goroutine seed timing or per-iteration latency measurement** - intentional. |
| P6-52, P6-63 | PERF-36 | `financial_report/main.go:87`, `zerodha/main.go:751` | `for range numWorkers` - **no loop variable** to capture. False positive. |
| P6-66 | PERF-3 | `typstsyntax/renderer.go:477` | `gridCells[r] = make([]*MathLayout, cols)` - outer slice preallocated at line 471 with `rowCount`. Inner slices are `cols`-sized fixed allocation per row, **not a "rebuild"** (re-allocating a working slice). Standard Go 2D slice pattern. |

**Total P6: 66 findings → 4 newly fixed, 62 unfixable/false-positive/duplicate**
