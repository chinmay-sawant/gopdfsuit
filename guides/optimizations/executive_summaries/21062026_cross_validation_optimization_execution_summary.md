# Optimization Execution Summary - 2026-06-21 Cross-Validation

## Date & Scope

Six-agent cross-validation of the **Zerodha 15K optimization plan** on **2026-06-21**, using fresh pprof artifacts from `make bench-gopdflib-zerodha-x10-pprof`. Profile inputs: `cpu_zerodha_run3.prof` (11.27s, 1771% CPU) and `heap_zerodha.prof` (563 MB live). Workload: Zerodha gold-standard mix (80% retail / 15% active / 5% HFT). **Target:** 15,000 ops/sec from idle baseline **9,009 ops/sec** (gap: **5,991 ops/sec**). This pass validates individual agent analyses (A1–A6) against a unified phased backlog (P26–P40) and compliance constraints before Phase A implementation begins.

## Key Outcomes

- **6/6 agent consensus:** Phase C (HFT struct tail) is **mandatory** for 15K; retail-only or active-only work cannot close the mixed-workload gap.
- **First implementation item locked:** **P26** (sRGB ICC cache fix) - unanimous quick win, ~+550 ops/sec mid estimate, lowest risk, all three formats.
- **Benchmark discipline clarified:** Measure all phase gates against **idle 9,009 ops/sec**, not the load-depressed **7,852 ops/sec** from the 2026-06-21 loaded x10 run (~12% WSL2 load tax).
- **Phased throughput model validated:** A6 milestone math (with ~10% overlap discount) closes the 5,991 gap - **11K** after Phase A, **13K** after Phase B, **15K** after Phase C.
- **Compliance red lines confirmed:** No TR→TD collapse, no key-based cross-request caches, HFT output must remain ~2,291,942 bytes, veraPDF 6/6 after every change.
- **Seven approaches ruled out** as insufficient or non-compliant (retail-only ceiling ~11.5K, wrap optimizations at 0% CPU, etc.).

## Work Completed

### Six-Agent Roster

| ID | Specialization | Primary artifacts reviewed |
|----|----------------|---------------------------|
| **A1** | HFT TR→TD structure path | `structure.go`, `hft_algo.json`, HFT pprof subtree |
| **A2** | Retail signing + PDF/A | `signature.go`, `retail_investor.json`, `metadata.go` |
| **A3** | Active trader format | `active_trader.json`, `bench.go` active template |
| **A4** | Memory / heap / GC | `heap_zerodha.prof`, pool implementations, x10 variance |
| **A5** | Table rendering | `draw.go`, `drawTable` call tree, shared layout |
| **A6** | Cross-format shared bottlenecks | All-format CPU ranking, phased path model |

### Validation Process

1. Each agent produced independent bottleneck rankings and gain estimates from the same pprof session.
2. Findings were bucketed into **unanimous (6/6)**, **majority (4–5/6)**, and **agent-unique** categories.
3. Unique claims were cross-checked against at least one other agent's data (e.g., A2 PKCS#7 alloc confirmed by A4 heap; A3 SharedRowLayout eligibility confirmed by A5 code review).
4. Individual gain estimates were summed (5,500–7,100 ops/sec raw) and discounted ~10% for overlap to validate the phased model.
5. A unified backlog (P26–P40) was ranked by agent support count, estimated mid gain, phase assignment, and compliance risk.
6. Fresh benchmark sessions were compared against the user idle baseline to establish measurement gates.

## Findings

### Unanimous (6/6 Agreement)

| # | Finding | Key evidence | Implication |
|---|---------|--------------|-------------|
| 1 | **HFT dominates despite 5% doc share** | 75–85% mean latency; 40–60% CPU; 17× per-doc vs active; `drawTable` 93.7% HFT-weighted | Phase C mandatory |
| 2 | **sRGB ICC cache leak - P9 incomplete → P26** | `buildSRGBICCProfile` 4.17% cum; hits all PDFs; #1 quick win per A6 | **Start P26 first** |
| 3 | **Memory bandwidth wall blocks next tier** | 90% heap = `growSlice` (47%) + arena (44%); `memmove` 10.1%; pdfBuffer 263 MB + arena 247 MB | Phase B (P31–P35, P40) before Phase C realizes full gain |
| 4 | **Compliance shortcuts off the table** | No TR→TD collapse; no key-based caches; HFT ~2,291,942 B; veraPDF 6/6 gate | Constrains all backlog items |

