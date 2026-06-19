# PyPDFSuit Benchmark Profile and Optimization Checklist

**Date:** 2026-06-19  
**Harness:** `make bench-pypdfsuit-zerodha` and `make bench-pypdfsuit-zerodha-x5`  
**Profile harness:** `make bench-pypdfsuit-profile` / `sampledata/gopdflib/zerodha/pypdfsuit_profile.py`  
**Workload:** Zerodha weighted mix, 80% retail / 15% active / 5% HFT  
**Benchmark mode:** full execution path only; every benchmark operation rebuilds the Python template dict and JSON payload before crossing the FFI boundary.  
**Raw artifacts:** `guides/cursor/baselines/pypdfsuit_pprof_runs/`

This plan keeps the headline benchmark honest: benchmark runs must continue to generate the template dict and JSON from scratch for every PDF. JSON payload caching and cache-only benchmark targets were removed because they made the result look like benchmark tuning rather than actual execution.

---

## Fresh Benchmark Results

Command:

```bash
make bench-pypdfsuit-zerodha-x5
```

Configuration:

| Setting | Value |
|---------|------:|
| Iterations per run | 5,000 |
| Workers | 48 |
| JSON cache | removed |
| Python | 3.13.10 |
| Machine | WSL2, i7-13700HX, 24 logical CPUs, 7.6 GiB RAM |

Results:

| Run | Throughput | Avg latency | Max latency |
|-----|-----------:|------------:|------------:|
| 1 | 205.09 ops/sec | 202.787 ms | 4,782.265 ms |
| 2 | 228.70 ops/sec | 184.129 ms | 3,635.044 ms |
| 3 | 228.19 ops/sec | 187.800 ms | 3,004.876 ms |
| 4 | 222.95 ops/sec | 189.331 ms | 3,393.410 ms |
| 5 | 222.07 ops/sec | 190.729 ms | 3,237.903 ms |

Summary:

| Metric | Value |
|--------|------:|
| Best | 228.70 ops/sec |
| Worst | 205.09 ops/sec |
| Mean | 221.40 ops/sec |
| Median | 222.95 ops/sec |
| Stddev | 9.60 ops/sec |
| Operating band | about 205-229 ops/sec |

Interpretation:

- The original pypdfsuit baseline was about **221 ops/sec mean** on the weighted Zerodha workload.
- Tail latency is dominated by the HFT case: only 218 of 5,000 operations were HFT in each run, but those 1.1 MB JSON / 2.4 MB PDF operations push max latency above 3 seconds.
- The benchmark measures the full Python API path: template object, `to_dict`, `json.dumps`, ctypes call, Go JSON decode, PDF generation, and copy-back.

---

## Fresh Profile Results

Raw profile:

```text
guides/cursor/baselines/pypdfsuit_pprof_runs/pypdfsuit_profile.txt
```

Single-thread phase breakdown:

| Template | JSON size | PDF size | Total mean | `to_dict` | `json_dumps` | `cgo_call` | Serialize share |
|----------|----------:|---------:|-----------:|----------:|-------------:|-----------:|----------------:|
| Retail | 11,389 B | 61,290 B | 1.794 ms | 0.288 ms | 0.060 ms | 1.434 ms | 19.4% |
| Active | 20,977 B | 76,072 B | 2.345 ms | 1.129 ms | 0.123 ms | 1.071 ms | 53.4% |
| HFT | 1,131,142 B | 2,424,782 B | 107.867 ms | 66.889 ms | 5.196 ms | 34.897 ms | 66.8% |

cProfile, weighted full path:

| Hot path | Calls / time | Meaning |
|----------|-------------:|---------|
| `types._to_dict` | 572,750 calls, 1.516 s cumulative | dominant Python tree walk |
| `types._python_to_json_key` | 417,690 calls, 0.660 s cumulative | repeated field-name mapping |
| `_bindings.call_bytes_result` | 200 calls, 0.630 s cumulative | ctypes + Go work |
| `json.dumps` | 200 calls, 0.067 s cumulative | not the main serializer cost |
| `ctypes.string_at` | 200 calls, 0.014 s cumulative | copy-back is negligible except HFT |

## 2026-06-19 Implementation Update

Scope completed in Python/Makefile only. No `.go` files were changed.

Changes:

- Precomputed Python dataclass field-to-JSON-key mappings so `_python_to_json_key` is no longer called for every field in every PDF.
- Switched `serialize_template()` to compact JSON separators, reducing bytes sent through the Go JSON path without changing the schema.
- Removed automatic JSON payload caching from `generate_pdf()`.
- Removed `generate_pdf(use_cache=...)`, `template_json=...`, and `invalidate_template_cache()`.
- Removed cache benchmark controls from the Makefile.
- Added `PAYLOAD_SCENARIO=weighted|retail_only|active_only|hft_only` to `pypdfsuit_bench.py`.
- Added p50/p95/p99 latency reporting to `pypdfsuit_bench.py`.

