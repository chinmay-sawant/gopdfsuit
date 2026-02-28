"""
Integration tests for pypdfsuit — mirrors the Go integration_test.go.

Each test uses the same sample data as the Go suite and writes output
files with a _python.pdf (or _python.zip / _python.png) suffix into
the corresponding sampledata/ directory.
"""

import json
import os
from pathlib import Path

import pytest

from pypdfsuit import (
    generate_pdf,
    merge_pdfs,
    split_pdf,
    fill_pdf_with_xfdf,
    convert_html_to_pdf,
    convert_html_to_image,
    PDFTemplate,
    SplitSpec,
    HtmlToPDFRequest,
    HtmlToImageRequest,
)
from pypdfsuit._bindings import get_lib, call_bytes_result

# Resolve paths relative to the repo root
_REPO_ROOT = Path(__file__).resolve().parents[3]
_SAMPLEDATA = _REPO_ROOT / "sampledata"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

_MATH_FONT_CANDIDATES = [
    "/usr/share/fonts/truetype/noto/NotoSansMath-Regular.ttf",
    "/usr/share/fonts/opentype/noto/NotoSansMath-Regular.otf",
    "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
    "/usr/share/fonts/truetype/liberation2/LiberationSans-Regular.ttf",
]


def _resolve_math_font() -> str | None:
    for path in _MATH_FONT_CANDIDATES:
        if os.path.isfile(path):
            return path
    return None


def _generate_pdf_from_dict(template_dict: dict) -> bytes:
    """Generate a PDF from a raw JSON-compatible dict (bypasses dataclass construction)."""
    lib = get_lib()
    template_json = json.dumps(template_dict).encode("utf-8")
    return call_bytes_result(lib.GeneratePDF, template_json)


def _has_chrome() -> bool:
    """Check whether a Chrome/Chromium binary is available."""
    import shutil
    return any(
        shutil.which(name) is not None
        for name in ("google-chrome", "google-chrome-stable", "chromium", "chromium-browser")
    )


requires_chrome = pytest.mark.skipif(
    not _has_chrome(),
    reason="Chrome/Chromium not found on PATH",
)


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------


class TestGenerateTemplatePDF:
    """Mirrors Go TestGenerateTemplatePDF."""

    def test_generate_from_json(self):
        json_path = _SAMPLEDATA / "editor" / "financial_digitalsignature.json"
        if not json_path.exists():
            pytest.skip(f"Sample JSON not found: {json_path}")

        template_dict = json.loads(json_path.read_text())
        pdf_bytes = _generate_pdf_from_dict(template_dict)

        assert pdf_bytes is not None
        assert len(pdf_bytes) > 0
        assert pdf_bytes[:5] == b"%PDF-"

        out_path = _SAMPLEDATA / "editor" / "temp_editor_python.pdf"
        out_path.write_bytes(pdf_bytes)

        expected_path = _SAMPLEDATA / "editor" / "generated.pdf"
        if expected_path.exists():
            exp_size = expected_path.stat().st_size
            gen_size = out_path.stat().st_size
            assert abs(gen_size - exp_size) <= 1024, (
                f"Size difference {abs(gen_size - exp_size)} exceeds tolerance 1024"
            )


class TestMergePDFs:
    """Mirrors Go TestMergePDFs."""

    def test_merge_three_pdfs(self):
        base = _SAMPLEDATA / "merge"
        files = ["em-16.pdf", "em-19.pdf", "em-51.pdf"]

        pdf_list = []
        for fname in files:
            fpath = base / fname
            if not fpath.exists():
                pytest.skip(f"Sample PDF not found: {fpath}")
            pdf_list.append(fpath.read_bytes())

        merged = merge_pdfs(pdf_list)

        assert merged is not None
        assert len(merged) > 0
        assert merged[:5] == b"%PDF-"

        out_path = base / "temp_merge_python.pdf"
        out_path.write_bytes(merged)

        expected_path = base / "generated.pdf"
        if expected_path.exists():
            assert out_path.stat().st_size == expected_path.stat().st_size


class TestFillPDF:
    """Mirrors Go TestFillPDF."""

    def test_fill_xfdf(self):
        base = _SAMPLEDATA / "filler"
        pdf_path = base / "us_hospital_encounter_acroform.pdf"
        xfdf_path = base / "us_hospital_encounter_data.xfdf"

        if not pdf_path.exists():
            pytest.skip("Sample PDF not found")
        if not xfdf_path.exists():
            pytest.skip("Sample XFDF not found")

        filled = fill_pdf_with_xfdf(pdf_path.read_bytes(), xfdf_path.read_bytes())

        assert filled is not None
        assert len(filled) > 0
        assert filled[:5] == b"%PDF-"

        out_path = base / "temp_filler_python.pdf"
        out_path.write_bytes(filled)

        expected_path = base / "generated.pdf"
        if expected_path.exists():
            assert out_path.stat().st_size == expected_path.stat().st_size


