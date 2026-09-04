VERSION ?= 6.0.0
DOCKERUSERNAME ?= chinmaysawant

docker:
	docker build -f dockerfolder/Dockerfile --build-arg VERSION=$(VERSION) -t gopdfsuit:$(VERSION) .
	docker run -d -p 8080:8080 gopdfsuit:$(VERSION)

dockertag:
	docker tag gopdfsuit:$(VERSION) $(DOCKERUSERNAME)/gopdfsuit:$(VERSION)
	docker tag gopdfsuit:$(VERSION) $(DOCKERUSERNAME)/gopdfsuit:latest
	docker login
	docker push $(DOCKERUSERNAME)/gopdfsuit:$(VERSION)
	docker push $(DOCKERUSERNAME)/gopdfsuit:latest

pull:
	docker pull $(DOCKERUSERNAME)/gopdfsuit:$(VERSION)
	docker run -d -p 8080:8080 $(DOCKERUSERNAME)/gopdfsuit:$(VERSION)

build: test-go
	mkdir -p bin
	go build -o bin/app ./cmd/gopdfsuit

# ── Test harness (OOM-safe, filterable) ────────────────────────────────────
# Defaults cap package parallelism so a 24-CPU / 8GB box does not OOM:
#   go test -p limits concurrent package binaries (biggest RAM saver),
#   -parallel limits parallel subtests inside a package,
#   GOMEMLIMIT makes the Go runtime GC earlier under pressure,
#   VERIFY_PDFS_JOBS caps concurrent veraPDF JVMs (each ~0.5-1GB).
#
# Usage (tiers - cheapest first):
#   make test-fast                # tier 1 (<2 min): SHORT=1 unit + python, no verify
#   make test-prepush             # tier 2: test-fast + ./test integration (run before push)
#   make test                     # tier 3 (full): go + python + verify (sequential)
#   make test-nightly             # tier 4: full test + scan-all-compliance + race
#   make test-go                  # go only (unit pkgs, then ./test integration)
#   make test-unit                # go only, skip ./test integration package
#   make test-integration         # ./test split: suite, then zerodha (standalone)
#   make test-integration-suite   # only TestIntegrationSuite process
#   make test-integration-zerodha # only TestZerodhaPDFCompliance process
#   make test-python              # python only
#   make test-verify-pdfs         # veraPDF verify only (test-verify is an alias)
#   make test-fast                # tier 1: SHORT=1 unit + python, no verify
#   make test-prepush             # tier 2: test-fast + test-integration
#   make test-nightly             # tier 4: full test + scan-all-compliance + race
#   make test-list                # list packages and test names
#
# Filters (all optional, overridable on CLI):
#   make test-go P=1 PKG=./internal/pdf T=TestWrap
#   make test-go JOBS=4 RUN=TestMerge ./... (JOBS/RUN are aliases of P/T)
#   make test-python PY_K=test_merge
#   make test SHORT=1 V=1 COUNT=2
JOBS ?= 2
P ?= $(JOBS)
RUN ?=
T ?= $(RUN)
PKG ?= ./...
PY_K ?= $(T)
SHORT ?= 0
V ?= 0
COUNT ?= 1
TEST_PARALLEL ?= 2
TEST_TIMEOUT ?= 10m
TEST_GOMEMLIMIT ?= 4GiB
TEST_GOMAXPROCS ?=
PYTEST_FLAGS ?=
ifeq ($(SHORT),1)
GO_SHORT_FLAG := -short
else
GO_SHORT_FLAG :=
endif
ifeq ($(V),1)
GO_V_FLAG := -v
PY_V_FLAG := -v
else
GO_V_FLAG :=
PY_V_FLAG := -q
endif
GO_RUN_FLAG := $(if $(strip $(T)),-run '$(T)')
PY_K_FLAG := $(if $(strip $(PY_K)),-k '$(PY_K)')
GO_TEST_ENV := GOMEMLIMIT=$(TEST_GOMEMLIMIT) $(if $(strip $(TEST_GOMAXPROCS)),GOMAXPROCS=$(TEST_GOMAXPROCS) ,)