Post-change validation:

| Command | Scenario | Cache | Throughput | P50 | P95 | P99 |
|---------|----------|-------|-----------:|----:|----:|----:|
| `make bench-pypdfsuit-zerodha` | weighted 80/15/5 | removed | **565.16 ops/sec** | 40.052 ms | 270.051 ms | 481.097 ms |

Post-removal profile from `guides/cursor/baselines/pypdfsuit_pprof_runs/pypdfsuit_profile_no_json_cache_20260619.txt`:

| Template | JSON size | Total | `to_dict` | `json_dumps` | `cgo_call` | Serialize share |
|----------|----------:|------:|----------:|-------------:|-----------:|----------------:|
| Retail | 10,956 B | 1.937 ms | 0.089 ms | 0.074 ms | 1.753 ms | 8.4% |
| Active | 19,531 B | 2.045 ms | 0.342 ms | 0.163 ms | 1.512 ms | 24.7% |
| HFT | 1,054,811 B | 76.033 ms | 19.743 ms | 6.680 ms | 48.079 ms | 34.8% |

Weighted cProfile after JSON-cache removal:

| Hot path | Cumulative time |
|----------|----------------:|
| `types._to_dict` | 1.009 s |
| `_bindings.call_bytes_result` | 0.825 s |
| `json.dumps` / encoder | 0.099 s |
| `ctypes.string_at` | 0.019 s |

P2 serializer pass:

| Command | Scenario | Cache | Throughput | P50 | P95 | P99 |
|---------|----------|-------|-----------:|----:|----:|----:|
| `make bench-pypdfsuit-zerodha` | weighted 80/15/5 | removed | **736.55 ops/sec** | 33.512 ms | 194.619 ms | 316.176 ms |

P2 profile from `guides/cursor/baselines/pypdfsuit_pprof_runs/pypdfsuit_p2_serializer_profile_20260619.txt`:

| Template | Total | `to_dict` | `json_dumps` | `cgo_call` | Serialize share |
|----------|------:|----------:|-------------:|-----------:|----------------:|
| Retail | 2.553 ms | 0.077 ms | 0.105 ms | 2.337 ms | 7.1% |
| Active | 2.145 ms | 0.168 ms | 0.188 ms | 1.753 ms | 16.6% |
| HFT | 76.834 ms | 11.167 ms | 8.624 ms | 55.497 ms | 25.8% |

P2 cProfile shifted the hot path from Python tree walking to the FFI/render boundary:

| Hot path | Cumulative time |
|----------|----------------:|
| `_bindings.call_bytes_result` | 0.996 s |
| `PDFTemplate.to_dict` and child serializers | 0.169 s |
| `json.dumps` / encoder | 0.109 s |

Existing Go retail phase profile from `guides/cursor/baselines/pypdfsuit_pprof_runs/pypdfsuit_go_retail_profile_20260619.txt`:

| Path | Mean |
|------|-----:|
| Native `GeneratePDF` | 1.168 ms |
| `json.Unmarshal` | 0.116 ms |
| Generate after JSON unmarshal | 1.132 ms |
| Combined JSON round-trip | 1.248 ms |

P3 x10 benchmark from `guides/cursor/baselines/pypdfsuit_pprof_runs/pypdfsuit_p3_x10_20260619.txt`:

| Metric | Value |
|--------|------:|
| Runs | 10 |
| Best throughput | **927.19 ops/sec** |
| Worst throughput | 642.15 ops/sec |
| Mean throughput | **750.78 ops/sec** |
| Median throughput | 739.34 ops/sec |
| Stddev throughput | 98.17 ops/sec |
| Mean avg latency | 58.121 ms |
| Mean p50 latency | 34.304 ms |
| Mean p95 latency | 193.663 ms |
| Mean p99 latency | 317.655 ms |

| Run | Throughput | Avg latency | P95 | P99 |
|-----|-----------:|------------:|----:|----:|
| 1 | 766.39 ops/sec | 53.564 ms | 175.677 ms | 280.109 ms |
| 2 | **927.19 ops/sec** | 46.927 ms | 152.231 ms | 253.882 ms |
| 3 | 864.14 ops/sec | 50.131 ms | 163.154 ms | 273.756 ms |
| 4 | 681.85 ops/sec | 63.784 ms | 217.419 ms | 359.128 ms |
| 5 | 657.66 ops/sec | 66.198 ms | 228.843 ms | 371.554 ms |
| 6 | 796.71 ops/sec | 53.737 ms | 180.186 ms | 273.052 ms |
| 7 | 712.29 ops/sec | 61.052 ms | 200.246 ms | 335.058 ms |
| 8 | 642.15 ops/sec | 67.385 ms | 228.551 ms | 393.933 ms |
| 9 | 649.65 ops/sec | 65.483 ms | 217.439 ms | 345.050 ms |
| 10 | 809.73 ops/sec | 52.953 ms | 172.887 ms | 291.025 ms |

