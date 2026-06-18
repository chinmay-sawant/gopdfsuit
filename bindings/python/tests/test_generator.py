"""
Tests for PDF generation functionality.
"""

import pytest
import json

from pypdfsuit import (
    generate_pdf,
    get_available_fonts,
    invalidate_template_cache,
    serialize_template,
    PDFTemplate,
    Config,
    Title,
    Element,
    Table,
    Row,
    Cell,
    Spacer,
    Footer,
)


class TestGeneratePDF:
    """Tests for generate_pdf function."""

    def test_basic_generation(self):
        """Test basic PDF generation with minimal template."""
        template = PDFTemplate(
            config=Config(page="A4", page_alignment=1),
            title=Title(
                props="Helvetica:18:100:center:0:0:0:0",
                text="Test Document",
            ),
            elements=[],
        )

        pdf_bytes = generate_pdf(template)

        assert pdf_bytes is not None
        assert len(pdf_bytes) > 0
        assert pdf_bytes.startswith(b"%PDF-")

    def test_generation_with_table(self):
        """Test PDF generation with a table element."""
        template = PDFTemplate(
            config=Config(page="A4", page_alignment=1),
            title=Title(
                props="Helvetica:18:100:center:0:0:0:0",
                text="Table Test",
            ),
            elements=[
                Element(
                    type="table",
                    table=Table(
                        max_columns=2,
                        column_widths=[1.0, 1.0],
                        rows=[
                            Row(
                                row=[
                                    Cell(
                                        props="Helvetica:12:100:left:1:1:1:1",
                                        text="Column 1",
                                    ),
                                    Cell(
                                        props="Helvetica:12:000:left:1:1:1:1",
                                        text="Column 2",
                                    ),
                                ]
                            ),
                            Row(
                                row=[
                                    Cell(
                                        props="Helvetica:12:000:left:1:1:1:1",
                                        text="Value 1",
                                    ),
                                    Cell(
                                        props="Helvetica:12:000:left:1:1:1:1",
                                        text="Value 2",
                                    ),
                                ]
                            ),
                        ],
                    ),
                )
            ],
        )

        pdf_bytes = generate_pdf(template)

        assert pdf_bytes is not None
        assert len(pdf_bytes) > 0
        assert pdf_bytes.startswith(b"%PDF-")

    def test_generation_with_spacer(self):
        """Test PDF generation with a spacer element."""
        template = PDFTemplate(
            config=Config(page="A4", page_alignment=1),
            title=Title(
                props="Helvetica:18:100:center:0:0:0:0",
                text="Spacer Test",
            ),
            elements=[
                Element(type="spacer", spacer=Spacer(height=50)),
            ],
        )

        pdf_bytes = generate_pdf(template)

        assert pdf_bytes is not None
        assert pdf_bytes.startswith(b"%PDF-")

    def test_generation_with_footer(self):
        """Test PDF generation with a footer."""
        template = PDFTemplate(
            config=Config(page="A4", page_alignment=1),
            title=Title(
                props="Helvetica:18:100:center:0:0:0:0",
                text="Footer Test",
            ),
            elements=[],
            footer=Footer(font="Helvetica:10", text="Page {page} of {pages}"),
        )

        pdf_bytes = generate_pdf(template)

        assert pdf_bytes is not None
        assert pdf_bytes.startswith(b"%PDF-")

    def test_landscape_orientation(self):
        """Test PDF generation with landscape orientation."""
        template = PDFTemplate(
            config=Config(page="A4", page_alignment=2),  # Landscape
            title=Title(
                props="Helvetica:18:100:center:0:0:0:0",
                text="Landscape Test",
            ),
            elements=[],
        )

        pdf_bytes = generate_pdf(template)

        assert pdf_bytes is not None
        assert pdf_bytes.startswith(b"%PDF-")

    def test_different_page_sizes(self):
        """Test PDF generation with different page sizes."""
        page_sizes = ["A4", "Letter", "Legal"]

        for page_size in page_sizes:
            template = PDFTemplate(
                config=Config(page=page_size, page_alignment=1),
                title=Title(
                    props="Helvetica:18:100:center:0:0:0:0",
                    text=f"{page_size} Test",
                ),
                elements=[],
            )

            pdf_bytes = generate_pdf(template)

            assert pdf_bytes is not None
            assert pdf_bytes.startswith(b"%PDF-")

    def test_template_json_cache_reuses_payload(self):
        """Repeated calls should not re-serialize the same template."""
        template = PDFTemplate(
            config=Config(page="A4", page_alignment=1),
            title=Title(
                props="Helvetica:18:100:center:0:0:0:0",
                text="Cache Test",
            ),
            elements=[],
        )

        first = serialize_template(template)
        second = serialize_template(template)

        assert first is second

    def test_generate_pdf_uses_cached_json(self, monkeypatch):
        """generate_pdf should serialize a template only once by default."""
        template = PDFTemplate(
            config=Config(page="A4", page_alignment=1),
            title=Title(
                props="Helvetica:18:100:center:0:0:0:0",
                text="Cache Test",
            ),
            elements=[],
        )
        dumps_calls = 0
        original_dumps = json.dumps

        def counting_dumps(*args, **kwargs):
            nonlocal dumps_calls
            dumps_calls += 1
            return original_dumps(*args, **kwargs)

        monkeypatch.setattr(json, "dumps", counting_dumps)

        pdf_one = generate_pdf(template)
        pdf_two = generate_pdf(template)

        assert pdf_one.startswith(b"%PDF-")
        assert pdf_two.startswith(b"%PDF-")
        assert dumps_calls == 1

    def test_invalidate_template_cache_forces_reserialize(self, monkeypatch):
        template = PDFTemplate(
            config=Config(page="A4", page_alignment=1),
            title=Title(
                props="Helvetica:18:100:center:0:0:0:0",
                text="Cache Test",
            ),
            elements=[],
        )
        dumps_calls = 0
        original_dumps = json.dumps

        def counting_dumps(*args, **kwargs):
            nonlocal dumps_calls
            dumps_calls += 1
            return original_dumps(*args, **kwargs)

        monkeypatch.setattr(json, "dumps", counting_dumps)

        generate_pdf(template)
        invalidate_template_cache(template)
        generate_pdf(template)

        assert dumps_calls == 2

    def test_generate_pdf_accepts_pre_serialized_json(self):
        template = PDFTemplate(
            config=Config(page="A4", page_alignment=1),
            title=Title(
                props="Helvetica:18:100:center:0:0:0:0",
                text="Pre-serialized Test",
            ),
            elements=[],
        )
        payload = serialize_template(template, use_cache=False)

        pdf_bytes = generate_pdf(template, template_json=payload)

        assert pdf_bytes.startswith(b"%PDF-")


class TestGetAvailableFonts:
    """Tests for get_available_fonts function."""

    def test_get_fonts(self):
        """Test getting available fonts."""
        fonts = get_available_fonts()

        assert fonts is not None
        assert isinstance(fonts, list)
        assert len(fonts) > 0

        # Check that each font has the expected attributes
        for font in fonts:
            assert hasattr(font, "id")
            assert hasattr(font, "name")
            assert hasattr(font, "display_name")
            assert hasattr(font, "reference")
