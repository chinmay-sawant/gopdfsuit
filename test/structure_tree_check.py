#!/usr/bin/env python3
"""Strict structure-tree consistency checks that veraPDF does not enforce.

Validates ParentTree[MCID] -> owning TD/TH (not TR) and TR /Pg consistency with
child TD pages. These are the checks PAC/Adobe-style validators apply and that
caught the Zerodha ParentTree regression (2026-06).
"""

from __future__ import annotations

import argparse
import json
import re
import sys
from dataclasses import dataclass, field
from pathlib import Path


@dataclass
class Finding:
    code: str
    message: str
    count: int = 1


@dataclass
class StructureTreeReport:
    pdf: str
    ok: bool
    parent_tree_total: int = 0
    parent_tree_tr_refs: int = 0
    parent_tree_td_refs: int = 0
    tr_td_page_mismatches: int = 0
    empty_sect_count: int = 0
    findings: list[Finding] = field(default_factory=list)

    def to_dict(self) -> dict:
        return {
            "pdf": self.pdf,
            "ok": self.ok,
            "parent_tree_total": self.parent_tree_total,
            "parent_tree_tr_refs": self.parent_tree_tr_refs,
            "parent_tree_td_refs": self.parent_tree_td_refs,
            "tr_td_page_mismatches": self.tr_td_page_mismatches,
            "empty_sect_count": self.empty_sect_count,
            "findings": [
                {"code": f.code, "message": f.message, "count": f.count}
                for f in self.findings
            ],
        }


def parse_objects(data: bytes) -> dict[int, str]:
    objs: dict[int, str] = {}
    for match in re.finditer(rb"(\d+) (\d+) obj\s*<<(.*?)>>\s*endobj", data, re.S):
        objs[int(match.group(1))] = match.group(3).decode("latin1", "replace")
    return objs


def struct_type(body: str) -> str | None:
    match = re.search(r"/S\s*/(\w+)", body)
    return match.group(1) if match else None


def page_ref(body: str) -> int | None:
    match = re.search(r"/Pg\s+(\d+)\s+0\s+R", body)
    return int(match.group(1)) if match else None


def parent_tree_refs(objs: dict[int, str]) -> list[int]:
    root = next((body for body in objs.values() if "/Type /StructTreeRoot" in body), None)
    if not root:
        return []
    match = re.search(r"/ParentTree\s+(\d+)\s+0\s+R", root)
    if not match:
        return []
    pt_body = objs.get(int(match.group(1)), "")
    return [int(n) for n in re.findall(r"(\d+)\s+0\s+R", pt_body)]


def analyze_pdf(pdf_path: str) -> StructureTreeReport:
    report = StructureTreeReport(pdf=pdf_path, ok=True)
    data = Path(pdf_path).read_bytes()
    objs = parse_objects(data)

    refs = parent_tree_refs(objs)
    report.parent_tree_total = len(refs)
    for ref in refs:
        body = objs.get(ref, "")
        stype = struct_type(body)
        if stype == "TR":
            report.parent_tree_tr_refs += 1
        elif stype in {"TD", "TH"}:
            report.parent_tree_td_refs += 1

    if report.parent_tree_tr_refs > 0:
        report.ok = False
        report.findings.append(
            Finding(
                code="parent_tree_tr_ref",
                message="ParentTree maps MCID slots to TR instead of TD/TH",
                count=report.parent_tree_tr_refs,
            )
        )

    for _obj_id, body in objs.items():
        if "/Type /StructElem" not in body or "/S /TR" not in body:
            continue
        tr_pg = page_ref(body)
        if tr_pg is None:
            continue
        kids_match = re.search(r"/K\s*\[(.*?)\]", body, re.S)
        if not kids_match:
            continue
        for kid_ref in re.findall(r"(\d+)\s+0\s+R", kids_match.group(1)):
            td_body = objs.get(int(kid_ref), "")
            td_pg = page_ref(td_body)
            if td_pg is not None and td_pg != tr_pg:
                report.tr_td_page_mismatches += 1

    if report.tr_td_page_mismatches > 0:
        report.ok = False
        report.findings.append(
            Finding(
                code="tr_td_page_mismatch",
                message="TR /Pg does not match child TD /Pg on multi-page tables",
                count=report.tr_td_page_mismatches,
            )
        )

    for body in objs.values():
        if (
            "/Type /StructElem" in body
            and "/S /Sect" in body
            and "/K " not in body
        ):
            report.empty_sect_count += 1

    if report.empty_sect_count > 0:
        report.findings.append(
            Finding(
                code="empty_sect",
                message="Bookmark Sect elements without /K (informational; PAC may flag)",
                count=report.empty_sect_count,
            )
        )

    return report


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("pdfs", nargs="+", help="PDF files to check")
    parser.add_argument("--format", choices=("text", "json"), default="text")
    parser.add_argument("--fail-on-empty-sect", action="store_true")
    args = parser.parse_args()

    reports: list[StructureTreeReport] = []
    exit_code = 0
    for pdf in args.pdfs:
        if not Path(pdf).is_file():
            print(f"FAIL structure-tree: file not found: {pdf}", file=sys.stderr)
            exit_code = 1
            continue
        report = analyze_pdf(pdf)
        if args.fail_on_empty_sect:
            for finding in report.findings:
                if finding.code == "empty_sect":
                    report.ok = False
        if not report.ok:
            exit_code = 1
        reports.append(report)

    if args.format == "json":
        json.dump([r.to_dict() for r in reports], sys.stdout, indent=2)
        sys.stdout.write("\n")
    else:
        for report in reports:
            rel = report.pdf
            if report.ok:
                print(
                    f"ParentTree TD={report.parent_tree_td_refs}/{report.parent_tree_total}, "
                    f"TR mismatches=0"
                )
            else:
                print(f"FAIL structure-tree: {rel}")
                for finding in report.findings:
                    if finding.code == "empty_sect" and not args.fail_on_empty_sect:
                        print(f"  INFO [{finding.code}] ({finding.count}x) {finding.message}")
                        continue
                    print(f"  • [{finding.code}] ({finding.count}x) {finding.message}")

    return exit_code


if __name__ == "__main__":
    raise SystemExit(main())