P3 profile from `guides/cursor/baselines/pypdfsuit_pprof_runs/pypdfsuit_p3_profile_20260619.txt`:

| Template | Total | `to_dict` | `json_dumps` | `cgo_call` | Copy-back | Serialize share |
|----------|------:|----------:|-------------:|-----------:|----------:|----------------:|
| Retail | 4.123 ms | 0.197 ms | 0.208 ms | 3.624 ms | 0.077 ms | 9.8% |
| Active | 3.362 ms | 0.353 ms | 0.302 ms | 2.624 ms | 0.064 ms | 19.5% |
| HFT | 82.838 ms | 12.523 ms | 8.933 ms | 59.395 ms | 1.937 ms | 25.9% |

P3 cProfile:

| Hot path | Cumulative time |
|----------|----------------:|
| `_bindings.call_bytes_result` | 0.918 s |
| `PDFTemplate.to_dict` and child serializers | 0.127 s |
| `json.dumps` / encoder | 0.091 s |
| `ctypes.string_at` | 0.018 s |

Interpretation:

- The previous **3,621.65 ops/sec** cache-only retail result was removed from the active benchmark surface because it bypassed per-call JSON construction.
- The active pypdfsuit benchmark now measures actual execution: template object -> `to_dict` -> compact `json.dumps` -> ctypes -> Go JSON decode/render -> PDF copy-back.
- The weighted full-execution path improved from the earlier **221 ops/sec mean** to **565 ops/sec** after serializer cleanup and JSON-cache removal.
- Reaching 3000 ops/sec on the full weighted HFT mix without Go changes is not currently supported by the profile evidence; HFT now spends **48.079 ms** in `cgo_call` and **26.423 ms** in Python serialization per single-thread call.
- After P2 serializers, weighted throughput improved again to **736.55 ops/sec**. HFT Python serialization dropped to **19.791 ms**, but HFT `cgo_call` rose to **55.497 ms** in the latest sample and now dominates.
- The available Go retail profile shows JSON decode is small for retail. HFT-specific Go phase separation would require a Go-side HFT profile harness; no Go files were edited in this pass.
- The P3 x10 run shows the current honest full-path operating band is roughly **640-930 ops/sec**, with a stable center around **740-750 ops/sec**. This is the number to publish for the current CGO Python benchmark, not cache-only numbers.
- The remaining Python-side budget is small in the weighted cProfile: `to_dict` plus JSON encoding is about **0.218 s cumulative** versus **0.918 s** in the FFI/render path. Further Python-only work can improve tail latency, but large throughput jumps require reducing `cgo_call` cost or changing the workload/API contract.

P4 exact `make bench-pypdfsuit-zerodha` x5 after compact UTF-8 JSON from `guides/cursor/baselines/pypdfsuit_pprof_runs/pypdfsuit_p4_make_x5_20260619.txt`:

| Metric | Value |
|--------|------:|
| Runs | 5 |
| Best throughput | **855.55 ops/sec** |
| Worst throughput | 804.13 ops/sec |
| Mean throughput | **835.56 ops/sec** |
| Median throughput | 843.17 ops/sec |
| Stddev throughput | 22.20 ops/sec |
| Mean avg latency | 51.620 ms |
| Mean p50 latency | 31.758 ms |
| Mean p95 latency | 170.258 ms |
| Mean p99 latency | 269.198 ms |

| Run | Throughput | Avg latency | P95 | P99 |
|-----|-----------:|------------:|----:|----:|
| 1 | 855.55 ops/sec | 50.343 ms | 161.608 ms | 258.910 ms |
| 2 | 853.58 ops/sec | 50.120 ms | 169.211 ms | 260.317 ms |
| 3 | 843.17 ops/sec | 50.950 ms | 166.494 ms | 268.432 ms |
| 4 | 821.37 ops/sec | 52.784 ms | 178.168 ms | 267.385 ms |
| 5 | 804.13 ops/sec | 53.903 ms | 175.807 ms | 290.945 ms |

User-observed local reference runs before this refresh reached **1108.97** and **1113.40 ops/sec** with the same no-cache benchmark configuration. This session's P4 x5 band is lower, so publish it as the current measured artifact for this shell and treat the user's 1100+ numbers as a higher local reference, not the canonical artifact unless captured into `guides/cursor/baselines/`.

P4 profile from `guides/cursor/baselines/pypdfsuit_pprof_runs/pypdfsuit_p4_make_profile_20260619.txt`:

