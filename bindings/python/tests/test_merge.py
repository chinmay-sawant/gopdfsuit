"""
Tests for PDF merging functionality.
"""

import pytest
from pypdfsuit import (
    merge_pdfs,
    generate_pdf,
    PDFTemplate,
    Config,
    Title,
)
from pypdfsuit._bindings import GoPDFSuitError


def create_simple_pdf(title: str) -> bytes:
    """Helper to create a simple PDF for testing."""
    template = PDFTemplate(
        config=Config(page="A4", page_alignment=1),
        title=Title(
            props="Helvetica:18:100:center:0:0:0:0",
            text=title,
        ),
        elements=[],
    )
    return generate_pdf(template)


class TestMergePDFs:
    """Tests for merge_pdfs function."""

    def test_merge_two_pdfs(self):
        """Test merging two PDFs."""
        pdf1 = create_simple_pdf("Document 1")
        pdf2 = create_simple_pdf("Document 2")

        merged = merge_pdfs([pdf1, pdf2])

        assert merged is not None
        assert len(merged) > 0
        assert merged.startswith(b"%PDF-")

    def test_merge_multiple_pdfs(self):
        """Test merging multiple PDFs."""
        pdfs = [create_simple_pdf(f"Document {i}") for i in range(5)]

        merged = merge_pdfs(pdfs)

        assert merged is not None
        assert merged.startswith(b"%PDF-")

    def test_merge_single_pdf(self):
        """Test merging a single PDF (should return same content)."""
        pdf = create_simple_pdf("Single Document")

        merged = merge_pdfs([pdf])

        assert merged is not None
        assert merged.startswith(b"%PDF-")

    def test_merge_empty_list(self):
        """Test merging with empty list raises error."""
        with pytest.raises(ValueError):
            merge_pdfs([])
