# Optimization Execution Summary — 2026-06-19

## Date & Scope

Honest full-execution PyPDFSuit benchmark work on `feat/optimization-5.5-medium`: remove JSON-cache benchmark tuning, restore per-call `to_dict` + `json.dumps` semantics, and ship Python-side serializer wins (precomputed field keys, direct dataclass serializers, compact UTF-8 JSON) on the Zerodha weighted mix (80% retail / 15% active / 5% HFT).

## Key Outcomes

- **Baseline (pre-pass, no-cache x5):** **221.40 ops/s** mean (205–229 band); HFT tail drives p99 above 3 s.
- **P0 serializers + cache removal:** **565.16 ops/s** single run; HFT `to_dict` dropped **60.34 → 19.74 ms** per call.
- **P2 direct row/cell serializers:** **736.55 ops/s** single run; Python serialization no longer dominates cProfile.
- **P3 x10 honest band:** **750.78 ops/s** mean, best **927.19 ops/s** (640–930 operating range).
- **P4 compact UTF-8 JSON (canonical x5):** **835.56 ops/s** mean, best **855.55 ops/s** — publishable no-cache number.
- **Bottleneck shift:** FFI/render (`cgo_call`) now **73–92%** of per-template latency; further Python-only gains are incremental.

## Work Completed

- Precomputed dataclass field-to-JSON-key mappings; removed `_python_to_json_key` hot-path cost.
- Implemented specialized serializers for `PDFTemplate`, `Table`, `Row`, `Cell`, `Config`, and related hot types.
- Removed automatic JSON payload caching, `use_cache`, and cache benchmark Makefile targets.
- Added `PAYLOAD_SCENARIO` and p50/p95/p99 latency reporting to `pypdfsuit_bench.py`.
- Switched `serialize_template()` to compact JSON separators and UTF-8 rupee symbols (HFT JSON **~1.055 → 1.043 MB**).
- Added `bindings/python/tests/test_serializer_schema.py` for Go-facing JSON key parity.

## Findings / Bottlenecks

| Bucket | Post-P4 share | Notes |
|--------|---------------|-------|
| `_bindings.call_bytes_result` (CGO + Go) | **~92%** retail, **~82%** active, **~73%** HFT | Hard ceiling for current API |
| `to_dict` + `json.dumps` | **6–26%** by template | Down from **50–67%** pre-pass |
| `ctypes.string_at` copy-back | <2% | Negligible except HFT tail |

- The prior **~1,505 ops/s** cache-only figure was removed from the active benchmark surface as harness tuning, not execution throughput.
- **3,000 ops/s** on the full weighted HFT mix without Go changes is not supported by profile evidence.

## Open Items / Next Steps

| Priority | Item | Est. impact | Status |
|----------|------|-------------|--------|
| **P1** | HFT row-block or handle-based signature references | Lower HFT tail | Planned (must stay honest) |
| **P2** | Go `GeneratePDFFromTemplateHandle` / skip `json.Unmarshal` | **+10–30%** post-serializer | Planned |
| **P3** | Go render-boundary work for HFT `cgo_call` | Major step-change | Requires `.go` changes |
| **P4** | HTTP-first production path (Gin k6) | Match native scale | Architecture path |

## Source Documents

| File | Role |
|------|------|
| `20260619_pypdfsuit_benchmark_optimization_checklist.md` | Primary checklist |
| `guides/cursor/baselines/pypdfsuit_pprof_runs/` | x5, x10, and profile artifacts |
| `bindings/python/pypdfsuit/generator.py` | Serializer and cache-removal changes |
| `bindings/python/tests/test_serializer_schema.py` | Schema parity tests |