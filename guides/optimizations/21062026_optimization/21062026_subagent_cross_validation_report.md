# Six-Agent Cross-Validation Report - Zerodha 15K Optimization

**Date:** 2026-06-21  
**Profile:** `cpu_zerodha_run3.prof` (11.27s, 1771% CPU), `heap_zerodha.prof` (563 MB)  
**Benchmark:** `make bench-gopdflib-zerodha-x10-pprof` (fresh run)  
**Target:** 15,000 ops/sec from idle baseline 9,009 ops/sec

---

## Agent Roster

| ID | Specialization | Primary artifacts reviewed |
|----|----------------|---------------------------|
| A1 | HFT TR→TD structure path | `structure.go`, `hft_algo.json`, HFT pprof subtree |
| A2 | Retail signing + PDF/A | `signature.go`, `retail_investor.json`, `metadata.go` |
| A3 | Active trader format | `active_trader.json`, `bench.go` active template |
| A4 | Memory / heap / GC | `heap_zerodha.prof`, pool implementations, x10 variance |
| A5 | Table rendering | `draw.go`, `drawTable` call tree, shared layout |
| A6 | Cross-format shared bottlenecks | All-format CPU ranking, phased path model |

---

## Unanimous Findings (6/6 Agreement)

### 1. HFT dominates despite 5% document share

| Agent | HFT latency share estimate | HFT CPU share estimate |
|-------|---------------------------:|----------------------:|
| A1 | 75–85% of mean latency | 50–60% of CPU samples |
| A2 | (implicit via mixed math) | ~42% from HFT tail |
| A3 | - | HFT vs active 17× per-doc |
| A4 | - | Arena + pdfBuffer HFT-weighted |
| A5 | - | drawTable 93.7% HFT |
| A6 | - | 40–50% of total CPU time |

**Validated conclusion:** Phase C (HFT tail) is **mandatory** for 15K. No agent disputes this.

### 2. sRGB ICC cache leak (P9 incomplete)

| Agent | Observation |
|-------|-------------|
| A2 | `buildSRGBICCProfile` 4.17% cum, hits all PDFs |
| A4 | Listed in alloc hot path |
| A6 | **#1 quick win**, 4.5% CPU, all 3 formats |
| A1 | Included in P26 |
| A3 | +60–90 ops/sec active share |
| A5 | Not table-related, agreed defer to P26 |

**Validated conclusion:** P26 is the **first item to implement** in Phase A.

### 3. Memory bandwidth wall blocks next tier

| Agent | Key metric |
|-------|-----------|
| A4 | 90% heap = growSlice (47%) + arena (44%) |
| A1 | pdfBuffer 263 MB + arena 247 MB |
| A5 | memmove 10.1% from cache replay |
| A6 | S1+S2 = 33% CPU combined |
| A2 | embed memclr on 61 KB retail PDF |
| A3 | Active fits 262 KB pool; minor grow |

**Validated conclusion:** Phase B (P31–P35) must land before Phase C items realize full gain.

### 4. Compliance shortcuts are off the table

All six agents independently flagged:
- No TR→TD collapse
- No key-based cross-request caches
- HFT output must stay ~2,291,942 bytes
- veraPDF 6/6 gate after every change

---

## Majority Findings (4–5/6 Agreement)

### 5. `beginTableRowArena` is the top HFT CPU hotspot

| Agent | Rank | % CPU |
|-------|------|------:|
| A1 | #3 HFT-specific | 7.5% |
| A5 | #1 child of drawTable | 7.45% |
| A4 | GC correlate | 5.8% flat |
| A6 | S3 shared ranking | 7.5% |
| A2 | Noted as HFT-only | - |
| A3 | Below 512 threshold for active | 0% |

**Validated:** P36 + P37 are the highest-impact HFT CPU items in Phase C.

### 6. Retail signature is significant but insufficient alone

| Agent | Signature % CPU | Retail-only 15K possible? |
|-------|----------------:|:------------------------:|
| A2 | 11.4% | No (~11.5K mixed ceiling) |
| A6 | 6.3% | No |
| A4 | In alloc (PKCS7 21 MB) | - |
| A1 | Retail-only, doesn't block HFT | - |
| A3 | Active has no signing | - |
| A5 | Not in drawTable path | - |

**Validated:** P35 + P42 in Phase B/C; retail opts are necessary but not sufficient.

### 7. Active trader is low-leverage for mixed 15K

| Agent | Active CPU share | Top active item |
|-------|-----------------:|-----------------|
| A3 | 7.5% | P29 SharedRowLayout +250–350 |
| A5 | P22 extension safe | +20–50 from active SharedRowLayout |
| A1 | Defer vs HFT | - |
| A2 | 15% docs, faster than retail | - |
| A4 | Minor buffer grow | - |
| A6 | P28 font precompute helps all | - |

**Validated:** P29 is worth doing in Phase A (low risk, stacks with other wins) but
will not move the 15K needle alone.

---

## Agent-Specific Unique Findings (Cross-Checked)

### A1 unique: HFT per-doc latency model

```
Retail:  ~1.0 ms/doc × 4000 = 4.0 ms weighted
Active:  ~0.5 ms/doc × 750  = 0.4 ms weighted
HFT:    ~80 ms/doc  × 250   = 20.0 ms weighted  ← 83% of mean
```

