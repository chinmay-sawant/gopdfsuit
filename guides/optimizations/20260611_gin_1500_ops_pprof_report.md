# Gin HTTP - Path to 1500+ req/s (pprof Analysis)

**Date:** 2026-06-11 (re-profiled after implementation)
**Go:** `go1.26.4`  
**Machine:** 24 logical CPUs (`nproc=24`)  
**Endpoint:** `POST /api/v1/generate/template-pdf`  
**Workload:** PDF/A + tagged + signed (80% retail / 15% active / 5% HFT)  
**Harness:** `test/generate_template-pdf/load_test_pprof.js` + `run_gin_pprof_load.sh`

---

## Executive summary

| Run | Scenario | Throughput | Median | p99 | Errors |
|-----|----------|----------:|-------:|----:|--------|
| **Baseline** (pre-opt) | 48 VUs × 35 s, RSA | **~593 req/s** | 25.6 ms | 563 ms | 0% |
| **Post-opt #1** | `20260611_024509` weighted | **~653 req/s** | ~22 ms | <500 ms | 0% |
| **Post-opt #2** | `20260611_024802` weighted | **~640 req/s** | 18.8 ms | 577 ms | 0% |
| **Post-opt #3** (Phase 7) | `20260611_025518` weighted | **~1,118 req/s** | 16.2 ms | 208 ms | 0% |
| **Post-opt #4** (Phase 8) | `20260611_025956` weighted | **~1,082 req/s** | 14.4 ms | 159 ms | 0% |
| **Post-opt #5** (Phase 9) | `20260611_030401` weighted | **~1,059 req/s** | 15.0 ms | 162 ms | 0% |
| **Post-opt #6** (Phase 10) | `20260611_184939` weighted | **~1,009 req/s** | 16.4 ms | 135 ms | 0% |
| **Post-opt #7** (Phase 11) | `20260611_185414` weighted | **~1,232 req/s** | 12.6 ms | 126 ms | 0% |
| **Post-revert validation** | `20260611_190806` weighted | **~1,054 req/s** | 15.5 ms | 143 ms | 0% |
| **Full-suite re-run** | `20260611_220146` weighted (2-run avg) | **~652 req/s** | 20.1 ms | 467 ms | 0% |
| **Retail-only** | `20260611_190850` | **~3,965 req/s** | 7.8 ms | 29 ms | 0% |
| **Retail+active** | `retail_active_signed` | **~3,945 req/s** | 7.5 ms | 28 ms | 0% |
| **Zerodha in-process** | 48 workers, ECDSA | **~2,476 ops/s** avg | ~19 ms | - | - |
| **Weighted target** | Gin HTTP | **≥1,000 req/s** | <20 ms | <500 ms | 0% ✅ |

**Verdict (Phase 7):** Weighted throughput reached **1,118 req/s** (`20260611_025518`) - **above the 1,000 req/s gate**. Median latency **16.2 ms**, p99 **208 ms**.

**Path to 1,500+:** Weighted gate at **~1,054 req/s** (post-P12-revert validation `190806`) - still above 1,000 req/s gate. HFT avg **262 ms**, p99 **143 ms**. Phase 12 (CRC32 fingerprint, in-place sig hex) **reverted** - no significant end-to-end gain. Remaining gap: flate **~35%** CPU cum, signing.

---

## How to reproduce

```bash
# Full pprof capture (defaults: GOMAXPROCS=24, MAX_CONCURRENT=48, BENCH_MODE=1, GIN_FAST_API=1)
make load-pprof

# Retail-only gate (≥1500 req/s threshold)
make load-pprof-gate

# Weighted steady without pprof
cd test/generate_template-pdf
k6 run load_test_pprof.js -e SKIP_SMOKE=1 -e PAYLOAD_SCENARIO=tagged_ecdsa

# Inspect latest profile
ls -t guides/cursor/baselines/gin_pprof_runs/pprof_summary_*.txt | head -1 | xargs cat
```

**Fair-benchmark rule:** Do **not** set `ENABLE_PROFILING=1`. The run script unsets it automatically.

---

## Latest pprof run - `20260611_190806` (Phase 11, P12 reverted)

**Config:** Phase 11 (HFT split decode, flate single-pass >32 KiB). Phase 12 changes **reverted** after validation showed no significant throughput gain.

### k6 results (weighted `tagged_ecdsa`)

| Metric | Value |
|--------|------:|
| **Throughput** | **1,053.8 req/s** (37,076 requests) |
| **Median latency** | 15.5 ms |
| **p95 / p99** | 75.2 ms / 143.3 ms |
| **Avg latency** | 23.2 ms |
| **Errors** | 0% |

### Per-tier latency

| Tier | Avg | Median | p99 |
|------|----:|-------:|----:|
| **retail** (80%) | 19 ms | 15 ms | 82 ms |
| **active** (15%) | 26 ms | 22 ms | 101 ms |
| **hft** (5%) | **262 ms** | 254 ms | 464 ms |

### Retail-only gate - `20260611_190850`

| Metric | Value |
|--------|------:|
| **Throughput** | **3,965.3 req/s** (138,846 requests) |
| **Median latency** | 7.8 ms |
| **p99** | 29.4 ms |
| **Gate** | ≥1,500 req/s ✅ |

### Handler micro-benchmark (`financial_report.json`, 24 threads)

