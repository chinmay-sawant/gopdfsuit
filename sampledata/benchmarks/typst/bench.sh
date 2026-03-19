#!/bin/bash
# Typst benchmark runner with PDF/A-2b compliance (parallel)
# Highest PDF standard supported by the bundled typst 0.12.0 binary is a-2b (PDF/A-2b).
# Upgrade to --pdf-standard a-3b requires typst 0.13+.

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

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

PIDS=()
TIMING_FILES=()

echo "=== Typst Data Benchmark ==="
echo "Iterations: $ITERATIONS | Workers: $ITERATIONS (parallel)"

TOTAL_START=$(date +%s%N)
for ((RUN_INDEX=1; RUN_INDEX<=ITERATIONS; RUN_INDEX++)); do
	TIMING_FILE="$TMP_DIR/timing_${RUN_INDEX}.txt"
	TIMING_FILES+=("$TIMING_FILE")
	(
		START=$(date +%s%N)
		"$TYPST_BIN" compile --pdf-standard a-2b \
			"$SCRIPT_DIR/benchmark.typ" \
			"$SCRIPT_DIR/output_typst_${RUN_INDEX}.pdf" >/dev/null 2>&1
		END=$(date +%s%N)
		awk "BEGIN {printf \"%.2f\", ($END - $START) / 1000000}" > "$TIMING_FILE"
	) &
	PIDS+=($!)
done

for PID in "${PIDS[@]}"; do
	wait "$PID"
done
TOTAL_END=$(date +%s%N)

TIMINGS=()
for TIMING_FILE in "${TIMING_FILES[@]}"; do
	TIMINGS+=("$(cat "$TIMING_FILE")")
done

# Keep only the last run's PDF as the canonical output
for ((RUN_INDEX=1; RUN_INDEX<ITERATIONS; RUN_INDEX++)); do
	rm -f "$SCRIPT_DIR/output_typst_${RUN_INDEX}.pdf"
done
if [ -f "$SCRIPT_DIR/output_typst_${ITERATIONS}.pdf" ]; then
	mv "$SCRIPT_DIR/output_typst_${ITERATIONS}.pdf" "$SCRIPT_DIR/output_typst.pdf"
fi

for ((i=0; i<${#TIMINGS[@]}; i++)); do
	echo "Run $((i+1)): ${TIMINGS[$i]} ms"
done

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
echo "PDF Standard: PDF/A-2b (ISO 19005-2:2011) — highest supported by bundled typst 0.12.0"
