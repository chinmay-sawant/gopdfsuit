"""Integration test: generate PDF from sampledata/financialreport/financial_report.json."""

import json
from pathlib import Path

import pytest

from pypdfsuit._bindings import call_bytes_result, get_lib

_REPO = Path(__file__).resolve().parents[3]
_JSON = _REPO / "sampledata" / "financialreport" / "financial_report.json"
_OUT = _REPO / "sampledata" / "financialreport" / "financial_report.pdf"


def _generate_pdf_from_dict(template_dict: dict) -> bytes:
    lib = get_lib()
    template_json = json.dumps(template_dict).encode("utf-8")
    return call_bytes_result(lib.GeneratePDF, template_json)


@pytest.mark.skipif(not _JSON.exists(), reason="financial_report.json not found")
def test_generate_financial_report_pdf():
    template_dict = json.loads(_JSON.read_text())
    pdf_bytes = _generate_pdf_from_dict(template_dict)

    assert pdf_bytes[:5] == b"%PDF-"
    assert len(pdf_bytes) > 1000

    _OUT.write_bytes(pdf_bytes)
    assert _OUT.stat().st_size > 1000