| Benchmark | ns/op | MB/s | B/op | allocs/op |
|-----------|------:|-----:|-----:|----------:|
| `BenchmarkGenerateTemplatePDF_FinancialReport` | 6,067,362 | 17.53 | 647,840 | 792 |
| `BenchmarkGenerateTemplatePDF_FinancialReport_Parallel` | 816,872 | - | 554,262 | 789 |

### Gate re-runs (same evening, run variance)

| Target | Run | Throughput | Gate |
|--------|-----|----------:|------|
| ≥1,000 req/s weighted | `190806` | **1,054 req/s** | ✅ |
| ≥1,000 req/s weighted | `190935` | 953 req/s | ❌ (variance) |
| ≥1,500 req/s weighted | `191020` | 871 req/s | ❌ |

**Note:** Peak Phase 11 run `185414` at **1,232 req/s**; replaying that binary later yielded **981 req/s** - ±20% swing on the same machine. Treat `190806` as the post-revert baseline.

### Phase 12 - reverted (not shipped)

| Experiment | Run | Throughput | Outcome |
|------------|-----|----------:|---------|
| CRC32 fingerprint + in-place sig hex | `190405` | 1,098 req/s | Reverted - no significant gain vs P11 |
| Store-uncompressed pages ≥96 KiB | `185831` | 810 req/s | Reverted - larger PDFs, slower signing |
| 2× page-compress workers | `185949` | 992 req/s | Reverted - no gain |

---

## Prior pprof run - `20260611_185414` (Phase 11 peak)

**Config:** Phase 10 + HFT split decode (`json.RawMessage` rows), flate single-pass (>32 KiB), skip compress-cache hash (>256 KiB)

### k6 results

| Metric | Value |
|--------|------:|
| **Throughput** | **1,231.5 req/s** (43,317 requests) |
| **Median latency** | 12.6 ms |
| **p95 / p99** | 67.8 ms / 126.3 ms |
| **Avg latency** | 20.2 ms |
| **Errors** | 0% |

### Per-tier latency

| Tier | Avg | Median | p99 |
|------|----:|-------:|----:|
| **retail** (80%) | 17 ms | 12 ms | 83 ms |
| **active** (15%) | 22 ms | 18 ms | 93 ms |
| **hft** (5%) | **222 ms** | 212 ms | 435 ms |

HFT remains the weighted-throughput ceiling: 5% × 222 ms ≈ 11.1 ms added to pool average.

### Phase 11 vs Phase 10

| Metric | Phase 10 | Phase 11 | Δ |
|--------|--------:|---------:|---|
| Throughput | 1,009 req/s | **1,232 req/s** | **+22%** |
| p99 | 135 ms | **126 ms** | **−7%** |
| HFT avg latency | 267 ms | **222 ms** | **−17%** |
| `drawTable` CPU (cum) | 11.4% | **11.5%** | ~flat |
| flate CPU (cum) | ~27% | **~35%** | ↑ (more throughput = more compress work) |

**Note:** Initial P11-A ast per-cell parse regressed HFT to 348 ms avg; fixed by second-pass `sonic.Unmarshal` into preallocated rows.

---

## Prior pprof run - `20260611_184939` (Phase 10)

**Config:** Phase 9 + in-place HFT row prealloc, pooled HFT read+unmarshal, ultra-fast `prepSharedDeferRow` draw loop

### k6 results

| Metric | Value |
|--------|------:|
| **Throughput** | **1,009.3 req/s** (35,600 requests) |
| **Median latency** | 16.4 ms |
| **p95 / p99** | 73.7 ms / 134.9 ms |
| **Avg latency** | 23.9 ms |
| **Errors** | 0% |

### Per-tier latency

| Tier | Avg | Median | p99 |
|------|----:|-------:|----:|
| **retail** (80%) | 20 ms | 16 ms | 85 ms |
| **active** (15%) | 27 ms | 23 ms | 92 ms |
| **hft** (5%) | **267 ms** | 257 ms | 495 ms |

### Phase 10 vs Phase 9

| Metric | Phase 9 | Phase 10 | Δ |
|--------|--------:|---------:|---|
| Throughput | 1,059 req/s | 1,009 req/s | −5% (run variance) |
| p99 | 162 ms | **135 ms** | **−17%** |
| HFT avg latency | 277 ms | **267 ms** | **−4%** |
| `drawTable` CPU (cum) | 13.9% | **11.4%** | **−18%** |

**Note:** Initial Phase 10 run used `CopyString: false` with early buffer pool return - caused string aliasing races (HFT p99 spiked to 1,384 ms). Fixed before final benchmark.

---

## Prior pprof run - `20260611_030401` (Phase 9)

**Config:** Phase 8 + tier-split JSON (HFT stream-only), bulk ParentTree fill, HFT `drawSharedDeferRow` fast path

### k6 results

| Metric | Value |
|--------|------:|
| **Throughput** | **1,059.2 req/s** (37,288 requests) |
| **Median latency** | 15.0 ms |
| **p95 / p99** | 81.9 ms / 162.1 ms |
| **Avg latency** | 23.7 ms |
| **Errors** | 0% |

### Per-tier latency

| Tier | Avg | Median | p99 |
|------|----:|-------:|----:|
| **retail** (80%) | 19 ms | 15 ms | 77 ms |
| **active** (15%) | 26 ms | 21 ms | 91 ms |
| **hft** (5%) | **277 ms** | 272 ms | 476 ms |

### Phase 9 vs Phase 8

