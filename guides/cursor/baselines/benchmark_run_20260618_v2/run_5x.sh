#!/usr/bin/env bash
# Run benchmarks 5x and record best throughput (or best ns/op for go test).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
OUT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

RUNS=5
export GOMAXPROCS_BENCH=24
export BENCH_ITERATIONS=5000
export BENCH_WORKERS=48

log() { echo "[$(date +%H:%M:%S)] $*" | tee -a "$OUT/run.log"; }

run_5x_make() {
  local name="$1"
  local target="$2"
  local out="$OUT/${name}_5x.txt"
  : > "$out"
  log "=== $name ($target) x$RUNS ==="
  for i in $(seq 1 "$RUNS"); do
    echo "--- run $i/$RUNS ---" >> "$out"
    make "$target" >> "$out" 2>&1 || echo "RUN $i FAILED" >> "$out"
    echo "" >> "$out"
  done
}

run_5x_go_bench() {
  local name="$1"
  local benchpat="$2"
  local pkg="$3"
  local extra="${4:-}"
  local out="$OUT/${name}_5x.txt"
  : > "$out"
  log "=== $name x$RUNS ==="
  for i in $(seq 1 "$RUNS"); do
    echo "--- run $i/$RUNS ---" >> "$out"
    # shellcheck disable=SC2086
    GOMAXPROCS=24 go test $extra -run='^$' -bench="$benchpat" -benchmem -benchtime=5s -count=1 "$pkg" >> "$out" 2>&1 || echo "RUN $i FAILED" >> "$out"
    echo "" >> "$out"
  done
}

log "Environment: $(go version)"
log "Branch: $(git branch --show-current) $(git rev-parse --short HEAD)"

make bench-setup >> "$OUT/run.log" 2>&1 || true
make bench-gopdfkit-setup >> "$OUT/run.log" 2>&1 || true

# Already has x5 target - run once via makefile x5
log "=== gopdflib-zerodha (make bench-gopdflib-zerodha-x5) ==="
make bench-gopdflib-zerodha-x5 > "$OUT/gopdflib_zerodha_x5.txt" 2>&1 || true

# sampledata/benchmarks + makefile harness (5x each)
for t in bench-gopdflib-data bench-gopdfsuit-zerodha bench-pypdfsuit-zerodha bench-pypdfsuit-legacy \
         bench-fpdf bench-jspdf bench-pdfkit-lib bench-pdflib; do
  run_5x_make "${t#bench-}" "$t"
done

# typst manual 5x (bundled 0.11 lacks --pdf-standard)
log "=== typst-manual x5 ==="
: > "$OUT/typst_manual_5x.txt"
TYPST="$ROOT/sampledata/benchmarks/typst-x86_64-unknown-linux-musl/typst"
for i in $(seq 1 "$RUNS"); do
  echo "--- run $i/$RUNS ---" >> "$OUT/typst_manual_5x.txt"
  START=$(date +%s%N)
  "$TYPST" compile "$ROOT/sampledata/benchmarks/typst/benchmark.typ" "/tmp/typst_bench_${i}.pdf" >> "$OUT/typst_manual_5x.txt" 2>&1 || echo "RUN $i FAILED" >> "$OUT/typst_manual_5x.txt"
  END=$(date +%s%N)
  MS=$(awk "BEGIN {printf \"%.2f\", ($END - $START) / 1000000}")
  echo "Elapsed: ${MS} ms (~$(awk "BEGIN {printf \"%.2f\", 1000/$MS}") ops/s single-doc)" >> "$OUT/typst_manual_5x.txt"
done

# Go test benchmarks (5x, best = min ns/op)
run_5x_go_bench "handler-all" 'BenchmarkGenerateTemplatePDF_FinancialReport' ./test
run_5x_go_bench "handler-parallel" 'BenchmarkGenerateTemplatePDF_FinancialReport_Parallel' ./test
run_5x_go_bench "pdf-macro" '^BenchmarkGenerateTemplatePDF$' ./internal/pdf
run_5x_go_bench "pdf-wrap" '^BenchmarkGenerateTemplatePDF_WrapEnabled$' ./internal/pdf
run_5x_go_bench "pdf-gopdfsuit" '^BenchmarkGoPdfSuit$' ./internal/pdf
run_5x_go_bench "pdf-typst" '^BenchmarkTypst$' ./internal/pdf "-tags=compare"

# gopdfkit compare (5x)
log "=== gopdfkit-compare x$RUNS ==="
: > "$OUT/gopdfkit_compare_5x.txt"
for i in $(seq 1 "$RUNS"); do
  echo "--- run $i/$RUNS ---" >> "$OUT/gopdfkit_compare_5x.txt"
  cd "$ROOT/sampledata/benchmarks/gopdfkit_compare"
  GOMAXPROCS=24 go test -run='^$' -bench='^BenchmarkGoPDF(Kit|Lib)$' -benchmem -benchtime=5s -count=1 >> "$OUT/gopdfkit_compare_5x.txt" 2>&1 || echo "RUN $i FAILED" >> "$OUT/gopdfkit_compare_5x.txt"
  cd "$ROOT"
done

# bench-pdf-micro already count=10 in makefile - run once
log "=== pdf-micro (count=10) ==="
make bench-pdf-micro > "$OUT/pdf_micro_10x.txt" 2>&1 || true

# k6 variants (5x best req/s)
for t in bench-k6 bench-k6-light bench-k6-retail; do
  run_5x_make "${t#bench-}" "$t"
done

# gotenberg (5x)
run_5x_make "gotenberg" "bench-gotenberg"

log "Done. Parse with parse_results.py"