**HFT latency model (A1, cross-checked A2/A4):**

```
Retail:  ~1.0 ms/doc × 4000 =  4.0 ms weighted
Active:  ~0.5 ms/doc ×  750 =  0.4 ms weighted
HFT:    ~80 ms/doc  ×  250 = 20.0 ms weighted  ← 83% of mean
```

### Majority (4–5/6 Agreement)

| # | Finding | Agent support | Backlog items |
|---|---------|---------------|---------------|
| 5 | **`beginTableRowArena` = top HFT CPU hotspot** (~7.5% cum) | A1, A4, A5, A6 (+ A2 noted HFT-only) | **P36 + P37** highest-impact Phase C CPU |
| 6 | **Retail signature significant but insufficient alone** (~11.4% CPU; mixed ceiling ~11.5K) | A2, A4, A6 | **P35 + P42** in Phase B/C; necessary not sufficient |
| 7 | **Active trader low-leverage for mixed 15K** (~7.5% CPU share) | A3, A5, A6 (+ others defer vs HFT) | **P29** worth doing in Phase A (+300 mid) but won't move needle alone |

### Agent-Unique Findings (Cross-Checked)

| Agent | Unique finding | Cross-check result |
|-------|----------------|-------------------|
| **A1** | HFT per-doc latency model (83% weighted mean) | Consistent with A2 serial proxy (41.9% CPU) and A4 cold-start memory pattern (>1,200 MB → <7,500 ops/sec) |
| **A2** | Signature pipeline: `embedSignatureInPlace` 51%, `CreateSignatureField` 39%, ECDSA 22% | A4 confirms `asn1.MarshalWithParams` 21.54 MB alloc → **P42 validated** |
| **A3** | Active 41-row table eligible for `tableSupportsSharedRowLayout` | A5 code review confirms; A1 confirms compliance preserved (pool TR→TD, not arena) → **P29 safe** |
| **A4** | Peak alloc >1,220 MB → 5,573–7,414 ops/sec; <1,000 MB → 8,384–9,182 | A6 WSL2 load tax ~12% (7,852 vs 9,009) - **P0 idle-machine gate mandatory** |
| **A5** | `WrapTextInto` = 0% in top 400 nodes | A1/A3 confirm - table opt = structure + font + stream copy, not wrapping |
| **A6** | Phased model: 11K → 13K → 15K with P26–P40 | All agents' gains sum to 5,500–7,100 ops/sec - sufficient with overlap discount |

### What Will NOT Get Us to 15K

| Approach | Why it fails | Consensus |
|----------|--------------|:---------:|
| Retail-only optimizations | Mixed ceiling ~11,500 even at zero retail cost | 6/6 |
| Disabling HFT TR→TD | veraPDF FAIL; 748 KB non-compliant output | 6/6 |
| Key-based row render cache | k6 regression documented 2026-06-17 | 6/6 |
| Skipping signing on retail | 80% workload; compliance violation | 6/6 |
| Text wrap optimizations | 0% CPU on Zerodha workload | 5/6 |
| Compression tuning | P7 done; 1.2% cum, not a gate | 5/6 |
| Benchmark harness tuning only | Doesn't improve `GeneratePDF` itself | 4/6 |

### Benchmark: Idle Baseline vs Loaded Run

| Session | Mean | Best | Worst | σ | Notes |
|---------|-----:|-----:|------:|--:|-------|
| **User idle baseline** | **9,009** | 10,659 | 6,943 | 1,333 | Reference target for all gates |
| 2026-06-21 loaded x10 | 7,852 | 9,182 | 5,573 | 1,112 | This pprof run (WSL2 load) |
| 2026-06-21 pprof x5 mean | 8,856 | 9,576 | 7,181 | 859 | Profile subset |

