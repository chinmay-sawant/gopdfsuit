## Summary

- Rebuilds the frontend site shell and tool workspaces (viewer, editor, filler, merge, split, compress, HTML convert) around shared preview and wasm-first loaders.
- Adds a shared 3-minute content-cache TTL (`internal/cachettl`, `GOPDFSUIT_CACHE_TTL` / `gopdflib.SetCacheTTL`) across font, props, image, signer, and template-data caches, plus a BOPS bypass-cache benchmark (`make bench-gopdflib-bops-x10`) and a `sampledata/caching` example.

---

## Motivation / context

- Plans: `plans/frontend/frontend-update.md`, `plans/review/go-structure-performance-review-2026-09-05.md`
- Issues: see **Related issues**

---

## Changes

### Frontend shell and workspaces

- New site header/footer, tool showcase content, GitHub stars hook with StrictMode fix.
- Shared `PdfPreview` plus conditional result layout across Merge/Split/Compress/Filler/Redact/Viewer/HtmlConvert.
- Wasm-first HTML convert, redact engine fixes (new `redact/ctm.go`), filler viewport fit, split-zip download helper.
- Layer-1 editor tests plus split-zip test; removed legacy home/documentation components.

### Engine cache TTL

- New `internal/cachettl` package: 3-minute default, env and code override, lazy expiry on lookup, no background goroutine.
- TTL wired into font subset, page compress, font object, props, image, signer/PEM, and template-data caches.
- Public API in `pkg/gopdflib/bops.go`: `ClearBOPSCaches`, `SetCacheTTL`, `CacheTTL`, `DefaultCacheTTL`.
- Docs updated in `documentation/CACHING_AND_MEMORY_LIFECYCLE.md`.

### BOPS benchmarking and example

- `internal/benchmarktemplates` BOPS runner plus `internal/pdf/bops_test.go` standard benchmarks.
- `sampledata/gopdflib/zerodha_bops` signed BOPS bench with `run_bops_x10.sh` stats (`guides/cursor/baselines/zerodha_bops_x10_wsl_stats_latest.txt`, smoke scale mean ~1231 ops/sec) and `make bench-gopdflib-bops-x10`.
- `sampledata/caching` example generating a verified 2-page `caching_example.pdf` with cold vs warm timing output.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | Warm throughput unchanged; BOPS gives cold-path truth (~1.2k ops/sec smoke scale signed retail vs ~8.8k warm). Expired entries recompute on demand, so p99 may rise just after TTL boundaries. |
| **Memory** | Bounded growth for long-lived servers: previously unbounded font-object maps now expire; other caches keep size caps plus TTL. |
| **Behavior / correctness** | Cache hits can only serve entries younger than TTL; rotated signer certs re-parse after expiry. No template-layout behavior change. |
| **API (`/api/v1/*`) / UI** | No API shape change; UI rebuilt around shared preview components. |
| **Dependencies** | None added. |
| **Binary size / build time** | Negligible (one small package, timestamp fields on cache entries). |
| **PDF compliance (PDF/A-4, PDF/UA-2)** | No emitter change; BOPS fixtures generate signed compliant retail docs. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | - |

Operators who relied on indefinite caching can set `GOPDFSUIT_CACHE_TTL=0` to restore size-only eviction.

---

## Test plan

- [x] `go vet` plus `gofmt` clean on all touched packages
- [x] `golangci-lint -E revive,gocritic,gocyclo,goconst` on `internal/cachettl`, `pkg/gopdflib`, `internal/benchmarktemplates`, `internal/pdf/font`, `internal/pdf/signature`
- [x] `go test` on `internal/cachettl`, `pkg/gopdflib`, `internal/pdf`, `internal/pdf/font`, `internal/pdf/signature`, `internal/handlers` (includes new `TestPropsCacheExpiresAfterTTL` and existing cache race/bounds tests)
- [ ] `make test` (full: `go test ./...` plus Python bindings plus `test/verify_pdfs.sh`) - only targeted packages run so far
- [ ] `make test-integration` (`go test -count=1 -v ./test`)
- [ ] `make lint` full including `frontend npm run lint`
- [ ] `make build` (`go build -o bin/app ./cmd/gopdfsuit`)
- [ ] `make test-verify-pdfs` for the new sample PDFs
- [x] BOPS x10 script at smoke scale (`BENCH_ITERATIONS=100 BENCH_WORKERS=4`), exit 0, stats file written
- [ ] Full-scale `make bench-gopdflib-bops-x10` (defaults 5000/48)

### Commands

```sh
make fmt && make lint
make test
# plus when relevant:
make test-integration
make test-verify-pdfs
```

---

## Screenshots / sample output

```
Cold:    26.688 ms  (6795 bytes, caches cleared first)
Warm:     0.113 ms  (6795 bytes, subset plus compress plus props reuse)
Saved: caching/caching_example.pdf (6795 bytes, 2 pages)
```

Fixtures: `sampledata/caching/caching_example.pdf`, `guides/cursor/baselines/zerodha_bops_x10_wsl_stats_latest.txt`.

---

## Related issues

- No pre-existing tracking issue for caching/TTL/BOPS exists (verified `gh issue list` plus keyword search; closest items #87 and #201 are unrelated to this work). No keyword applied rather than a wrong link.

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied
- [ ] Related issues filled with real ticket IDs (none exists; see above)
- [x] Filled body committed under `plans/PR/pr-<slug>.md` when process-gated

---

## Follow-ups (out of scope)

- Full-scale BOPS run (`make bench-gopdflib-bops-x10` at 5000/48) for the true WOPS/BOPS ratio.
- Full `make test`, `make test-integration`, frontend lint/build, and veraPDF verification before merge.
- Pre-existing branch hygiene (not from this session's commit): committed `docs/` build output (31 files, normally CI-owned) and 3 stray root PDFs (`zerodha_*_output.pdf` from `0b8b387`) should be reverted or moved under `sampledata/` before merge.

---

## Reviewer checklist

- [ ] Behavior matches summary and test plan
- [ ] No unrelated changes in diff
- [ ] Public API (`pkg/gopdflib`, `/api/v1/*`) or UI changes documented in `guides/` when needed
- [ ] New engine behavior has fixture coverage in `sampledata/` when applicable
- [ ] PR has assignee and labels
- [ ] Related issues use correct Closes/Relates keywords
- [ ] No secrets, certs, `.env`, `verapdf/` binaries, or generated `docs/` edits committed