test: test-go test-python test-verify

test-go:
ifeq ($(PKG),./...)
	@$(MAKE) test-unit P=$(P) TEST_PARALLEL=$(TEST_PARALLEL) COUNT=$(COUNT) V=$(V) SHORT=$(SHORT) T='$(T)'
	@$(MAKE) test-integration P=$(P) TEST_PARALLEL=$(TEST_PARALLEL) COUNT=$(COUNT) V=$(V) SHORT=$(SHORT) T='$(T)'
else
	$(GO_TEST_ENV) go test -count=$(COUNT) -timeout $(TEST_TIMEOUT) -p $(P) -parallel $(TEST_PARALLEL) $(GO_SHORT_FLAG) $(GO_V_FLAG) $(GO_RUN_FLAG) $(PKG)
endif

test-unit:
	$(GO_TEST_ENV) go test -count=$(COUNT) -timeout $(TEST_TIMEOUT) -p $(P) -parallel $(TEST_PARALLEL) $(GO_SHORT_FLAG) $(GO_V_FLAG) $(GO_RUN_FLAG) $$(go list ./... | grep -v '/test$$')

# Integration runs each top-level test in its own process, sequentially, so
# peak RAM stays flat: TestIntegrationSuite (25 sequential HTTP/library
# subtests) first, then TestZerodhaPDFCompliance (6 veraPDF JVM subtests
# capped by -parallel). Pass T= to filter to a single process instead.
test-integration:
ifeq ($(strip $(T)),)
	@$(MAKE) test-integration-suite P=$(P) TEST_PARALLEL=$(TEST_PARALLEL) COUNT=$(COUNT) V=$(V) SHORT=$(SHORT)
	@$(MAKE) test-integration-zerodha P=$(P) TEST_PARALLEL=$(TEST_PARALLEL) COUNT=$(COUNT) V=$(V) SHORT=$(SHORT)
else
	$(GO_TEST_ENV) go test -count=$(COUNT) -timeout $(TEST_TIMEOUT) -p 1 -parallel $(TEST_PARALLEL) $(GO_SHORT_FLAG) -v $(GO_RUN_FLAG) ./test
endif

test-integration-suite:
	$(GO_TEST_ENV) go test -count=$(COUNT) -timeout $(TEST_TIMEOUT) -p 1 -parallel 1 $(GO_SHORT_FLAG) -v $(if $(strip $(T)),$(GO_RUN_FLAG),-run 'TestIntegrationSuite') ./test

test-integration-zerodha:
	$(GO_TEST_ENV) go test -count=$(COUNT) -timeout $(TEST_TIMEOUT) -p 1 -parallel $(TEST_PARALLEL) $(GO_SHORT_FLAG) -v $(if $(strip $(T)),$(GO_RUN_FLAG),-run 'TestZerodhaPDFCompliance') ./test

test-python:
	cd bindings/python && python3 -m pytest tests $(PY_V_FLAG) $(PY_K_FLAG) $(PYTEST_FLAGS)

# Canonical verify target. test-verify stays as an alias so existing CI
# and muscle memory keep working (row 5.5: one name, one alias).
test-verify-pdfs:
	VERIFY_PDFS_JOBS=$(P) bash test/verify_pdfs.sh

test-verify: test-verify-pdfs

test-fast:
	@$(MAKE) test-unit SHORT=1 P=$(P) PKG='$(PKG)' T='$(T)' COUNT=$(COUNT) V=$(V)
	@$(MAKE) test-python PY_K='$(PY_K)' PYTEST_FLAGS='$(PYTEST_FLAGS)'

test-prepush: test-fast test-integration

test-nightly: test test-scan-pdfs-compliance test-race

