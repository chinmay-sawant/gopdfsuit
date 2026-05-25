#!/usr/bin/env bash
# Run GoPDFLib data benchmark 5000× with CPU/heap pprof (PDF/A, 48 workers).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
OUT="$REPO_ROOT/guides/cursor/baselines/gopdflib_pprof_runs"
BIN="$OUT/gopdflib_bench"
mkdir -p "$OUT"

export BENCH_ITERATIONS="${BENCH_ITERATIONS:-5000}"
export BENCH_WORKERS="${BENCH_WORKERS:-48}"
export BENCH_QUIET=1
export GOWORK=off

echo "GoPDFLib pprof suite: iterations=$BENCH_ITERATIONS workers=$BENCH_WORKERS"
echo "Output: $OUT"

cd "$SCRIPT_DIR"
go build -mod=mod -o "$BIN" .

echo "=== Timing run (no profile) ==="
"$BIN" data 2>&1 | tee "$OUT/bench_gopdflib_5000_timing.txt"

for i in 1 2 3 4 5; do
  echo "=== CPU profile run $i / 5 ==="
  "$BIN" -cpuprofile="$OUT/cpu_gopdflib_data_run${i}.prof" data 2>&1 | tee "$OUT/bench_gopdflib_5000_run${i}.txt"
done

echo "=== Heap profile run ==="
"$BIN" -memprofile="$OUT/heap_gopdflib_data.prof" data 2>&1 | tee "$OUT/bench_gopdflib_5000_heap.txt"

echo "Done. Profiles in $OUT"
