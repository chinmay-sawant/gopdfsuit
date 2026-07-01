# Agent 4 — Findings 451–600 Validation (Manual Review)

> Exported from prior subagent session. Domain-agnostic evaluation per `CHUNK_VALIDATOR.md`.

## Summary

- Total findings analyzed: 150
- Correctly Fired (True Positives): 134
- Incorrectly Fired (False Positives): 16
- FP rate: 10.7%

## Per-Finding Table

| Finding | Rule | Correctly Fired? | Reason |
|---------|------|------------------|--------|
| 451 | PERF-6 | Yes | `fmt.Sprintf` inside `for _, item := range` loop |
| 452 | PERF-6 | Yes | `fmt.Sprintf` inside `for _, r := range s` loop |
| 453 | PERF-6 | Yes | `fmt.Sprintf` inside same `for _, r := range` loop |
| 454 | PERF-32 | Yes | `[]byte(name)` string-to-byte conversion in loop |
| 455 | PERF-6 | Yes | `fmt.Sprintf` inside `for i, name := range names` loop |
| 456 | PERF-6 | Yes | `fmt.Sprintf` inside same names loop |
| 457 | PERF-6 | Yes | `fmt.Sprintf` inside same names loop |
| 458 | PERF-6 | Yes | `fmt.Sprintf` inside same names loop |
| 459 | PERF-192 | Yes | `make(map[int]string)` without size hint |
| 460 | PERF-123 | Yes | `make([]AnnotStructElem, 0)` uses redundant zero |
| 461 | PERF-192 | Yes | `make(map[string]NamedDest)` without size hint |
| 462 | PERF-27 | Yes | New `bytes.Buffer{}` allocated on each page add |
| 463 | PERF-35 | Yes | `fmt.Sprintf` boxes args through `interface{}` |
| 464 | PERF-35 | Yes | `fmt.Sprintf` boxes args through `interface{}` |
| 465 | PERF-32 | Yes | `[]byte(tag.sig)` conversion inside loop |
| 466 | BP-1 | Yes | `_ = zlibWriter.Close()` discards error return |
| 467 | BP-1 | Yes | `_ = zlibWriter.Close()` discards error return |
| 468 | BP-1 | Yes | `_ = zlibWriter.Close()` discards error return |
| 469 | BP-1 | Yes | `_ = zlibWriter.Close()` discards error return |
| 470 | PERF-32 | Yes | `[]byte(tag.sig)` conversion inside loop |
| 471 | PERF-109 | No | Loop writes slice data; no map key computation |
| 472 | PERF-35 | Yes | `fmt.Sprintf` boxes args through `interface{}` |
| 473 | CWE-916 | Yes | MD5 used in password/key derivation loop |
| 474 | CWE-328 | Yes | MD5 used for security-sensitive password digest |
| 475 | PERF-48 | Yes | `bytes.Equal` on 32-byte values in hot path |
| 476 | PERF-109 | No | MD5 iteration loop; no map key recomputation |
| 477 | PERF-119 | Yes | Three consecutive `append` calls to same slice |
| 478 | PERF-128 | Yes | Three independent `append` calls can be combined |
| 479 | PERF-32 | Yes | `[]byte('/Encrypt')` conversion inside loop |
| 480 | BP-1 | Yes | `_ = r.Close()` discards error return |
| 481 | BP-1 | Yes | `_ = r.Close()` discards error return |
| 482 | PERF-188 | Yes | `fmt.Sscanf` inside `for _, p := range parts` loop |
| 483 | PERF-32 | Yes | `[]byte(...)` conversion inside loop |
| 484 | PERF-32 | Yes | Second `[]byte(...)` conversion on same line |
| 485 | PERF-1 | Yes | `regexp.MustCompile` inside loop body |
| 486 | PERF-109 | Yes | Loop contains `fmt.Sprintf` map key construction |
| 487 | PERF-1 | Yes | `regexp.MustCompile` inside nested loop |
| 488 | PERF-35 | Yes | `fmt.Sprintf` boxes args in loop body |
| 489 | PERF-6 | Yes | `fmt.Sprintf` formatting inside loop body |
| 490 | PERF-46 | Yes | `strings.TrimSpace` allocates on request path |
| 491 | PERF-112 | Yes | `strings.ToLower` before string comparison |
| 492 | PERF-35 | Yes | `fmt.Errorf` boxes args through `interface{}` |
| 493 | PERF-109 | No | Outer words loop; no expensive map key built |
| 494 | PERF-112 | Yes | `strings.ToLower` before comparison in loop |
| 495 | BP-1 | Yes | `_ = os.RemoveAll(tmpDir)` discards error |
| 496 | PERF-123 | Yes | `make([]ocrWord, 0)` has redundant zero arg |
| 497 | PERF-6 | Yes | `fmt.Sprintf` inside page loop |
| 498 | PERF-15 | Yes | `strconv.Itoa` inside page loop |
| 499 | PERF-15 | Yes | Second `strconv.Itoa` inside same loop |
| 500 | BP-1 | Yes | `_ = imgFile.Close()` discards error return |
| 501 | PERF-55 | Yes | `bufio.NewScanner` without explicit buffer sizing |
| 502 | PERF-47 | Yes | `strings.Split` allocates inside scanner loop |
| 503 | PERF-192 | Yes | `make(map[string][]byte)` without size hint |
| 504 | PERF-1 | Yes | `regexp.MustCompile` inside matches loop |
| 505 | PERF-1 | Yes | `regexp.MustCompile` inside nested loop body |
| 506 | PERF-188 | Yes | `fmt.Sscanf` inside parsing loop |
| 507 | PERF-186 | Yes | `strings.Fields` used in hot parsing path |
| 508 | PERF-109 | Yes | Inner loop builds `fmt.Sprintf` map keys |
| 509 | PERF-188 | Yes | `fmt.Sscanf` inside hot parsing loop |
| 510 | PERF-188 | Yes | `fmt.Sscanf` inside hot parsing loop |
| 511 | PERF-188 | Yes | `fmt.Sscanf` inside nested parsing loop |
| 512 | PERF-35 | Yes | `fmt.Sprintf` boxes args for map key |
| 513 | PERF-6 | Yes | `fmt.Sprintf` inside loop body |
| 514 | BP-2 | Yes | Bare `return err` without wrapping context |
| 515 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 516 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 517 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 518 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 519 | BP-2 | Yes | Bare `return err` without wrapping context |
| 520 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 521 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 522 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 523 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 524 | BP-1 | No | `_` discards offset int, not an error return |
| 525 | BP-1 | No | `scanErr` is checked; `_` is scan count |
| 526 | PERF-188 | Yes | `fmt.Sscanf` inside `for key, body := range` loop |
| 527 | PERF-48 | Yes | `bytes.Equal` on object bodies in loop |
| 528 | PERF-6 | Yes | `fmt.Fprintf` inside `for _, obj := range` loop |
| 529 | PERF-6 | Yes | `fmt.Sprintf` inside nested xref loop |
| 530 | PERF-35 | Yes | `fmt.Errorf` boxes `err` through `interface{}` |
| 531 | BP-1 | No | `_` discards raw stream bytes, not error |
| 532 | PERF-119 | Yes | Two consecutive `append` calls to `combined` |
| 533 | BP-1 | Yes | `_` discards `extractPageContent` error return |
| 534 | PERF-46 | Yes | `strings.TrimSpace(strings.ToLower(...))` allocates |
| 535 | PERF-112 | Yes | `strings.ToLower` before mode comparison |
| 536 | BP-1 | Yes | `_` discards `NewRedactor` error return |
| 537 | PERF-109 | Yes | `strings.ToLower(term)` used as map key in loop |
| 538 | PERF-46 | Yes | `strings.TrimSpace` inside search loop |
| 539 | PERF-112 | Yes | `strings.ToLower` before substring comparison |
| 540 | PERF-186 | Yes | `strings.Fields` in normalization hot path |
| 541 | PERF-2 | Yes | `joined += " " + part` concatenation in loop |
| 542 | PERF-119 | No | Single `append`; not consecutive appends |
| 543 | PERF-192 | Yes | `make(map[int][]...)` without size hint |
| 544 | PERF-109 | No | Uses `r.PageNum` directly; not expensive key |
| 545 | PERF-35 | Yes | `fmt.Sprintf` boxes args in page loop |
| 546 | PERF-6 | Yes | `fmt.Sprintf` inside page loop |
| 547 | PERF-6 | Yes | `fmt.Sprintf` inside page loop |
| 548 | PERF-4 | Yes | `make(map[string]bool)` allocated inside loop |
| 549 | PERF-192 | Yes | `make(map[string]bool)` without size hint |
| 550 | BP-1 | No | `_` discards raw stream, not error return |
| 551 | PERF-1 | Yes | `regexp.MustCompile` inside child-refs loop |
| 552 | PERF-119 | Yes | Three consecutive `append` calls to `newObj` |
| 553 | PERF-128 | Yes | Three independent `append` calls combinable |
| 554 | PERF-32 | Yes | `[]byte(fmt.Sprintf(...))` copies string data |
| 555 | PERF-46 | Yes | `strings.TrimSpace` inside match loop |
| 556 | PERF-32 | Yes | `[]byte(out.String())` copies builder output |
| 557 | BP-1 | Yes | `_, _ = fmt.Fprintf` discards write error |
| 558 | PERF-6 | Yes | `fmt.Fprintf` inside `for _, r := range` loop |
| 559 | BP-1 | Yes | `_, _ = fmt.Fprintf` discards write error |
| 560 | PERF-6 | Yes | `fmt.Fprintf` inside second char loop |
| 561 | BP-1 | Yes | `_, _ = fmt.Fprintf` discards write error |
| 562 | PERF-6 | Yes | `fmt.Fprintf` inside segments loop |
| 563 | BP-1 | Yes | `_, _ = fmt.Fprintf` discards write error |
| 564 | PERF-6 | Yes | `fmt.Fprintf` inside nested hex loop |
| 565 | PERF-192 | Yes | `make(map[int][]...)` without size hint |
| 566 | PERF-109 | No | Uses `rect.PageNum` directly as map key |
| 567 | BP-1 | Yes | `_, _ = fmt.Sscanf` discards parse error |
| 568 | PERF-188 | Yes | `fmt.Sscanf` inside `for k := range objMap` |
| 569 | PERF-35 | Yes | `fmt.Errorf` boxes args through `interface{}` |
| 570 | PERF-6 | Yes | `fmt.Sprintf` inside rects loop |
| 571 | PERF-6 | Yes | `fmt.Sprintf` inside page loop |
| 572 | PERF-6 | Yes | `fmt.Sprintf` inside page loop |
| 573 | PERF-32 | Yes | `[]byte(streamObj)` string-to-byte conversion |
| 574 | BP-1 | No | `pem.Decode` second value is remainder, not error |
| 575 | PERF-32 | Yes | `[]byte(config.CertificatePEM)` copies string |
| 576 | PERF-42 | Yes | `fmt.Errorf` with static string, no format args |
| 577 | PERF-35 | Yes | `fmt.Errorf` boxes `err` through `interface{}` |
| 578 | BP-1 | No | `pem.Decode` remainder discarded, not error |
| 579 | PERF-32 | Yes | `[]byte(config.PrivateKeyPEM)` copies string |
| 580 | BP-1 | No | `pem.Decode` remainder discarded, not error |
| 581 | PERF-32 | Yes | `[]byte(chainPEM)` conversion inside loop |
| 582 | PERF-40 | No | Only one `time.Now()` call in function |
| 583 | BP-1 | No | `now.Zone()` returns name/offset, not error |
| 584 | BP-3 | Yes | `panic(err)` outside `main` or test file |
| 585 | PERF-32 | Yes | `[]byte(newByteRange)` copies formatted string |
| 586 | PERF-32 | Yes | `[]byte(sigHex)` copies hex string to bytes |
| 587 | PERF-192 | Yes | `make(map[int]int)` without size hint |
| 588 | PERF-192 | Yes | `make(map[int][]*StructElem)` without hint |
| 589 | PERF-192 | Yes | `make(map[int]int)` without size hint |
| 590 | PERF-123 | Yes | `make([]*StructElem, 0)` redundant zero arg |
| 591 | PERF-123 | Yes | `make([]*StructElem, 0)` redundant zero arg |
| 592 | PERF-44 | No | Only one type assertion on `Kids[0]` |
| 593 | PERF-192 | Yes | `make(map[int]*StructElem)` without size hint |
| 594 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 595 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 596 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 597 | PERF-192 | Yes | `make(map[string]xml.StartElement)` no hint |
| 598 | PERF-4 | Yes | `make(map[string]string)` inside token loop |
| 599 | PERF-192 | Yes | Loop-body map created without size hint |
| 600 | PERF-6 | Yes | `fmt.Fprintf` inside XML token loop |

## Notable FP Patterns

- **PERF-109** (5) — loop flagged without map key recomputation
- **BP-1** (11) — non-error `_` bindings (`pem.Decode`, raw bytes, `Zone()`, scan count)
- **PERF-119** (1), **PERF-40** (1), **PERF-44** (1) — single-instance misfires