test-list:
	@go list ./...
	@echo '---'
	@$(GO_TEST_ENV) go test -p $(P) ./... -list '.*' 2>/dev/null | grep -E '^(ok|Testing|Test|Benchmark|Example)' | head -n 100

# D1 triple-gate: frontend schema self-check against the golden fixture
# (also covered by Go TestTemplateSchemaGolden and Python
# test_golden_template.py). CI runs this in the frontend-lint job.
test-schema:
	cd frontend && npm run test:schema

# D4 smoke: one quick slice per headless harness, total well under 5 min.
# k6/Gotenberg/gopdfkit-compare need live infra (server :8080, Docker,
# network setup) - run bench-k6-smoke / bench-gotenberg-smoke /
# bench-gopdfkit-compare-test against running infra instead.
bench-smoke:
	cd $(ZERODHA_DIR) && BENCH_SEED=42 BENCH_ITERATIONS=20 BENCH_WORKERS=2 BENCH_SKIP_WRITE=1 $(GO_BENCH) run .
	cd $(ZERODHA_DIR) && PAYLOAD_SCENARIO=retail_only BENCH_ITERATIONS=20 BENCH_WORKERS=2 python3 pypdfsuit_bench.py
	GOMAXPROCS=$(GOMAXPROCS_BENCH) $(GO_BENCH) test -bench='BenchmarkGenerateTemplatePDF_FinancialReport$$' -benchtime=10x -count=1 ./test

test-race:
	$(GO_TEST_ENV) go test -race -p $(P) -parallel $(TEST_PARALLEL) $(GO_RUN_FLAG) ./internal/pdf/... ./internal/handlers/...

install-verapdf:
	bash test/install_verapdf.sh

install-pdf-validators:
	bash test/install_pdf_validators.sh

test-scan-pdfs:
	VERIFY_PDFS_JOBS=$(P) bash test/verify_pdfs.sh --scan-all

test-scan-pdfs-compliance:
	VERIFY_PDFS_JOBS=$(P) bash test/verify_pdfs.sh --scan-all-compliance

test-zerodha-compliance:
	VERIFY_PDFS_JOBS=$(P) bash test/verify_pdfs.sh --zerodha-only

clean:
	rm -rf bin/

# Go 1.24+ ships wasm_exec.js under lib/wasm; older toolchains use misc/wasm.
WASM_EXEC := $(shell go env GOROOT)/lib/wasm/wasm_exec.js
ifeq ($(wildcard $(WASM_EXEC)),)
WASM_EXEC := $(shell go env GOROOT)/misc/wasm/wasm_exec.js
endif

# In-browser compressor: GOOS=js GOARCH=wasm, no gin / browser deps.
# HTML conversion is server-side pure-Go via gowkhtmltopdf, no Chrome needed.
# Output goes to frontend/public only; sampledata harnesses copy it locally
# for runs (gitignored, row 5.6 - wasm is never committed under sampledata/).
wasm-compress:
	mkdir -p frontend/public
	GOOS=js GOARCH=wasm go build -o frontend/public/compress.wasm ./cmd/wasmcompress
	cp "$(WASM_EXEC)" frontend/public/wasm_exec.js
	file frontend/public/compress.wasm

# Full browser bundle: Generate, Merge, Split, Compress, Fill, text-path Redact.
# HTML conversion and the OCR subprocess stay server-side and are never
# referenced by ./cmd/wasm (see plans/wasm/01-full-wasm-port.md Phase 2.1).
wasm:
	mkdir -p frontend/public
	GOOS=js GOARCH=wasm go build -o frontend/public/gopdfsuit.wasm ./cmd/wasm
	cp "$(WASM_EXEC)" frontend/public/wasm_exec.js
	file frontend/public/gopdfsuit.wasm

.PHONY: wasm wasm-compress

