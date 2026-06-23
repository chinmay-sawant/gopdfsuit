# Release Prep Guide

Reusable checklist for preparing a GoPdfSuit release. Based on the **v6.0.0** prep (June 2026).

Replace `X.Y.Z` / `vN` below with your target version (e.g. `6.0.0` / `v6`).

---

## PyPI auto-publish on tag

Pushing a git tag matching `v*` triggers CI (`.github/workflows/frontend-build-commit.yml`):

| Job | When | Action |
|-----|------|--------|
| `python-build` | Tag push `refs/tags/v*` | Builds CGO lib + sdist (Linux) + wheels (macOS, Windows) |
| `publish-pypi` | After `python-build` succeeds | Uploads to PyPI if `pypdfsuit==X.Y.Z` is not already published |

**Requirement:** `bindings/python/pyproject.toml` version must match the tag (e.g. tag `v6.0.0` → `version = "6.0.0"`).

Manual publish (without tag):

```bash
# GitHub Actions → CI/CD Pipeline → Run workflow → publish-pypi
```

---

## 1. Version bump (major release: `/vN` → `/vN+1`)

### Files to update

| Area | Files |
|------|-------|
| **Go module** | `go.mod`, all `*.go` imports, nested `go.mod` files |
| **Makefile** | `makefile` → `VERSION ?= X.Y.Z` |
| **Python** | `bindings/python/pyproject.toml`, `bindings/python/pypdfsuit/__init__.py` |
| **Sample data modules** | `sampledata/go.mod`, `sampledata/benchmarks/gopdfkit_compare/go.mod` |
| **Nested Go modules** | `frontend/go.mod`, `guides/go.mod`, `certs/go.mod`, `dockerfolder/go.mod`, `bkp/go.mod`, `screenshots/go.mod` |
| **Docs (source)** | `guides/GETTING_STARTED_GOPDFLIB.md`, `guides/MAKEFILE.md`, `frontend/src/components/documentation/content/getting-started.js` |
| **Docs (built)** | `docs/**` - regenerate via frontend build (do not hand-edit) |
| **Generated** | `internal/handlers/mocks/mock_services.go`, `bindings/python/pypdfsuit/lib/libgopdfsuit.so` |

### Bulk module-path migration (v5 → v6 example)

```bash
cd /path/to/gopdfsuit

# Go imports + go.mod module paths
find . -type f \( -name '*.go' -o -name 'go.mod' \) \
  ! -path './.git/*' ! -path './sampledata/benchmarks/node_modules/*' \
  -exec sed -i 's|github.com/chinmay-sawant/gopdfsuit/v5|github.com/chinmay-sawant/gopdfsuit/v6|g' {} +

# Sampledata require lines
sed -i 's|github.com/chinmay-sawant/gopdfsuit/v5 v5.0.0|github.com/chinmay-sawant/gopdfsuit/v6 v6.0.0|g' \
  sampledata/go.mod sampledata/benchmarks/gopdfkit_compare/go.mod

# App + Python version strings
sed -i 's|VERSION ?= 5.0.0|VERSION ?= 6.0.0|' makefile
sed -i 's|version = "5.0.0"|version = "6.0.0"|' bindings/python/pyproject.toml
sed -i 's|__version__ = "5.0.0"|__version__ = "6.0.0"|' bindings/python/pypdfsuit/__init__.py

# Frontend doc source + guides
sed -i 's|gopdfsuit/v5|gopdfsuit/v6|g; s|@v5.0.0|@v6.0.0|g; s|v5.0.0|v6.0.0|g' \
  frontend/src/components/documentation/content/getting-started.js
sed -i 's|gopdfsuit/v5|gopdfsuit/v6|g; s|@v5.0.0|@v6.0.0|g; s|v5.0.0|v6.0.0|g' \
  guides/GETTING_STARTED_GOPDFLIB.md
sed -i 's|5\.0\.0|6.0.0|g' guides/MAKEFILE.md

go mod tidy
```

### Verify no stale version references

```bash
# Should return no matches in maintained source
grep -r 'gopdfsuit/v5' --include='*.go' --include='go.mod' --include='*.js' --include='*.toml' --include='*.py' \
  . --exclude-dir=.git --exclude-dir=node_modules --exclude-dir=docs

grep -r 'v5\.0\.0' frontend/src guides bindings/python README.md CONTRIBUTING.md makefile
```

For a **patch/minor** release (same major module path), skip the bulk `sed` and only bump `VERSION`, `pyproject.toml`, and `__init__.py`.

---

## 2. Code quality

```bash
make fmt && make lint
```

Fix any lint issues before building artifacts. Common fixes from v6 prep:

- Remove unused vars (`unused`)
- Handle or wrap unchecked `Write` calls in hot paths (`errcheck`)
- Add `//nolint:goconst` where string literals are intentional

---

## 3. Regenerate Go mocks

```bash
go generate ./internal/handlers/...
```

Updates `internal/handlers/mocks/mock_services.go` when the module path changes.

---

## 4. Build Python bindings (`pypdfsuit`)

