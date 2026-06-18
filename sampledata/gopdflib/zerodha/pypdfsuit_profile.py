#!/usr/bin/env python3
"""
Profile where PyPDFSuit Zerodha benchmark time is spent.

Phases measured per generate_pdf() call:
  1. to_dict      - Python dataclass → dict tree
  2. json_dumps   - dict → UTF-8 JSON bytes
  3. cgo_call     - ctypes GeneratePDF (Go: C.GoString + json.Unmarshal + PDF + malloc)
  4. copy_back    - ctypes.string_at from C buffer
  5. free_result  - FreeBytesResult

Also runs cProfile on a weighted mix and a cached-JSON control (skips to_dict/json_dumps).
"""

from __future__ import annotations

import cProfile
import io
import json
import pstats
import random
import statistics
import sys
import time
from dataclasses import dataclass
from pathlib import Path

import ctypes

REPO_ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(REPO_ROOT / "bindings" / "python"))
sys.path.insert(0, str(Path(__file__).resolve().parent))

from pypdfsuit._bindings import get_lib, call_bytes_result  # noqa: E402
from pypdfsuit.types import PDFTemplate  # noqa: E402
from pypdfsuit_bench import (  # noqa: E402
    build_active_trader_template,
    build_hft_template,
    build_retail_template,
)


WARMUP = 3
SINGLE_THREAD_ITERS = 50
PROFILE_ITERS = 200


@dataclass
class PhaseStats:
    name: str
    samples_ms: list[float]

    @property
    def mean(self) -> float:
        return statistics.mean(self.samples_ms)

    @property
    def p50(self) -> float:
        return statistics.median(self.samples_ms)

    @property
    def p95(self) -> float:
        ordered = sorted(self.samples_ms)
        idx = max(0, int(len(ordered) * 0.95) - 1)
        return ordered[idx]

    @property
    def total(self) -> float:
        return sum(self.samples_ms)


def profile_generate(template: PDFTemplate, iterations: int) -> dict[str, PhaseStats]:
    lib = get_lib()
    phases = {
        "to_dict": PhaseStats("to_dict", []),
        "json_dumps": PhaseStats("json_dumps", []),
        "cgo_call": PhaseStats("cgo_call", []),
        "copy_back": PhaseStats("copy_back", []),
        "free_result": PhaseStats("free_result", []),
        "total": PhaseStats("total", []),
    }
    json_sizes: list[int] = []
    pdf_sizes: list[int] = []

    for _ in range(WARMUP):
        _timed_generate(lib, template)

    for _ in range(iterations):
        t0 = time.perf_counter()
        payload = template.to_dict()
        t1 = time.perf_counter()
        template_json = json.dumps(payload).encode("utf-8")
        t2 = time.perf_counter()
        result = lib.GeneratePDF(template_json)
        t3 = time.perf_counter()
        if result.error:
            raise RuntimeError(result.error.decode("utf-8"))
        if result.data and result.length > 0:
            pdf = ctypes.string_at(result.data, result.length)
        else:
            pdf = b""
        t4 = time.perf_counter()
        lib.FreeBytesResult(result)
        t5 = time.perf_counter()

        phases["to_dict"].samples_ms.append((t1 - t0) * 1000)
        phases["json_dumps"].samples_ms.append((t2 - t1) * 1000)
        phases["cgo_call"].samples_ms.append((t3 - t2) * 1000)
        phases["copy_back"].samples_ms.append((t4 - t3) * 1000)
        phases["free_result"].samples_ms.append((t5 - t4) * 1000)
        phases["total"].samples_ms.append((t5 - t0) * 1000)
        json_sizes.append(len(template_json))
        pdf_sizes.append(len(pdf))

    return {
        "phases": phases,
        "json_bytes_mean": int(statistics.mean(json_sizes)),
        "pdf_bytes_mean": int(statistics.mean(pdf_sizes)),
    }


def _timed_generate(lib, template: PDFTemplate) -> bytes:
    template_json = json.dumps(template.to_dict()).encode("utf-8")
    return call_bytes_result(lib.GeneratePDF, template_json)


def profile_cached_json(template: PDFTemplate, iterations: int) -> dict[str, PhaseStats]:
    """Control: skip Python serialization; measure pure FFI + Go work."""
    lib = get_lib()
    cached = json.dumps(template.to_dict()).encode("utf-8")
    phases = {
        "cgo_call": PhaseStats("cgo_call", []),
        "copy_back": PhaseStats("copy_back", []),
        "free_result": PhaseStats("free_result", []),
        "total": PhaseStats("total", []),
    }

    for _ in range(WARMUP):
        _timed_cached(lib, cached)

    for _ in range(iterations):
        t0 = time.perf_counter()
        result = lib.GeneratePDF(cached)
        t1 = time.perf_counter()
        if result.error:
            raise RuntimeError(result.error.decode("utf-8"))
        if result.data and result.length > 0:
            ctypes.string_at(result.data, result.length)
        t2 = time.perf_counter()
        lib.FreeBytesResult(result)
        t3 = time.perf_counter()

        phases["cgo_call"].samples_ms.append((t1 - t0) * 1000)
        phases["copy_back"].samples_ms.append((t2 - t1) * 1000)
        phases["free_result"].samples_ms.append((t3 - t2) * 1000)
        phases["total"].samples_ms.append((t3 - t0) * 1000)

    return {"phases": phases, "json_bytes_mean": len(cached)}


