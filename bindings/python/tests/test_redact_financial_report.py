"""
Tests for redaction against the sample financial report.
"""

from pathlib import Path

from pypdfsuit import apply_redactions_advanced, find_text_occurrences


def _repo_root() -> Path:
    # tests/ -> python/ -> bindings/ -> repo root
    return Path(__file__).resolve().parents[3]


class TestFinancialReportRedaction:
    """Redaction tests using the sample financial report PDF."""

    def test_financial_report_text_redaction(self):
        """Redact text and persist output PDF at repository root for inspection."""
        repo_root = _repo_root()
        pdf_path = repo_root / "sampledata" / "financialreport" / "financial_report.pdf"
        pdf_bytes = pdf_path.read_bytes()

        out = apply_redactions_advanced(
            pdf_bytes,
            {
                "mode": "visual_allowed",
                "textSearch": [{"text": "SEC"}, {"text": "COM"}],
            },
        )

        assert out is not None
        assert len(out) > 0

        output_path = repo_root / "financial_report_redacted_pypdfsuit_test_output.pdf"
        output_path.write_bytes(out)
        assert output_path.exists()

        assert out != pdf_bytes

    def test_financial_report_page2_text_redaction(self):
        """Ensure SECTION C can be located on page 2 for targeted redaction."""
        repo_root = _repo_root()
        pdf_path = repo_root / "sampledata" / "financialreport" / "financial_report.pdf"
        pdf_bytes = pdf_path.read_bytes()

        rects = find_text_occurrences(pdf_bytes, "SECTION C")

        assert len(rects) > 0
        assert any(r.get("pageNum") == 2 for r in rects)