```bash
cd bindings/python
rm -rf build dist *.egg-info
./build.sh
pip install build
python3 -m build
ls -lh dist/
```

**Outputs:**

- `pypdfsuit/lib/libgopdfsuit.so` (or `.dylib` / `.dll` on other OS)
- `dist/pypdfsuit-X.Y.Z.tar.gz`
- `dist/pypdfsuit-X.Y.Z-*.whl`

**Test:**

```bash
python3 -m pytest bindings/python/tests -q
```

---

## 5. Rebuild frontend → `docs/`

Source lives in `frontend/src/`; the Go server and GitHub Pages serve the **built** bundle in `docs/`.

```bash
cd frontend
npm ci          # first time or after lockfile changes
npm run lint
npm run build
cd ..
```

**Verify built docs:**

```bash
grep -r 'gopdfsuit/v6' docs/assets/
grep -r 'gopdfsuit/v5' docs/   # should be empty
```

---

## 6. Full validation (recommended before tag)

```bash
make fmt && make lint
make test
go test -count=1 -v ./test
```

CI runs `make test` on every PR and branch push via the **Backend Test** job (`.github/workflows/frontend-build-commit.yml`). Enable **Branch protection → Require status checks → Backend Test** on `master` to block merges when tests fail.

Optional Docker smoke test:

```bash
VERSION=6.0.0 make docker
# or
docker build -f dockerfolder/Dockerfile --build-arg VERSION=6.0.0 -t gopdfsuit:6.0.0 .
docker run --rm -p 8080:8080 gopdfsuit:6.0.0
```

---

## 7. Publish

```bash
git add -A
git commit -m "chore: release prep v6.0.0"
git tag v6.0.0
git push origin HEAD
git push origin v6.0.0
```

**CI will then:**

1. Lint Go + frontend (on branch push, not tag-only paths)
2. Build Python wheels on Linux/macOS/Windows
3. Publish `pypdfsuit==6.0.0` to PyPI (if not already published)

Create a GitHub Release with upgrade notes (breaking change: `/v5` → `/v6` imports).

---

## v6.0.0 prep summary (what we did)

### Docs & contributor updates

| File | Change |
|------|--------|
| `README.md` | Go 1.26.4 requirement, Prerequisites, WSL note, FAQ, link to CONTRIBUTING |
| `CONTRIBUTING.md` | New contributor guide |

### Lint fixes (before version bump)

| File | Change |
|------|--------|
| `internal/pdf/font/compress_cache.go` | Removed unused `compressShardSeq` |
| `internal/pdf/generator.go` | Removed dead struct-elem pool helpers; added write helpers for errcheck |
| `internal/pdf/draw.go` | Added `//nolint:goconst` on font/alignment literals |

### Version bump v5.0.0 → v6.0.0

| File | Change |
|------|--------|
| `go.mod` | Module → `github.com/chinmay-sawant/gopdfsuit/v6` |
| All `*.go` | Imports `/v5` → `/v6` |
| `makefile` | `VERSION ?= 6.0.0` |
| `bindings/python/pyproject.toml` | `version = "6.0.0"` |
| `bindings/python/pypdfsuit/__init__.py` | `__version__ = "6.0.0"` |
| `sampledata/go.mod`, `sampledata/benchmarks/gopdfkit_compare/go.mod` | `/v6 v6.0.0` |
| `frontend/go.mod` + nested `go.mod` files | `/v6` module paths |
| `guides/GETTING_STARTED_GOPDFLIB.md`, `guides/MAKEFILE.md` | v6 install examples |
| `frontend/src/.../getting-started.js` | v6 code examples |
| `docs/**` | Rebuilt via `npm run build` |
| `bindings/python/pypdfsuit/lib/libgopdfsuit.so` | Rebuilt via `./build.sh` |
| `bindings/python/dist/*` | Generated via `python -m build` |

### Commands run (in order)

```bash
make fmt && make lint

# Version bump (see section 1)
go mod tidy
go generate ./internal/handlers/...

cd bindings/python
rm -rf build dist *.egg-info
./build.sh
pip install build
python3 -m build

python3 -m pytest bindings/python/tests -q

cd ../frontend
npm run build

make lint
```

### Post-tag checks

- [ ] PyPI: https://pypi.org/project/pypdfsuit/6.0.0/
- [ ] GitHub Actions: `python-build` + `publish-pypi` jobs green
- [ ] Docs site shows `go get .../v6@v6.0.0`
- [ ] Docker image tagged if publishing: `chinmaysawant/gopdfsuit:6.0.0`

---

## Related docs

- [RELEASE_CHECKLIST_5.0.0.md](RELEASE_CHECKLIST_5.0.0.md) - v5 major-version checklist (template for v6+)
- [MAKEFILE.md](MAKEFILE.md) - build, test, Docker, benchmark targets
- [GETTING_STARTED_GOPDFLIB.md](GETTING_STARTED_GOPDFLIB.md) - Go library install examples
- [CONTRIBUTING.md](../CONTRIBUTING.md) - dev setup and PR workflow
