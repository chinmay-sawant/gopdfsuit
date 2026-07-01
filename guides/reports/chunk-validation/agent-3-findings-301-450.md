# Agent 3 — Findings 301–450 Validation (Manual Review)

> Exported from prior subagent session. Domain-agnostic evaluation per `CHUNK_VALIDATOR.md`.

## Summary

- Total findings analyzed: 150
- Correctly Fired (True Positives): 132
- Incorrectly Fired (False Positives): 18
- FP rate: 12.0%

## Per-Finding Table

| Finding | Rule | Correctly Fired? | Reason |
|---------|------|-------------------|--------|
| 301 | PERF-192 | Yes | `make(map[string]bool)` has no size hint |
| 302 | PERF-109 | No | Iterates slice IDs, not recomputing map keys |
| 303 | PERF-119 | Yes | Three consecutive `append` calls per loop iteration |
| 304 | BP-1 | Yes | `_ = zlibWriter.Close()` discards error return |
| 305 | BP-1 | Yes | `_ = zlibWriter.Close()` discards error return |
| 306 | PERF-192 | Yes | `make(map[string]bool)` has no size hint |
| 307 | PERF-6 | Yes | `fmt.Sprintf` inside `for _, imgObj := range` loop |
| 308 | PERF-6 | Yes | `fmt.Sprintf` inside `for _, imgObj := range` loop |
| 309 | PERF-6 | Yes | `fmt.Sprintf` inside `for _, imgObj := range` loop |
| 310 | PERF-40 | Yes | Multiple `time.Now()` calls in same function |
| 311 | BP-1 | No | Discards `int` ID, not an error return |
| 312 | BP-1 | No | Discards `int` ID, not an error return |
| 313 | BP-1 | No | Discards `string` from `Zone()`, not error |
| 314 | PERF-53 | No | Uses `crypto/rand.Read`, not `math/rand` |
| 315 | CWE-328 | Yes | `md5.Sum(randomBytes)` uses weak MD5 hash |
| 316 | PERF-44 | No | One `*StructElem` assertion per loop item |
| 317 | PERF-6 | Yes | `fmt.Sprintf` inside loop building parent tree |
| 318 | PERF-6 | Yes | `fmt.Sprintf` inside loop building parent tree |
| 319 | PERF-6 | Yes | `fmt.Sprintf` inside loop building parent tree |
| 320 | PERF-6 | Yes | `fmt.Sprintf` inside nested loops |
| 321 | PERF-6 | Yes | `fmt.Sprintf` inside nested loops |
| 322 | PERF-6 | Yes | `fmt.Sprintf` inside nested loops |
| 323 | PERF-6 | Yes | `fmt.Sprintf` inside `for elemIdx, elem` loop |
| 324 | PERF-6 | Yes | `fmt.Sprintf` inside `for elemIdx, elem` loop |
| 325 | PERF-6 | Yes | `fmt.Sprintf` inside `for elemIdx, elem` loop |
| 326 | PERF-6 | Yes | `fmt.Sprintf` inside elements loop |
| 327 | PERF-6 | Yes | `fmt.Sprintf` inside `for tableIdx, table` loop |
| 328 | PERF-6 | Yes | `fmt.Sprintf` inside `for i, image` loop |
| 329 | PERF-192 | Yes | `make(map[string]bool)` has no size hint |
| 330 | PERF-192 | Yes | `make(map[string]bool)` has no size hint |
| 331 | PERF-192 | Yes | `make(map[string]struct{})` has no size hint |
| 332 | PERF-32 | Yes | `[]byte(`/Encrypt`)` is string-to-bytes conversion |
| 333 | BP-1 | Yes | `_ = r.Close()` discards error return |
| 334 | BP-1 | Yes | `_ = r.Close()` discards error return |
| 335 | PERF-188 | Yes | `fmt.Sscanf` inside `for _, p := range parts` loop |
| 336 | PERF-32 | Yes | `[]byte(...)` literals passed to `bytesIndex` |
| 337 | PERF-32 | Yes | `[]byte(...)` literals passed to `bytesIndex` |
| 338 | PERF-1 | Yes | `regexp.MustCompile` inside outer `for` loop |
| 339 | PERF-109 | No | Numeric `for pos` loop, not map key computation |
| 340 | PERF-1 | Yes | `regexp.MustCompile` inside nested loop |
| 341 | PERF-35 | Yes | `fmt.Sprintf` boxes arguments via `interface{}` |
| 342 | PERF-6 | Yes | `fmt.Sprintf` inside nested `for pos` loop |
| 343 | PERF-35 | Yes | `fmt.Sprintf` boxes `float64` through `interface{}` |
| 344 | PERF-192 | Yes | `make(map[uint64]*ImageObject)` has no size hint |
| 345 | PERF-192 | Yes | `make(map[uint64]*ImageObject)` has no size hint |
| 346 | BP-1 | Yes | `_ = zlibWriter.Close()` discards error return |
| 347 | BP-1 | Yes | `_ = zlibWriter.Close()` discards error return |
| 348 | BP-1 | Yes | `_ = zlibWriter.Close()` discards error return |
| 349 | BP-1 | Yes | `_ = zlibWriter.Close()` discards error return |
| 350 | PERF-44 | No | Distinct `*NRGBA`/`*RGBA` assertions, not repeated |
| 351 | BP-1 | No | Discards alpha `uint32`, not an error return |
| 352 | PERF-119 | Yes | Consecutive `append` calls to same `b` slice |
| 353 | PERF-128 | Yes | Three-plus independent `append` calls to `b` |
| 354 | PERF-192 | Yes | `make(map[int]bool)` has no size hint |
| 355 | PERF-42 | Yes | `fmt.Errorf("cannot merge encrypted PDF")` has no verbs |
| 356 | PERF-4 | Yes | `make(map[int][]byte)` allocated inside file loop |
| 357 | PERF-192 | Yes | `make(map[int][]byte)` has no size hint |
| 358 | PERF-4 | Yes | `make(map[string][]byte)` allocated inside file loop |
| 359 | PERF-192 | Yes | `make(map[string][]byte)` has no size hint |
| 360 | PERF-188 | Yes | `fmt.Sscanf` inside `for k, v := range` loop |
| 361 | PERF-188 | Yes | `fmt.Sscanf` inside `for _, f := range files` loop |
| 362 | PERF-1 | Yes | `regexp.MustCompile` inside file-processing loop |
| 363 | PERF-1 | Yes | `regexp.MustCompile` inside file-processing loop |
| 364 | PERF-1 | Yes | `regexp.MustCompile` inside nested loop |
| 365 | PERF-109 | No | Iterates `pagesFromTree` slice, not map keys |
| 366 | PERF-2 | Yes | `catalogDict += " "` inside `for i, fieldNum` loop |
| 367 | PERF-2 | Yes | `catalogDict += fmt.Sprintf(...)` inside loop |
| 368 | PERF-35 | Yes | `fmt.Sprintf` boxes `fieldNum` via `interface{}` |
| 369 | PERF-6 | Yes | `fmt.Sprintf` inside `for i, fieldNum` loop |
| 370 | PERF-6 | Yes | `fmt.Sprintf` inside `for _, p := range` loop |
| 371 | PERF-1 | Yes | `regexp.MustCompile` inside `for _, a := range` loop |
| 372 | PERF-6 | Yes | `fmt.Sprintf` inside `for i := 1; i <= maxObj` loop |
| 373 | PERF-32 | Yes | `out[i] = []byte(s)` copies string to bytes |
| 374 | BP-1 | Yes | `on, _ := strconv.Atoi(...)` discards error |
| 375 | PERF-32 | Yes | `[]byte(fmt.Sprintf(...))` copies formatted string |
| 376 | PERF-6 | Yes | `fmt.Sprintf` inside `ReplaceAllFunc` callback in loop |
| 377 | BP-1 | Yes | `on, _ := strconv.Atoi(...)` discards error |
| 378 | PERF-192 | Yes | `make(map[int]bool)` has no size hint |
| 379 | PERF-1 | Yes | `regexp.MustCompile` inside `for _, body := range` loop |
| 380 | PERF-1 | Yes | `regexp.MustCompile` inside `for _, body := range` loop |
| 381 | PERF-192 | Yes | `make(map[int]bool)` has no size hint |
| 382 | PERF-192 | Yes | `make(map[int]bool)` has no size hint |
| 383 | BP-2 | Yes | Bare `return err` without wrapping context |
| 384 | PERF-35 | Yes | `fmt.Errorf` boxes `ref` through `interface{}` |
| 385 | BP-1 | Yes | `on, _ := strconv.Atoi(...)` discards error |
| 386 | PERF-119 | Yes | Three consecutive `append` calls to `result` |
| 387 | PERF-128 | Yes | Three independent `append` calls to `result` |
| 388 | PERF-192 | Yes | `make(map[int]bool)` has no size hint |
| 389 | PERF-1 | Yes | `regexp.MustCompile` inside `for _, ref := range` loop |
| 390 | PERF-42 | Yes | `fmt.Errorf("no PDF files provided")` has no verbs |
| 391 | PERF-109 | No | Iterates `files` slice, not recomputing map keys |
| 392 | PERF-35 | Yes | `fmt.Sprintf` boxes version string via `interface{}` |
| 393 | BP-1 | Yes | `pagesNum, _ = strconv.Atoi(...)` discards error |
| 394 | BP-1 | Yes | `pagesNum, _ := strconv.Atoi(...)` discards error |
| 395 | BP-1 | Yes | `kidNum, _ := strconv.Atoi(...)` discards error |
| 396 | PERF-192 | Yes | `make(map[int]bool)` has no size hint |
| 397 | PERF-192 | Yes | `make(map[int]bool)` has no size hint |
| 398 | PERF-15 | Yes | `strconv.FormatInt` inside `for i := 1` loop |
| 399 | PERF-119 | No | Single-char `append` per inner-loop iteration only |
| 400 | PERF-128 | No | Not three-plus independent appends at flagged line |
| 401 | BP-1 | Yes | `major, _ := strconv.Atoi(...)` discards error |
| 402 | BP-1 | Yes | `minor, _ = strconv.Atoi(...)` discards error |
| 403 | BP-1 | Yes | `objNum, _ := strconv.Atoi(...)` discards error |
| 404 | BP-1 | Yes | `genNum, _ := strconv.Atoi(...)` discards error |
| 405 | PERF-45 | Yes | `append(results, ...)` in loop without capacity hint |
| 406 | PERF-192 | Yes | `make(map[int][]byte)` has no size hint |
| 407 | BP-1 | Yes | `numObjects, _ := strconv.Atoi(...)` discards error |
| 408 | BP-1 | Yes | `firstOffset, _ := strconv.Atoi(...)` discards error |
| 409 | BP-1 | Yes | `_ = reader.Close()` discards error return |
| 410 | PERF-192 | Yes | `make(map[int]bool)` has no size hint |
| 411 | BP-1 | Yes | `a, _ := strconv.Atoi(m[1])` discards error |
| 412 | BP-1 | Yes | `b, _ := strconv.Atoi(m[2])` discards error |
| 413 | PERF-121 | No | Integer comparison, not struct-literal conversion |
| 414 | PERF-35 | Yes | `fmt.Errorf` boxes `p` through `interface{}` |
| 415 | BP-1 | Yes | `n, _ := strconv.Atoi(p)` discards error |
| 416 | PERF-45 | Yes | `append(pages, k)` in loop without capacity hint |
| 417 | PERF-42 | Yes | `fmt.Errorf("empty file")` has no format verbs |
| 418 | PERF-192 | Yes | `make(map[int]bool)` has no size hint |
| 419 | PERF-119 | Yes | `append(groups, slice)` inside chunking loop |
| 420 | PERF-192 | Yes | `make(map[int]bool)` has no size hint |
| 421 | PERF-109 | No | Iterates `deps` slice, not recomputing map keys |
| 422 | BP-1 | Yes | `refNum, _ := strconv.Atoi(...)` discards error |
| 423 | PERF-192 | Yes | `make(map[int]bool)` has no size hint |
| 424 | PERF-192 | Yes | `make(map[int]int)` has no size hint |
| 425 | PERF-192 | Yes | `make(map[int]bool)` has no size hint |
| 426 | PERF-192 | Yes | `make(map[int]int)` has no size hint |
| 427 | PERF-192 | Yes | `make(map[int][]int)` has no size hint |
| 428 | PERF-192 | Yes | `make(map[int][]byte)` has no size hint |
| 429 | PERF-192 | Yes | `make(map[int][]int)` has no size hint |
| 430 | PERF-192 | Yes | `make(map[int][]int)` has no size hint |
| 431 | PERF-151 | No | Simple 6-line constructor, not complex/non-inlinable |
| 432 | PERF-35 | Yes | `fmt.Sprintf` boxes `part` through `interface{}` |
| 433 | PERF-46 | Yes | `strings.TrimSpace(kw)` inside keyword loop |
| 434 | PERF-6 | Yes | `fmt.Sprintf` inside `for _, kw := range` loop |
| 435 | PERF-32 | Yes | `[]byte(xmpContent)` copies string to bytes |
| 436 | BP-1 | Yes | `_ = zlibWriter.Close()` discards error return |
| 437 | BP-1 | Yes | `_ = zlibWriter.Close()` discards error return |
| 438 | PERF-35 | Yes | `fmt.Sprintf` boxes int via `interface{}` |
| 439 | PERF-6 | Yes | `fmt.Sprintf` inside `for _, bm := range` loop |
| 440 | PERF-109 | No | Iterates `outlineItems` slice, not map keys |
| 441 | PERF-32 | Yes | `[]byte(item.Title)` copies string to bytes |
| 442 | PERF-6 | Yes | `fmt.Sprintf` inside outline-item loop |
| 443 | PERF-6 | Yes | `fmt.Sprintf` inside outline-item loop |
| 444 | PERF-6 | Yes | `fmt.Sprintf` inside outline-item loop |
| 445 | PERF-6 | Yes | `fmt.Sprintf` inside outline-item loop |
| 446 | PERF-6 | Yes | `fmt.Sprintf` inside outline-item loop |
| 447 | PERF-6 | Yes | `fmt.Sprintf` inside outline-item loop |
| 448 | PERF-6 | Yes | `fmt.Sprintf` inside outline-item loop |
| 449 | PERF-6 | Yes | `fmt.Sprintf` inside outline-item loop |
| 450 | PERF-6 | Yes | `fmt.Sprintf` inside outline-item loop |

## Notable FP Patterns

- **PERF-109** (6) — slice iteration flagged as map key recomputation
- **BP-1** (4) — non-error discards (`int` ID, `Zone()` string, alpha `uint32`)
- **PERF-44** (2), **PERF-119/128** (3) — append/type-assertion edge cases
- Single misfires: PERF-53, PERF-121, PERF-151