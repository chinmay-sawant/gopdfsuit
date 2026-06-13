# gopdfsuit vs Gotenberg — k6 weighted benchmark (2026-06-13)

**Environment:** WSL2, i7-13700HX, 24 logical CPUs, `GOMAXPROCS=24`  
**Harness:** 48 VUs × 35s steady, `PAYLOAD_SCENARIO=tagged_ecdsa` (80% retail / 15% active / 5% HFT)  
**Runs:** back-to-back same evening (`21:50`–`21:52` IST)

| Stack | Command | Run ID | Peak | Latest avg | http med | http p99 | Errors |
|-------|---------|--------|-----:|-----------:|---------:|---------:|-------:|
| **gopdflib** | Zerodha `go run .` | 30-run `20260614` | **2,953 ops/s** | **2,646 ops/s** | — | — | — |
| **gopdfsuit** | `make bench-k6` | 5-run `20260614` | **859 req/s** | **825 req/s** | 16.0 ms | 347 ms | 0% |
| **Gotenberg** | `make bench-gotenberg` | `20260613_215127` | — | **10.3 req/s** | 4.26 s | 8.22 s | 0% |

**gopdfsuit vs Gotenberg (avg):** **~80×** throughput (825 ÷ 10.3).  
**gopdflib vs gopdfsuit HTTP (avg):** **~3.2×** in-process (2,646 ÷ 825).

## Per-tier latency (custom k6 metrics)

| Tier | gopdfsuit avg / med / p99 | Gotenberg avg / med / p99 |
|------|---------------------------|---------------------------|
| **Retail (80%)** | 22 ms / 17 ms / 111 ms | 4,210 ms / 4,241 ms / 8,003 ms |
| **Active (15%)** | 32 ms / 24 ms / 167 ms | 3,874 ms / 4,141 ms / 8,077 ms |
| **HFT (5%)** | 576 ms / 444 ms / 3,759 ms | 6,247 ms / 5,956 ms / 8,397 ms |

## Configuration notes

| | gopdfsuit | Gotenberg |
|---|-----------|-----------|
| **Workers** | `MAX_CONCURRENT=48` | `chromium-max-concurrency=6` (v8.x hard cap) |
| **Input** | JSON template POST | HTML multipart Chromium render |
| **PDF/A + ECDSA** | Yes | No (visual HTML only) |
| **Gotenberg tuning** | — | `skipNetworkIdleEvent=true`, port `:3010` |

Gotenberg completed only **413 iterations** in 35s at 48 VUs because each Chromium conversion takes **~4–6 s** and the container allows **6** concurrent Chromium jobs. Remaining VUs queue behind the API.

## Artifacts

| Stack | k6 log | Summary |
|-------|--------|---------|
| gopdfsuit | `guides/cursor/baselines/gin_pprof_runs/k6_gin_20260613_215040.txt` | `pprof_summary_20260613_215040.txt` |
| Gotenberg | `guides/cursor/baselines/gotenberg_runs/k6_gotenberg_20260613_215127.txt` | `summary_20260613_215127.txt` |

## Commands to reproduce

```bash
make bench-k6
make bench-gotenberg
```