# GoPdfSuit Performance Implementation Plan

**Source audit:** [PERFORMANCE_AUDIT.md](./PERFORMANCE_AUDIT.md)  
**Status:** Pass 1 ✅ | Pass 2 ✅ | Pass 3 ✅ - All phases complete

---

## Pass 3 - Advanced ✅ COMPLETE

| ID | Task | Files | Status |
|----|------|-------|--------|
| P3-01 | Allocation-free `WrapText` | `utils.go`, `draw.go`, `utils_wrap_test.go` | ✅ |
| P3-02 | `ExtraObjects map[int][]byte` | `pagemanager.go`, `generator.go`, `outline.go` | ✅ |
| P3-03 | Redact parser unification | `redact/*.go`, `redact_parser_test.go` | ✅ |
| P3-04 | Struct packing (`ImageObject` reorder) | `image.go` | ✅ partial |
| P3-05 | `StructElem.Kids []StructKid` | `structure.go`, `generator.go` | ✅ |

### P3-04 Note

`ImageObject` fields reordered. `TTFFont.Flags` / `FontObjectIDs` deferred (low impact vs. font pipeline risk).

### Pass 3 Impact

- Wrap path: **~11% fewer allocs** at 10K rows (`WrapTextInto` buffer reuse)
- Redact: merge-based object scanner (no catastrophic regex on full PDF)
- PDF/UA: typed `StructKid` eliminates interface boxing in structure tree

---

## Pass 1 - Low-Hanging Fruit ✅ (10/10)

See [PASS1_BLUEPRINTS.md](./PASS1_BLUEPRINTS.md). Baseline: [baselines/bench_pass1_20260525.txt](./baselines/bench_pass1_20260525.txt).

---

## Pass 2 - Architecture Changes ✅ (12/12)

Baseline: [baselines/bench_pass2_20260525.txt](./baselines/bench_pass2_20260525.txt).

---

## Change Log

| Date | Change |
|------|--------|
| 2026-05-25 | Pass 1–3 implemented; all tests pass |

---

## Validation

```bash
go test ./internal/pdf/... ./pkg/gopdflib/...
go test -run='^$' -bench=BenchmarkGenerateTemplatePDF_WrapEnabled -benchmem ./internal/pdf/
```
