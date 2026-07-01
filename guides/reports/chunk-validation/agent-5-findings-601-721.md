# Agent 5 — Findings 601–721 Validation (Manual Review)

> Exported from prior subagent session. Domain-agnostic evaluation per `CHUNK_VALIDATOR.md`.

## Summary

- Total findings analyzed: 121
- Correctly Fired (True Positives): 104
- Incorrectly Fired (False Positives): 17
- FP rate: 14.0%

## Per-Finding Table

| Finding | Rule | Correctly Fired? | Reason |
|---------|------|------------------|--------|
| 601 | PERF-6 | Yes | `fmt.Fprintf` inside XML token loop |
| 602 | PERF-109 | No | Attribute iteration, not map key recomputation |
| 603 | PERF-35 | Yes | `fmt.Sprintf` boxes args inside outer token loop |
| 604 | PERF-6 | Yes | `fmt.Sprintf` used inside the token loop |
| 605 | PERF-192 | Yes | `make(map)` lacks size hint despite known `len(se.Attr)` |
| 606 | PERF-47 | Yes | `strings.SplitN` called inside style-parts loop |
| 607 | PERF-122 | No | `HasPrefix` followed by array slice, not string trim |
| 608 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 609 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 610 | PERF-6 | Yes | `fmt.Fprintf` inside transform-parts loop |
| 611 | PERF-122 | No | `HasPrefix` followed by array slice, not string trim |
| 612 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 613 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 614 | PERF-6 | Yes | `fmt.Fprintf` inside transform-parts loop |
| 615 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 616 | PERF-6 | Yes | `fmt.Fprintf` inside transform-parts loop |
| 617 | PERF-122 | No | `HasPrefix` followed by array slice, not string trim |
| 618 | PERF-6 | Yes | `fmt.Fprintf` inside transform-parts loop |
| 619 | BP-1 | Yes | `strconv.ParseInt` error discarded with `_` |
| 620 | BP-1 | Yes | `strconv.ParseInt` error discarded with `_` |
| 621 | BP-1 | Yes | `strconv.ParseInt` error discarded with `_` |
| 622 | BP-1 | Yes | `strconv.ParseInt` error discarded with `_` |
| 623 | BP-1 | Yes | `strconv.ParseInt` error discarded with `_` |
| 624 | BP-1 | Yes | `strconv.ParseInt` error discarded with `_` |
| 625 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 626 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 627 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 628 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 629 | PERF-6 | Yes | `fmt.Fprintf` inside path-token loop |
| 630 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 631 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 632 | PERF-6 | Yes | `fmt.Fprintf` inside path-token loop |
| 633 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 634 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 635 | PERF-6 | Yes | `fmt.Fprintf` inside path-token loop |
| 636 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 637 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 638 | PERF-6 | Yes | `fmt.Fprintf` inside path-token loop |
| 639 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 640 | PERF-6 | Yes | `fmt.Fprintf` inside path-token loop |
| 641 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 642 | PERF-6 | Yes | `fmt.Fprintf` inside path-token loop |
| 643 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 644 | PERF-6 | Yes | `fmt.Fprintf` inside path-token loop |
| 645 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 646 | PERF-6 | Yes | `fmt.Fprintf` inside path-token loop |
| 647 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 648 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 649 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 650 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 651 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 652 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 653 | PERF-6 | Yes | `fmt.Fprintf` inside path-token loop |
| 654 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 655 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 656 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 657 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 658 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 659 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 660 | PERF-6 | Yes | `fmt.Fprintf` inside path-token loop |
| 661 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 662 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 663 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 664 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 665 | PERF-6 | Yes | `fmt.Fprintf` inside path-token loop |
| 666 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 667 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 668 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 669 | BP-1 | Yes | `strconv.ParseFloat` error discarded with `_` |
| 670 | PERF-6 | Yes | `fmt.Fprintf` inside path-token loop |
| 671 | BP-1 | Yes | `strconv.Atoi` error discarded with `_` |
| 672 | PERF-2 | Yes | `testLine += " "` concatenates inside word loop |
| 673 | PERF-119 | Yes | Consecutive `append` calls to same slice per iteration |
| 674 | PERF-128 | No | Only two appends per iteration, not three or more |
| 675 | PERF-7 | Yes | `defer` appears lexically inside `for` loop body |
| 676 | BP-11 | Yes | `defer` inside loop defers until function returns |
| 677 | PERF-35 | No | `fmt.Errorf` on rare error path, not hot path |
| 678 | BP-1 | Yes | `resp.Body.Close()` error explicitly discarded |
| 679 | BP-5 | Yes | `Close()` return value ignored |
| 680 | BP-1 | Yes | `os.Remove` error discarded with `_` |
| 681 | BP-1 | Yes | `os.Remove` error discarded with `_` |
| 682 | PERF-35 | Yes | `fmt.Sprintf` boxes args inside generation loop |
| 683 | PERF-6 | Yes | `fmt.Sprintf` inside record-generation loop |
| 684 | PERF-6 | Yes | `fmt.Sprintf` inside record-generation loop |
| 685 | PERF-35 | No | `fmt.Sprintf` called once, not on hot path |
| 686 | CWE-497 | Yes | Returns OS, arch, CPU, Go version to callers |
| 687 | PERF-148 | No | Channel is buffered (`make(chan int, iterations)`) |
| 688 | PERF-36 | No | `for range numWorkers` has no loop variable capture |
| 689 | PERF-7 | Yes | `defer wg.Done()` lexically inside worker loop |
| 690 | BP-11 | Yes | `defer` inside loop defers until function returns |
| 691 | PERF-40 | Yes | `time.Now` called repeatedly in same function |
| 692 | PERF-35 | No | `fmt.Sprintf` called once, not on hot path |
| 693 | CWE-497 | Yes | Exposes host environment details to callers |
| 694 | PERF-109 | No | Slice iteration, not map key recomputation |
| 695 | PERF-6 | Yes | `fmt.Sprintf` inside trade-generation loop |
| 696 | PERF-6 | Yes | `fmt.Sprintf` inside table-row loop |
| 697 | PERF-6 | Yes | `fmt.Sprintf` inside table-row loop |
| 698 | PERF-6 | Yes | `fmt.Sprintf` inside table-row loop |
| 699 | PERF-6 | Yes | `fmt.Sprintf` inside table-row loop |
| 700 | PERF-6 | Yes | `fmt.Sprintf` inside table-row loop |
| 701 | PERF-6 | Yes | `fmt.Sprintf` inside table-row loop |
| 702 | PERF-6 | Yes | `fmt.Sprintf` inside table-row loop |
| 703 | PERF-148 | No | Channel is buffered (`make(chan int, iterations)`) |
| 704 | PERF-36 | No | `for range numWorkers` has no loop variable capture |
| 705 | PERF-7 | Yes | `defer wg.Done()` lexically inside worker loop |
| 706 | BP-11 | Yes | `defer` inside loop defers until function returns |
| 707 | PERF-40 | Yes | `time.Now` used inside per-job goroutine loop |
| 708 | PERF-192 | No | No `len(src)` available before map is populated |
| 709 | BP-1 | Yes | `json.Marshal` error discarded with `_` |
| 710 | PERF-119 | No | Mutually exclusive branches, not consecutive appends |
| 711 | PERF-123 | Yes | `make([]MathElement, 0)` uses redundant zero length |
| 712 | PERF-119 | No | Loop append not consecutive with later standalone append |
| 713 | PERF-123 | Yes | `make([]MathElement, 0)` uses redundant zero length |
| 714 | PERF-123 | Yes | `make([]MathElement, 0)` uses redundant zero length |
| 715 | PERF-123 | Yes | `make([]MathElement, 0)` uses redundant zero length |
| 716 | PERF-3 | Yes | `make([]*MathLayout, cols)` inside row loop |
| 717 | PERF-46 | Yes | `strings.TrimSpace` allocates inside children loop |
| 718 | PERF-123 | Yes | `make([]MathElement, 0)` uses redundant zero length |
| 719 | PERF-123 | Yes | `make([]MathElement, 0)` uses redundant zero length |
| 720 | PERF-123 | Yes | `make([]MathElement, 0)` uses redundant zero length |
| 721 | PERF-109 | No | Path-point iteration, not map key recomputation |

## Notable FP Patterns

- **PERF-122** (3) — `HasPrefix` + slice, not string trim
- **PERF-148 / PERF-36** (4) — buffered channels, no loop-variable capture
- **PERF-109** (3) — slice/path loops without map keys
- **PERF-35** (3) — one-shot fmt on non-hot paths
- **PERF-119 / PERF-128** (3) — non-consecutive appends
- **PERF-192** (1) — no known size before populate