# 🛠️ Makefile Reference

Complete guide to the Makefile targets for building, testing, and deploying GoPdfSuit.

---

## Table of Contents

- [Overview](#overview)
- [Variables](#variables)
- [Docker Targets](#docker-targets)
- [Build Targets](#build-targets)
- [Development Targets](#development-targets)
- [Benchmark Targets](#benchmark-targets)
- [Quick Reference](#quick-reference)

---

## Overview

The Makefile provides convenient shortcuts for common development and deployment tasks. All targets are designed to work from the repository root directory.

---

## Variables

Customize behavior using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `VERSION` | `5.0.0` | Application version tag |
| `DOCKERUSERNAME` | `chinmaysawant` | Docker Hub username |
| `GO_BENCH` | `go` | Go binary for benchmarks (`go1.26.4` recommended) |
| `GOMAXPROCS_BENCH` | `24` | CPU limit for Go benchmark processes |
| `BENCH_ITERATIONS` | `5000` | Zerodha / data-table iteration count |
| `BENCH_WORKERS` | `48` | Concurrent workers for Zerodha harnesses |
| `BENCH_COUNT` | `1` | `go test -count` repetitions |
| `BENCH_TIME` | `5s` | `go test -benchtime` for GoPDFKit compare |
| `LOAD_VUS` | `48` | k6 virtual users (`run_gin_pprof_load.sh`) |
| `PROFILE_SECONDS` | `35` | CPU profile duration during k6 load |
| `PAYLOAD_SCENARIO` | `tagged_ecdsa` | k6 payload mix (`retail_only_signed` for gate) |
| `THROUGHPUT_GATE` | `0` | Minimum req/s gate for k6 script (0 = off) |
| `BASE_URL` | `http://127.0.0.1:8080` | Target URL for k6 / pprof script |

### Setting Variables

```bash
# Inline
VERSION=5.0.0 make docker

# Export for session
export VERSION=5.0.0
export DOCKERUSERNAME=myusername
make docker
```

---

## Docker Targets

### `make docker`

Build and run the Docker container locally.

```bash
make docker
```

**What it does:**
1. Builds Docker image from `dockerfolder/Dockerfile`
2. Tags image as `gopdfsuit:<VERSION>`
3. Runs container on port 8080

**Equivalent commands:**
```bash
docker build -f dockerfolder/Dockerfile --build-arg VERSION=5.0.0 -t gopdfsuit:5.0.0 .
docker run -d -p 8080:8080 gopdfsuit:5.0.0
```

---

### `make dockertag`

Tag and push the Docker image to Docker Hub.

```bash
make dockertag
```

**What it does:**
1. Tags image with version number
2. Tags image as `latest`
3. Prompts for Docker Hub login
4. Pushes both tags to Docker Hub

**Equivalent commands:**
```bash
docker tag gopdfsuit:5.0.0 chinmaysawant/gopdfsuit:5.0.0
docker tag gopdfsuit:5.0.0 chinmaysawant/gopdfsuit:latest
docker login
docker push chinmaysawant/gopdfsuit:5.0.0
docker push chinmaysawant/gopdfsuit:latest
```

---

### `make pull`

Pull and run the latest image from Docker Hub.

```bash
make pull
```

**What it does:**
1. Pulls image from Docker Hub
2. Runs container on port 8080

**Equivalent commands:**
```bash
docker pull chinmaysawant/gopdfsuit:5.0.0
docker run -d -p 8080:8080 chinmaysawant/gopdfsuit:5.0.0
```

---

## Build Targets

### `make build`

Compile the Go application.

```bash
make build
```

**What it does:**
1. Creates `bin/` directory
2. Compiles to `bin/app`

**Output:** `bin/app`

---

### `make run`

Build frontend and run the application locally.

```bash
make run
```

**What it does:**
1. Builds React frontend (`npm run build`)
2. Runs Go application

**Access:** `http://localhost:8080`

---

### `make clean`

Remove build artifacts.

```bash
make clean
```

**What it does:**
- Deletes `bin/` directory

---

## Development Targets

### `make test`

Run all Go tests.

```bash
make test
```

**Equivalent:**
```bash
go test ./...
```

---

### `make fmt`

Format Go source code.

```bash
make fmt
```

**Equivalent:**
```bash
go fmt ./...
```

---

### `make vet`

Run Go static analysis.

```bash
make vet
```

**Equivalent:**
```bash
go vet ./...
```

---

### `make mod`

Tidy Go module dependencies.

```bash
make mod
```

**Equivalent:**
```bash
go mod tidy
```

---

## Benchmark Targets

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

### gopdfsuit HTTP (k6 + pprof)

The `load-pprof*` and `bench-k6*` targets (except `bench-k6-load`, `bench-k6-smoke`, `bench-k6-spike`, `bench-k6-soak`) build the server, start it on port 8080, run k6, and capture CPU/heap profiles under `guides/cursor/baselines/gin_pprof_runs/`.

| Target | Description |
|--------|-------------|
| `make load-pprof` | Weighted `tagged_ecdsa` workload, 48 VU × 35s + pprof (alias: `bench-k6`) |
| `make load-pprof-gate` | Retail-only signed fast path, 1500 req/s gate (alias: `bench-k6-retail`) |
| `make load-pprof-1k` | Weighted workload, 1000 req/s gate (alias: `bench-k6-1k`) |
| `make load-pprof-1500` | Weighted workload, 1500 req/s gate (alias: `bench-k6-1500`) |
| `make bench-k6-load` | k6 `load_test.js` only — start server yourself |
| `make bench-k6-smoke` | Quick `smoke_test.js` (10s, 1 VU) |
| `make bench-k6-spike` | Traffic spike simulation (`spike_test.js`) |
| `make bench-k6-soak` | 30-minute stability test (`soak_test.js`) |

**k6-only workflow** (no auto server):

```bash
go run ./cmd/gopdfsuit &
make bench-k6-smoke
```

**Custom gate example:**

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
| `make bench-pypdfsuit-zerodha-x2` | Two sequential runs |

**Override iterations/workers:**

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

> **Note:** `bench-pypdfsuit-legacy` uses the older `sampledata/benchmarks/pypdfsuit` harness. For Zerodha parity against gopdflib, prefer `bench-pypdfsuit-zerodha`.

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

Long-running; run when you want a full regression pass.

| Target | Includes |
|--------|----------|
| `make bench-suite` | Zerodha gopdflib + pypdfsuit + GoPDFKit compare + handler benches + k6/pprof |
| `make bench-suite-x2` | Two passes of each harness in `bench-suite` |
| `make bench-suite-full` | `bench-suite` + `bench-all-libraries` |

---

## Quick Reference

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
| `make bench-gopdflib-zerodha` | Zerodha gold standard | Library throughput |

---

## Common Workflows

### Development Cycle

```bash
make fmt          # Format code
make vet          # Check for issues
make test         # Run tests
make run          # Start development server
```

### Release Workflow

```bash
export VERSION=5.0.0
make docker       # Build and test locally
make dockertag    # Push to Docker Hub
```

### Fresh Start

```bash
make clean        # Remove old builds
make mod          # Update dependencies
make build        # Compile fresh binary
```

### Benchmark Regression

```bash
make bench-help                              # See all targets
make bench-setup                             # Typst + data.json (multi-library)
make bench-gopdfkit-setup                    # GoPDFKit symlink (compare only)
make bench-gopdflib-zerodha-x2               # Quick Zerodha check (2 runs)
make bench-handler-all GO_BENCH=go1.26.4     # Handler micro-bench
make load-pprof                              # Full HTTP + pprof (several minutes)
```

For a full pass before publishing numbers:

```bash
make bench-suite-x2 GO_BENCH=go1.26.4
```

---

## Troubleshooting

### Port Already in Use

```bash
# Find and kill existing container
docker ps
docker stop <container_id>

# Or use different port
docker run -d -p 9090:8080 gopdfsuit:5.0.0
```

### Docker Login Issues

```bash
# Manual login
docker login

# Then retry
make dockertag
```

### Build Failures

```bash
# Clean and rebuild
make clean
make mod
make build
```

### Benchmark Failures

**Port 8080 in use** (k6 / `load-pprof`):

```bash
fuser -k 8080/tcp
# or stop the conflicting container/process, then retry
make load-pprof
```

**GoPDFKit compare: empty PDF / module not found:**

```bash
make bench-gopdfkit-setup
make bench-gopdfkit-compare-test
```

**Multi-library Typst / data.json missing:**

```bash
make bench-setup
make bench-typst
```

**k6 not installed:**

```bash
make bench-k6-install
k6 version
```
