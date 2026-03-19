#!/usr/bin/env python3

import math
import multiprocessing
import sys
import time
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(REPO_ROOT / "sampledata" / "gopdflib" / "zerodha"))
sys.path.insert(0, str(REPO_ROOT / "bindings" / "python"))

from pypdfsuit import generate_pdf  # noqa: E402
from pypdfsuit_bench import build_retail_template  # noqa: E402

ITERATIONS = 10
NUM_WORKERS = min(48, ITERATIONS)


def run_once(run_index):
    """Each worker builds its own template and generates the PDF independently."""
    start = time.perf_counter()
    generate_pdf(build_retail_template())
    return (time.perf_counter() - start) * 1000


def main() -> None:
    print("=== PyPDFSuit Concurrent Zerodha Benchmark ===")
    print(f"Iterations: {ITERATIONS} | Workers: {NUM_WORKERS}")

    total_start = time.perf_counter()
    with multiprocessing.Pool(processes=NUM_WORKERS) as pool:
        all_times = pool.map(run_once, range(1, ITERATIONS + 1))
    total_seconds = time.perf_counter() - total_start

    for i, elapsed in enumerate(all_times, start=1):
        print(f"Run {i}: {elapsed:.2f} ms")

    sorted_times = sorted(all_times)
    p95_idx = max(0, math.ceil(len(sorted_times) * 0.95) - 1)

    print()
    print(f"Min: {min(all_times):.2f} ms")
    print(f"Avg: {sum(all_times)/len(all_times):.2f} ms")
    print(f"P95: {sorted_times[p95_idx]:.2f} ms")
    print(f"Max: {max(all_times):.2f} ms")
    print(f"Throughput: {ITERATIONS / total_seconds:.2f} ops/sec")


if __name__ == "__main__":
    main()