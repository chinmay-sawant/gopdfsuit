#!/usr/bin/env bash
# Zerodha Gold Standard: 5000 iterations × 5 runs + optional pprof.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
OUT="$REPO_ROOT/guides/cursor/baselines/zerodha_pprof_runs"
ZERODHA="$REPO_ROOT/sampledata/gopdflib/zerodha"
BIN="$OUT/zerodha_bench"

mkdir -p "$OUT"
export BENCH_ITERATIONS="${BENCH_ITERATIONS:-5000}"
export BENCH_WORKERS="${BENCH_WORKERS:-48}"

echo "Zerodha benchmark: iterations=$BENCH_ITERATIONS workers=$BENCH_WORKERS"
echo "Output: $OUT"

cd "$REPO_ROOT"
go build -o "$BIN" ./sampledata/gopdflib/zerodha/main.go

for i in 1 2 3 4 5; do
  echo "=== Run $i / 5 (timing) ==="
  "$BIN" 2>&1 | tee "$OUT/zerodha_run${i}.txt"
done

for i in 1 2 3 4 5; do
  echo "=== CPU profile run $i / 5 ==="
  "$BIN" -cpuprofile="$OUT/cpu_zerodha_run${i}.prof" 2>&1 | tee "$OUT/zerodha_cpu_run${i}.txt"
done

echo "=== Heap profile run ==="
"$BIN" -memprofile="$OUT/heap_zerodha.prof" 2>&1 | tee "$OUT/zerodha_heap.txt"

echo "Done."
