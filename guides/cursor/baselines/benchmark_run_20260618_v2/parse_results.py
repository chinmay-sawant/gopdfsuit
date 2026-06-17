#!/usr/bin/env python3
"""Parse 5x benchmark logs and emit JSON summary (best-of-5)."""
from __future__ import annotations

import json
import re
import sys
from pathlib import Path

OUT = Path(__file__).parent


def best_throughput(text: str) -> float | None:
    vals = [float(x) for x in re.findall(r"Throughput:\s*([\d.]+)\s*ops/sec", text)]
    return max(vals) if vals else None


def best_k6_reqs(text: str) -> float | None:
    vals = []
    for m in re.finditer(r"http_reqs\.*:.*?\s([\d.]+)/s", text):
        vals.append(float(m.group(1)))
    return max(vals) if vals else None


def best_zerodha(text: str) -> float | None:
    vals = [float(x) for x in re.findall(r"Throughput:\s*([\d.]+)\s*ops/sec", text)]
    return max(vals) if vals else None


def best_go_bench(text: str, name: str) -> dict | None:
    pat = re.compile(
        rf"Benchmark{name}[^\n]*\s+\d+\s+([\d.]+)\s+ns/op(?:\s+([\d.]+)\s+MB/s)?(?:\s+([\d.]+)\s+B/op)?(?:\s+([\d.]+)\s+allocs/op)?"
    )
    rows = pat.findall(text)
    if not rows:
        # sub-benchmarks
        pat2 = re.compile(
            rf"(Benchmark{name}/[^\s]+)[^\n]*\s+\d+\s+([\d.]+)\s+ns/op(?:\s+([\d.]+)\s+MB/s)?(?:\s+([\d.]+)\s+B/op)?(?:\s+([\d.]+)\s+allocs/op)?"
        )
        subs = {}
        for m in pat2.finditer(text):
            subs[m.group(1)] = {
                "ns_per_op": float(m.group(2)),
                "mb_per_s": float(m.group(3)) if m.group(3) else None,
                "b_per_op": float(m.group(4)) if m.group(4) else None,
                "allocs_per_op": float(m.group(5)) if m.group(5) else None,
            }
        if subs:
            return {"sub": subs}
        return None
    best = min(rows, key=lambda r: float(r[0]))
    return {
        "ns_per_op": float(best[0]),
        "mb_per_s": float(best[1]) if best[1] else None,
        "b_per_op": float(best[2]) if best[2] else None,
        "allocs_per_op": float(best[3]) if best[3] else None,
        "ops_per_s": 1e9 / float(best[0]),
    }


def best_gopdfkit_compare(text: str) -> dict:
    pat = re.compile(r"(BenchmarkGoPDF(?:Kit|Lib)/[^\s]+)[^\n]*\s+\d+\s+[\d.]+\s+ns/op\s+([\d.]+)\s+pdf/s")
    best: dict[str, float] = {}
    for m in pat.finditer(text):
        k, v = m.group(1), float(m.group(2))
        best[k] = max(best.get(k, 0), v)
    return best


def parse_typst_manual(text: str) -> float | None:
    vals = []
    for m in re.finditer(r"Elapsed: ([\d.]+) ms", text):
        ms = float(m.group(1))
        if ms > 0:
            vals.append(1000.0 / ms)
    return max(vals) if vals else None


def main() -> None:
    summary: dict = {}

    mapping = {
        "gopdflib-data": "gopdflib-data_5x.txt",
        "gopdfsuit-zerodha": "gopdfsuit-zerodha_5x.txt",
        "pypdfsuit-zerodha": "pypdfsuit-zerodha_5x.txt",
        "pypdfsuit-legacy": "pypdfsuit-legacy_5x.txt",
        "fpdf": "fpdf_5x.txt",
        "jspdf": "jspdf_5x.txt",
        "pdfkit-lib": "pdfkit-lib_5x.txt",
        "pdflib": "pdflib_5x.txt",
    }
    for key, fname in mapping.items():
        p = OUT / fname
        if p.exists():
            v = best_throughput(p.read_text())
            if v:
                summary[key] = {"best_ops_per_s": v, "runs": 5}

    p = OUT / "gopdflib_zerodha_x5.txt"
    if p.exists():
        v = best_zerodha(p.read_text())
        if v:
            summary["gopdflib-zerodha"] = {"best_ops_per_s": v, "runs": 5, "source": "bench-gopdflib-zerodha-x5"}

    for key, fname, parser in [
        ("k6", "k6_5x.txt", best_k6_reqs),
        ("k6-light", "k6-light_5x.txt", best_k6_reqs),
        ("k6-retail", "k6-retail_5x.txt", best_k6_reqs),
        ("gotenberg", "gotenberg_5x.txt", best_k6_reqs),
    ]:
        p = OUT / fname
        if p.exists():
            v = parser(p.read_text())
            if v:
                summary[key] = {"best_req_per_s": v, "runs": 5}

    go_map = {
        "handler-serial": ("handler-all_5x.txt", "GenerateTemplatePDF_FinancialReport"),
        "handler-parallel": ("handler-parallel_5x.txt", "GenerateTemplatePDF_FinancialReport_Parallel"),
        "pdf-macro": ("pdf-macro_5x.txt", "GenerateTemplatePDF"),
        "pdf-wrap": ("pdf-wrap_5x.txt", "GenerateTemplatePDF_WrapEnabled"),
        "pdf-gopdfsuit": ("pdf-gopdfsuit_5x.txt", "GoPdfSuit"),
        "pdf-typst": ("pdf-typst_5x.txt", "Typst"),
    }
    for key, (fname, bname) in go_map.items():
        p = OUT / fname
        if p.exists():
            r = best_go_bench(p.read_text(), bname)
            if r:
                summary[key] = {**r, "runs": 5}

    p = OUT / "gopdfkit_compare_5x.txt"
    if p.exists():
        summary["gopdfkit-compare"] = {"best_pdf_per_s": best_gopdfkit_compare(p.read_text()), "runs": 5}

    p = OUT / "typst_manual_5x.txt"
    if p.exists():
        v = parse_typst_manual(p.read_text())
        if v:
            summary["typst-manual"] = {"best_ops_per_s": v, "runs": 5}

    p = OUT / "pdf_micro_10x.txt"
    if p.exists():
        r = best_go_bench(p.read_text(), "GoPdfSuit")
        if r:
            summary["pdf-micro"] = {**r, "runs": 10, "note": "Rows2000 + GoPdfSuit, count=10"}
        r2 = best_go_bench(p.read_text(), "GenerateTemplatePDF/Rows2000")
        if r2 and "sub" in r2:
            summary["pdf-micro-rows2000"] = {**r2["sub"], "runs": 10}

    if (OUT / ".done_pdf-gopdfsuit").exists():
        p = OUT / "pdf-gopdfsuit_5x.txt"
        if p.exists():
            r = best_go_bench(p.read_text(), "GoPdfSuit")
            if r:
                summary["pdf-gopdfsuit"] = {**r, "runs": 5}

    out_path = OUT / "summary.json"
    out_path.write_text(json.dumps(summary, indent=2))
    print(out_path)
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()