| Template | JSON size | Total | `to_dict` | `json_dumps` | `cgo_call` | Copy-back | Serialize share |
|----------|----------:|------:|----------:|-------------:|-----------:|----------:|----------------:|
| Retail | 10,935 B | 1.579 ms | 0.032 ms | 0.071 ms | 1.460 ms | 0.013 ms | 6.5% |
| Active | 19,279 B | 1.210 ms | 0.076 ms | 0.128 ms | 0.991 ms | 0.011 ms | 16.9% |
| HFT | 1,042,811 B | 44.739 ms | 5.141 ms | 6.287 ms | 32.815 ms | 0.472 ms | 25.5% |

P4 cProfile:

| Hot path | Cumulative time |
|----------|----------------:|
| `_bindings.call_bytes_result` | 0.594 s |
| `PDFTemplate.to_dict` and child serializers | about 0.127 s |
| `json.dumps` / encoder | 0.079 s |
| UTF-8 string encode | 0.012 s |
| `ctypes.string_at` | 0.013 s |

P4 interpretation:

- Compact UTF-8 JSON is a valid full-execution optimization: every call still rebuilds `to_dict` and `json.dumps`, but rupee symbols are emitted as UTF-8 instead of `\u20b9`, reducing HFT JSON from about **1.055 MB** to **1.043 MB**.
- The profile now points at FFI/render cost as the hard ceiling for the current API: `cgo_call` is **92.5%** of retail, **81.8%** of active, and **73.3%** of HFT single-thread latency.
- The largest remaining Python-only hotspot is HFT serialization at **11.428 ms**, but even removing all HFT serialization would leave a **32.815 ms** HFT `cgo_call`.
- Benchmark harness cleanup such as precomputing the deterministic workload sequence and removing the counter lock could improve reported ops/sec slightly, but it would not improve `generate_pdf()` itself. Keep that separate from production optimization to avoid benchmark-only tuning.
- Major throughput beyond the current no-cache band requires one of: Go render-boundary work, a separate handle/batch/service API, or a schema/content representation change for large HFT tables. Those are not the same benchmark as `make bench-pypdfsuit-zerodha`.

---

## Guardrails

- [x] Keep `make bench-pypdfsuit-zerodha` and `make bench-pypdfsuit-zerodha-x5` on the full execution path.
- [x] Remove cache-enabled pypdfsuit benchmark targets from the active benchmark surface.
- [x] Every no-cache benchmark operation must build the Python template dict and JSON payload from scratch before `GeneratePDF`.
- [x] Keep output PDFs generated on every run for sanity checks: retail, active, and HFT.
- [x] Do not use JSON caching to claim pypdfsuit throughput.

---

## P0 Checklist - Make Serialization Cheaper Without Changing Baseline Semantics

- [x] **Move JSON key mapping out of the hot path.**
  - Evidence: `_python_to_json_key` costs 0.660 s cumulative in the 200-call weighted profile.
  - Plan: precompute field-to-JSON-key tuples per dataclass type, for example `[(field_name, json_key), ...]`.
  - Constraint: no-cache benchmark still walks each object and emits fresh JSON each call.
  - Expected impact: lower active/HFT `to_dict` CPU without changing API behavior.

- [x] **Replace generic `_to_dict` recursion for core table types with specialized methods.**
  - Evidence: `Cell.to_dict`, `Row.to_dict`, and generic `_to_dict` account for hundreds of thousands of calls per profile.
  - Plan: implement direct serializers for `PDFTemplate`, `Element`, `Table`, `Row`, `Cell`, and `Config`, using local variables and explicit key names.
  - Constraint: preserve current JSON schema, including compatibility names such as `chequebox`, `maxcolumns`, and `columnwidths`.
  - Result: direct serializers now cover `PDFTemplate`, `Element`, `Table`, `Row`, `Cell`, `Config`, `SignatureConfig`, `Bookmark`, `Image`, and related hot schema types.

- [x] **Add a generated serializer test fixture.**
  - Plan: compare old `template.to_dict()` output against the new fast serializer for retail, active, HFT, and existing Python tests.
  - Acceptance: `bindings/python/tests/test_serializer_schema.py` checks the Go-facing JSON key schema for specialized serializers, including `chequebox`, `bgcolor`, `textcolor`, `mathEnabled`, `certificatePem`, `privateKeyPem`, and `certificateChain`.

- [x] **Benchmark the full path after serializer changes.**
  - Commands:
    - `make bench-pypdfsuit-zerodha`
    - `make bench-pypdfsuit-profile`
    - `cd bindings/python && python3 -m pytest tests`
  - Acceptance: no-cache mean improves from 221 ops/sec while JSON is regenerated on every call.

---

## P1 Checklist - Remove Benchmark JSON Caching