def _timed_cached(lib, cached: bytes) -> bytes:
    result = lib.GeneratePDF(cached)
    try:
        if result.error:
            raise RuntimeError(result.error.decode("utf-8"))
        if result.data and result.length > 0:
            return ctypes.string_at(result.data, result.length)
        return b""
    finally:
        lib.FreeBytesResult(result)


def print_phase_report(label: str, result: dict, iterations: int) -> None:
    phases: dict[str, PhaseStats] = result["phases"]
    total_mean = phases["total"].mean
    print(f"\n=== {label} ({iterations} iters, single-thread) ===")
    if "json_bytes_mean" in result:
        print(f"  JSON payload (mean): {result['json_bytes_mean']:,} bytes")
    if "pdf_bytes_mean" in result:
        print(f"  PDF output (mean):   {result['pdf_bytes_mean']:,} bytes")
    print(f"  End-to-end mean:     {total_mean:.3f} ms  ({1000 / total_mean:.1f} ops/s)")
    print()
    print(f"  {'Phase':<14} {'Mean ms':>9} {'P50 ms':>9} {'P95 ms':>9} {'% total':>8}")
    print(f"  {'-' * 14} {'-' * 9} {'-' * 9} {'-' * 9} {'-' * 8}")
    for key in ("to_dict", "json_dumps", "cgo_call", "copy_back", "free_result"):
        if key not in phases:
            continue
        p = phases[key]
        pct = (p.mean / total_mean * 100) if total_mean else 0
        print(f"  {p.name:<14} {p.mean:>9.3f} {p.p50:>9.3f} {p.p95:>9.3f} {pct:>7.1f}%")
    if "to_dict" in phases and "json_dumps" in phases:
        serialize_mean = phases["to_dict"].mean + phases["json_dumps"].mean
        print(f"  {'serialize':<14} {serialize_mean:>9.3f} {'':>9} {'':>9} {serialize_mean / total_mean * 100:>7.1f}%")


def run_cprofile_weighted(iterations: int) -> str:
    retail = build_retail_template()
    active = build_active_trader_template()
    hft = build_hft_template()
    templates = (
        ([retail] * 80) + ([active] * 15) + ([hft] * 5)
    ) * max(1, iterations // 100)

    def workload() -> None:
        for template in templates[:iterations]:
            call_bytes_result(get_lib().GeneratePDF, json.dumps(template.to_dict()).encode("utf-8"))

    profiler = cProfile.Profile()
    profiler.enable()
    workload()
    profiler.disable()

    stream = io.StringIO()
    stats = pstats.Stats(profiler, stream=stream)
    stats.strip_dirs()
    stats.sort_stats("cumulative")
    stats.print_stats(25)
    return stream.getvalue()


def run_cprofile_cached(iterations: int) -> str:
    retail = build_retail_template()
    active = build_active_trader_template()
    hft = build_hft_template()
    cached = {
        "retail": json.dumps(retail.to_dict()).encode("utf-8"),
        "active": json.dumps(active.to_dict()).encode("utf-8"),
        "hft": json.dumps(hft.to_dict()).encode("utf-8"),
    }
    lib = get_lib()

    def workload() -> None:
        rng = random.Random(0)
        for _ in range(iterations):
            roll = rng.randint(0, 99)
            if roll < 80:
                payload = cached["retail"]
            elif roll < 95:
                payload = cached["active"]
            else:
                payload = cached["hft"]
            call_bytes_result(lib.GeneratePDF, payload)

    profiler = cProfile.Profile()
    profiler.enable()
    workload()
    profiler.disable()

    stream = io.StringIO()
    stats = pstats.Stats(profiler, stream=stream)
    stats.strip_dirs()
    stats.sort_stats("cumulative")
    stats.print_stats(20)
    return stream.getvalue()


def main() -> None:
    print("=== PyPDFSuit Phase Profiler ===")
    print(f"Python: {sys.version.split()[0]}")
    print(f"Repo:   {REPO_ROOT}")

    templates = {
        "retail": build_retail_template(),
        "active": build_active_trader_template(),
        "hft": build_hft_template(),
    }

    results: dict[str, dict] = {}
    for name, template in templates.items():
        iters = 20 if name == "hft" else SINGLE_THREAD_ITERS
        results[name] = profile_generate(template, iters)
        print_phase_report(name.upper(), results[name], iters)

        cached = profile_cached_json(template, iters)
        print_phase_report(f"{name.upper()} (cached JSON control)", cached, iters)

    # Weighted mix estimate from per-template means
    print("\n=== Weighted Mix Estimate (80/15/5, single-thread) ===")
    retail_t = results["retail"]["phases"]["total"].mean
    active_t = results["active"]["phases"]["total"].mean
    hft_t = results["hft"]["phases"]["total"].mean
    weighted = 0.80 * retail_t + 0.15 * active_t + 0.05 * hft_t
    print(f"  Expected mean latency: {weighted:.3f} ms  ({1000 / weighted:.1f} ops/s theoretical max, 1 thread)")
    print(f"  Observed bench (~253 ops/s @ 48 workers) implies ~{1000/253:.1f} ms effective + contention")

    print("\n=== cProfile: weighted mix (full path) ===")
    print(run_cprofile_weighted(PROFILE_ITERS))

    print("\n=== cProfile: weighted mix (cached JSON) ===")
    print(run_cprofile_cached(PROFILE_ITERS))


if __name__ == "__main__":
    main()