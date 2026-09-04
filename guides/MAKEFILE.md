# Makefile reference

> Frozen archive (see `plans/adr-2026-09-04-doc-homes.md`). Do not update;
> current make targets are documented in the makefile itself (`make bench-help`)
> and `documentation/`.

Complete guide to the Makefile targets for building, testing, and deploying GoPdfSuit.

---

## Table of contents

- [Overview](#overview)
- [Variables](#variables)
- [Docker targets](#docker-targets)
- [Build targets](#build-targets)
- [Development targets](#development-targets)
- [Benchmark targets](#benchmark-targets)
- [Quick reference](#quick-reference)

---

## Overview

I use these targets so I do not have to remember long docker, go, and k6 commands. They wrap the repeatable steps for builds, tests, releases, and benchmarks. Run them from the repo root.

---

## Variables

Customize behavior using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `VERSION` | `6.0.0` | Application version tag |
| `DOCKERUSERNAME` | `chinmaysawant` | Docker Hub username |
| `GO_BENCH` | `go` | Go binary for benchmarks (`go1.26.4` recommended) |
| `GOMAXPROCS_BENCH` | `24` | CPU limit for Go benchmark processes |
| `BENCH_ITERATIONS` | `5000` | Zerodha / data-table iteration count |
| `BENCH_WORKERS` | `48` | Concurrent workers for Zerodha harnesses |
| `BENCH_COUNT` | `1` | `go test -count` repetitions |
| `BENCH_TIME` | `5s` | `go test -benchtime` for GoPDFKit compare |
| `LOAD_VUS` | `48` | k6 virtual users (`run_gin_pprof_load.sh`) |
| `PROFILE_SECONDS` | `35` | CPU profile duration during k6 load |
| `K6_LIGHT_VUS` | `24` | Virtual users for `bench-k6-light` |
| `K6_LIGHT_SECONDS` | `15` | Steady duration for `bench-k6-light` |
| `K6_LIGHT_MAX_CONCURRENT` | `24` | Server concurrency cap for `bench-k6-light` |
| `K6_LIGHT_GOMAXPROCS` | `12` | `GOMAXPROCS` for `bench-k6-light` server process |
| `PAYLOAD_SCENARIO` | `tagged_ecdsa` | k6 payload mix (`retail_only_signed` for gate) |
| `THROUGHPUT_GATE` | `0` | Minimum req/s gate for k6 script (0 = off) |
| `BASE_URL` | `http://127.0.0.1:8080` | Target URL for k6 / pprof script |

### Setting variables

```bash
# Inline
VERSION=6.0.0 make docker

# Export for session
export VERSION=6.0.0
export DOCKERUSERNAME=myusername
make docker
```

---

## Docker targets

### `make docker`

Build and run the Docker container locally.

```bash
make docker
```

**What it does.** It builds the image from `dockerfolder/Dockerfile`, tags it as `gopdfsuit:<VERSION>`, and runs it on port 8080.

**Same commands.**

```bash
docker build -f dockerfolder/Dockerfile --build-arg VERSION=6.0.0 -t gopdfsuit:6.0.0 .
docker run -d -p 8080:8080 gopdfsuit:6.0.0
```

---

### `make dockertag`

Tag and push the Docker image to Docker Hub.

```bash
make dockertag
```

**What it does.** It tags the image with the version number and `latest`, prompts for Docker Hub login, and pushes both tags.

**Same commands.**

```bash
docker tag gopdfsuit:6.0.0 chinmaysawant/gopdfsuit:6.0.0
docker tag gopdfsuit:6.0.0 chinmaysawant/gopdfsuit:latest
docker login
docker push chinmaysawant/gopdfsuit:6.0.0
docker push chinmaysawant/gopdfsuit:latest
```

---

### `make pull`

Pull and run the latest image from Docker Hub.

```bash
make pull
```

**What it does.** It pulls the image from Docker Hub and runs it on port 8080.

**Same commands.**

```bash
docker pull chinmaysawant/gopdfsuit:6.0.0
docker run -d -p 8080:8080 chinmaysawant/gopdfsuit:6.0.0
```

---

## Build targets

### `make build`

Compile the Go application.

```bash
make build
```

**What it does.** It creates `bin/` and compiles the binary to `bin/app`.

**Output.** `bin/app`.

---

### `make wasm-compress`

Build the in-browser PDF compressor (`GOOS=js GOARCH=wasm`).

```bash
make wasm-compress
```

**What it does.** It compiles `cmd/wasmcompress` to `frontend/public/compress.wasm`, copies Go's `wasm_exec.js` next to it, and copies both files into `sampledata/compress-js/` for the JS sample.

Check the result with `file frontend/public/compress.wasm`. It should report `WebAssembly`.

The `/compress` page and `cd sampledata/compress-js && node run.mjs` use this file. It is not a CLI.

---

### `make run`

Build frontend and run the application locally.

```bash
make run
```

**What it does.** It builds the React frontend with `npm run build` and runs the Go application.

Open `http://localhost:8080` in a browser.

---

### `make clean`

Remove build artifacts.

```bash
make clean
```

**What it does.** It deletes the `bin/` directory.

---

## Development targets

### `make test`

Run all Go tests.

```bash
make test
```

**Same command.**

```bash
go test ./...
```

---

### `make fmt`

Format Go source code.

```bash
make fmt
```

**Same command.**

```bash
go fmt ./...
```

---

### `make vet`

Run Go static analysis.

```bash
make vet
```

**Same command.**

```bash
go vet ./...
```

---

### `make mod`

Tidy Go module dependencies.

```bash
make mod
```

**Same command.**

```bash
go mod tidy
```

---

## Benchmark targets

All benchmark targets run from the repository root. List every target:

```bash
make bench-help
```

For result interpretation and latest numbers, see [INTEGRATION_AND_BENCHMARK_TESTS.md](./INTEGRATION_AND_BENCHMARK_TESTS.md).

### Setup

| Target | Description |
|--------|-------------|
| `make bench-setup` | Download Typst binary and generate `sampledata/benchmarks/data.json` |
| `make bench-k6-install` | Install k6 on Debian/Ubuntu WSL (`test/generate_template-pdf/install_k6.sh`) |

### Gopdfsuit HTTP (k6 + pprof)

The `load-pprof*` and `bench-k6*` targets (except `bench-k6-load`, `bench-k6-smoke`, `bench-k6-spike`, `bench-k6-soak`) build the server, start it on port 8080, run k6, and capture CPU/heap profiles under `guides/cursor/baselines/gin_pprof_runs/`.

Use `make bench-k6-light` when you run on WSL with limited RAM, alongside other benchmarks, or when the full 48 VU by 35s run kills the server mid-run.

| Target | Description |
|--------|-------------|
| `make load-pprof` | Weighted `tagged_ecdsa` workload, 48 VU × 35s + pprof (alias: `bench-k6`) |
| `make bench-k6-light` | Same harness at **24 VU × 15s**, `MAX_CONCURRENT=24`, `GOMAXPROCS=12` - lower CPU/RAM |
| `make load-pprof-gate` | Retail-only signed fast path, 1500 req/s gate (alias: `bench-k6-retail`) |
| `make load-pprof-1k` | Weighted workload, 1000 req/s gate (alias: `bench-k6-1k`) |
| `make load-pprof-1500` | Weighted workload, 1500 req/s gate (alias: `bench-k6-1500`) |
| `make bench-k6-load` | k6 `load_test.js` only - start server yourself |
| `make bench-k6-smoke` | Quick `smoke_test.js` (10s, 1 VU) |
| `make bench-k6-spike` | Traffic spike simulation (`spike_test.js`) |
| `make bench-k6-soak` | 30-minute stability test (`soak_test.js`) |

k6-only workflow without auto server start:

```bash
go run ./cmd/gopdfsuit &
make bench-k6-smoke
```

Light run for WSL or a shared machine:

```bash
make bench-k6-light
# or tune further:
make bench-k6-light K6_LIGHT_VUS=16 K6_LIGHT_SECONDS=10
```

Custom gate example:

```bash
make load-pprof THROUGHPUT_GATE=1500 LOAD_VUS=48 GO_BENCH=go1.26.4
```

Inspect profiles after `load-pprof`:

```bash
go tool pprof -http=:8081 guides/cursor/baselines/gin_pprof_runs/cpu_gin_*.prof
```

### Zerodha gold standard

High-volume contract-note workload in `sampledata/gopdflib/zerodha` (80% retail / 15% active / 5% HFT).

| Target | Description |
|--------|-------------|
| `make bench-gopdflib-zerodha` | gopdflib via `go run .` |
| `make bench-gopdflib-zerodha-x2` | Two sequential runs |
| `make bench-gopdflib-zerodha-x5` | Five timing runs + CPU/heap pprof (`run_bench_x5.sh`) |
| `make bench-gopdflib-zerodha-x10` | Ten sequential timing runs (`run_bench_x10.sh`) |
| `make bench-pypdfsuit-zerodha` | Python parity via `pypdfsuit_bench.py` |
| `make bench-pypdfsuit-zerodha-nocomply` | Same workload with `BENCH_NOCOMPLY=1` |
| `make bench-pypdfsuit-zerodha-nocomply-x10` | Ten sequential non-compliant runs (`run_pypdfsuit_bench_x10_nocomply.sh`) |
| `make bench-pypdfsuit-zerodha-x2` | Two sequential runs |
| `make bench-pypdfsuit-zerodha-x5` | Five timing runs + phase profile (`run_pypdfsuit_bench_x5.sh`) |
| `make bench-pypdfsuit-zerodha-x10` | Ten sequential timing runs (`run_pypdfsuit_bench_x10.sh`) |

Override iterations and workers:

```bash
make bench-gopdflib-zerodha BENCH_ITERATIONS=1000 BENCH_WORKERS=24
```

### GoPDFLib data-table

Tabular PDF/A workload in `sampledata/benchmarks/gopdflib` (distinct from Zerodha single-document).

| Target | Description |
|--------|-------------|
| `make bench-gopdflib-data` | `go run . data` with configurable workers |
| `make bench-gopdflib-data-pprof` | 5000× run + 5 CPU profiles + 1 heap profile |

### Multi-library suite

Cross-engine comparisons in `sampledata/benchmarks`. Run `make bench-setup` first for Typst and `data.json`.

| Target | Engine | Harness |
|--------|--------|---------|
| `make bench-all-libraries` | All below | `run_all_benchmarks.sh` (sequential) |
| `make bench-gopdflib-data` | GoPDFLib | `gopdflib/go run . data` |
| `make bench-gopdflib-zerodha` | GoPDFLib | `gopdflib/go run .` |
| `make bench-gopdfsuit-zerodha` | GoPDFSuit | `gopdfsuit/go run .` |
| `make bench-pypdfsuit-legacy` | PyPDFSuit | `pypdfsuit/bench.py` |
| `make bench-fpdf` | FPDF2 | `fpdf/bench.py` |
| `make bench-jspdf` | jsPDF | `jspdf/bench.js` |
| `make bench-pdfkit-lib` | PDFKit (Node) | `pdfkit/bench.js` |
| `make bench-pdflib` | pdf-lib | `pdflib/bench.js` |
| `make bench-typst` | Typst | `typst/bench.sh` |

**Note.** `bench-pypdfsuit-legacy` uses the older `sampledata/benchmarks/pypdfsuit` harness. For Zerodha parity against gopdflib, prefer `bench-pypdfsuit-zerodha`.

### GoPDFKit apples-to-apples

Module: `sampledata/benchmarks/gopdfkit_compare`. Requires a real gopdfkit checkout symlinked at `/tmp/gopdfkit-real/...`.

| Target | Description |
|--------|-------------|
| `make bench-gopdfkit-setup` | Download gopdfkit v0.5.2 and create symlink |
| `make bench-gopdfkit-compare-test` | Verify both libraries emit valid PDFs before timing |
| `make bench-gopdfkit-compare` | `BenchmarkGoPDFKit` vs `BenchmarkGoPDFLib` workloads |
| `make bench-gopdfkit-compare-x2` | Two sequential compare runs |
| `make bench-gopdfkit-html` | Opt-in HTML subset (needs Chrome; set `GOPDFKIT_COMPARE_HTML=1`) |

### Go test benchmarks

| Target | Package | Benchmarks |
|--------|---------|------------|
| `make bench-handler-all` | `./test` | Gin handler serial + parallel (`financial_report.json`) |
| `make bench-handler` | `./test` | Serial handler only |
| `make bench-handler-parallel` | `./test` | Parallel handler only |
| `make bench-pdf-micro` | `./internal/pdf` | `Rows2000` + `data.json` (`BenchmarkGoPdfSuit`) |
| `make bench-pdf-macro` | `./internal/pdf` | Synthetic tables at 2000 / 10000 / 25000 rows |
| `make bench-pdf-typst` | `./internal/pdf` | Typst compile (`-tags=compare`; needs `bench-setup`) |

### Full suites

Long-running. Run these when you want a full regression pass.

| Target | Includes |
|--------|----------|
| `make bench-suite` | Zerodha gopdflib + pypdfsuit + GoPDFKit compare + handler benches + k6/pprof |
| `make bench-suite-x2` | Two passes of each harness in `bench-suite` |
| `make bench-suite-full` | `bench-suite` + `bench-all-libraries` |

---

## Quick reference

| Command | Description | Use Case |
|---------|-------------|----------|
| `make docker` | Build & run Docker container | Local Docker testing |
| `make dockertag` | Push to Docker Hub | Release deployment |
| `make pull` | Pull & run from Docker Hub | Production deployment |
| `make build` | Compile Go binary | Local builds |
| `make run` | Build frontend & run app | Development |
| `make test` | Run tests | CI/CD, pre-commit |
| `make test-integration` | Go + Python integration suite | Pre-release validation |
| `make clean` | Remove build artifacts | Cleanup |
| `make fmt` | Format code | Pre-commit |
| `make vet` | Static analysis | Code quality |
| `make mod` | Tidy dependencies | After adding packages |
| `make bench-help` | List all benchmark targets | Discover harnesses |
| `make bench-suite` | Core benchmark regression | Performance check |
| `make load-pprof` | k6 HTTP load + pprof | End-to-end throughput |
| `make bench-k6-light` | k6 HTTP load + pprof (24 VU × 15s) | WSL / constrained runs |
| `make bench-gopdflib-zerodha` | Zerodha gold standard | Library throughput |

---

## Common workflows

### Development cycle

```bash
make fmt          # Format code
make vet          # Check for issues
make test         # Run tests
make run          # Start development server
```

### Release workflow

```bash
export VERSION=6.0.0
make docker       # Build and test locally
make dockertag    # Push to Docker Hub
```

### Fresh start

```bash
make clean        # Remove old builds
make mod          # Update dependencies
make build        # Compile fresh binary
```

### Benchmark regression

```bash
make bench-help                              # See all targets
make bench-setup                             # Typst + data.json (multi-library)
make bench-gopdfkit-setup                    # GoPDFKit symlink (compare only)
make bench-gopdflib-zerodha-x2               # Quick Zerodha check (2 runs)
make bench-handler-all GO_BENCH=go1.26.4     # Handler micro-bench
make load-pprof                              # Full HTTP + pprof (several minutes)
make bench-k6-light                          # Same harness, lower load (WSL-friendly)
```

For a full pass before publishing numbers:

```bash
make bench-suite-x2 GO_BENCH=go1.26.4
```

---

## Troubleshooting

### Port already in use

```bash
# Find and kill existing container
docker ps
docker stop <container_id>

# Or use different port
docker run -d -p 9090:8080 gopdfsuit:6.0.0
```

### Docker login issues

```bash
# Manual login
docker login

# Then retry
make dockertag
```

### Build failures

```bash
# Clean and rebuild
make clean
make mod
make build
```

### Benchmark failures

k6 stops at about 70% with `connection reset` or `connection refused`.

The server process died mid-run, often from OOM on WSL when other benchmarks run in parallel. Stop competing jobs first, then run k6 alone with the lighter harness:

```bash
# stop competing jobs first, then:
make bench-k6-light
```

Port 8080 in use for k6 or `load-pprof`:

```bash
fuser -k 8080/tcp
# or stop the conflicting container/process, then retry
make load-pprof
```

GoPDFKit compare fails with empty PDF or module not found:

```bash
make bench-gopdfkit-setup
make bench-gopdfkit-compare-test
```

Multi-library run misses Typst or `data.json`:

```bash
make bench-setup
make bench-typst
```

k6 is not installed:

```bash
make bench-k6-install
k6 version
```
