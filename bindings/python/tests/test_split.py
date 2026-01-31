"""
Tests for PDF splitting functionality.
"""

import pytest
from pypdfsuit import (
    split_pdf,
    parse_page_spec,
    generate_pdf,
    merge_pdfs,
    PDFTemplate,
    Config,
    Title,
    SplitSpec,
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


def create_multi_page_pdf(num_pages: int) -> bytes:
    """Helper to create a multi-page PDF by merging single-page PDFs."""
    pdfs = [create_simple_pdf(f"Page {i+1}") for i in range(num_pages)]
    return merge_pdfs(pdfs)


class TestParsePageSpec:
    """Tests for parse_page_spec function."""

    def test_single_page(self):
        """Test parsing a single page number."""
        pages = parse_page_spec("1", 10)
        assert pages == [1]

    def test_multiple_pages(self):
        """Test parsing multiple page numbers."""
        pages = parse_page_spec("1,3,5", 10)
        assert pages == [1, 3, 5]

    def test_page_range(self):
        """Test parsing a page range."""
        pages = parse_page_spec("1-3", 10)
        assert pages == [1, 2, 3]

    def test_mixed_pages_and_ranges(self):
        """Test parsing mixed pages and ranges."""
        pages = parse_page_spec("1-3,5,7-9", 10)
        assert pages == [1, 2, 3, 5, 7, 8, 9]

    def test_empty_spec(self):
        """Test parsing empty spec."""
        pages = parse_page_spec("", 10)
        assert pages is None or pages == []

    def test_invalid_page_number(self):
        """Test parsing invalid page number."""
        with pytest.raises(GoPDFSuitError):
            parse_page_spec("0", 10)

    def test_page_exceeds_total(self):
        """Test parsing page that exceeds total."""
        with pytest.raises(GoPDFSuitError):
            parse_page_spec("15", 10)


class TestSplitPDF:
    """Tests for split_pdf function."""

    def test_split_specific_pages(self):
        """Test splitting specific pages."""
        pdf = create_multi_page_pdf(3)
        spec = SplitSpec(pages=[1, 2])

        parts = split_pdf(pdf, spec)

        assert parts is not None
        assert len(parts) == 1  # Single output with specified pages
        assert parts[0].startswith(b"%PDF-")

    def test_split_max_per_file(self):
        """Test splitting with max pages per file."""
        pdf = create_multi_page_pdf(5)
        spec = SplitSpec(max_per_file=2)

        parts = split_pdf(pdf, spec)

        assert parts is not None
        assert len(parts) == 3  # 5 pages / 2 per file = 3 files
        for part in parts:
            assert part.startswith(b"%PDF-")

    def test_split_empty_spec(self):
        """Test splitting with empty spec (should return all pages)."""
        pdf = create_multi_page_pdf(3)
        spec = SplitSpec()

        parts = split_pdf(pdf, spec)

        assert parts is not None
        assert len(parts) == 1
        assert parts[0].startswith(b"%PDF-")

    def test_split_empty_pdf(self):
        """Test splitting empty PDF raises error."""
        with pytest.raises(ValueError):
            split_pdf(b"", SplitSpec())
