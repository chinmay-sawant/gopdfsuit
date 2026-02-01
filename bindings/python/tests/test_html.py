"""
Tests for HTML conversion functionality.

Note: These tests require Chrome/Chromium to be installed on the system.
They may be skipped if Chrome is not available.
"""

import pytest
from pypdfsuit import (
    convert_html_to_pdf,
    convert_html_to_image,
    HtmlToPDFRequest,
    HtmlToImageRequest,
)
from pypdfsuit._bindings import GoPDFSuitError


# Mark tests that require Chrome
requires_chrome = pytest.mark.skipif(
    True,  # Set to False if Chrome is available
    reason="Chrome/Chromium not available for testing",
)


class TestConvertHTMLToPDF:
    """Tests for convert_html_to_pdf function."""

    def test_missing_html_and_url(self):
        """Test that missing both html and url raises error."""
        request = HtmlToPDFRequest()
        with pytest.raises(ValueError):
            convert_html_to_pdf(request)

    @requires_chrome
    def test_convert_html_string(self):
        """Test converting HTML string to PDF."""
        request = HtmlToPDFRequest(
            html="<html><body><h1>Hello World</h1></body></html>",
            page_size="A4",
            orientation="Portrait",
        )

        pdf_bytes = convert_html_to_pdf(request)

        assert pdf_bytes is not None
        assert pdf_bytes.startswith(b"%PDF-")

    @requires_chrome
    def test_convert_with_margins(self):
        """Test converting HTML with custom margins."""
        request = HtmlToPDFRequest(
            html="<html><body><h1>Margin Test</h1></body></html>",
            page_size="A4",
            margin_top="20mm",
            margin_right="20mm",
            margin_bottom="20mm",
            margin_left="20mm",
        )

        pdf_bytes = convert_html_to_pdf(request)

        assert pdf_bytes is not None
        assert pdf_bytes.startswith(b"%PDF-")


class TestConvertHTMLToImage:
    """Tests for convert_html_to_image function."""

    def test_missing_html_and_url(self):
        """Test that missing both html and url raises error."""
        request = HtmlToImageRequest()
        with pytest.raises(ValueError):
            convert_html_to_image(request)

    @requires_chrome
    def test_convert_html_to_png(self):
        """Test converting HTML to PNG image."""
        request = HtmlToImageRequest(
            html="<html><body><h1>Hello World</h1></body></html>",
            format="png",
            width=800,
            height=600,
        )

        img_bytes = convert_html_to_image(request)

        assert img_bytes is not None
        # PNG magic bytes
        assert img_bytes[:8] == b"\x89PNG\r\n\x1a\n"

    @requires_chrome
    def test_convert_html_to_jpeg(self):
        """Test converting HTML to JPEG image."""
        request = HtmlToImageRequest(
            html="<html><body><h1>Hello World</h1></body></html>",
            format="jpg",
            width=800,
            height=600,
        )

        img_bytes = convert_html_to_image(request)

        assert img_bytes is not None
        # JPEG magic bytes
        assert img_bytes[:2] == b"\xff\xd8"
