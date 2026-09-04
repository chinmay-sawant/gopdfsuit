"""Triple-gate anchor (D1): the same golden fixture checked by Go
internal/models TestTemplateSchemaGolden and frontend npm run test:schema
must generate a valid PDF through the Python bindings."""

import json
from pathlib import Path

import pytest

from pypdfsuit._bindings import call_bytes_result, get_lib

_REPO = Path(__file__).resolve().parents[3]
_SCHEMA = _REPO / "frontend" / "template.schema.json"
_GOLDEN = _REPO / "sampledata" / "editor" / "financial_report.json"


@pytest.mark.skipif(not _GOLDEN.exists(), reason="golden financial_report.json not found")
def test_golden_template_generates_pdf():
    template_dict = json.loads(_GOLDEN.read_text())
    lib = get_lib()
    pdf_bytes = call_bytes_result(lib.GeneratePDF, json.dumps(template_dict).encode("utf-8"))

    assert pdf_bytes[:5] == b"%PDF-"
    assert len(pdf_bytes) > 1000


@pytest.mark.skipif(not _SCHEMA.exists(), reason="template.schema.json not found")
def test_golden_top_level_keys_match_schema():
    schema = json.loads(_SCHEMA.read_text())
    golden = json.loads(_GOLDEN.read_text())
    properties = set(schema.get("properties", {}))

    assert schema.get("title") == "GoPdfSuit PDFTemplate"
    assert properties, "schema has no top-level properties"
    assert set(golden) <= properties, (
        f"golden keys drifted from schema: {sorted(set(golden) - properties)}"
    )
