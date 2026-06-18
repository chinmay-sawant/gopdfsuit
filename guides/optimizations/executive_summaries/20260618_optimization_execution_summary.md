# Optimization Execution Summary — 2026-06-18

## Date & Scope

Profile and optimization work on the **PyPDFSuit Python binding** (`feat/optimization-5.5-medium`), using the Zerodha gold-standard workload (80% retail / 15% active / 5% HFT) on WSL2 (i7-13700HX, 48 workers, Python 3.13.10). Goal: explain why Python throughput (~253 ops/s) lags native Go (~11,721 ops/s) and ship the highest-impact fix.

## Key Outcomes

- **Root cause identified:** Python latency is dominated by `to_dict()` tree walks and JSON+CGO overhead — not the PDF renderer itself.
- **P0 JSON cache implemented:** Throughput **253 → 1,505 ops/s** (~**5.9×**, **+494%**) at 48 workers, 5000 iterations.
- **Cached JSON (concurrent):** **1,029 ops/s** (500 iter) / **870 ops/s** (200 iter) — **~4.1×** vs full path (**222 ops/s**).
- **Single-thread cached JSON:** Retail **776 ops/s** (−29% latency), Active **1,160 ops/s** (−60%), HFT **32 ops/s** (−68%).
- **Ceiling after P0:** Python remains **~11× below** native Go (11,721 ops/s) on the same weighted mix.
- **Retail-only benchmark:** **1,323 ops/s** (24 workers) vs **222 ops/s** weighted full path.

## Work Completed

- Ran phase profiler (`make bench-pypdfsuit-profile`) with per-call breakdown: `to_dict`, `json_dumps`, `cgo_call`, `copy_back`, `free_result`.
- Ran production benchmark (`make bench-pypdfsuit-zerodha`, 5000 iter, 48 workers) and Go-side comparison (`pypdfsuit_go_profile.go`).
- **Implemented P0:** `serialize_template()` JSON payload cache in `bindings/python/pypdfsuit/generator.py`; `generate_pdf()` reuses cached UTF-8; `invalidate_template_cache()` for mutations.
- Documented prioritized optimization plan (P0–P3), expected throughput by phase, and verification commands.

## Findings / Bottlenecks

| Bucket | Weighted latency share | Notes |
|--------|------------------------|-------|
| Python `to_dict()` | **~50–67%** (active/HFT); ~15% retail | **2,864 `_to_dict` calls/PDF** |
| CGO + Go `json.Unmarshal` + render | **~81%** retail; ~45% active | HFT JSON **~1.1 MB** |
| FFI `string_at` copy-back | <1% retail/active; ~1% HFT | Negligible |

**Phase breakdown (single-thread, ms):**

| Template | Total | `to_dict` | `json_dumps` | `cgo_call` | Serialize % |
|----------|------:|----------:|-------------:|-----------:|------------:|
| Retail | 1.81 | 0.27 | 0.05 | 1.47 | 18% |
| Active | 2.17 | 1.08 | 0.11 | 0.98 | 55% |
| HFT | 97.70 | 60.34 | 4.74 | 31.78 | 67% |

- **GIL limits parallelism** during CPU-bound `to_dict` + `json.dumps`; native Go uses real goroutines.
- **PEM keys** inflate retail JSON to **11 KB** on every call.

## Open Items / Next Steps

| Priority | Item | Est. impact | Status |
|----------|------|-------------|--------|
| **P0** | Pre-serialize in `pypdfsuit_bench.py` (pass cached bytes) | Same as cache | Not done (~1 hr) |
| **P1** | Fast `to_dict`: `orjson`, `slots=True`, dict cache | **+50–100%** active/HFT | Planned |
| **P1** | Skip Go `json.Unmarshal` via `GeneratePDFFromTemplateHandle` CGO | **+10–30%** post-cache | Planned |
| **P2** | Concurrency tuning; zero-copy PDF return (`memoryview`) | **+20–40%** / **+5%** HFT | Planned |
| **P3** | HTTP-first production (Gin **7,515 req/s** retail k6) | Match native scale | Architecture path |

**Projected weighted 48-worker ops/s:** P0 ~1,000 → P0+P1 ~1,200–1,400 → P0+P1+skip unmarshal ~1,500–2,000 (vs Go **11,721**).

## Source Documents

| File | Role |
|------|------|
| `20260618_pypdfsuit_profile_optimization_report.md` | Primary report |
| `guides/cursor/baselines/pypdfsuit_profile_20260618.txt` | Raw profile output |
| `sampledata/gopdflib/zerodha/pypdfsuit_profile.py` | Phase profiler harness |
| `sampledata/gopdflib/zerodha/pypdfsuit_go_profile.go` | Go native vs JSON comparison |
| `bindings/python/pypdfsuit/generator.py` | P0 JSON cache implementation |