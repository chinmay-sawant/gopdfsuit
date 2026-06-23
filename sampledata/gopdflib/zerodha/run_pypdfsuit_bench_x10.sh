#!/usr/bin/env bash
# PyPDFSuit Zerodha Gold Standard: 5000 iterations × 10 sequential timing runs (WSL native).
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
OUT="$REPO_ROOT/guides/cursor/baselines/pypdfsuit_bench_x10_wsl"
mkdir -p "$OUT"

export BENCH_ITERATIONS="${BENCH_ITERATIONS:-5000}"
export BENCH_WORKERS="${BENCH_WORKERS:-48}"

echo "PyPDFSuit Zerodha benchmark ×10 (sequential): iterations=$BENCH_ITERATIONS workers=$BENCH_WORKERS"
echo "Output: $OUT"

cd "$(dirname "${BASH_SOURCE[0]}")"
for i in $(seq 1 10); do
  echo "=== Run $i / 10 ==="
  python3 pypdfsuit_bench.py 2>&1 | tee "$OUT/pypdfsuit_run${i}.txt"
done

STATS="$REPO_ROOT/guides/cursor/baselines/pypdfsuit_bench_x10_wsl_stats_latest.txt"
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
for path in sorted(out.glob("pypdfsuit_run*.txt"), key=lambda p: int(re.search(r"(\d+)", p.stem).group(1))):
    text = path.read_text()
    throughput = re.search(r"Throughput:\s+([0-9.]+) ops/sec", text)
    avg_latency = re.search(r"Avg Latency:\s+([0-9.]+) ms", text)
    p50_latency = re.search(r"P50 Latency:\s+([0-9.]+) ms", text)
    p95_latency = re.search(r"P95 Latency:\s+([0-9.]+) ms", text)
    p99_latency = re.search(r"P99 Latency:\s+([0-9.]+) ms", text)
    min_latency = re.search(r"Min Latency:\s+([0-9.]+) ms", text)
    max_latency = re.search(r"Max Latency:\s+([0-9.]+) ms", text)
    if throughput:
        runs.append((
            path.name,
            float(throughput.group(1)),
            float(avg_latency.group(1)) if avg_latency else 0.0,
            float(p50_latency.group(1)) if p50_latency else 0.0,
            float(p95_latency.group(1)) if p95_latency else 0.0,
            float(p99_latency.group(1)) if p99_latency else 0.0,
            float(min_latency.group(1)) if min_latency else 0.0,
            float(max_latency.group(1)) if max_latency else 0.0,
        ))

lines = [
    "# PyPDFSuit Zerodha x10 latest summary",
    "",
    f"Runs: {len(runs)}",
]
if runs:
    throughputs = [r[1] for r in runs]
    latencies = [r[2] for r in runs]
    p50s = [r[3] for r in runs]
    p95s = [r[4] for r in runs]
    p99s = [r[5] for r in runs]
    mins = [r[6] for r in runs]
    maxes = [r[7] for r in runs]
    mean = stats.mean(throughputs)
    median = stats.median(throughputs)
    lines.extend([
        f"Best throughput: {max(throughputs):.2f} ops/sec",
        f"Worst throughput: {min(throughputs):.2f} ops/sec",
        f"Mean throughput: {mean:.2f} ops/sec",
        f"Median throughput: {median:.2f} ops/sec",
        f"Stddev throughput: {(stats.stdev(throughputs) if len(throughputs) > 1 else 0.0):.2f} ops/sec",
        f"Mean avg latency: {stats.mean(latencies):.3f} ms",
        f"Mean p50 latency: {stats.mean(p50s):.3f} ms",
        f"Mean p95 latency: {stats.mean(p95s):.3f} ms",
        f"Mean p99 latency: {stats.mean(p99s):.3f} ms",
        f"Worst max latency: {max(maxes):.3f} ms",
        f"Best min latency: {min(mins):.3f} ms",
        "",
        "| Run | Throughput | Avg latency | P50 | P95 | P99 | Min | Max |",
        "|-----|-----------:|------------:|----:|----:|----:|----:|----:|",
    ])
    lines.extend(
        (
            f"| {name} | {throughput:.2f} ops/sec | {avg:.3f} ms | "
            f"{p50:.3f} ms | {p95:.3f} ms | {p99:.3f} ms | "
            f"{min_latency:.3f} ms | {max_latency:.3f} ms |"
        )
        for name, throughput, avg, p50, p95, p99, min_latency, max_latency in runs
    )

stats_path.write_text("\n".join(lines) + "\n")
print(f"Summary: {stats_path}")
if runs and mean_gate > 0 and mean < mean_gate:
    raise SystemExit(f"mean throughput gate failed: {mean:.2f} < {mean_gate:.2f}")
if runs and median_gate > 0 and median < median_gate:
    raise SystemExit(f"median throughput gate failed: {median:.2f} < {median_gate:.2f}")
PY

echo "Done. Stats: $STATS"
