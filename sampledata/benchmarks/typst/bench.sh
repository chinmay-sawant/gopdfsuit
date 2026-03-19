#!/bin/bash
# Typst benchmark runner with PDF/A-2b compliance

set -euo pipefail

ITERATIONS=10
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUNDLED_TYPST="$SCRIPT_DIR/../typst-x86_64-unknown-linux-musl/typst"

if command -v typst >/dev/null 2>&1; then
	TYPST_BIN="$(command -v typst)"
elif [ -x "$BUNDLED_TYPST" ]; then
	TYPST_BIN="$BUNDLED_TYPST"
else
	echo "Typst executable not found"
	exit 1
fi

TIMINGS=()

echo "=== Typst Data Benchmark ==="
echo "Iterations: $ITERATIONS"

TOTAL_START=$(date +%s%N)
for ((RUN_INDEX=1; RUN_INDEX<=ITERATIONS; RUN_INDEX++)); do
	START=$(date +%s%N)
	"$TYPST_BIN" compile --pdf-standard a-2b "$SCRIPT_DIR/benchmark.typ" "$SCRIPT_DIR/output_typst.pdf" >/dev/null 2>&1
	END=$(date +%s%N)
	ELAPSED=$(awk "BEGIN {printf \"%.2f\", ($END - $START) / 1000000}")
	TIMINGS+=("$ELAPSED")
	echo "Run $RUN_INDEX: ${ELAPSED} ms"
done
TOTAL_END=$(date +%s%N)

SUMMARY=$(printf '%s\n' "${TIMINGS[@]}" | awk '
BEGIN {min = -1; max = 0; sum = 0; count = 0}
{
	value = $1 + 0
	if (min < 0 || value < min) min = value
	if (value > max) max = value
	sum += value
	count += 1
}
END {
	totalSeconds = ('"$TOTAL_END"' - '"$TOTAL_START"') / 1000000000
	printf "%.2f %.2f %.2f %.2f", min, sum / count, max, count / totalSeconds
}')

read -r MIN AVG MAX THROUGHPUT <<< "$SUMMARY"

echo ""
echo "Min: ${MIN} ms"
echo "Avg: ${AVG} ms"
echo "Max: ${MAX} ms"
echo "Throughput: ${THROUGHPUT} ops/sec"
echo "PDF Standard: PDF/A-2b (ISO 19005-2:2011)"
