"""
PDF splitting functionality.
"""

import json
from typing import List, Optional

from .types import SplitSpec
from ._bindings import get_lib, call_bytes_result, call_bytes_array_result, json_payload, pdf_args


def split_pdf(pdf_data: bytes, spec: SplitSpec) -> List[bytes]:
    """
    Split a PDF into multiple parts based on the specification.

    Args:
        pdf_data: The PDF file content as bytes
        spec: SplitSpec defining how to split the PDF

    Returns:
        List[bytes]: List of PDF parts as bytes

    Raises:
        GoPDFSuitError: If splitting fails
        ValueError: If pdf_data is empty

    Example:
        >>> from pypdfsuit import split_pdf, SplitSpec
        >>> with open("document.pdf", "rb") as f:
        ...     pdf_data = f.read()
        >>> # Split specific pages
        >>> spec = SplitSpec(pages=[1, 3, 5])
        >>> parts = split_pdf(pdf_data, spec)
        >>> # Or split every N pages
        >>> spec = SplitSpec(max_per_file=5)
        >>> parts = split_pdf(pdf_data, spec)
    """
    lib = get_lib()
    return call_bytes_array_result(lib.SplitPDF, *pdf_args(pdf_data), json_payload(spec))


def parse_page_spec(spec: str, total_pages: int = 0) -> List[int]:
    """
    Parse a page specification string into a sorted list of page numbers.

    Args:
        spec: Page specification string like "1-3,5,7-9"
        total_pages: Total number of pages for validation (0 = no validation)

    Returns:
        List[int]: Sorted list of 1-based page numbers

    Raises:
        GoPDFSuitError: If the specification is invalid

    Example:
        >>> from pypdfsuit import parse_page_spec
        >>> pages = parse_page_spec("1-3,5,7-9", 10)
        >>> print(pages)
        [1, 2, 3, 5, 7, 8, 9]
    """
    lib = get_lib()
    data = call_bytes_result(
        lib.ParsePageSpec,
        spec.encode("utf-8"),
        total_pages,
    )

    if not data:
        return []

    return json.loads(data.decode("utf-8"))