- [x] **Remove automatic JSON payload caching from `generate_pdf()`.**
  - Current implementation: `generate_pdf(template)` always calls `serialize_template(template)` and builds fresh bytes.
  - Acceptance: tests assert repeated `generate_pdf()` calls serialize twice.

- [x] **Remove cache benchmark controls.**
  - Removed `bench-pypdfsuit-zerodha-cache`, `bench-pypdfsuit-retail-cache`, and `bench-pypdfsuit-retail-cache-x2`.
  - Acceptance: `make bench-help` no longer advertises cache targets.

- [x] **Remove cached controls from profiling.**
  - `pypdfsuit_profile.py` now reports only full-path phase timings and weighted full-path cProfile.

---

## P1 Checklist - Reduce HFT Tail Latency

- [ ] **Create HFT row-block serialization helpers.**
  - Evidence after cache removal: HFT spends 26.423 ms in serialization, 34.8% of total single-thread latency.
  - Plan: build row JSON chunks with stable column schema and only substitute changing values.
  - Constraint: every run must still generate JSON from scratch; caching row JSON between benchmark calls is not allowed for the baseline.
  - Status: not implemented as row-block JSON chunks because that would move toward preformatted row JSON and risks violating the full object-to-JSON execution rule. P2 direct row/cell serializers were implemented instead.

- [ ] **Avoid repeated certificate/key serialization when signatures are constant.**
  - Evidence: retail JSON includes PEM key and cert chain every call; this is smaller than HFT but affects the dominant 80% case.
  - Plan: investigate a Go-side signature profile handle for fixed benchmark certs or a Python-side compact signature reference.
  - Constraint: the benchmark must still construct the final JSON payload from scratch; handle-based APIs must be separately benchmarked.
  - Acceptance: any handle path is opt-in and does not change existing JSON API behavior.

- [x] **Add p95/p99 latency reporting to `pypdfsuit_bench.py`.**
  - Evidence: max latency is noisy and HFT-driven; average latency hides the tail.
  - Plan: report p50, p95, and p99 from `durations`.
  - Acceptance: x5 artifacts show throughput plus tail percentiles for every run.

---

## P2 Checklist - Reduce Go Boundary Cost After Python Serialization Is Fixed

- [ ] **Profile Go shared-library `GeneratePDF` under pypdfsuit HFT.**
  - Status: partially evaluated using the existing retail Go profile harness; true HFT split remains deferred because this pass intentionally did not edit Go files.
  - Retail evidence: `json.Unmarshal` is 0.116 ms, native render is 1.168 ms, and JSON round-trip render is 1.248 ms.
  - HFT evidence from Python profile: HFT `cgo_call` is 55.497 ms, 72.2% of HFT total latency.
  - Constraint: do not use cached/pre-parsed JSON handles for the pypdfsuit benchmark.

- [x] **Evaluate Python-side specialized serializers before any Go-boundary work.**
  - Plan: add direct serializers for `PDFTemplate`, `Element`, `Table`, `Row`, `Cell`, `Config`, and signature structs.
  - Rationale: this preserves actual per-call execution while reducing Python dispatch and dict-walk overhead.
  - Result: weighted benchmark improved from **565.16** to **736.55 ops/sec** without cache controls.
  - Profile result: weighted cProfile `to_dict` path dropped to **0.169 s** cumulative while `_bindings.call_bytes_result` became **0.996 s** cumulative.
  - Current expected ceiling: even eliminating HFT Python serialization entirely would leave the 55 ms HFT `cgo_call`, so this alone is unlikely to reach 3000 ops/sec on the weighted HFT mix.

- [x] **Evaluate batch/handle APIs only as separate production APIs, not benchmark baselines.**
  - Status: rejected for the current benchmark goal because these would bypass per-call full JSON construction.
  - Constraint: any future handle/batch API must be reported separately from `make bench-pypdfsuit-zerodha`.

---

## P3 Checklist - Highest Honest CGO Python Potential

- [x] **Run full-path pypdfsuit x10 benchmark.**
  - Command: `make bench-pypdfsuit-zerodha-x10`.
  - Artifact: `guides/cursor/baselines/pypdfsuit_pprof_runs/pypdfsuit_p3_x10_20260619.txt`.
  - Result: mean **750.78 ops/sec**, median **739.34 ops/sec**, best **927.19 ops/sec**.

- [x] **Run full-path pypdfsuit profiler after P2 serializers.**
  - Command: `python3 sampledata/gopdflib/zerodha/pypdfsuit_profile.py`.
  - Artifact: `guides/cursor/baselines/pypdfsuit_pprof_runs/pypdfsuit_p3_profile_20260619.txt`.
  - Result: `_bindings.call_bytes_result` is now the dominant profile entry; Python serialization is no longer the primary bottleneck.