## Open Items / Next Steps

### Immediate Actions

1. **Start P26** (sRGB ICC cache fix) - 6/6 consensus, lowest risk, all formats.
2. **Re-run idle-machine x10** to confirm 9,009 baseline before measuring Phase A wins.
3. **Implement Phase A sequentially:** P26 → P28 → P29 → P30 with veraPDF gate after each item.
4. **Re-profile after Phase A;** expect `buildSRGBICCProfile` eliminated from top-40 CPU.

### Prioritized Unified Backlog (P26–P40, Post Cross-Validation)

| Rank | Item | Phase | Gain (mid) | Agents | Risk |
|------|------|-------|----------:|--------|------|
| 1 | **P26** sRGB ICC fix | A | +550 | 6/6 | Low |
| 2 | **P31** pdfBuffer zero-grow | B | +1,000 | 5/6 | Low–Med |
| 3 | **P36** arena TD template | C | +750 | 4/6 | Med |
| 4 | **P38** batch TD leaf write | C | +1,000 | 3/6 | Med |
| 5 | **P32** arena slab sizing | B | +700 | 4/6 | Med |
| 6 | **P37** stripe-batch arena | C | +575 | 3/6 | Med |
| 7 | **P35** signature embed | B | +500 | 3/6 | Med |
| 8 | **P29** active SharedRowLayout | A | +300 | 3/6 | Low |
| 9 | **P39** MarkCharsUsed batch | C | +300 | 3/6 | Low |
| 10 | **P28** font precompute | A | +275 | 2/6 | Low |
| - | **P30** (Phase A quick win) | A | - | - | Low |
| - | **P34** page content stream caps | B | +350–550 | - | Med |
| - | **P40** row stream direct append | B | +300–500 | - | Med |
| - | **P33** xref offset slice pooling | B | +150–300 | - | Low–Med |

### Phased Milestones (A6 Model)

| Milestone | Items | Projected mean | Gate |
|-----------|-------|---------------:|------|
| Baseline (idle) | P0–P25 done | **9,009** | measured |
| **11,000** | P26, P28, P27a/P30, P29 | 10,500–11,000 | veraPDF 6/6 |
| **13,000** | P31, P32, P34, P40, P35 | 12,500–13,500 | peak alloc ≤ 750 MB |
| **15,000** | P36–P39 (+ P41, P42) | 14,500–15,500 | HFT 2,291,942 ± 5% B |

```
9,009 ──Phase A──► ~10,500–11,000 ──Phase B──► ~13,000 ──Phase C──► ~15,000
```

### Phase Acceptance Gates

| Phase | Throughput gate | Memory / quality gate |
|-------|-----------------|----------------------|
| A | x10 mean ≥ 10,400 ops/sec (idle) | `buildSRGBICCProfile` out of top-40 |
| B | x10 mean ≥ **13,000** ops/sec | peak alloc ≤ 750 MB; `growSlice` ≤ 50 MB |
| C | x10 mean ≥ **15,000** ops/sec; best ≥ 16,000 | peak alloc ≤ 650 MB; `drawTable` cum ≤ 20% |

## Source Documents

| File | Role |
|------|------|
| `21062026_optimization/21062026_subagent_cross_validation_report.md` | Primary report - agent roster, unanimous/majority/unique findings, unified backlog, benchmark comparison |
| `21062026_optimization/21062026_zerodha_15k_optimization_checklist.md` | Full P26–P42 implementation checklist, per-format maps, phase gates, verification commands |
| `guides/cursor/baselines/zerodha_pprof_runs/cpu_zerodha_run3.prof` | CPU profile (11.27s, 1771% CPU) |
| `guides/cursor/baselines/zerodha_pprof_runs/heap_zerodha.prof` | Heap profile (563 MB) |
| `20260617_k6_bench_regression_analysis.md` | Key-based cache regression precedent (context for compliance red line) |