run: test-integration lint wasm-compress
	export VITE_IS_CLOUD_RUN=false;\
	export VITE_ENVIRONMENT=local;\
	export VITE_API_URL=http://localhost:8080;\
	cd frontend && npm run build && cd ..
	go run cmd/gopdfsuit/main.go

fmt:
	go fmt ./...

vet:
	go vet ./...

mod:
	go mod tidy

lint:
	golangci-lint run --timeout 10m -E revive,gocritic,gocyclo,goconst ./...
	cd frontend && npm run lint
	cd .. 

gdocker: test-integration wasm-compress
	cd frontend && npm run build && cd ..
	docker rm -f gopdfsuit
	docker build -t gopdfsuit . 

gdocker-run:
	docker run --rm -p 8080:8080 -d --name gopdfsuit gopdfsuit

gdocker-push:
	export VITE_IS_CLOUD_RUN=true;\
	export VITE_ENVIRONMENT=cloudrun;\
	gcloud builds submit --tag us-east1-docker.pkg.dev/gopdfsuit/gopdfsuit/gopdfsuit-app .	
	gcloud run deploy gopdfsuit-service \
    --image us-east1-docker.pkg.dev/gopdfsuit/gopdfsuit/gopdfsuit-app \
    --region us-east1 \
    --platform managed \
    --allow-unauthenticated \
    --max-instances 1 \
    --concurrency 80 \
    --cpu 1 \
    --memory 512Mi \
	--env-vars-file .env

gengine-deploy: test-integration
	export VITE_IS_CLOUD_RUN=true;\
	export VITE_ENVIRONMENT=cloudrun;\
	export DISABLE_PROFILING=true;\
	cd frontend && npm run build && cd ..
	gcloud app deploy

# ── Benchmark defaults (override on CLI, e.g. make bench-gopdflib-zerodha BENCH_ITERATIONS=1000) ──
GO_BENCH ?= go
GOMAXPROCS_BENCH ?= 24
BENCH_ITERATIONS ?= 5000
BENCH_WORKERS ?= 48
BENCH_COUNT ?= 1
BENCH_TIME ?= 5s
ZERODHA_DIR := sampledata/gopdflib/zerodha
GPDF_ZERODHA_DIR := sampledata/gpdf/zerodha
BENCHMARKS_DIR := sampledata/benchmarks
GOPDFKIT_COMPARE_DIR := $(BENCHMARKS_DIR)/gopdfkit_compare
GOTENBERG_DIR := $(BENCHMARKS_DIR)/gotenberg
K6_DIR := test/generate_template-pdf
# bench-k6-light defaults (override on CLI, e.g. K6_LIGHT_VUS=16 make bench-k6-light)
K6_LIGHT_VUS ?= 24
K6_LIGHT_SECONDS ?= 15
K6_LIGHT_MAX_CONCURRENT ?= 24
K6_LIGHT_GOMAXPROCS ?= 12