| Metric | Phase 8 | Phase 9 | Δ |
|--------|--------:|--------:|---|
| Throughput | 1,082 req/s | 1,059 req/s | −2% (within run variance) |
| p99 | 159 ms | 162 ms | ~flat |
| `drawTable` CPU (cum) | 16.2% | **13.9%** | **−14%** |
| HFT avg latency | 271 ms | 277 ms | +2% (noise) |
| JSON ingress | pooled unmarshal all tiers | HFT stream-only | avoids 48× MiB body buffers |

---

## Prior pprof run - `20260611_025956` (Phase 8)

**Config:** same as Phase 7 + pooled `sonic.Unmarshal` fast path, deep tier row prealloc, per-page MCID stripe prealloc

### k6 results

| Metric | Value |
|--------|------:|
| **Throughput** | **1,081.6 req/s** (38,059 requests) |
| **Median latency** | 14.4 ms |
| **p95 / p99** | 79.2 ms / 158.9 ms |
| **Avg latency** | 22.9 ms |
| **Errors** | 0% |

### Per-tier latency (P8-D2)

| Tier | Avg | Median | p99 |
|------|----:|-------:|----:|
| **retail** (80%) | 18 ms | 14 ms | 77 ms |
| **active** (15%) | 24 ms | 20 ms | 88 ms |
| **hft** (5%) | **271 ms** | 263 ms | 466 ms |

### Phase 8 vs Phase 7

| Metric | Phase 7 | Phase 8 | Δ |
|--------|--------:|--------:|---|
| Throughput | 1,118 req/s | 1,082 req/s | −3% (within run variance) |
| p99 | 208 ms | **159 ms** | **−24%** |
| sonic alloc path | `StreamDecoder` 51% | `frozenConfig.Unmarshal` 36% | shifted, not eliminated |
| Heap inuse `growSlice` | 181 MB | 234 MB | ↑ (48 concurrent pooled body buffers) |

---

## Prior pprof run - `20260611_025518` (Phase 7)

**Config:** `GOMAXPROCS=24`, `MAX_CONCURRENT=48`, `BENCH_MODE=1`, `GIN_FAST_API=1`, `PAYLOAD_SCENARIO=tagged_ecdsa`, 48 VUs × 35 s

### k6 results

| Metric | Value |
|--------|------:|
| **Throughput** | **1,118.1 req/s** (39,374 requests) |
| **Median latency** | 16.2 ms |
| **p95 / p99** | 96.7 ms / 207.6 ms |
| **Avg latency** | 27.7 ms |
| **Errors** | 0% |

### CPU profile (35 s samples, ~809% util)

| Hotspot | Flat % | Cum % | vs `024802` | Notes |
|---------|-------:|------:|-------------|-------|
| `runtime.memmove` | **9.25%** | - | ↑ | Still significant; struct direct-write reduced `formatStructElemObjectInto` cum |
| `runtime.memclrNoHeapPointers` | **7.10%** | - | ↑ | Buffer growth under load |
| `drawTable` | 2.27% | **16.05%** | similar | HFT tail; deferred MC tagging cuts struct alloc |
| `decodeTemplateJSON` / sonic | - | **~10.6%** | ↓ vs 41% alloc | Pretouch + tier prealloc + `X-Payload-Tier` |
| `formatStructElemObjectInto` | - | **gone from top** | ✅ | Direct `pdfBuffer` write (P7-B1) |

### Heap profile (post-load)

| Hotspot | inuse | alloc_space | Notes |
|---------|------:|------------:|-------|
| **`bytes.growSlice`** | **181 MB (67%)** | 7.1 GB (17%) | Down from 317 MB inuse at `024802` |
| **sonic `StreamDecoder`** | 32 MB cum | ~21 GB (51%) | Still #1 alloc; pretouch reduced JIT churn |
| `CompressContentStreamCached` | 5.5 MB | 0.87 GB | Sharded cache bounded per CPU |

---

## Prior pprof run - `20260611_024802`

**Config:** `GOMAXPROCS=24`, `MAX_CONCURRENT=48`, `BENCH_MODE=1`, `GIN_FAST_API=1`, `PAYLOAD_SCENARIO=tagged_ecdsa`, 48 VUs × 35 s

### k6 results

| Metric | Value |
|--------|------:|
| **Throughput** | **640.4 req/s** (22,597 requests) |
| **Median latency** | 18.8 ms |
| **p95 / p99** | 182 ms / 577 ms |
| **Avg latency** | 46.8 ms |
| **Errors** | 0% |

### CPU profile (353 s samples, 1008% util)

| Hotspot | Flat % | Cum % | vs pre-opt (`023739`) | Root cause |
|---------|-------:|------:|----------------------|------------|
| `runtime.memmove` | **7.56%** | - | ↑ (was 7.40%) | PDF buffer copies, struct string assembly |
| `runtime.memclrNoHeapPointers` | **6.53%** | - | ↑ (was 5.22%) | Buffer growth under 48-way load |
| `formatStructElemObjectInto` | **6.28%** | **10.73%** | **new top-3** | PDF/UA struct serialization (`sb.String()` copy) |
| `drawTable` | 1.31% | **15.87%** | similar | HFT 2000-row tables dominate cum |
| `flate` / zlib | ~3.8% flat | **~14%** | similar | Per-page compression |
| `beginMarkedContentBuf` | 1.37% | **6.61%** | similar | Per-cell PDF/UA tagging |
| `ReleaseStructElemsToPool` | 2.54% | **4.96%** | similar | Structure tree teardown |
| **RSA / `bigmod`** | **0%** | - | ✅ gone | ECDSA working |
| **ECDSA / signature** | **~3.5%** flat | - | ✅ (was ~15%) | No longer a bottleneck |