- [x] **Add a pypdfsuit x10 summary artifact with p50/p95/p99 columns.**
  - Current x10 summary only records throughput and average latency.
  - Plan: extend `run_pypdfsuit_bench_x10.sh` parser to include p50, p95, p99, min, and max latency from each run.
  - Acceptance: `guides/cursor/baselines/pypdfsuit_bench_x10_wsl_stats_latest.txt` now includes p50, p95, p99, min, and max latency columns.

- [x] **Remove serializer-only benchmark from P3 scope.**
  - Rationale: P3 should focus only on the real `make bench-pypdfsuit-zerodha` full PDF generation path.
  - Result: removed `pypdfsuit_serializer_microbench.py` and its generated artifact.

- [x] **Investigate HFT-specific payload reduction without caching.**
  - Plan: reduce Python object churn in `build_hft_template()` and inspect whether repeated static cell values can be represented more compactly while still generating full JSON every call.
  - Constraint: no JSON cache, no pre-serialized row cache, no handle baseline.
  - Result: compact JSON already keeps the HFT payload at about **1.055 MB** versus the earlier **1.131 MB** profile payload. Further reduction would require changing schema, row representation, or rendered content, so no additional HFT payload rewrite was made.

- [x] **Keep Go-boundary improvements as non-baseline unless the API contract changes.**
  - Current profile says FFI/render dominates the benchmark after P2.
  - Any batch, handle, pre-parsed-template, or service-mode design must be labeled as a new production API, not `make bench-pypdfsuit-zerodha`.
  - Acceptance: P3 docs keep full-path pypdfsuit separate from any future batch/handle/service mode.

- [x] **Publish current honest benchmark as best-in-class pypdfsuit CGO baseline.**
  - Recommended headline: **~740-750 ops/sec mean**, **927 ops/sec best**, weighted 80/15/5 Zerodha full execution, 48 workers, no JSON cache.
  - Include tail latency: mean p95 **193.663 ms**, mean p99 **317.655 ms**.
  - Include scope: Python CGO binding over the optimized Go PDF engine, full Python template serialization on every operation.

## P4 Checklist - Exact Make Benchmark Refresh and Remaining Headroom

- [x] **Run exact `make bench-pypdfsuit-zerodha` five times after P3.**
  - Artifact: `guides/cursor/baselines/pypdfsuit_pprof_runs/pypdfsuit_p4_make_x5_20260619.txt`.
  - Result in this session: mean **835.56 ops/sec**, median **843.17 ops/sec**, best **855.55 ops/sec**, p95 mean **170.258 ms**, p99 mean **269.198 ms**.
  - External local reference from the user: **1108.97** and **1113.40 ops/sec** with the same no-cache harness.

- [x] **Run full-path profiler after exact make refresh.**
  - Artifact: `guides/cursor/baselines/pypdfsuit_pprof_runs/pypdfsuit_p4_make_profile_20260619.txt`.
  - Result: `cgo_call` dominates retail, active, and HFT; Python serialization is now secondary.

- [x] **Emit compact UTF-8 JSON without caching.**
  - Implementation: `serialize_template()` uses `ensure_ascii=False` plus compact separators.
  - Constraint preserved: each `generate_pdf(template)` call still runs fresh `template.to_dict()` and `json.dumps(...)`.
  - Result: HFT JSON payload is **1,042,811 bytes** in the current profile, down from the prior compact escaped payload of about **1,054,811 bytes**.

- [ ] **Do not precompute benchmark workload selection unless explicitly labeled as harness cleanup.**
  - Current benchmark still does deterministic per-operation template selection and counter updates inside the worker path.
  - This is not a production `generate_pdf()` cost, so optimizing it may raise reported ops/sec without making PDF generation faster.
  - If implemented later, publish it as a harness-noise cleanup, not a pypdfsuit renderer optimization.

- [ ] **Investigate optional faster JSON encoder only if dependency policy allows it.**
  - Candidate: `orjson` can reduce `json_dumps` time for large HFT payloads while still regenerating JSON every call.
  - Constraint: this adds a dependency and should not be introduced silently into the package or benchmark.
  - Expected ceiling: bounded by HFT `cgo_call` and retail/active render cost; not enough alone for 3000 ops/sec weighted.

- [ ] **Profile HFT render/decode inside the Go shared-library boundary only if Go source work becomes allowed.**
  - Current Python profile can only see `cgo_call`, not JSON decode versus PDF render versus signing/font/table layout internals.
  - This is the most likely place for large remaining gains, but it violates the current "do not touch Go files" guardrail unless explicitly reopened.

## P5 Checklist - FFI/Render-Side Profiling

Python can time the `ctypes` envelope, but it cannot introspect Go stacks while `GeneratePDF` is executing inside `libgopdfsuit.so`. The Python-side `cgo_call` bucket includes:

1. `C.GoString(jsonTemplate)` copy inside the exported Go function.
2. `json.Unmarshal([]byte(goTemplate), &template)`.
3. `gopdflib.GeneratePDF(template)`.
4. C allocation and copy-back setup with `C.malloc` + `C.memcpy`.

The exact split below was gathered without changing Go source files:

- Existing retail JSON-round-trip profiler: `sampledata/gopdflib/zerodha/pypdfsuit_go_profile.go`.
- Native render CPU profile: `guides/cursor/baselines/pypdfsuit_pprof_runs/pypdfsuit_cgo_render_weighted_cpu_20260619.prof`.
- Native render heap profile: `guides/cursor/baselines/pypdfsuit_pprof_runs/pypdfsuit_cgo_render_weighted_heap_20260619.prof`.
- Text summary: `guides/cursor/baselines/pypdfsuit_pprof_runs/pypdfsuit_cgo_render_weighted_pprof_summary_20260619.txt`.

Retail JSON-round-trip split:

| Phase | Mean |
|-------|-----:|
| Native `GeneratePDF` | 1.222 ms |
| `json.Unmarshal` | 0.102 ms |
| Generate after JSON unmarshal | 1.156 ms |
| Combined JSON round-trip | 1.259 ms |

Native weighted render profile with CPU/heap profiling enabled:

| Metric | Value |
|--------|------:|
| Iterations | 5,000 |
| Workers | 48 |
| Throughput | 7,160.62 ops/sec |
| Avg latency | 6.572 ms |
| Max latency | 133.039 ms |
| Workload | 4,000 retail / 750 active / 250 HFT |

Important caveats:

- This **7,160.62 ops/sec** number is a profiled run, not the native Go baseline. The current published Go Zerodha artifact in `guides/BENCHMARKS.md` reports **11,721 ops/sec** best for `bench-gopdflib-zerodha`, and the raw x10 WSL artifacts are generally around **10,829-11,775 ops/sec**.
- This native Go render benchmark is not byte-for-byte identical to pypdfsuit HFT. The native HFT warm-up output was **748,163 bytes**, while pypdfsuit HFT is **2,424,782 bytes**.
- Therefore this profile is useful for render hotspot families, but it is not sufficient as the exact pypdfsuit FFI bottleneck proof. A true pypdfsuit JSON-round-trip pprof harness is still needed before making render-side optimization claims for pypdfsuit HFT.

CPU pprof hotspots:

| Hotspot | CPU evidence | Interpretation |
|---------|-------------:|----------------|
| `runtime.memclrNoHeapPointers` | 11.08% flat | zeroing large buffers/objects |
| `runtime.memmove` | 9.10% flat | copy-heavy PDF assembly path |
| `generateAllContentWithImages` | 20.67% cumulative | content stream generation |
| `drawTable` | 17.80% cumulative | table layout/render path remains central |
| `GenerateGrayICCProfileObject` | 12.76% cumulative | output intent / ICC object generation |
| `runtime.mallocgc` | 12.36% cumulative | allocation pressure |
| `signature.UpdatePDFWithSignatureBuffer` | 8.41% cumulative | retail signing cost |
| `bytes.(*Buffer).grow` / `bytes.growSlice` | 8.11% / 7.91% cumulative | buffer growth and copy pressure |
| `signature.embedSignatureInPlace` | 7.52% cumulative | signing update path |
| `formatStructElemObjectTo` | 7.32% cumulative | PDF/UA structure tree formatting |
| `compress/zlib` / `compress/flate` close/write path | about 6.73% cumulative | stream compression cost |
| `drawSharedLayoutRow` | 5.84% cumulative | HFT/shared-row table path |

Heap alloc-space hotspots:

| Hotspot | Allocation evidence | Interpretation |
|---------|--------------------:|----------------|
| `bytes.growSlice` | 334.56 MB, 34.69% | dominant allocation family |
| `drawTable` | 253.07 MB, 26.24% cumulative | table rendering allocates heavily |
| `compress/flate.NewWriter` | 128.39 MB, 13.31% cumulative | per-stream compression writer setup |
| `GenerateGrayICCProfileObject` | 110.64 MB, 11.47% cumulative | ICC/profile object allocation |
| `GenerateTemplatePDFBorrowed.func7` | 77.42 MB, 8.03% cumulative | final object/output assembly path |
| `formatStructElemObjectTo` | 41.01 MB, 4.25% cumulative | structure tree formatting |
| `signature.(*PDFSigner).createPKCS7SignedData` | 40.04 MB, 4.15% cumulative | signing allocation |

- [x] **Confirm Python cannot split `cgo_call` internally.**
  - Result: Python `cProfile` stops at `_bindings.call_bytes_result`; the Go shared library must be profiled separately for stack-level attribution.