.PHONY: build test test-go test-unit test-integration test-integration-suite test-integration-zerodha test-python test-verify test-fast test-prepush test-nightly test-list test-schema bench-smoke test-race install-verapdf install-pdf-validators test-verify-pdfs test-scan-pdfs test-scan-pdfs-compliance test-zerodha-compliance clean run fmt vet mod lint \
	load-pprof load-pprof-gate load-pprof-1k load-pprof-1500 \
	bench-help bench-setup \
	bench-k6 bench-k6-light bench-k6-retail bench-k6-1k bench-k6-1500 bench-k6-load \
	bench-k6-smoke bench-k6-spike bench-k6-soak bench-k6-install \
	bench-gotenberg bench-gotenberg-load bench-gotenberg-smoke bench-gotenberg-start \
	bench-gopdflib-zerodha bench-gopdflib-zerodha-nocomply bench-gopdflib-zerodha-x2 bench-gopdflib-zerodha-x5 bench-gopdflib-zerodha-x10 bench-gopdflib-zerodha-x10-pprof \
	bench-gopdflib-zerodha-nocomply-x10 \
	bench-gpdf-zerodha bench-gpdf-zerodha-nocomply bench-gpdf-zerodha-x10 \
	bench-gopdflib-data bench-gopdflib-data-pprof \
	bench-gopdfsuit-zerodha bench-pypdfsuit-zerodha bench-pypdfsuit-zerodha-x2 \
	bench-pypdfsuit-zerodha-nocomply bench-pypdfsuit-zerodha-nocomply-x10 \
	bench-pypdfsuit-zerodha-x5 bench-pypdfsuit-zerodha-x10 bench-pypdfsuit-zerodha-x10-pprof \
	bench-pypdfsuit-profile bench-pypdfsuit-legacy \
	bench-fpdf bench-jspdf bench-pdfkit-lib bench-pdflib bench-typst bench-all-libraries \
	bench-gopdfkit-setup bench-gopdfkit-compare bench-gopdfkit-compare-x2 \
	bench-gopdfkit-compare-test bench-gopdfkit-html \
	bench-handler bench-handler-parallel bench-handler-all \
	bench-pdf-micro bench-pdf-macro bench-pdf-typst \
	bench-suite bench-suite-x2 bench-suite-full

bench-help:
	@echo "Benchmark targets (Go 1.26.4 recommended: GO_BENCH=go1.26.4)"
	@echo ""
	@echo "  Overrides: GO_BENCH GOMAXPROCS_BENCH BENCH_ITERATIONS BENCH_WORKERS BENCH_COUNT BENCH_TIME"
	@echo "             LOAD_VUS PROFILE_SECONDS PAYLOAD_SCENARIO THROUGHPUT_GATE BASE_URL (k6/pprof script)"
	@echo "             K6_LIGHT_VUS K6_LIGHT_SECONDS K6_LIGHT_MAX_CONCURRENT K6_LIGHT_GOMAXPROCS (bench-k6-light)"
	@echo ""
	@echo "  Setup:"
	@echo "    make bench-setup                 # Typst binary + data.json (sampledata/benchmarks)"
	@echo "    make bench-k6-install            # install k6 on Debian/Ubuntu WSL"
	@echo ""
	@echo "  Targets below are generated from this makefile (row 5.5: no hand-kept list):"
	@echo ""
	@grep -E '^(bench|load)-[A-Za-z0-9-]+:' makefile | sed -E 's/:.*//' | sort -u | \
		awk -F'-' '{key = $$2; sub(/^bench$$/, "misc", key); print key" "$$0}' | sort | \
		awk '{if ($$1 != prev) {printf "  %s:\n", $$1; prev = $$1} printf "    make %s\n", $$2}'

# ── gopdfsuit: k6 HTTP load + pprof ──────────────────────────────────────────

# Steady-state k6 + CPU/heap pprof for /api/v1/generate/template-pdf
load-pprof: bench-k6

bench-k6:
	bash $(K6_DIR)/run_gin_pprof_load.sh

# Reduced load for WSL / shared machines / running alongside other benchmarks
bench-k6-light:
	GOMAXPROCS=$(K6_LIGHT_GOMAXPROCS) MAX_CONCURRENT=$(K6_LIGHT_MAX_CONCURRENT) \
		LOAD_VUS=$(K6_LIGHT_VUS) PROFILE_SECONDS=$(K6_LIGHT_SECONDS) \
		bash $(K6_DIR)/run_gin_pprof_load.sh

# Gate run: retail-only sanity (≥1500 req/s target on fast path)
load-pprof-gate: bench-k6-retail

bench-k6-retail:
	PAYLOAD_SCENARIO=retail_only_signed THROUGHPUT_GATE=1500 bash $(K6_DIR)/run_gin_pprof_load.sh

# Weighted workload gate toward 1000+ req/s
load-pprof-1k: bench-k6-1k

