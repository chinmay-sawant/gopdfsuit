#!/usr/bin/env python3
"""Parse veraPDF JSON reports and render compliance results for verify_pdfs.sh."""

from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
from dataclasses import dataclass, field
from typing import Any


FLAVOUR_LABELS = {
    "1a": "PDF/A-1a",
    "1b": "PDF/A-1b",
    "2a": "PDF/A-2a",
    "2b": "PDF/A-2b",
    "2u": "PDF/A-2u",
    "3a": "PDF/A-3a",
    "3b": "PDF/A-3b",
    "3u": "PDF/A-3u",
    "4": "PDF/A-4",
    "4f": "PDF/A-4f",
    "4e": "PDF/A-4e",
    "ua1": "PDF/UA-1",
    "ua2": "PDF/UA-2",
    "wt1r": "PDF/WT-1r",
    "wt1a": "PDF/WT-1a",
}


@dataclass
class FailureItem:
    clause: str
    message: str
    count: int = 1


@dataclass
class ComplianceResult:
    pdf: str
    flavour: str
    profile: str
    compliant: bool
    passed_rules: int = 0
    failed_rules: int = 0
    passed_checks: int = 0
    failed_checks: int = 0
    summary: str = ""
    failures: list[FailureItem] = field(default_factory=list)
    error: str = ""

    def to_dict(self) -> dict[str, Any]:
        return {
            "pdf": self.pdf,
            "flavour": self.flavour,
            "profile": self.profile,
            "compliant": self.compliant,
            "passed_rules": self.passed_rules,
            "failed_rules": self.failed_rules,
            "passed_checks": self.passed_checks,
            "failed_checks": self.failed_checks,
            "summary": self.summary,
            "failures": [
                {"clause": f.clause, "message": f.message, "count": f.count}
                for f in self.failures
            ],
            "error": self.error,
        }


def flavour_label(flavour: str) -> str:
    return FLAVOUR_LABELS.get(flavour.lower(), flavour.upper())


def _collapse_ws(text: str) -> str:
    return re.sub(r"\s+", " ", text).strip()


def _shorten(text: str, limit: int = 96) -> str:
    text = _collapse_ws(text)
    if len(text) <= limit:
        return text
    return text[: limit - 3].rstrip() + "..."


def _rule_failure_message(rule: dict[str, Any]) -> str:
    checks = rule.get("checks") or []
    for check in checks:
        if check.get("status") == "failed":
            msg = check.get("errorMessage") or ""
            if msg:
                return _collapse_ws(msg)
    return _collapse_ws(rule.get("description") or "Validation rule failed")


def extract_failures(validation_result: dict[str, Any], max_items: int = 5) -> list[FailureItem]:
    details = validation_result.get("details") or {}
    failures: list[FailureItem] = []
    for rule in details.get("ruleSummaries") or []:
        if rule.get("status") != "failed":
            continue
        clause = str(rule.get("clause") or rule.get("testNumber") or "?")
        message = _rule_failure_message(rule)
        count = int(rule.get("failedChecks") or 1)
        failures.append(FailureItem(clause=clause, message=message, count=count))
        if len(failures) >= max_items:
            break
    return failures


def build_summary(result: ComplianceResult, max_items: int = 3) -> str:
    if result.error:
        return _shorten(result.error, 120)
    if result.compliant:
        return f"compliant ({result.passed_rules} rules, {result.passed_checks} checks)"
    if not result.failures:
        return f"{result.failed_rules} rule(s) failed ({result.failed_checks} checks)"
    parts = [_shorten(item.message, 72) for item in result.failures[:max_items]]
    summary = "; ".join(parts)
    remaining = result.failed_rules - min(len(result.failures), max_items)
    if remaining > 0:
        summary += f"; +{remaining} more rule(s)"
    return _shorten(summary, 140)


