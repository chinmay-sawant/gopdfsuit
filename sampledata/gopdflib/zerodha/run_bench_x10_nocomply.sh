#!/usr/bin/env bash
# Zerodha Gold Standard (non-compliant): 5000 iterations × 10 sequential timing runs.
# Same 80/15/5 workload with PDF/A, tagging, signing, and font embedding disabled.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
OUT="$REPO_ROOT/guides/cursor/baselines/zerodha_bench_x10_nocomply_wsl"
mkdir -p "$OUT"

export BENCH_ITERATIONS="${BENCH_ITERATIONS:-5000}"
export BENCH_WORKERS="${BENCH_WORKERS:-48}"

echo "Zerodha non-compliant benchmark ×10 (sequential): iterations=$BENCH_ITERATIONS workers=$BENCH_WORKERS"
echo "Output: $OUT"
echo "Cache override example: GOCACHE=/tmp/gopdfsuit-go-build-cache GOMODCACHE=/tmp/gopdfsuit-go-mod-cache make bench-gopdflib-zerodha-nocomply-x10"

cd "$(dirname "${BASH_SOURCE[0]}")"
for i in $(seq 1 10); do
  echo "=== Run $i / 10 ==="
  go run -tags nocomply . 2>&1 | tee "$OUT/zerodha_nocomply_run${i}.txt"
done

STATS="$REPO_ROOT/guides/cursor/baselines/zerodha_bench_x10_nocomply_wsl_stats_latest.txt"
python3 - "$OUT" "$STATS" "${BENCH_X10_MEAN_GATE:-0}" "${BENCH_X10_MEDIAN_GATE:-0}" <<'PY'
from pathlib import Path
import re
import statistics as stats
import sys

out = Path(sys.argv[1])
stats_path = Path(sys.argv[2])
mean_gate = float(sys.argv[3])
median_gate = float(sys.argv[4])
runs = []
for path in sorted(out.glob("zerodha_nocomply_run*.txt"), key=lambda p: int(re.search(r"(\d+)", p.stem).group(1))):
    text = path.read_text()
    throughput = re.search(r"Throughput:\s+([0-9.]+) ops/sec", text)
    latency = re.search(r"Avg Latency:\s+([0-9.]+) ms", text)
    allocated = re.search(r"Max Memory Allocated:\s+([0-9.]+) MB", text)
    if throughput:
        runs.append((
            path.name,
            float(throughput.group(1)),
            float(latency.group(1)) if latency else 0.0,
            float(allocated.group(1)) if allocated else 0.0,
        ))

lines = [
    "# Zerodha x10 non-compliant latest summary",
    "",
    f"Runs: {len(runs)}",
]
if runs:
    throughputs = [r[1] for r in runs]
    latencies = [r[2] for r in runs]
    allocations = [r[3] for r in runs]
    mean = stats.mean(throughputs)
    median = stats.median(throughputs)
    lines.extend([
        f"Best throughput: {max(throughputs):.2f} ops/sec",
        f"Worst throughput: {min(throughputs):.2f} ops/sec",
        f"Mean throughput: {mean:.2f} ops/sec",
        f"Median throughput: {median:.2f} ops/sec",
        f"Stddev throughput: {(stats.stdev(throughputs) if len(throughputs) > 1 else 0.0):.2f} ops/sec",
        f"Mean avg latency: {stats.mean(latencies):.3f} ms",
        f"Mean peak allocated: {stats.mean(allocations):.2f} MB",
        "",
        "| Run | Throughput | Avg latency | Peak allocated |",
        "|-----|-----------:|------------:|---------------:|",
    ])
    lines.extend(
        f"| {name} | {throughput:.2f} ops/sec | {latency:.3f} ms | {allocated:.2f} MB |"
        for name, throughput, latency, allocated in runs
    )

stats_path.write_text("\n".join(lines) + "\n")
print(f"Summary: {stats_path}")
if runs and mean_gate > 0 and mean < mean_gate:
    raise SystemExit(f"mean throughput gate failed: {mean:.2f} < {mean_gate:.2f}")
if runs and median_gate > 0 and median < median_gate:
    raise SystemExit(f"median throughput gate failed: {median:.2f} < {median_gate:.2f}")
PY

echo "Done. Stats: $STATS"