bench-k6-1k:
	THROUGHPUT_GATE=1000 bash $(K6_DIR)/run_gin_pprof_load.sh

# Weighted workload gate toward 1500+ req/s
load-pprof-1500: bench-k6-1500

bench-k6-1500:
	THROUGHPUT_GATE=1500 bash $(K6_DIR)/run_gin_pprof_load.sh

# k6 only (no pprof script); start server separately: go run ./cmd/gopdfsuit
bench-k6-load:
	cd $(K6_DIR) && k6 run load_test.js

bench-k6-smoke:
	cd $(K6_DIR) && k6 run smoke_test.js

bench-k6-spike:
	cd $(K6_DIR) && k6 run spike_test.js

bench-k6-soak:
	cd $(K6_DIR) && k6 run soak_test.js

bench-k6-install:
	bash $(K6_DIR)/install_k6.sh

# ── Gotenberg: k6 HTML→PDF (Chromium) ────────────────────────────────────────

bench-gotenberg:
	bash $(GOTENBERG_DIR)/run_gotenberg_load.sh

bench-gotenberg-load:
	cd $(GOTENBERG_DIR) && k6 run load_test.js

bench-gotenberg-smoke:
	cd $(GOTENBERG_DIR) && k6 run smoke_test.js

bench-gotenberg-start:
	bash $(GOTENBERG_DIR)/start_gotenberg.sh

# go tool pprof -http=:8081 "http://localhost:8080/debug/pprof/profile?seconds=30"
# go tool pprof -http=:8081 "http://localhost:8080/debug/pprof/heap"

# ── Benchmark data setup (Typst + data.json) ─────────────────────────────────

bench-setup:
	bash $(BENCHMARKS_DIR)/setup_benchmarks.sh

# ── gopdflib: Zerodha gold standard ──────────────────────────────────────────

bench-gopdflib-zerodha:
	cd $(ZERODHA_DIR) && GOMAXPROCS=$(GOMAXPROCS_BENCH) BENCH_ITERATIONS=$(BENCH_ITERATIONS) BENCH_WORKERS=$(BENCH_WORKERS) $(GO_BENCH) run .

bench-gopdflib-zerodha-nocomply:
	cd $(ZERODHA_DIR) && GOMAXPROCS=$(GOMAXPROCS_BENCH) BENCH_ITERATIONS=$(BENCH_ITERATIONS) BENCH_WORKERS=$(BENCH_WORKERS) $(GO_BENCH) run -tags nocomply .

bench-gopdflib-zerodha-nocomply-x10:
	bash $(ZERODHA_DIR)/run_bench_x10_nocomply.sh

bench-gopdflib-zerodha-x2:
	@for i in 1 2; do \
		echo "=== gopdflib zerodha run $$i / 2 ==="; \
		$(MAKE) bench-gopdflib-zerodha; \
	done

bench-gopdflib-zerodha-x5:
	bash $(ZERODHA_DIR)/run_bench_x5.sh

bench-gopdflib-zerodha-x10:
	bash $(ZERODHA_DIR)/run_bench_x10.sh

bench-gopdflib-zerodha-x10-pprof: bench-gopdflib-zerodha-x10 bench-gopdflib-zerodha-x5

# ── gpdf: Zerodha gold standard ──────────────────────────────────────────────

bench-gpdf-zerodha:
	cd $(GPDF_ZERODHA_DIR) && GOMAXPROCS=$(GOMAXPROCS_BENCH) BENCH_ITERATIONS=$(BENCH_ITERATIONS) BENCH_WORKERS=$(BENCH_WORKERS) $(GO_BENCH) run .

bench-gpdf-zerodha-nocomply:
	cd $(GPDF_ZERODHA_DIR) && GOMAXPROCS=$(GOMAXPROCS_BENCH) BENCH_ITERATIONS=$(BENCH_ITERATIONS) BENCH_WORKERS=$(BENCH_WORKERS) $(GO_BENCH) run -tags nocomply .

