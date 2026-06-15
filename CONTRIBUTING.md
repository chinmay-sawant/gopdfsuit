# Contributing to GoPdfSuit

Thank you for your interest in contributing to **GoPdfSuit**. This guide covers setup, development workflow, and what we expect in pull requests.

## Prerequisites

| Requirement | Version / notes |
|-------------|-----------------|
| **Go** | **1.26.4** (required — matches `go.mod`) |
| **Make** | Required for build, test, and lint targets |
| **Google Chrome** | Required for HTML→PDF/Image conversion |
| **Node.js + npm** | Frontend build (Node 18+ recommended) |
| **Python 3.8+** | Python bindings tests (`pypdfsuit`) |
| **Java 11+** | Optional — needed to install veraPDF for PDF/A validation |
| **golangci-lint** | v1.64.8+ (matches CI) |

### Windows

On Windows, use **WSL (Windows Subsystem for Linux)** for the best compatibility. The project relies on **Make** and Unix shell scripts (`test/verify_pdfs.sh`, benchmark harnesses, etc.) that are not available in PowerShell or CMD. Native Windows builds are supported only for Python wheel packaging in CI.

### First-time setup

```bash
git clone https://github.com/chinmay-sawant/gopdfsuit.git
cd gopdfsuit

# Optional guided setup (frontend .env, npm install, go mod tidy)
bash setup-auth.sh

# Or manual setup:
go mod tidy
cd frontend && npm ci && cp .env.example .env && cd ..
make install-verapdf   # optional, for PDF/A validation in tests
```

Install Google Chrome on Linux:

```bash
sudo apt install -y google-chrome-stable
```

## Project overview

GoPdfSuit ships three products from one repository:

| Component | Path | Use case |
|-----------|------|----------|
| **gopdfsuit** | `cmd/gopdfsuit` | REST API + embedded React UI |
| **gopdflib** | `pkg/gopdflib` | Importable Go library |
| **pypdfsuit** | `bindings/python` | Python bindings via CGO |

Core PDF logic lives in `internal/` (private). External Go consumers should import only `pkg/gopdflib`.

## Development workflow

### Common commands

```bash
make fmt              # Format Go code
make vet              # Go static analysis
make lint             # golangci-lint (Go) + ESLint (frontend)
make test             # Go unit tests + Python pytest + veraPDF checks
make test-integration # Full test suite + HTTP integration tests
make build            # test-integration + compile binary to bin/app
make run              # test-integration + lint + frontend build + dev server
```

Recommended pre-commit cycle:

```bash
make fmt && make lint && make test
```

See [guides/MAKEFILE.md](guides/MAKEFILE.md) for the full list of Makefile targets including Docker, deployment, and benchmarks.

### Running the server

**Full dev loop** (tests + lint + frontend build):

```bash
make run
# → http://localhost:8080
```

**Split frontend/backend** (faster iteration):

```bash
# Terminal 1
go run cmd/gopdfsuit/main.go

# Terminal 2
cd frontend && npm run dev
# → http://localhost:5173
```

### Direct commands (without Make)

```bash
go build -o bin/gopdfsuit ./cmd/gopdfsuit
go test ./...
go test -count=1 -v ./test
python3 -m pytest bindings/python/tests -v
```

## Code quality

### Go