def run_verapdf_check(verapdf_bin: str, pdf: str, flavour: str, max_items: int = 5) -> ComplianceResult:
    profile = flavour_label(flavour)
    result = ComplianceResult(pdf=pdf, flavour=flavour, profile=profile, compliant=False)

    if not os.path.isfile(pdf):
        result.error = "file not found"
        result.summary = result.error
        return result

    proc = subprocess.run(
        [verapdf_bin, "-f", flavour, "--format", "json", "--loglevel", "0", pdf],
        capture_output=True,
        text=True,
    )

    stdout = (proc.stdout or "").strip()
    if not stdout:
        stderr = _collapse_ws(proc.stderr or "")
        result.error = stderr or f"veraPDF produced no output (exit {proc.returncode})"
        result.summary = _shorten(result.error, 140)
        return result

    try:
        payload = json.loads(stdout)
    except json.JSONDecodeError as exc:
        result.error = f"invalid veraPDF JSON: {exc}"
        result.summary = result.error
        return result

    try:
        job = payload["report"]["jobs"][0]
        validation = job["validationResult"][0]
    except (KeyError, IndexError, TypeError) as exc:
        result.error = f"unexpected veraPDF report shape: {exc}"
        result.summary = result.error
        return result

    details = validation.get("details") or {}
    result.compliant = bool(validation.get("compliant"))
    result.passed_rules = int(details.get("passedRules") or 0)
    result.failed_rules = int(details.get("failedRules") or 0)
    result.passed_checks = int(details.get("passedChecks") or 0)
    result.failed_checks = int(details.get("failedChecks") or 0)
    result.failures = extract_failures(validation, max_items=max_items)
    result.summary = build_summary(result, max_items=3)
    return result


class Colors:
    def __init__(self, enabled: bool) -> None:
        self.enabled = enabled

    def wrap(self, code: str, text: str) -> str:
        if not self.enabled:
            return text
        return f"\033[{code}m{text}\033[0m"

    def bold(self, text: str) -> str:
        return self.wrap("1", text)

    def green(self, text: str) -> str:
        return self.wrap("32", text)

    def red(self, text: str) -> str:
        return self.wrap("31", text)

    def yellow(self, text: str) -> str:
        return self.wrap("33", text)

    def cyan(self, text: str) -> str:
        return self.wrap("36", text)

    def dim(self, text: str) -> str:
        return self.wrap("2", text)


def _display_path(path: str, sampledata: str | None = None) -> str:
    if sampledata:
        prefix = sampledata.rstrip("/") + "/"
        if path.startswith(prefix):
            return path[len(prefix) :]
    return path


def format_status(result: ComplianceResult, colors: Colors) -> str:
    if result.compliant:
        return colors.green("PASS")
    return colors.red(colors.bold("FAIL"))


def print_result_lines(result: ComplianceResult, colors: Colors, sampledata: str | None = None) -> None:
    rel = _display_path(result.pdf, sampledata)
    status = format_status(result, colors)
    profile = colors.cyan(result.profile)
    print(f"{status} {profile}: {result.summary}")
    if not result.compliant and result.failures:
        for item in result.failures:
            count_suffix = f" ({item.count}×)" if item.count > 1 else ""
            line = f"  {colors.red('•')} [{item.clause}]{count_suffix} {item.message}"
            print(line)


def print_table(results: list[ComplianceResult], colors: Colors, sampledata: str | None = None) -> None:
    if not results:
        return

    rows: list[tuple[str, str, str, str]] = []
    for result in results:
        rows.append(
            (
                _display_path(result.pdf, sampledata),
                result.profile,
                "PASS" if result.compliant else "FAIL",
                result.summary,
            )
        )

    headers = ("File", "Profile", "Status", "Summary")
    widths = [len(headers[0]), len(headers[1]), len(headers[2]), len(headers[3])]
    for row in rows:
        for idx, cell in enumerate(row):
            widths[idx] = max(widths[idx], len(cell))

    widths[0] = min(widths[0], 56)
    widths[3] = min(max(widths[3], 24), 80)

    def fit(text: str, width: int) -> str:
        if len(text) <= width:
            return text
        if width <= 3:
            return text[:width]
        return text[: width - 3] + "..."

    def color_cell(column: int, text: str) -> str:
        if column == 2:
            if text == "PASS":
                return colors.green(text)
            if text == "FAIL":
                return colors.red(colors.bold(text))
        if column == 1:
            return colors.cyan(text)
        return text

    separator = "+" + "+".join("-" * (w + 2) for w in widths) + "+"

    def render_row(cells: tuple[str, str, str, str], header: bool = False) -> str:
        rendered = []
        for idx, (cell, width) in enumerate(zip(cells, widths)):
            value = fit(cell, width)
            if not header:
                value = color_cell(idx, value)
            rendered.append(f" {value:<{width}} ")
        return "|" + "|".join(rendered) + "|"

    print("")
    print(colors.bold("veraPDF compliance summary"))
    print(separator)
    print(render_row(headers, header=True))
    print(separator)
    for row in rows:
        print(render_row(row))
    print(separator)

    failed = sum(1 for result in results if not result.compliant)
    passed = len(results) - failed
    print(
        f"Totals: {colors.green(str(passed))} passed, "
        f"{colors.red(str(failed))} failed, {len(results)} checks"
    )


