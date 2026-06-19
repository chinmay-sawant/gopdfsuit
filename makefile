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

build: test-integration
	mkdir -p bin
	go build -o bin/app ./cmd/gopdfsuit

test:
	go test ./...
	cd bindings/python && python3 -m pytest tests
	bash test/verify_pdfs.sh

install-verapdf:
	bash test/install_verapdf.sh

test-verify-pdfs:
	bash test/verify_pdfs.sh

test-scan-pdfs:
	bash test/verify_pdfs.sh --scan-all

test-integration: test
	go test -count=1 -v ./test

clean:
	rm -rf bin/

run: test-integration lint
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
	golangci-lint run -E revive,gocritic,gocyclo,goconst ./...
	cd frontend && npm run lint
	cd .. 

gdocker: test-integration
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
BENCHMARKS_DIR := sampledata/benchmarks
GOPDFKIT_COMPARE_DIR := $(BENCHMARKS_DIR)/gopdfkit_compare
GOTENBERG_DIR := $(BENCHMARKS_DIR)/gotenberg
K6_DIR := test/generate_template-pdf
# bench-k6-light defaults (override on CLI, e.g. K6_LIGHT_VUS=16 make bench-k6-light)
K6_LIGHT_VUS ?= 24
K6_LIGHT_SECONDS ?= 15
K6_LIGHT_MAX_CONCURRENT ?= 24
K6_LIGHT_GOMAXPROCS ?= 12

.PHONY: build test install-verapdf test-verify-pdfs test-scan-pdfs clean run fmt vet mod lint \
	load-pprof load-pprof-gate load-pprof-1k load-pprof-1500 \
	bench-help bench-setup \
	bench-k6 bench-k6-light bench-k6-retail bench-k6-1k bench-k6-1500 bench-k6-load \
	bench-k6-smoke bench-k6-spike bench-k6-soak bench-k6-install \
	bench-gotenberg bench-gotenberg-load bench-gotenberg-smoke bench-gotenberg-start \
	bench-gopdflib-zerodha bench-gopdflib-zerodha-x2 bench-gopdflib-zerodha-x5 bench-gopdflib-zerodha-x10 bench-gopdflib-zerodha-x10-pprof \
	bench-gopdflib-data bench-gopdflib-data-pprof \
	bench-gopdfsuit-zerodha bench-pypdfsuit-zerodha bench-pypdfsuit-zerodha-x2 \
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
	@echo "  gopdfsuit HTTP (k6 + pprof — server started by script unless noted):"
	@echo "    make load-pprof                  # weighted tagged_ecdsa, 48 VU x 35s + CPU/heap pprof"
	@echo "    make bench-k6                    # alias for load-pprof"
	@echo "    make bench-k6-light              # 24 VU x 15s, lighter CPU/RAM (WSL / shared machine)"
	@echo "    make load-pprof-gate             # retail-only signed (bench-k6-retail)"
	@echo "    make load-pprof-1k               # weighted, 1000 req/s gate"
	@echo "    make load-pprof-1500             # weighted, 1500 req/s gate"
	@echo "    make bench-k6-load               # k6 load_test.js only (start server yourself)"
	@echo "    make bench-k6-smoke              # quick smoke_test.js"
	@echo "    make bench-k6-spike              # spike_test.js"
	@echo "    make bench-k6-soak               # 30 min soak_test.js"
	@echo ""
	@echo "  Gotenberg HTML→PDF (Docker + k6, sampledata/benchmarks/gotenberg):"
	@echo "    make bench-gotenberg             # weighted tagged_ecdsa, 48 VU x 35s"
	@echo "    make bench-gotenberg-load        # k6 only (start Gotenberg yourself)"
	@echo "    make bench-gotenberg-smoke       # 1 VU smoke_test.js"
	@echo "    make bench-gotenberg-start       # docker run Gotenberg on :3010"
	@echo ""
	@echo "  Zerodha gold standard (sampledata/gopdflib/zerodha):"
	@echo "    make bench-gopdflib-zerodha"
	@echo "    make bench-gopdflib-zerodha-x2"
	@echo "    make bench-gopdflib-zerodha-x5   # 5 runs + CPU/heap pprof"
	@echo "    make bench-gopdflib-zerodha-x10  # 10 sequential timing runs"
	@echo "    make bench-gopdflib-zerodha-x10-pprof # x10 timing + CPU/heap pprof"
	@echo "    make bench-pypdfsuit-zerodha"
	@echo "    make bench-pypdfsuit-zerodha-x2"
	@echo "    make bench-pypdfsuit-zerodha-x5   # 5 runs + phase profile (run_pypdfsuit_bench_x5.sh)"
	@echo "    make bench-pypdfsuit-zerodha-x10  # 10 sequential timing runs"
	@echo "    make bench-pypdfsuit-zerodha-x10-pprof # x10 timing + x5/profile"
	@echo "    make bench-pypdfsuit-profile      # phase breakdown (pypdfsuit_profile.py)"
	@echo ""
	@echo "  GoPDFLib data-table + pprof (sampledata/benchmarks/gopdflib):"
	@echo "    make bench-gopdflib-data"
	@echo "    make bench-gopdflib-data-pprof   # 5000x + 5 CPU + 1 heap profile"
	@echo ""
	@echo "  Multi-library suite (sampledata/benchmarks — run bench-setup first):"
	@echo "    make bench-all-libraries         # all engines sequentially"
	@echo "    make bench-gopdfsuit-zerodha     # GoPDFSuit via benchmarktemplates"
	@echo "    make bench-pypdfsuit-legacy      # pypdfsuit bench.py (benchmarks dir)"
	@echo "    make bench-fpdf bench-jspdf bench-pdfkit-lib bench-pdflib bench-typst"
	@echo ""
	@echo "  GoPDFKit apples-to-apples (sampledata/benchmarks/gopdfkit_compare):"
	@echo "    make bench-gopdfkit-setup        # symlink real gopdfkit module"
	@echo "    make bench-gopdfkit-compare-test # PDF output sanity before timing"
	@echo "    make bench-gopdfkit-compare"
	@echo "    make bench-gopdfkit-compare-x2"
	@echo "    make bench-gopdfkit-html         # HTML subset (needs Chrome; opt-in)"
	@echo ""
	@echo "  Go test benchmarks:"
	@echo "    make bench-handler-all           # Gin handler financial_report.json"
	@echo "    make bench-handler               # serial only"
	@echo "    make bench-handler-parallel      # parallel only"
	@echo "    make bench-pdf-micro             # Rows2000 + data.json (internal/pdf)"
	@echo "    make bench-pdf-macro             # Rows2000/10000/25000 synthetic tables"
	@echo "    make bench-pdf-typst             # Typst compile (requires bench-setup + compare tag)"
	@echo ""
	@echo "  Full suites (sequential — allow several minutes):"
	@echo "    make bench-suite                 # zerodha + pypdfsuit + gopdfkit + handler + k6"
	@echo "    make bench-suite-x2              # two passes each harness in bench-suite"
	@echo "    make bench-suite-full            # bench-suite + multi-library run_all_benchmarks.sh"

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