class TestHtmlToPDF:
    """Mirrors Go TestHtmlToPDF."""

    @requires_chrome
    def test_url_to_pdf(self):
        pdf_bytes = convert_html_to_pdf(
            HtmlToPDFRequest(url="https://en.wikipedia.org/wiki/Ana_de_Armas")
        )

        assert pdf_bytes is not None
        assert len(pdf_bytes) > 0

        out_dir = _SAMPLEDATA / "htmltopdf"
        out_dir.mkdir(parents=True, exist_ok=True)
        out_path = out_dir / "temp_htmltopdf_python.pdf"
        out_path.write_bytes(pdf_bytes)

        assert out_path.stat().st_size > 0


class TestHtmlToImage:
    """Mirrors Go TestHtmlToImage."""

    @requires_chrome
    def test_url_to_png(self):
        img_bytes = convert_html_to_image(
            HtmlToImageRequest(
                url="https://en.wikipedia.org/wiki/Ana_de_Armas",
                format="png",
            )
        )

        assert img_bytes is not None
        assert len(img_bytes) > 0

        out_dir = _SAMPLEDATA / "htmltoimg"
        out_dir.mkdir(parents=True, exist_ok=True)
        out_path = out_dir / "temp_htmltoimage_python.png"
        out_path.write_bytes(img_bytes)

        assert out_path.stat().st_size > 0


class TestSplitPDF:
    """Mirrors Go TestSplitPDF."""

    def test_split_single_page(self):
        base = _SAMPLEDATA / "split"
        pdf_path = base / "em.pdf"
        if not pdf_path.exists():
            pytest.skip("Sample PDF not found")

        result = split_pdf(pdf_path.read_bytes(), SplitSpec(pages=[10]))

        assert len(result) == 1
        assert result[0][:5] == b"%PDF-"

        out_path = base / "temp_split_python.pdf"
        out_path.write_bytes(result[0])

        expected_path = base / "split.pdf"
        if expected_path.exists():
            assert out_path.stat().st_size == expected_path.stat().st_size

    def test_split_page_range(self):
        base = _SAMPLEDATA / "split"
        pdf_path = base / "em.pdf"
        if not pdf_path.exists():
            pytest.skip("Sample PDF not found")

        result = split_pdf(pdf_path.read_bytes(), SplitSpec(ranges=[(10, 12)]))

        assert len(result) == 1
        assert result[0][:5] == b"%PDF-"

        out_path = base / "temp_split_range_python.pdf"
        out_path.write_bytes(result[0])

        expected_path = base / "split_range.pdf"
        if expected_path.exists():
            assert out_path.stat().st_size == expected_path.stat().st_size

    def test_split_max_per_file(self):
        base = _SAMPLEDATA / "split"
        pdf_path = base / "em.pdf"
        if not pdf_path.exists():
            pytest.skip("Sample PDF not found")

        result = split_pdf(
            pdf_path.read_bytes(),
            SplitSpec(ranges=[(10, 12)], max_per_file=1),
        )

        # With max_per_file=1 and 3 pages, expect 3 separate PDFs
        assert len(result) == 3
        for part in result:
            assert part[:5] == b"%PDF-"

        # Write first part as reference
        out_path = base / "temp_split_maxperfile_python.pdf"
        out_path.write_bytes(result[0])


class TestGenerateTypstMathShowcasePDF:
    """Mirrors Go TestGenerateTypstMathShowcasePDF."""

    def test_typst_math_showcase(self):
        base = _SAMPLEDATA / "typstsyntax"
        json_path = base / "typst_math_showcase.json"
        if not json_path.exists():
            pytest.skip("typst_math_showcase.json not found")

        template_dict = json.loads(json_path.read_text())
        pdf_bytes = _generate_pdf_from_dict(template_dict)

        assert pdf_bytes is not None
        assert len(pdf_bytes) > 0
        assert pdf_bytes[:5] == b"%PDF-"

        out_path = base / "typst_math_showcase_python.pdf"
        out_path.write_bytes(pdf_bytes)

        assert out_path.stat().st_size > 0


class TestGenerateTypstSamplePDF:
    """Mirrors Go TestGenerateTypstSamplePDF."""

    def test_typst_sample(self):
        math_font = _resolve_math_font()
        if math_font is None:
            pytest.skip(
                "No unicode math-capable font found "
                "(install fonts-dejavu-core or fonts-noto-math)"
            )

        base = _SAMPLEDATA / "typstsyntax"
        json_path = base / "typst_sample.json"
        if not json_path.exists():
            pytest.skip("typst_sample.json not found")

        template_dict = json.loads(json_path.read_text())

        # Inject customFonts with the resolved math font, matching Go test
        template_dict.setdefault("config", {})["customFonts"] = [
            {"name": "MathUnicode", "filePath": math_font}
        ]

        pdf_bytes = _generate_pdf_from_dict(template_dict)

        assert pdf_bytes is not None
        assert len(pdf_bytes) > 0
        assert pdf_bytes[:5] == b"%PDF-"

        out_path = base / "typst_sample_python.pdf"
        out_path.write_bytes(pdf_bytes)

        assert out_path.stat().st_size > 0