**Cross-check A2:** Serial proxy gives HFT 41.9% of CPU time - consistent with
latency-weighted model (HFT docs are slower but fewer).

**Cross-check A4:** Peak memory >1,200 MB correlates with runs <7,500 ops/sec - HFT
cold-start pattern confirmed.

### A2 unique: Signature pipeline breakdown

| Stage | Per retail doc | % of sign chain |
|-------|---------------:|----------------:|
| CreateSignatureField | 125 µs | 39% |
| embedSignatureInPlace | 163 µs | 51% |
| ECDSA SignASN1 | 70 µs | 22% |
| ASN.1 Marshal | 48 µs | 15% |

**Cross-check A4:** `encoding/asn1.MarshalWithParams` 21.54 MB alloc in 20260620
heap - confirms A2's PKCS#7 alloc concern. P42 validated.

### A3 unique: Active SharedRowLayout eligibility

41-row trade table passes `tableSupportsSharedRowLayout`:
- Uniform `Props` per column ✓
- Only `Text`/`TextColor` vary ✓
- No Wrap, images, forms ✓

**Cross-check A5:** `tableSupportsSharedRowLayout` code confirms; `BgColor`/`TextColor`
handled in `drawSharedDeferRow`. **P29 is safe.**

**Cross-check A1:** Active at 292 struct elems stays below arena threshold (512) -
SharedRowLayout routes through HFT fast draw path but uses pool-based TR→TD, not arena.
Compliance preserved.

### A4 unique: Variance correlation

| Peak allocated | Throughput (x10 runs) |
|---------------:|----------------------:|
| >1,220 MB | 5,573–7,414 ops/sec |
| <1,000 MB | 8,384–9,182 ops/sec |

**Cross-check A6:** WSL2 load tax ~12% (7,852 loaded vs 9,009 idle) - both factors
contribute. **P0 idle-machine gate is mandatory.**

### A5 unique: Text wrap is not a gate

`WrapTextInto` = 0% in top 400 pprof nodes. HFT explicitly disables wrap.

**Cross-check A1, A3:** Confirmed - table optimization = structure + font + stream copy,
not wrapping.

### A6 unique: Phased throughput model

| Milestone | Items | Projected mean |
|-----------|-------|---------------:|
| 11,000 | P26, P28, P27a | 10,500–11,000 |
| 13,000 | P30, P31, P32 | 12,500–13,500 |
| 15,000 | P36–P40 | 14,500–15,500 |

**Cross-check all agents:** Individual gain estimates sum to 5,500–7,100 ops/sec -
sufficient to close 5,991 gap with ~10% overlap discount.

---

## Prioritized Unified Backlog (Post Cross-Validation)

| Rank | Item | Phase | Gain (mid) | Agents | Risk |
|------|------|-------|----------:|--------|------|
| 1 | P26 sRGB ICC fix | A | +550 | 6/6 | Low |
| 2 | P31 pdfBuffer zero-grow | B | +1,000 | 5/6 | Low–Med |
| 3 | P36 arena TD template | C | +750 | 4/6 | Med |
| 4 | P38 batch TD leaf write | C | +1,000 | 3/6 | Med |
| 5 | P32 arena slab sizing | B | +700 | 4/6 | Med |
| 6 | P37 stripe-batch arena | C | +575 | 3/6 | Med |
| 7 | P35 signature embed | B | +500 | 3/6 | Med |
| 8 | P29 active SharedRowLayout | A | +300 | 3/6 | Low |
| 9 | P39 MarkCharsUsed batch | C | +300 | 3/6 | Low |
| 10 | P28 font precompute | A | +275 | 2/6 | Low |

---

## What Will NOT Get Us to 15K

| Approach | Why it fails | Agent consensus |
|----------|--------------|:---------------:|
| Retail-only optimizations | Mixed ceiling ~11,500 even at zero retail cost | 6/6 |
| Disabling HFT TR→TD | veraPDF FAIL; 748 KB non-compliant output | 6/6 |
| Key-based row render cache | k6 regression documented 2026-06-17 | 6/6 |
| Text wrap optimizations | 0% CPU on Zerodha workload | 5/6 |
| Compression tuning | P7 done; 1.2% cum, not a gate | 5/6 |
| Skipping signing on retail | 80% workload; compliance violation | 6/6 |
| Benchmark harness tuning only | Doesn't improve `GeneratePDF` itself | 4/6 |

---

## Fresh Benchmark vs Idle Baseline

| Session | Mean | Best | Worst | σ | Notes |
|---------|-----:|-----:|------:|--:|-------|
| User idle baseline | 9,009 | 10,659 | 6,943 | 1,333 | Reference target |
| 2026-06-21 loaded x10 | 7,852 | 9,182 | 5,573 | 1,112 | This pprof run |
| 2026-06-21 pprof x5 mean | 8,856 | 9,576 | 7,181 | 859 | Profile subset |

**Recommendation:** Re-establish idle baseline before starting Phase A implementation.
Compare all phase gates against **idle 9,009**, not load-depressed 7,852.

---

## Next Action

1. Start **P26** (sRGB ICC cache fix) - 6/6 agent consensus, lowest risk, all formats.
2. Run idle-machine x10 to confirm 9,009 baseline before measuring Phase A wins.
3. Implement Phase A items P26 → P28 → P29 → P30 sequentially with veraPDF gate.
4. Re-profile after Phase A; expect `buildSRGBICCProfile` eliminated from top-40 CPU.