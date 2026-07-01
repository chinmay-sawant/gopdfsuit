#!/usr/bin/env bash
# Zerodha Gold Standard: sequential timing runs from this directory (WSL/native).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
OUT="${OUT:-$REPO_ROOT/guides/cursor/baselines/zerodha_bench_x10_wsl}"
mkdir -p "$OUT"

export BENCH_ITERATIONS="${BENCH_ITERATIONS:-5000}"
export BENCH_WORKERS="${BENCH_WORKERS:-48}"
export GOMAXPROCS="${GOMAXPROCS:-24}"

echo "Zerodha benchmark ×10 (sequential)"
echo "  directory:   $SCRIPT_DIR"
echo "  iterations:  $BENCH_ITERATIONS"
echo "  workers:     $BENCH_WORKERS"
echo "  GOMAXPROCS:  $GOMAXPROCS"
echo "  output:      $OUT"
echo ""

cd "$SCRIPT_DIR"
for i in $(seq 1 10); do
  echo "=== Run $i / 10 ==="
  GOMAXPROCS="$GOMAXPROCS" BENCH_ITERATIONS="$BENCH_ITERATIONS" BENCH_WORKERS="$BENCH_WORKERS" \
    go run . 2>&1 | tee "$OUT/zerodha_run${i}.txt"
done

echo ""
echo "=== Done ==="
echo "Logs: $OUT"