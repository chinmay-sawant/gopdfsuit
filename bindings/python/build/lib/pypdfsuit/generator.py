"""
PDF generation functionality.
"""

import json
from typing import List

from .types import PDFTemplate, FontInfo
from ._bindings import get_lib, call_bytes_result


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
        >>> from pypdfsuit import generate_pdf, PDFTemplate, Config, Title
        >>> template = PDFTemplate(
        ...     config=Config(page="A4", page_alignment=1),
        ...     title=Title(props="Helvetica:18:100:center:0:0:0:0", text="My Document"),
        ...     elements=[]
        ... )
        >>> pdf_bytes = generate_pdf(template)
        >>> with open("output.pdf", "wb") as f:
        ...     f.write(pdf_bytes)
    """
    lib = get_lib()
    template_json = json.dumps(template.to_dict()).encode("utf-8")
    return call_bytes_result(lib.GeneratePDF, template_json)


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
