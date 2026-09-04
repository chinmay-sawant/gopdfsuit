"""
Coverage test for PDF compression (Python bindings).

Skips cleanly until a pypdfsuit.compress binding lands (Go already has
gopdflib.CompressPDF + POST /api/v1/compress + WASM).
"""

import pytest

compress = pytest.importorskip("pypdfsuit.compress", reason="no pypdfsuit.compress binding yet")

from pypdfsuit import (
    generate_pdf,
    PDFTemplate,
    Config,
    Title,
)


def _dummy_pdf(text: str) -> bytes:
    template = PDFTemplate(
        config=Config(page="A4", page_alignment=1),
        title=Title(
            props="Helvetica:12:000:left:0:0:0:0",
            text=text,
        ),
        elements=[],
    )
    return generate_pdf(template)


class TestCompress:
    def test_compress_dummy_pdf(self):
        src = _dummy_pdf("seeded dummy paragraph for compress")
        out = compress.compress_pdf(src, level="medium")
        assert out.startswith(b"%PDF-")
        assert b"%%EOF" in out

    def test_compress_heavy(self):
        src = _dummy_pdf("seeded dummy paragraph for heavy compress")
        out = compress.compress_pdf(src, level="heavy")
        assert out.startswith(b"%PDF-")

    def test_compress_rejects_non_pdf(self):
        with pytest.raises(Exception):
            compress.compress_pdf(b"not-a-pdf", level="medium")
