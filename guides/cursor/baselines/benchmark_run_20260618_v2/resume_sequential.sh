#!/usr/bin/env bash
# Resume benchmarks sequentially (one target at a time). Skips pdf-macro and pdf-wrap (too slow).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
OUT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

RUNS=5
export GOMAXPROCS_BENCH=24
export BENCH_ITERATIONS=5000
export BENCH_WORKERS=48

log() { echo "[$(date +%H:%M:%S)] $*" | tee -a "$OUT/resume.log"; }

run_one_5x() {
  local name="$1"
  local cmd="$2"
  local out="$OUT/${name}_5x.txt"
  if [[ -f "$OUT/.done_${name}" ]]; then
    log "SKIP $name (already done)"
    return 0
  fi
  : > "$out"
  log "START $name (5 runs, sequential)"
  for i in $(seq 1 "$RUNS"); do
    log "  $name run $i/$RUNS"
    echo "--- run $i/$RUNS ---" >> "$out"
    eval "$cmd" >> "$out" 2>&1 || echo "RUN $i FAILED" >> "$out"
    echo "" >> "$out"
  done
  touch "$OUT/.done_${name}"
  log "DONE $name"
}

log "=== Resume sequential (pdf-macro and pdf-wrap SKIPPED) ==="

# --- Remaining go test benchmarks ---
run_one_5x "pdf-gopdfsuit" \
  "GOMAXPROCS=24 go test -run='^\$' -bench='^BenchmarkGoPdfSuit\$' -benchmem -benchtime=5s -count=1 ./internal/pdf"

run_one_5x "pdf-typst" \
  "GOMAXPROCS=24 go test -tags=compare -run='^\$' -bench='^BenchmarkTypst\$' -benchmem -benchtime=5s -count=1 ./internal/pdf"

# pdf-micro uses Rows2000 only (makefile count=10) - single invocation
if [[ ! -f "$OUT/.done_pdf-micro" ]]; then
  log "START pdf-micro (Rows2000 + GoPdfSuit, count=10)"
  make bench-pdf-micro > "$OUT/pdf_micro_10x.txt" 2>&1 || true
  touch "$OUT/.done_pdf-micro"
  log "DONE pdf-micro"
else
  log "SKIP pdf-micro"
fi

# gopdfkit compare
run_one_5x "gopdfkit-compare" \
  "cd $ROOT/sampledata/benchmarks/gopdfkit_compare && GOMAXPROCS=24 go test -run='^\$' -bench='^BenchmarkGoPDF(Kit|Lib)\$' -benchmem -benchtime=5s -count=1"

# k6 / gotenberg (sequential, one full target before next)
run_one_5x "k6" "make bench-k6"
run_one_5x "k6-light" "make bench-k6-light"
run_one_5x "k6-retail" "make bench-k6-retail"
run_one_5x "gotenberg" "make bench-gotenberg"

log "=== All resume targets complete ==="
python3 "$OUT/parse_results.py" | tee -a "$OUT/resume.log"