def cmd_check(args: argparse.Namespace) -> int:
    result = run_verapdf_check(args.verapdf, args.pdf, args.flavour, max_items=args.max_failures)
    if args.json_out:
        os.makedirs(os.path.dirname(args.json_out) or ".", exist_ok=True)
        with open(args.json_out, "w", encoding="utf-8") as handle:
            json.dump(result.to_dict(), handle, indent=2)
            handle.write("\n")
    if args.format == "json":
        json.dump(result.to_dict(), sys.stdout, indent=2)
        sys.stdout.write("\n")
    else:
        colors = Colors(sys.stdout.isatty() and not args.no_color)
        print_result_lines(result, colors, sampledata=args.sampledata)
    return 0 if result.compliant and not result.error else 1


def cmd_table(args: argparse.Namespace) -> int:
    results: list[ComplianceResult] = []
    for path in args.results:
        with open(path, encoding="utf-8") as handle:
            payload = json.load(handle)
        results.append(
            ComplianceResult(
                pdf=payload.get("pdf", ""),
                flavour=payload.get("flavour", ""),
                profile=payload.get("profile", flavour_label(payload.get("flavour", ""))),
                compliant=bool(payload.get("compliant")),
                passed_rules=int(payload.get("passed_rules") or 0),
                failed_rules=int(payload.get("failed_rules") or 0),
                passed_checks=int(payload.get("passed_checks") or 0),
                failed_checks=int(payload.get("failed_checks") or 0),
                summary=payload.get("summary", ""),
                failures=[
                    FailureItem(
                        clause=item.get("clause", "?"),
                        message=item.get("message", ""),
                        count=int(item.get("count") or 1),
                    )
                    for item in payload.get("failures") or []
                ],
                error=payload.get("error", ""),
            )
        )

    colors = Colors(sys.stdout.isatty() and not args.no_color)
    if args.failed_only:
        results = [result for result in results if not result.compliant]
    print_table(results, colors, sampledata=args.sampledata)
    return 1 if any(not result.compliant for result in results) else 0


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)

    check = sub.add_parser("check", help="Run veraPDF and print a compliance result")
    check.add_argument("--verapdf", required=True, help="Path to veraPDF CLI")
    check.add_argument("--pdf", required=True, help="PDF file to validate")
    check.add_argument("--flavour", required=True, help="Validation profile flavour (e.g. 4, ua2)")
    check.add_argument("--json-out", help="Optional path to write JSON result")
    check.add_argument("--sampledata", help="sampledata/ prefix to strip from displayed paths")
    check.add_argument("--max-failures", type=int, default=5, help="Max failed rules to capture")
    check.add_argument("--format", choices=("text", "json"), default="text")
    check.add_argument("--no-color", action="store_true")
    check.set_defaults(func=cmd_check)

    table = sub.add_parser("table", help="Render a table from JSON result files")
    table.add_argument("results", nargs="+", help="JSON files produced by check --json-out")
    table.add_argument("--sampledata", help="sampledata/ prefix to strip from displayed paths")
    table.add_argument("--failed-only", action="store_true", help="Only show failed checks")
    table.add_argument("--no-color", action="store_true")
    table.set_defaults(func=cmd_table)

    return parser


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()
    return int(args.func(args))


if __name__ == "__main__":
    raise SystemExit(main())