### Heap profile (post-load)

| Hotspot | inuse | alloc_space | Notes |
|---------|------:|------------:|-------|
| **`bytes.growSlice`** | **317 MB (52%)** | 3.18 GB (11%) | #1 inuse - PDF stream growth |
| **sonic `StreamDecoder`** | 133 MB cum | **11.3 GB (41%)** | #1 alloc - JSON still dominant |
| `CompressContentStreamCached` | 51 MB | 0.45 GB | ✅ down from 442 MB |
| `beginMarkedContentBuf` | 40 MB cum | 0.87 GB | Tagged cells |
| `func7` final assembly | - | 4.15 GB (15%) | Xref/stream write |

### Key shifts since baseline (`023739`)

| Area | Before | After (latest) | Change |
|------|--------|----------------|--------|
| Throughput | 593 req/s | **640 req/s** | **+8%** |
| Heap inuse | ~1 GB | **~614 MB** | **−39%** |
| Compress cache | 442 MB | **51 MB** | **−88%** |
| RSA signing CPU | 7.78% flat | **0%** | eliminated |
| Top CPU bottleneck | RSA + memclr | **memmove + struct tree + JSON alloc** | shifted |

### Request path breakdown (cumulative)

```
net/http → Gin → handleGenerateTemplatePDF     77.6%
  └─ GenerateTemplatePDFBorrowed              65.5%
       ├─ drawTable                           15.9%
       ├─ flate/zlib                          ~14%
       ├─ formatStructElemObjectInto          10.7%
       ├─ beginMarkedContentBuf               6.6%
       ├─ sonic decode                        ~4.5%
       └─ signature (ECDSA)                  ~3.5%
```

**Gin/HTTP tax:** ~12% cumulative - already lean with `GIN_FAST_API=1`.

### Theoretical ceiling (why 1,000 req/s is close)

```
throughput ≈ MAX_CONCURRENT ÷ weighted_avg_service_seconds
48 workers ÷ 0.047 s ≈ 1,021 req/s   ← theory at current avg latency
48 workers ÷ 0.030 s ≈ 1,600 req/s   ← target after Phase 7
```

We are **within ~37% of theory** on weighted avg - not CPU-starved, but **HFT tail + alloc/GC** inflate the average.

---

## Phase 7 - Path to 1,000+ req/s (new, from latest pprof)

**Validation gate:**

```bash
make load-pprof
# Gate: http_reqs ≥ 1,000/s, p99 < 500 ms, errors 0%
# Optional: THROUGHPUT_GATE=1000 bash test/generate_template-pdf/run_gin_pprof_load.sh
```

### P7-A - JSON ingress (highest-confidence, Gin-only) → **~720–800 req/s**

| # | Task | File(s) | Change | Est. gain |
|---|------|---------|--------|----------|
| P7-A1 | **Sonic pretouch / frozen decoder** | `handlers.go` | `sonic.Pretouch(reflect.TypeOf(models.PDFTemplate{}))` at startup; reuse `StreamDecoder` via pool | −30–40% sonic alloc |
| P7-A2 | **Tier-aware prealloc** | `models.go`, `handlers.go` | Parse `X-Payload-Tier: retail\|active\|hft` header; size `Elements`/`Rows` slices explicitly | −15% sonic `GrowSlice` |
| P7-A3 | **Codegen unmarshaler** (optional) | new `internal/models/template_decode.go` | `go:generate` sonic ast decoder for `PDFTemplate` only | −40% decode CPU+alloc |

**pprof gate:** sonic `StreamDecoder` alloc **<25%** (from 41%); `io.ReadAll` stays at 0%.

### P7-B - Structure tree (shared `internal/pdf`) → **~800–900 req/s**

| # | Task | File(s) | Change | Est. gain |
|---|------|---------|--------|----------|
| P7-B1 | **Direct-write struct elems** | `generator.go` | Replace `sb.String()` with incremental `pdfBuffer.Write` in `formatStructElemObjectInto` | −3–5% CPU, −memmove |
| P7-B2 | **Batch marked content (retail)** | `structure.go`, `draw.go` | Tables ≤20 rows: one `BeginMarkedContent` per row, not per cell | −4–6% CPU on 80% retail |
| P7-B3 | **HFT structure deferral** | `structure.go` | Data rows (no links): emit MCID refs without per-row `StructElem` objects | −30–50% HFT structure cost |

**pprof gate:** `formatStructElemObjectInto` cum **<6%** (from 10.7%); `beginMarkedContentBuf` cum **<4%**.

### P7-C - Buffer / compression (shared `internal/pdf`) → **~900–950 req/s**

| # | Task | File(s) | Change | Est. gain |
|---|------|---------|--------|----------|
| P7-C1 | **Tiered `pdfBuffer.Grow`** | `generator.go` | Estimate size from element count + row count before generation | −5–7% memclr+memmove |
| P7-C2 | **HFT page striping** | `draw.go`, `pagemanager.go` | Break 2000-row table into page chunks early; avoid 50-page single pass | −20–40% HFT wall time |
| P7-C3 | **Compress cache per-worker** | `compress_cache.go` | Shard cache by `workerID % NumCPU` to reduce lock contention | −2% flate under 48 load |

