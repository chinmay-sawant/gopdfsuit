"""
PDF generation functionality.
"""

import json
from typing import List, Optional

from .types import PDFTemplate, FontInfo
from ._bindings import get_lib, call_bytes_result

_JSON_CACHE_ATTR = "_pypdfsuit_json_cache"


def serialize_template(template: PDFTemplate, *, use_cache: bool = True) -> bytes:
    """
    Serialize a template to UTF-8 JSON bytes for GeneratePDF.

    Results are cached on the template instance by default. Call
    :func:`invalidate_template_cache` after mutating a template in place.
    """
    if use_cache:
        cached: Optional[bytes] = getattr(template, _JSON_CACHE_ATTR, None)
        if cached is not None:
            return cached

    payload = json.dumps(template.to_dict()).encode("utf-8")

    if use_cache:
        setattr(template, _JSON_CACHE_ATTR, payload)

    return payload


def invalidate_template_cache(template: PDFTemplate) -> None:
    """Drop cached JSON for a template after in-place mutation."""
    if hasattr(template, _JSON_CACHE_ATTR):
        delattr(template, _JSON_CACHE_ATTR)


def generate_pdf(
    template: PDFTemplate,
    *,
    template_json: Optional[bytes] = None,
    use_cache: bool = True,
) -> bytes:
    """
    Generate a PDF from a template.

    Args:
        template: PDFTemplate object with configuration and content
        template_json: Optional pre-serialized JSON bytes (skips to_dict/json.dumps)
        use_cache: When True, reuse cached JSON bytes for repeated calls with the
            same template object. Set False if the template was mutated without
            calling :func:`invalidate_template_cache`.

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
    payload = (
        template_json
        if template_json is not None
        else serialize_template(template, use_cache=use_cache)
    )
    return call_bytes_result(lib.GeneratePDF, payload)


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