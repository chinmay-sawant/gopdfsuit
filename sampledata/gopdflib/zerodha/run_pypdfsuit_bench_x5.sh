#!/usr/bin/env bash
# PyPDFSuit Zerodha Gold Standard: 5000 iterations × 5 runs + phase profile.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
OUT="$REPO_ROOT/guides/cursor/baselines/pypdfsuit_pprof_runs"
ZERODHA="$REPO_ROOT/sampledata/gopdflib/zerodha"

mkdir -p "$OUT"
export BENCH_ITERATIONS="${BENCH_ITERATIONS:-5000}"
export BENCH_WORKERS="${BENCH_WORKERS:-48}"

echo "PyPDFSuit Zerodha benchmark: iterations=$BENCH_ITERATIONS workers=$BENCH_WORKERS"
echo "Output: $OUT"

cd "$ZERODHA"
for i in 1 2 3 4 5; do
  echo "=== Run $i / 5 (timing) ==="
  python3 pypdfsuit_bench.py 2>&1 | tee "$OUT/pypdfsuit_run${i}.txt"
done

echo "=== Phase profile run (cProfile + per-call breakdown) ==="
python3 pypdfsuit_profile.py 2>&1 | tee "$OUT/pypdfsuit_profile.txt"

echo "Done."