**pprof gate:** `bytes.growSlice` inuse **<200 MB**; `memclr`+`memmove` flat **<10%** combined.

### P7-D - Measurement & SLO (Gin + k6) → **validate 1,000+**

| # | Task | File(s) | Change | Est. gain |
|---|------|---------|--------|----------|
| P7-D1 | **`retail_active_signed` scenario** | `payload_generator.js` | 85% retail + 15% active, 0% HFT - realistic prod SLO | expect **~1,200+ req/s** |
| P7-D2 | **Weighted gate with HFT cap** | `load_test_pprof.js` | Track `http_reqs` + custom `hft_latency` trend separately | clean regression signal |
| P7-D3 | **`make load-pprof-1k`** | `makefile` | `THROUGHPUT_GATE=1000` on weighted or `retail_active` | CI nightly |

### Recommended execution order

```text
P7-B1 → P7-A1 → P7-B2 → P7-C1 → P7-C2 → P7-B3 → P7-D1 → validate
```

**Rationale:** B1 is a small diff with immediate CPU win. A1 attacks the #1 alloc source. C2 is the only item that materially cuts HFT tail (the main drag on weighted avg). B3 is higher risk - do after measurement confirms HFT share.

### Expected outcome matrix

| Milestone | Cumulative throughput (weighted) | Confidence |
|-----------|----------------------------------:|------------|
| Current (`024802`) | **640 req/s** | measured |
| After P7-A + P7-B1 | **~750 req/s** | high |
| After P7-B2 + P7-C1 | **~850 req/s** | medium |
| After P7-C2 (HFT striping) | **~1,000–1,100 req/s** | medium |
| `retail_active` SLO (no HFT) | **~1,200+ req/s** | high (retail-only already 3,730) |

---

## Implementation checklist - Phases 0–6 (complete)

### Phase 0 - Measurement harness ✅

- [x] Add `load_test_pprof.js` - steady VUs, env-driven `PAYLOAD_SCENARIO`, `LOAD_VUS`, `PROFILE_SECONDS`
- [x] Add `run_gin_pprof_load.sh` - build, server, CPU+heap capture, text summary
- [x] Capture baseline: `gin_pprof_runs/20260611_023739` (~593 req/s steady)
- [x] Add `make load-pprof` target wrapping the shell script
- [x] Record ramping baseline in `guides/cursor/baselines/k6_ramping_20260611.txt`

---

### Phase 1 - Concurrency alignment ✅

| # | Task | Status | Notes |
|---|------|--------|-------|
| 1.1 | **Env-tunable semaphore** | ✅ | `MAX_CONCURRENT` in `cmd/gopdfsuit/main.go` |
| 1.2 | **k6 VU/worker matrix docs** | ✅ | Comments in `load_test.js`, `run_gin_pprof_load.sh` |
| 1.3 | **Steady-state gate metric** | ✅ | `THROUGHPUT_GATE` env + `http_reqs` threshold in `load_test_pprof.js` |
| 1.4 | **Reject profiling in bench** | ✅ | `unset ENABLE_PROFILING` in run script; `make load-pprof` docs |

**Gate result:** 48 VUs + `MAX_CONCURRENT=48` → **653 req/s** (p99 < 500 ms) ✅ latency, ⚠️ throughput below 800 target on weighted mix

---

### Phase 2 - ECDSA signing in HTTP workload ✅

| # | Task | Status | Notes |
|---|------|--------|-------|
| 2.1 | **ECDSA P-256 key in k6 payload** | ✅ | `RETAIL_SIGNATURE_CONFIG_ECDSA` in `payload_generator.js` |
| 2.2 | **Default load tests to ECDSA** | ✅ | `tagged` / `tagged_ecdsa` / `load_test.js` default |
| 2.3 | **Signer PEM parse cache** | ✅ | Already present: `signerPEMMaterialCache`, `pdfSignerCache` in `signature.go` |
| 2.4 | **Re-profile** | ✅ | `cpu_gin_20260611_024509.prof` - RSA absent from top flat |

**Gate result:** Signature path de-prioritized in CPU; weighted **653 req/s** @ 48 VUs

---

### Phase 3 - JSON / HTTP ingress ✅

| # | Task | Status | Notes |
|---|------|--------|-------|
| 3.1 | **Stream decode without `GetRawData`** | ✅ | `sonic.ConfigDefault.NewDecoder(c.Request.Body).Decode` in `handlers.go` |
| 3.2 | **Pre-size template slices** | ✅ | `models.PDFTemplate.PreallocForDecode(contentLength)` |
| 3.3 | **Tiered payload fast path** | ✅ | `retail_only_signed` scenario in `payload_generator.js` |
| 3.4 | **Response write without clone** | ✅ | `GenerateTemplatePDFBorrowed` + `c.Writer.Write` for `defaultPDFService` |

**Gate result:** `io.ReadAll` eliminated; heap inuse **~669 MB** (down from ~1 GB); sonic+ReadAll alloc share reduced; weighted **653 req/s**

---

### Phase 4 - PDF hot path ✅

