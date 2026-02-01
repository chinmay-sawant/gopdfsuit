"""
Tests for PDF form filling functionality.
"""

import pytest
from pypdfsuit import fill_pdf_with_xfdf
from pypdfsuit._bindings import GoPDFSuitError


class TestFillPDFWithXFDF:
    """Tests for fill_pdf_with_xfdf function."""

    def test_empty_pdf_raises_error(self):
        """Test that empty PDF raises error."""
        xfdf = b"""<?xml version="1.0" encoding="UTF-8"?>
        <xfdf xmlns="http://ns.adobe.com/xfdf/">
            <fields>
                <field name="test"><value>test</value></field>
            </fields>
        </xfdf>"""

        with pytest.raises(ValueError):
            fill_pdf_with_xfdf(b"", xfdf)

    def test_empty_xfdf_raises_error(self):
        """Test that empty XFDF raises error."""
        with pytest.raises(ValueError):
            fill_pdf_with_xfdf(b"%PDF-1.4...", b"")
