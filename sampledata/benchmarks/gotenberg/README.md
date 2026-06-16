# Gotenberg k6 Benchmarks

HTML→PDF load tests for [Gotenberg](https://gotenberg.dev), mirroring the gopdfsuit k6 harness in `test/generate_template-pdf/`.

## Workload parity

| gopdfsuit | Gotenberg |
|-----------|-----------|
| JSON `POST /api/v1/generate/template-pdf` | `POST /forms/chromium/convert/html` (multipart HTML) |
| 80% retail / 15% active / 5% HFT | Same distribution via `html_payload_generator.js` |
| ECDSA signing + PDF/A + PDF/UA | **Not available** — visual HTML render only |

Use the same `PAYLOAD_SCENARIO` env vars as gopdfsuit (`tagged_ecdsa`, `retail_only_signed`, `retail_active_signed`).

## Prerequisites

- Docker
- k6 (`make bench-k6-install`)

## Quick start

```bash
# Smoke (Gotenberg must be running)
bash sampledata/benchmarks/gotenberg/start_gotenberg.sh
cd sampledata/benchmarks/gotenberg && k6 run smoke_test.js

# Full weighted load (starts/stops Gotenberg container)
make bench-gotenberg

# k6 only — start Gotenberg yourself
make bench-gotenberg-load
```

## Environment

| Variable | Default | Notes |
|----------|---------|-------|
| `BASE_URL` | `http://127.0.0.1:3010` | Gotenberg API (3010 avoids common :3000 conflicts) |
| `LOAD_VUS` | `48` | Match gopdfsuit bench |
| `PROFILE_SECONDS` | `35` | Steady-state duration |
| `PAYLOAD_SCENARIO` | `tagged_ecdsa` | Same names as gopdfsuit |
| `CHROMIUM_MAX_CONCURRENCY` | `6` | Gotenberg max per container (hard cap in v8.x) |
| `SKIP_NETWORK_IDLE` | `1` | `skipNetworkIdleEvent=true` (recommended) |
| `THROUGHPUT_GATE` | `0` | Min `http_reqs` rate threshold |
| `MANAGE_CONTAINER` | `1` | `0` = do not start/stop Docker |

## Files

| File | Role |
|------|------|
| `html_payload_generator.js` | Retail / active / HFT HTML (mirrors `payload_generator.js`) |
| `load_test_pprof.js` | Primary steady-state harness (tier latency metrics) |
| `load_test.js` | Shorter harness without smoke phase |
| `smoke_test.js` | 1 VU sanity check |
| `start_gotenberg.sh` | Docker run + health wait |
| `run_gotenberg_load.sh` | Full run + artifact capture |

Artifacts: `guides/cursor/baselines/gotenberg_runs/`

## Latest results (2026-06-13, same session as gopdfsuit `20260613_215040`)

| Metric | Gotenberg (`215127`) | gopdfsuit (`215040`) |
|--------|---------------------:|---------------------:|
| **Throughput (peak)** | — | **859 req/s** (5-run, `20260614_004108`) |
| **Throughput (avg)** | **10.3 req/s** | **825 req/s** (5-run) |
| **http median** | 4.26 s | 16.9 ms |
| **http p99** | 8.22 s | 374 ms |
| **Errors** | 0% | 0% |

gopdfsuit is **~80×** faster on the same weighted k6 mix (825 ÷ 10.3 avg). Gotenberg is capped at **6** Chromium workers per container; 48 VUs queue behind the API.

See [comparison_20260613.md](../../../guides/cursor/baselines/gotenberg_runs/comparison_20260613.md).