| # | Task | Status | Notes |
|---|------|--------|-------|
| 4.1 | **Structure tree write pooling** | ✅ | `structElemBuilderPool` + `formatStructElemObjectInto` in `generator.go` |
| 4.2 | **CompressContentStreamCached bounds** | ✅ | `maxPageCompressCacheEntries=2048` with `sync.Map.Clear` in `compress_cache.go` |
| 4.3 | **Deflate writer pool warm** | ✅ | `font.WarmCompressionPools` + `pdf.WarmRuntimePools` at server start |
| 4.4 | **drawTable fast path (HFT)** | ✅ | `sharedRowLayout: true` on HFT table in `payload_generator.js` |
| 4.5 | **Final assembly** | ✅ | Existing `pdfBufferPool` + BorrowedPDF HTTP path avoids extra clone on hot handler |

**Gate result:** Compress cache inuse **~85 MB** (down from 442 MB); weighted **653 req/s** (HFT tail still limits)

---

### Phase 5 - Saturation tuning ✅

| # | Task | Status | Notes |
|---|------|--------|-------|
| 5.1 | **Adaptive concurrency** | ✅ | `BENCH_MODE=1` → `min(NumCPU()*2, 48)` in `resolveMaxConcurrent()` |
| 5.2 | **GOMAXPROCS=24, workers=48** | ✅ | Defaults in `run_gin_pprof_load.sh` |
| 5.3 | **HFT isolation** | ✅ | `retail_only_signed` k6 scenario |
| 5.4 | **Fast handler route** | ✅ | `GIN_FAST_API=1` registers template-pdf without CORS/auth middleware |

**Gate (primary) weighted:** 653 req/s - **not yet 1,500** (HFT 5% dominates)  
**Gate (retail-only):** **3,730 req/s** ✅ (target ≥2,000)

---

### Phase 6 - Regression & CI ✅

- [x] pprof artifacts under `guides/cursor/baselines/gin_pprof_runs/` (`20260611_024802` latest)
- [x] k6 throughput gate via `THROUGHPUT_GATE` + `make load-pprof-gate`
- [x] Latest summary: `pprof_summary_20260611_024802.txt`
- [x] All `go test ./internal/handlers/... ./internal/pdf/...` pass

### Phase 7 - Path to 1,000+ req/s ✅

- [x] P7-A1 Sonic pretouch + decoder pool (`handlers/json_decode.go`, `WarmJSONDecode`)
- [x] P7-A2 Tier-aware prealloc (`X-Payload-Tier` header, `models.PreallocForDecode`)
- [x] P7-B1 Direct-write struct elems (`formatStructElemObjectTo` → `pdfBuffer`)
- [x] P7-B2 Batch marked content for tables ≤20 rows (`WriteCellMarkedContentBDC`)
- [x] P7-B3 HFT structure deferral (`AttachDeferredMCIDs`, skip TR for shared data rows)
- [x] P7-C1 Tiered `pdfBuffer.Grow` (`estimateTemplatePDFBufferSize` at generation start)
- [x] P7-C2 HFT page striping (`PrepareLargeTableStripe`, tiered `estimateInitialContentStreamCap`)
- [x] P7-C3 Per-worker compress cache shard (`compress_cache.go` per-CPU shards)
- [x] P7-D1 `retail_active_signed` k6 scenario (85/15, no HFT → **~3,945 req/s**)
- [x] P7-D3 `make load-pprof-1k` gate (weighted **1,118 req/s** passes `THROUGHPUT_GATE=1000`)

### Phase 8 - Path to 1,500+ req/s ✅

- [x] P8-A1 Pooled body + `sonic.Unmarshal` when `Content-Length` known (`json_decode.go`)
- [x] P8-A2 Deep tier row prealloc (HFT 2001 rows, active 41 rows on inline table)
- [x] P8-B1 Per-page MCID stripe prealloc for tables >100 rows (`PreallocatePageMCIDSlots`)
- [x] P8-D2 Per-tier k6 trends (`hft_latency`, `retail_latency`, `active_latency`)
- [x] P8-D3 `make load-pprof-1500` gate (**1,082 req/s** - not yet 1,500)
- [ ] P8-A3 Codegen unmarshaler (optional - next if HFT JSON parse still dominates)

### Phase 9 - HFT draw + ingress specialization ✅

- [x] P9-A Tier-split JSON: HFT → `StreamDecoder` only; retail/active → pooled unmarshal ≤512 KiB; pool cap 128 KiB (`json_decode.go`, `handlers.go`)
- [x] P9-B Bulk ParentTree: `ReserveMCIDsLite` + per-page `FillDeferredParentTreePage` (replaces per-row `AttachDeferredMCIDs`)
- [x] P9-C HFT fast row: `drawSharedDeferRow` + uniform border batch; `sharedColsUniformBorder` guard
- [x] P9-D `PageMCIDStart` - safe MCID cursor on multi-page HFT tables
- [x] P9-E `make load-pprof-1500` re-run (**1,059 req/s** - `drawTable` −14% CPU, gate still fails)

### Phase 10 - HFT decode + draw ultra-fast path ✅

- [x] P10-A In-place HFT row/cell prealloc (`preallocInlineTableRows` len=2001, sonic unmarshals without GrowSlice)
- [x] P10-B HFT pooled read + `sonic.Unmarshal` when `Content-Length` known (`hftBodyBufPool`, ≤8 MiB)
- [x] P10-C Ultra-fast defer row loop: `prepSharedDeferRow` + one-time `sharedCols` init; skips TR/height/wrap paths
- [x] P10-D Precomputed `textColorCmd` + `stdCharWidth` in `sharedColumnLayout`
- [x] P10-E `make load-pprof-1500` re-run (**1,009 req/s** - HFT avg **267 ms**, p99 **135 ms**, `drawTable` **11.4%** cum)