- [x] **Run the existing retail Go JSON-round-trip profiler.**
  - Command: `GOCACHE=/tmp/gopdfsuit-go-build-cache GOMODCACHE=/tmp/gopdfsuit-go-mod-cache go run pypdfsuit_go_profile.go`.
  - Result: retail JSON decode is only **0.102 ms**; retail cost is mostly render/sign/output work, not JSON unmarshal.

- [x] **Run render-side CPU and heap pprof without editing Go source files.**
  - Built a profiling binary under `guides/cursor/baselines/pypdfsuit_pprof_runs/zerodha_render_profile_bin`.
  - Ran 5,000 weighted native render operations with CPU and heap profile output.
  - Result: render-side evidence points to buffer growth/copying, table rendering, ICC/output intent generation, signing, compression writer setup, and structure-tree formatting.
  - Caveat: this is a profiled native-Go run. It does not replace the published **10k+ ops/sec** Go benchmark baseline and does not exactly reproduce pypdfsuit HFT bytes.

- [ ] **Add a true pypdfsuit HFT JSON-round-trip pprof harness if Go-file edits become allowed.**
  - Goal: feed the Python-generated compact UTF-8 HFT JSON into the same Go export sequence and collect CPU/heap pprof.
  - Why: current native render pprof has useful hotspot families, but its HFT output is smaller than the pypdfsuit HFT output.
  - Acceptance: split `C.GoString`, `json.Unmarshal`, `gopdflib.GeneratePDF`, C allocation, and C copy-back for retail, active, and HFT payloads.

- [ ] **Prioritize render-side candidates by evidence, not by hunch.**
  - P5-A: reduce `bytes.growSlice` / `bytes.Buffer.grow` pressure with more accurate pre-sizing for final PDF/object buffers.
  - P5-B: reduce repeated `GenerateGrayICCProfileObject` work or make output-intent object generation cheaper when PDF/A settings are identical.
  - P5-C: pool or reuse compression writers only if it preserves correctness and does not retain large HFT buffers.
  - P5-D: reduce structure-tree formatting allocations in `formatStructElemObjectTo` and MCID table-cell formatting.
  - P5-E: keep signing improvements separate from HFT table improvements; retail is signing-sensitive, HFT is table/buffer-sensitive.

---

## Validation Matrix

| Change class | Required commands | Required evidence |
|--------------|-------------------|-------------------|
| Python serializer optimization | `cd bindings/python && python3 -m pytest tests`; `make bench-pypdfsuit-profile`; `make bench-pypdfsuit-zerodha-x5` | Equivalent JSON/PDF outputs, lower `to_dict` cumulative time, better no-cache mean |
| Cache removal | `make bench-help`; `rg BENCH_USE_JSON_CACHE makefile sampledata/gopdflib/zerodha bindings/python` | No cache targets or JSON-cache controls |
| Go boundary optimization | pypdfsuit profile; Go JSON/profile harness if Go work is allowed later | Clear split of Python serialization vs FFI/render time |
| HFT tail improvement | `make bench-pypdfsuit-profile`; `make bench-pypdfsuit-zerodha-x5` | Lower HFT total ms, lower p95/p99/max in x5 runs |
| P3 benchmark stability | `make bench-pypdfsuit-zerodha-x10`; pypdfsuit profile | x10 mean/median/band plus full-path hotspots |
| P4 exact make refresh | `make bench-pypdfsuit-zerodha` repeated 5 times; pypdfsuit profile | exact command band plus latest full-path phase profile |

---

## Next Recommended Implementation Order

1. Add p50/p95/p99 latency to `pypdfsuit_bench.py`.
2. Precompute dataclass JSON field mappings and remove `_python_to_json_key` from per-field hot loops.
3. Add specialized serializers for `Cell`, `Row`, `Table`, `Element`, and `PDFTemplate`.
4. Keep profiling on full execution only; no JSON-cache controls.
5. Keep compact UTF-8 JSON enabled because it preserves full execution and reduces payload bytes.
6. Do not count workload preselection or counter-lock removal as renderer speedups unless reported as harness cleanup.
7. Add Go-side pypdfsuit JSON/render phase profiling for HFT only if Go work becomes allowed.
8. Treat pre-parsed handles, batch APIs, service mode, and schema-compressed HFT tables as separate production APIs, not `make bench-pypdfsuit-zerodha` baselines.
9. Keep future P4+ measurement centered on exact `make bench-pypdfsuit-zerodha` and its profile artifacts.

The highest-confidence Python-side production win has already landed: specialized serializers plus compact UTF-8 JSON. Remaining meaningful gains are mostly beyond Python serialization: either FFI/render-boundary profiling, a new API contract, or an explicitly labeled benchmark-harness cleanup.
