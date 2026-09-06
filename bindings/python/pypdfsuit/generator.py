"""
PDF generation functionality.
"""

import json
from typing import List, Dict, Any

from .types import PDFTemplate, FontInfo
from ._bindings import get_lib, call_bytes_result, json_payload


def serialize_template(template: PDFTemplate) -> bytes:
    """Serialize a template to fresh UTF-8 JSON bytes for GeneratePDF."""
    return json_payload(template)


def generate_pdf(template: PDFTemplate) -> bytes:
    """
    Generate a PDF from a template.

    Args:
        template: PDFTemplate object with configuration and content

    Returns:
        bytes: The generated PDF file content

    Raises:
        GoPDFSuitError: If PDF generation fails

    Example:
        >>> from pypdfsuit.builder import TemplateBuilder, Font
        >>> b = TemplateBuilder("A4", True)
        >>> b.add_title("My Document", font="Helvetica", size=18, bold=True)
        >>> tb = b.add_table(2, 1.0, 1.0)
        >>> tb.add_row(Font("Helvetica").size(12).bold().cell("Name"), Font("Helvetica").size(12).cell("John Doe"))
        >>> pdf_bytes = b.generate()
    """
    lib = get_lib()
    return call_bytes_result(lib.GeneratePDF, json_payload(template))


def generate_pdf_from_dict(template_dict: Dict[str, Any]) -> bytes:
    """Generate a PDF from a raw JSON-compatible template dict.

    Bypasses dataclass construction for raw JSON-template parity
    (mirrors the helper used in tests/test_integration.py).
    """
    lib = get_lib()
    return call_bytes_result(lib.GeneratePDF, json_payload(template_dict))


def get_available_fonts() -> List[FontInfo]:
    """
    Get the list of available fonts for PDF generation.

    Returns:
        List[FontInfo]: List of available fonts

    Raises:
        GoPDFSuitError: If getting fonts fails
    """
    lib = get_lib()
    data = call_bytes_result(lib.GetAvailableFonts)
    fonts_data = json.loads(data.decode("utf-8"))

    return [
        FontInfo(
            id=f.get("id", ""),
            name=f.get("name", ""),
            display_name=f.get("displayName", ""),
            reference=f.get("reference", ""),
        )
        for f in fonts_data
    ]
