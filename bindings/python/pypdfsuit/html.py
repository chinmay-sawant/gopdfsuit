"""
HTML to PDF/Image conversion functionality.

Note: These functions require Chrome/Chromium to be available on the system.
"""

import json

from .types import HtmlToPDFRequest, HtmlToImageRequest
from ._bindings import get_lib, call_bytes_result


def convert_html_to_pdf(request: HtmlToPDFRequest) -> bytes:
    """
    Convert HTML content or a URL to a PDF document.

    This function requires Chrome/Chromium to be available on the system.

    Args:
        request: HtmlToPDFRequest with HTML content or URL and conversion options

    Returns:
        bytes: The generated PDF content

    Raises:
        GoPDFSuitError: If conversion fails
        ValueError: If neither html nor url is provided

    Example:
        >>> from pypdfsuit import convert_html_to_pdf, HtmlToPDFRequest
        >>> # Convert HTML string
        >>> request = HtmlToPDFRequest(
        ...     html="<html><body><h1>Hello World</h1></body></html>",
        ...     page_size="A4",
        ...     orientation="Portrait",
        ... )
        >>> pdf_bytes = convert_html_to_pdf(request)
        >>> # Convert URL
        >>> request = HtmlToPDFRequest(
        ...     url="https://example.com",
        ...     page_size="Letter",
        ... )
        >>> pdf_bytes = convert_html_to_pdf(request)
    """
    if not request.html and not request.url:
        raise ValueError("Either html or url must be provided")

    lib = get_lib()
    request_json = json.dumps(request.to_dict()).encode("utf-8")

    return call_bytes_result(lib.ConvertHTMLToPDF, request_json)


def convert_html_to_image(request: HtmlToImageRequest) -> bytes:
    """
    Convert HTML content or a URL to an image.

    Supported formats: png, jpg/jpeg, svg (default: png).
    This function requires Chrome/Chromium to be available on the system.

    Args:
        request: HtmlToImageRequest with HTML content or URL and conversion options

    Returns:
        bytes: The generated image content

    Raises:
        GoPDFSuitError: If conversion fails
        ValueError: If neither html nor url is provided

    Example:
        >>> from pypdfsuit import convert_html_to_image, HtmlToImageRequest
        >>> request = HtmlToImageRequest(
        ...     html="<html><body><h1>Hello World</h1></body></html>",
        ...     format="png",
        ...     width=800,
        ...     height=600,
        ... )
        >>> img_bytes = convert_html_to_image(request)
        >>> with open("output.png", "wb") as f:
        ...     f.write(img_bytes)
    """
    if not request.html and not request.url:
        raise ValueError("Either html or url must be provided")

    lib = get_lib()
    request_json = json.dumps(request.to_dict()).encode("utf-8")

    return call_bytes_result(lib.ConvertHTMLToImage, request_json)