bench-gpdf-zerodha-x10:
	bash $(GPDF_ZERODHA_DIR)/run_bench_x10.sh

# ── GoPDFLib data-table (tabular workload) ───────────────────────────────────

bench-gopdflib-data:
	cd $(BENCHMARKS_DIR)/gopdflib && GOMAXPROCS=$(GOMAXPROCS_BENCH) GOWORK=off \
		BENCH_ITERATIONS=$(BENCH_ITERATIONS) BENCH_WORKERS=$(BENCH_WORKERS) \
		$(GO_BENCH) run . data

bench-gopdflib-data-pprof:
	bash $(BENCHMARKS_DIR)/gopdflib/run_pprof_bench.sh

# ── pypdfsuit: Zerodha gold standard (Python) ────────────────────────────────

bench-pypdfsuit-zerodha:
	cd $(ZERODHA_DIR) && BENCH_ITERATIONS=$(BENCH_ITERATIONS) BENCH_WORKERS=$(BENCH_WORKERS) python3 pypdfsuit_bench.py

bench-pypdfsuit-zerodha-nocomply:
	cd $(ZERODHA_DIR) && BENCH_NOCOMPLY=1 BENCH_ITERATIONS=$(BENCH_ITERATIONS) BENCH_WORKERS=$(BENCH_WORKERS) python3 pypdfsuit_bench.py

bench-pypdfsuit-zerodha-nocomply-x10:
	bash $(ZERODHA_DIR)/run_pypdfsuit_bench_x10_nocomply.sh

bench-pypdfsuit-zerodha-x2:
	@for i in 1 2; do \
		echo "=== pypdfsuit zerodha run $$i / 2 ==="; \
		$(MAKE) bench-pypdfsuit-zerodha; \
	done

bench-pypdfsuit-profile:
	cd $(ZERODHA_DIR) && python3 pypdfsuit_profile.py

bench-pypdfsuit-zerodha-x5:
	bash $(ZERODHA_DIR)/run_pypdfsuit_bench_x5.sh

bench-pypdfsuit-zerodha-x10:
	bash $(ZERODHA_DIR)/run_pypdfsuit_bench_x10.sh

bench-pypdfsuit-zerodha-x10-pprof: bench-pypdfsuit-zerodha-x10 bench-pypdfsuit-zerodha-x5

# ── Multi-library benchmarks (sampledata/benchmarks) ───────────────────────

bench-gopdfsuit-zerodha:
	cd $(BENCHMARKS_DIR)/gopdfsuit && GOMAXPROCS=$(GOMAXPROCS_BENCH) $(GO_BENCH) run .

bench-pypdfsuit-legacy:
	cd $(BENCHMARKS_DIR)/pypdfsuit && python3 bench.py

bench-fpdf:
	cd $(BENCHMARKS_DIR)/fpdf && python3 bench.py

bench-jspdf:
	cd $(BENCHMARKS_DIR)/jspdf && node bench.js

bench-pdfkit-lib:
	cd $(BENCHMARKS_DIR)/pdfkit && node bench.js

bench-pdflib:
	cd $(BENCHMARKS_DIR)/pdflib && node bench.js

bench-typst: bench-setup
	cd $(BENCHMARKS_DIR)/typst && bash bench.sh

bench-all-libraries:
	bash $(BENCHMARKS_DIR)/run_all_benchmarks.sh

# ── GoPDFKit vs gopdflib compare ─────────────────────────────────────────────

bench-gopdfkit-setup:
	$(GO_BENCH) mod download github.com/cssbruno/gopdfkit@v0.5.2
	mkdir -p /tmp/gopdfkit-real/github.com/cssbruno
	ln -sfn "$$($(GO_BENCH) list -m -f '{{.Dir}}' github.com/cssbruno/gopdfkit@v0.5.2)" \
		/tmp/gopdfkit-real/github.com/cssbruno/gopdfkit@v0.5.2

