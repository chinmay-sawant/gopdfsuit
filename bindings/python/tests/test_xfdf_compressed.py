"""Integration test: XFDF fill on zlib-compressed AcroForm PDF."""

from pathlib import Path

import pytest

from pypdfsuit import fill_pdf_with_xfdf

_SAMPLEDATA = Path(__file__).resolve().parents[3] / "sampledata"
_BASE = _SAMPLEDATA / "filler" / "compressed"


@pytest.mark.skipif(
    not (_BASE / "medical_form.pdf").exists(),
    reason="compressed medical_form.pdf not found",
)
@pytest.mark.skipif(
    not (_BASE / "medical_data.xfdf").exists(),
    reason="medical_data.xfdf not found",
)
def test_fill_compressed_xfdf():
    pdf_bytes = (_BASE / "medical_form.pdf").read_bytes()
    xfdf_bytes = (_BASE / "medical_data.xfdf").read_bytes()

    filled = fill_pdf_with_xfdf(pdf_bytes, xfdf_bytes)
    assert filled[:5] == b"%PDF-"
    assert len(filled) > 0

    out_path = _BASE / "temp_filler_compressed_python.pdf"
    out_path.write_bytes(filled)

    expected = _BASE / "generated.pdf"
    if expected.exists():
        assert abs(out_path.stat().st_size - expected.stat().st_size) < 500