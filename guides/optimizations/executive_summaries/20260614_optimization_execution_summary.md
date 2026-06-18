# Optimization Execution Summary — 2026-06-14

## Date & Scope

Multi-pass performance work on gopdfsuit's weighted `tagged_ecdsa` benchmark (`make bench-k6`: 80% retail / 15% active / 5% HFT, 48 VUs × 35s), targeting **~1,500+ req/s** while preserving PDF/A-4, PDF/UA-2, signing, and unique-document compliance.

## Key Outcomes

- **Fresh 2-run mean (latest):** **1,025.7 req/s** (runs: 998.7, 1,052.7) — up from **916.8 req/s** pre-pass (+12%) and **825 req/s** prior-session 5-run avg (+24%)
- **Best fresh run:** **1,052.7 req/s**, p99 **236.9 ms**, HFT avg **323.4 ms**, **0% errors**
- **Historical 5-run mean (committed HI-2 + MI-3 + MI-1):** **1,005.6 req/s**, p99 **269.6 ms**
- **Gap to targets:** ~87% of post-revert baseline (1,054), ~74% of Phase 11 peak (1,232), ~61% of **1,500 req/s** goal
- **Run stability improved:** spread tightened from **244.5 req/s** (794–1,039) to **~54 req/s** (998–1,053) after structure preallocation
- **Three experiments reverted** due to regression (HI-3 caps, generic structure writer, EW-2 Load-before-Store)

## Work Completed

**Shipped / committed:**
- **HI-2** — Bounded unbounded caches: `subsetCache` (1,024), `imgCache` (256), `propsCache` (8,192)
- **MI-3** — `strconv.AppendInt` + stack scratch buffers in structure-tree writer
- **MI-1 (partial)** — `appendEscapedPDFLiteral` + `Write([]byte)` for Title/Alt

**Additional same-session passes:**
- **HI-1/HI-3 support** — Stable shard selection for `CompressContentStreamCached`; removed per-call `hash/fnv` alloc
- **HI-4** — HFT table row preallocation; pooled `PDFTemplate` reset retains hot HFT arrays across `sync.Pool` reuse
- **Structure follow-up** — Pre-sized tagged structure-element backing storage from template shape

## Findings / Bottlenecks

| Hotspot | Profile share |
|---|---|
| `bytes.growSlice` | ~50–54% in-use heap, ~13% alloc_space (**296–325 MB**) |
| `compress/flate` / `CompressContentStreamCached` | ~20–21% cumulative CPU |
| `drawTable` / shared-layout row paths | ~15–16% cumulative CPU |
| `sonic` + JSON decode | ~16–18% alloc_space |
| `formatStructElemObjectTo` | ~9.8–10.5% CPU |
| `preallocInlineTableRows` | ~8–9% alloc_space |

Cache bounds and structure-writer micro-opts improved stability, but **compression, heap growth, JSON decode, and HFT row allocation** remain the primary ceiling.

## Open Items / Next Steps

**High impact:**
- **HI-1** — Flate compression tuning — est. **+5–8%**
- **HI-3** — Smarter buffer pre-sizing / page content buffer pooling — est. **+5–7%**, target `bytes.growSlice` **< 200 MB**
- **HI-4** — Codegen sonic AST decoder for `PDFTemplate` — est. **+5–10%** E2E on HFT

**Medium impact:**
- **MI-1** — Pool `StructKid` slices, concrete buffer path — est. **+3–5%**
- **MI-2** — HFT draw path (text-width cache, batched MCIDs) — est. **+3–5%**

**Acceptance criteria still open:** stable 5-run avg **~1,500+ req/s**, p99 **< 500 ms**, `bytes.growSlice` **< 200 MB**.

## Source Documents

- `20260614_remaining_optimizations_checklist.md` (primary)
- Referenced pprof summaries under `guides/cursor/baselines/gin_pprof_runs/`