bench-gopdfkit-compare-test: bench-gopdfkit-setup
	cd $(GOPDFKIT_COMPARE_DIR) && $(GO_BENCH) test -run '^TestComparableOutputsArePDF$$' -count=1

bench-gopdfkit-compare: bench-gopdfkit-setup
	cd $(GOPDFKIT_COMPARE_DIR) && GOMAXPROCS=$(GOMAXPROCS_BENCH) \
		$(GO_BENCH) test -run '^$$' -bench 'BenchmarkGoPDF(Kit|Lib)$$' -benchmem \
		-benchtime=$(BENCH_TIME) -count=$(BENCH_COUNT)

bench-gopdfkit-html: bench-gopdfkit-setup
	cd $(GOPDFKIT_COMPARE_DIR) && GOPDFKIT_COMPARE_HTML=1 GOMAXPROCS=$(GOMAXPROCS_BENCH) \
		$(GO_BENCH) test -run '^$$' -bench 'HTML' -benchmem \
		-benchtime=$(BENCH_TIME) -count=$(BENCH_COUNT)

bench-gopdfkit-compare-x2:
	@for i in 1 2; do \
		echo "=== gopdfkit compare run $$i / 2 ==="; \
		$(MAKE) bench-gopdfkit-compare; \
	done

# ── Go test: Gin handler + internal/pdf micro ───────────────────────────────

bench-handler:
	GOMAXPROCS=$(GOMAXPROCS_BENCH) $(GO_BENCH) test -bench='BenchmarkGenerateTemplatePDF_FinancialReport$$' \
		-benchmem -count=$(BENCH_COUNT) ./test

bench-handler-parallel:
	GOMAXPROCS=$(GOMAXPROCS_BENCH) $(GO_BENCH) test -bench='BenchmarkGenerateTemplatePDF_FinancialReport_Parallel$$' \
		-benchmem -count=$(BENCH_COUNT) ./test

bench-handler-all:
	GOMAXPROCS=$(GOMAXPROCS_BENCH) $(GO_BENCH) test -bench='BenchmarkGenerateTemplatePDF_FinancialReport' \
		-benchmem -count=$(BENCH_COUNT) ./test

bench-pdf-micro:
	$(GO_BENCH) test -run='^$$' -bench='BenchmarkGenerateTemplatePDF/Rows2000$$|BenchmarkGoPdfSuit$$' \
		-benchmem -count=10 ./internal/pdf

bench-pdf-macro:
	$(GO_BENCH) test -run='^$$' -bench='BenchmarkGenerateTemplatePDF$$|BenchmarkGenerateTemplatePDF_WrapEnabled$$' \
		-benchmem -count=$(BENCH_COUNT) ./internal/pdf

bench-pdf-typst: bench-setup
	$(GO_BENCH) test -tags=compare -run='^$$' -bench='BenchmarkTypst$$' \
		-benchmem -count=$(BENCH_COUNT) ./internal/pdf

# ── Full benchmark suites (sequential) ───────────────────────────────────────

bench-suite: bench-gopdflib-zerodha bench-pypdfsuit-zerodha bench-gopdfkit-compare bench-handler-all load-pprof
	@echo "bench-suite complete"

bench-suite-full: bench-suite bench-all-libraries
	@echo "bench-suite-full complete"

bench-suite-x2:
	@$(MAKE) bench-gopdflib-zerodha-x2
	@$(MAKE) bench-pypdfsuit-zerodha-x2
	@$(MAKE) bench-gopdfkit-compare-x2
	@for i in 1 2; do \
		echo "=== handler bench run $$i / 2 ==="; \
		$(MAKE) bench-handler-all; \
	done
	@$(MAKE) bench-k6
	@echo "=== k6 run 2 / 2 ==="
	@$(MAKE) bench-k6
	@echo "bench-suite-x2 complete"
