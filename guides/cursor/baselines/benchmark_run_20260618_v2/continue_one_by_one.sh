#!/usr/bin/env bash
# Continue benchmarks one-at-a-time with visible progress (% + run X/5).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
OUT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

RUNS=5
export GOMAXPROCS_BENCH=24
export BENCH_ITERATIONS=5000
export BENCH_WORKERS=48

# Remaining queue (pdf-micro LAST). Each target = RUNS steps except pdf-micro = 1 step.
TARGETS=(
  "gopdfkit-compare"
  "k6"
  "k6-light"
  "k6-retail"
  "gotenberg"
  "pdf-micro"
)

TOTAL_STEPS=0
for t in "${TARGETS[@]}"; do
  if [[ "$t" == "pdf-micro" ]]; then
    TOTAL_STEPS=$((TOTAL_STEPS + 1))
  else
    TOTAL_STEPS=$((TOTAL_STEPS + RUNS))
  fi
done
COMPLETED_STEPS=0

progress() {
  local target="$1"
  local run="$2"
  local run_max="$3"
  local pct=$(( COMPLETED_STEPS * 100 / TOTAL_STEPS ))
  local msg="[PROGRESS ${pct}%] ${target} run ${run}/${run_max} | completed ${COMPLETED_STEPS}/${TOTAL_STEPS} steps"
  echo "$msg" | tee -a "$OUT/continue.log"
}

log() {
  echo "[$(date '+%H:%M:%S')] $*" | tee -a "$OUT/continue.log"
}

run_cmd() {
  local name="$1"
  local cmd="$2"
  case "$name" in
    gopdfkit-compare)
      cd "$ROOT/sampledata/benchmarks/gopdfkit_compare"
      GOMAXPROCS=24 go test -run='^$' -bench='^BenchmarkGoPDF(Kit|Lib)$' -benchmem -benchtime=5s -count=1
      cd "$ROOT"
      ;;
    k6) make bench-k6 ;;
    k6-light) make bench-k6-light ;;
    k6-retail) make bench-k6-retail ;;
    gotenberg) make bench-gotenberg ;;
    pdf-micro) make bench-pdf-micro ;;
    *) eval "$cmd" ;;
  esac
}

run_target_5x() {
  local name="$1"
  local out="$OUT/${name}_5x.txt"
  if [[ -f "$OUT/.done_${name}" ]]; then
    log "SKIP ${name} (already done)"
    return 0
  fi
  : > "$out"
  log ">>> START ${name} (${RUNS} runs)"
  for i in $(seq 1 "$RUNS"); do
    progress "$name" "$i" "$RUNS"
    log "    RUNNING ${name} ${i}/${RUNS} ... (started $(date '+%H:%M:%S'))"
    echo "--- run $i/$RUNS ---" >> "$out"
    if run_cmd "$name" "" >> "$out" 2>&1; then
      log "    FINISHED ${name} ${i}/${RUNS} OK ($(date '+%H:%M:%S'))"
    else
      log "    FINISHED ${name} ${i}/${RUNS} FAILED ($(date '+%H:%M:%S'))"
      echo "RUN $i FAILED" >> "$out"
    fi
    echo "" >> "$out"
    COMPLETED_STEPS=$((COMPLETED_STEPS + 1))
    progress "$name" "$i" "$RUNS"
  done
  touch "$OUT/.done_${name}"
  log ">>> DONE ${name} (100% of this target)"
}

run_target_once() {
  local name="$1"
  local out="$OUT/pdf_micro_10x.txt"
  if [[ -f "$OUT/.done_${name}" ]]; then
    log "SKIP ${name} (already done)"
    return 0
  fi
  progress "$name" "1" "1"
  log ">>> START ${name} LAST (count=10, Rows2000)"
  log "    RUNNING ${name} ... (started $(date '+%H:%M:%S'))"
  if run_cmd "$name" "" > "$out" 2>&1; then
    log "    FINISHED ${name} OK ($(date '+%H:%M:%S'))"
  else
    log "    FINISHED ${name} FAILED ($(date '+%H:%M:%S'))"
  fi
  touch "$OUT/.done_${name}"
  COMPLETED_STEPS=$((COMPLETED_STEPS + 1))
  progress "$name" "1" "1"
  log ">>> DONE ${name}"
}

# Count already-finished steps from .done markers (for resume)
for t in "${TARGETS[@]}"; do
  if [[ -f "$OUT/.done_${t}" ]]; then
    if [[ "$t" == "pdf-micro" ]]; then
      COMPLETED_STEPS=$((COMPLETED_STEPS + 1))
    else
      COMPLETED_STEPS=$((COMPLETED_STEPS + RUNS))
    fi
  fi
done

log "=========================================="
log "BENCHMARK RUN - visible progress enabled"
log "Total steps remaining queue: ${TOTAL_STEPS}"
log "Already completed steps: ${COMPLETED_STEPS} ($(( COMPLETED_STEPS * 100 / TOTAL_STEPS ))%)"
log "=========================================="

for t in "${TARGETS[@]}"; do
  if [[ "$t" == "pdf-micro" ]]; then
    run_target_once "$t"
  else
    run_target_5x "$t"
  fi
done

log "=========================================="
log "ALL BENCHMARKS COMPLETE - 100%"
log "=========================================="
python3 "$OUT/parse_results.py" | tee -a "$OUT/continue.log"