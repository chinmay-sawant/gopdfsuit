#!/usr/bin/env python3

import statistics
import sys
import time
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(REPO_ROOT / "sampledata" / "gopdflib" / "zerodha"))
sys.path.insert(0, str(REPO_ROOT / "bindings" / "python"))

from pypdfsuit import generate_pdf  # noqa: E402
from pypdfsuit_bench import build_retail_template  # noqa: E402


def main() -> None:
    iterations = 5
    template = build_retail_template()
    timings = []

    print("=== PyPDFSuit Single Zerodha Benchmark ===")
    print(f"Iterations: {iterations}")

    total_start = time.perf_counter()
    for run_index in range(1, iterations + 1):
        start = time.perf_counter()
        generate_pdf(template)
        elapsed_ms = (time.perf_counter() - start) * 1000
        timings.append(elapsed_ms)
        print(f"Run {run_index}: {elapsed_ms:.2f} ms")
    total_seconds = time.perf_counter() - total_start

    average_ms = statistics.mean(timings)
    print()
    print(f"Min: {min(timings):.2f} ms")
    print(f"Avg: {average_ms:.2f} ms")
    print(f"Max: {max(timings):.2f} ms")
    print(f"Throughput: {iterations / total_seconds:.2f} ops/sec")


if __name__ == "__main__":
    main()