- **Format**: `make fmt` (`go fmt ./...`)
- **Lint**: `make lint` runs `golangci-lint run -E revive,gocritic,gocyclo,goconst ./...`
- **Layout**: Standard `cmd/` / `internal/` / `pkg/` structure
- **JSON hot paths**: Use [bytedance/sonic](https://github.com/bytedance/sonic), not stdlib `encoding/json`
- **Performance**: `sync.Pool`, tier-based preallocation, and buffer pooling are used in hot paths — preserve these patterns when adding code

### Frontend

- React 18 + Vite, JSX (not TypeScript)
- Lint: `cd frontend && npm run lint` (zero warnings allowed)
- Build output goes to `docs/` (served by the Go server and GitHub Pages)

### Python

- Package: `pypdfsuit` in `bindings/python/`
- Tests: `python3 -m pytest bindings/python/tests`
- Black/isort config exists in `pyproject.toml` but is not enforced in CI yet

## Testing

| Layer | Location | How to run |
|-------|----------|------------|
| Go unit tests | Colocated `*_test.go` in `internal/`, `pkg/` | `go test ./...` |
| HTTP integration | `test/` package | `go test -count=1 -v ./test` |
| Python bindings | `bindings/python/tests/` | `python3 -m pytest bindings/python/tests` |
| PDF validation | `test/verify_pdfs.sh` | `make test-verify-pdfs` |

Fixtures live under `sampledata/`. Tests write artifacts back with `temp_*` or `*_python.pdf` suffixes.

**Note:** CI currently runs lint and frontend build but does **not** run `go test` or integration tests. Run the full test suite locally before opening a PR.

## Branch naming

There is no strict policy, but prefix-based names are preferred:

| Prefix | Example | Purpose |
|--------|---------|---------|
| `feat/` | `feat/performance-improvements` | New features |
| `fix/` | `fix/xfdf-merge` | Bug fixes |
| `chore/` | `chore/update-readme` | Docs, CI, cleanup |

Base branch: **`master`**.

## Commit messages

Conventional Commits are preferred but not enforced:

```
feat(pdf): add Typst math support in table cells
fix(handlers): correct XFDF merge for compressed streams
perf(generator): reduce allocations in struct tree serialization
chore: update benchmark baselines
```

## Pull request checklist

Before opening a PR:

- [ ] `make fmt && make lint` passes
- [ ] `make test` passes (or `make test-integration` for API changes)
- [ ] New features include tests where practical
- [ ] API or template changes update relevant docs (see below)
- [ ] No secrets or credentials committed

### What CI checks

| Check | Workflow |
|-------|----------|
| Go lint (golangci-lint) | `frontend-build-commit.yml` |
| Frontend ESLint | `frontend-build-commit.yml` |
| Deslop scan | `deslop.yml` |
| Frontend auto-build → `docs/` commit | `frontend-build-commit.yml` |

## Documenting changes

| Change type | Update |
|-------------|--------|
| API endpoints / request shape | `frontend/src/components/documentation/content/api-reference.js` |
| Template JSON schema | `guides/TEMPLATE_REFERENCE.md`, `frontend/.../template-format.js` |
| gopdflib public API | `pkg/gopdflib/doc.go`, `example_test.go`, `guides/GETTING_STARTED_GOPDFLIB.md` |
| Python bindings | `bindings/python/README.md` |
| New sample templates | Add under `sampledata/<feature>/`, reference in web docs |
| Performance / benchmarks | `guides/INTEGRATION_AND_BENCHMARK_TESTS.md`, dated reports under `guides/optimizations/` |

After editing `frontend/src/`, run `npm run build` in `frontend/` and commit the generated `docs/` output, or let CI auto-commit it.

## Project structure

```
gopdfsuit/
├── cmd/gopdfsuit/           # HTTP server entrypoint
├── internal/
│   ├── handlers/            # Gin routes & HTTP handlers
│   ├── middleware/          # CORS, Google OAuth
│   ├── models/              # JSON template models
│   └── pdf/                 # Core PDF engine
├── pkg/gopdflib/            # Public Go library API
├── bindings/python/         # pypdfsuit CGO bindings
├── frontend/                # React + Vite SPA
├── docs/                    # Built frontend assets (generated)
├── test/                    # Integration tests + k6 load scripts
├── sampledata/              # Templates, fixtures, benchmarks
└── guides/                  # Reference docs and checklists
```

## Getting help

- [Web documentation](https://chinmay-sawant.github.io/gopdfsuit/#/documentation)
- [Template reference](guides/TEMPLATE_REFERENCE.md)
- [Makefile reference](guides/MAKEFILE.md)
- [FAQ](README.md#-faq)

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
