#!/usr/bin/env bash
# Zerodha Gold Standard: 5000 iterations × 10 sequential timing runs (WSL native).
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
OUT="$REPO_ROOT/guides/cursor/baselines/zerodha_bench_x10_wsl"
mkdir -p "$OUT"

export BENCH_ITERATIONS="${BENCH_ITERATIONS:-5000}"
export BENCH_WORKERS="${BENCH_WORKERS:-48}"

echo "Zerodha benchmark ×10 (sequential): iterations=$BENCH_ITERATIONS workers=$BENCH_WORKERS"
echo "Output: $OUT"

cd "$(dirname "${BASH_SOURCE[0]}")"
for i in $(seq 1 10); do
  echo "=== Run $i / 10 ==="
  go run . 2>&1 | tee "$OUT/zerodha_run${i}.txt"
done

echo "Done. Stats: guides/cursor/baselines/zerodha_bench_x10_wsl_stats_20260525.txt"