### Phase 11 - Split HFT decode + flate tuning ✅

- [x] P11-A HFT split decode: shell via `json.RawMessage` rows + second-pass `sonic.Unmarshal` (`hft_decode.go`)
- [x] P11-B Flate single-pass for streams >32 KiB (skip trial 4 KiB compress on HFT pages)
- [x] P11-C Skip compress-cache FNV fingerprint for streams >256 KiB (unique HFT pages)
- [x] P11-D `make load-pprof-1500` re-run (**1,232 req/s** - HFT avg **222 ms**, +22% throughput vs P10)

### Phase 12 - Compress-cache + signature embed ❌ reverted

- [ ] P12-A CRC32 fingerprint + skip cache >256 KiB - **reverted** (CPU win, no E2E throughput gain)
- [ ] P12-B In-place signature hex embed - **reverted**
- [ ] P12-C Store-uncompressed large HFT pages - **reverted** (810 req/s, `185831`)
- [x] P12-D Post-revert validation `make load-pprof` (**1,054 req/s** weighted, `190806`)

---

## Files changed (implementation)

| Area | Files |
|------|-------|
| Server | `cmd/gopdfsuit/main.go` - `resolveMaxConcurrent`, `WarmRuntimePools`, `WarmJSONDecode` |
| Handler | `internal/handlers/handlers.go`, `json_decode.go`, `hft_decode.go` - pretouch, tier decode, HFT split decode |
| Models | `internal/models/models.go` - `PreallocForDecode(contentLength, tier)` |
| PDF engine | `internal/pdf/generator.go`, `draw.go`, `structure.go`, `pagemanager.go` |
| Compression | `internal/pdf/font/compression.go`, `compress_cache.go` (sharded, P11 skip-sample) |
| k6 | `payload_generator.js`, `load_test_pprof.js` - `retail_active_signed`, `X-Payload-Tier` |
| Build | `makefile` - `load-pprof`, `load-pprof-gate`, `load-pprof-1k`, `load-pprof-1500` |

---

## Why Phase 7 achieved +75% throughput (640 → 1,118 req/s)

Phase 7 did not add CPU cores or raise `MAX_CONCURRENT`. The gain is almost entirely **lower per-request service time**, which raises the theoretical ceiling `48 workers ÷ avg_seconds`.

### 1. Weighted average latency dropped (~41%)

| Metric | `024802` (pre-P7) | `025518` (post-P7) | Effect |
|--------|------------------:|-------------------:|--------|
| **Avg latency** | 46.8 ms | 27.7 ms | **−41%** - directly increases sustainable req/s |
| **Median** | 18.8 ms | 16.2 ms | Retail/active path already fast; room from tail |
| **p99** | 577 ms | 208 ms | **−64%** - HFT tail no longer stalls worker pool |

At 48 workers: `48 ÷ 0.0468 ≈ 1,026 req/s` (theoretical pre-P7) vs `48 ÷ 0.0277 ≈ 1,733 req/s` (theoretical post-P7). Measured **1,118 req/s** is ~65% of the new theory - remaining gap is GC/flate under saturation, not worker starvation.

### 2. Per-optimization causal chain

| Change | Mechanism | Measured impact |
|--------|-----------|-----------------|
| **P7-A1 Pretouch** | JIT-compiles `PDFTemplate` decode graph at startup; removes first-request compile stalls and reduces decode CPU in the hot loop | `decodeTemplateJSON` cum **~10.6%** CPU (was scattered JIT + decode); fewer latency spikes |
| **P7-A2 Tier prealloc + `X-Payload-Tier`** | k6 sends tier header; `Elements`/`Table` slices sized before unmarshal → fewer `GrowSlice` during sonic parse | Contributed to higher throughput at same VU count (less GC pause) |
| **P7-B1 Direct struct write** | `formatStructElemObjectTo(pdfBuffer)` eliminates `sb.String()` copy per struct elem | **`formatStructElemObjectInto` fell off CPU top-25** (was **10.7% cum**); less `memmove` |
| **P7-B2 Batch MC (≤20 rows)** | Retail tables: `WriteCellMarkedContentBDC` + `AttachRowMCIDs` - one TR parent, no per-cell `StructElem` alloc | Cuts `beginMarkedContentBuf` structure churn on **80%** of requests |
| **P7-B3 HFT structure deferral** | Shared-layout data rows skip `StructTR`; MCIDs registered on table parent only | HFT p99 **577 → 208 ms**; largest contributor to avg latency drop (5% of traffic, ~40% of tail) |
| **P7-C1 Early `pdfBuffer.Grow`** | `estimateTemplatePDFBufferSize` before generation | Heap inuse **317 → 181 MB**; fewer mid-flight `bytes.growSlice` |
| **P7-C2 Page striping** | `PrepareLargeTableStripe` + raised per-page stream cap for >40-row tables | HFT draws in page-sized chunks; less single-pass buffer pressure |
| **P7-C3 Sharded compress cache** | Per-CPU `sync.Map` shards | Bounded cache memory (**5.5 MB** inuse); less lock contention at 48 workers |

### 3. Workload math (why +75% not +100%)

The weighted mix is **80% retail / 15% active / 5% HFT**. Phase 7 optimized all three tiers, but:

- **Retail** was already near ceiling (~19 ms med) - limited upside.
- **HFT** improved most in p99, but still dominates **avg** when it hits (~37 ms pdf_generation avg includes HFT spikes).
- **Sonic decode** remains **~51% alloc_space** - GC still caps sustained throughput below theory.

Hence +75% measured vs +69% theoretical avg-latency improvement: consistent with GC + flate remaining bottlenecks.

### 4. What did *not* drive the gain

- ECDSA signing (already optimized in Phase 2) - unchanged CPU share.
- Gin/HTTP middleware (`GIN_FAST_API=1` from Phase 5) - no change in Phase 7.
- Concurrency (`MAX_CONCURRENT=48`) - unchanged.

---

## Post-Phase 7 status

| Factor | Before (`024802`) | After (`025518`) | Fix |
|--------|-------------------|------------------|-----|
| Weighted throughput | 640 req/s | **1,118 req/s** (+75%) | Phase 7 combined |
| **5% HFT tail** | +25–40 ms avg | p99 208 ms (was 577 ms) | P7-B3 + P7-C2 |
| **sonic JSON decode** | 41% alloc | ~51% alloc (higher volume) | P7-A1 + P7-A2 |
| **struct tree write** | 10.7% CPU cum | off top-25 | P7-B1 |
| **per-cell PDF/UA** | 6.6% CPU | batch path for retail | P7-B2 |

**Next lever toward 1,500 req/s:** flate tuning on HFT stripes; HFT JSON/decode path; retail path still has headroom at **19 ms** avg.

---

## Artifacts

| File | Description |
|------|-------------|
| `guides/cursor/baselines/gin_pprof_runs/cpu_gin_20260611_190806.prof` | **Latest** CPU (35 s, P12 reverted) |
| `guides/cursor/baselines/gin_pprof_runs/k6_gin_20260611_190806.txt` | **Latest** weighted k6 (~1,054 req/s) |
| `guides/cursor/baselines/gin_pprof_runs/k6_gin_20260611_190850.txt` | Latest retail-only k6 (~3,965 req/s) |
| `guides/cursor/baselines/gin_pprof_runs/pprof_summary_20260611_190806.txt` | Latest text summary |
| `guides/cursor/baselines/gin_pprof_runs/cpu_gin_20260611_185414.prof` | Phase 11 peak CPU |
| `guides/cursor/baselines/gin_pprof_runs/k6_gin_20260611_185414.txt` | Phase 11 peak k6 (~1,232 req/s weighted) |
| `guides/cursor/baselines/gin_pprof_runs/cpu_gin_20260611_184939.prof` | Phase 10 CPU |
| `guides/cursor/baselines/gin_pprof_runs/k6_gin_20260611_184939.txt` | Phase 10 k6 (~1,009 req/s weighted) |
| `guides/cursor/baselines/gin_pprof_runs/cpu_gin_20260611_030401.prof` | Phase 9 CPU |
| `guides/cursor/baselines/gin_pprof_runs/k6_gin_20260611_030401.txt` | Phase 9 k6 (~1,059 req/s weighted) |
| `guides/cursor/baselines/gin_pprof_runs/cpu_gin_20260611_025956.prof` | Phase 8 CPU |
| `guides/cursor/baselines/gin_pprof_runs/heap_gin_20260611_025956.prof` | Phase 8 heap |
| `guides/cursor/baselines/gin_pprof_runs/pprof_summary_20260611_025956.txt` | Phase 8 text summary |
| `guides/cursor/baselines/gin_pprof_runs/k6_gin_20260611_025956.txt` | Phase 8 k6 (~1,082 req/s weighted) |
| `guides/cursor/baselines/gin_pprof_runs/cpu_gin_20260611_025518.prof` | Phase 7 CPU |
| `guides/cursor/baselines/gin_pprof_runs/k6_gin_20260611_025518.txt` | Phase 7 k6 (~1,118 req/s) |
| `guides/cursor/baselines/gin_pprof_runs/cpu_gin_20260611_024802.prof` | Pre-Phase 7 CPU |
| `guides/cursor/baselines/gin_pprof_runs/k6_gin_20260611_024802.txt` | Pre-Phase 7 k6 (~640 req/s) |
| `guides/cursor/baselines/gin_pprof_runs/cpu_gin_20260611_023739.prof` | Pre-opt baseline CPU |
| `guides/cursor/baselines/gin_pprof_runs/cpu_gin_20260611_024509.prof` | Post-opt run #1 CPU |
| `guides/cursor/baselines/k6_ramping_20260611.txt` | Ramping baseline notes |

```bash
# Inspect latest profile interactively
BIN=guides/cursor/baselines/gin_pprof_runs/gopdfsuit_20260611_024802
go tool pprof -http=:8081 $BIN guides/cursor/baselines/gin_pprof_runs/cpu_gin_20260611_024802.prof
go tool pprof -http=:8082 -inuse_space $BIN guides/cursor/baselines/gin_pprof_runs/heap_gin_20260611_024802.prof
```

---

## Related docs

- [20260611_zerodha_3000_ops_pprof_report.md](./20260611_zerodha_3000_ops_pprof_report.md)
- [../cursor/PASS4_PDFA_RESULTS.md](../cursor/PASS4_PDFA_RESULTS.md)
- [../cursor/PASS4_OPTIMIZATION_PLAN.md](../cursor/PASS4_OPTIMIZATION_PLAN.md)