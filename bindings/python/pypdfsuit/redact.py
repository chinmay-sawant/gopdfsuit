"""
PDF Redaction functionality.
"""

import json
from ._bindings import get_lib, call_bytes_result

def get_page_info(pdf_data: bytes) -> dict:
    """
    Get page count and dimensions from a PDF.

    Args:
        pdf_data: The PDF content as bytes

    Returns:
        dict: Page information (totalPages, pages list with dimensions)
    """
    if not pdf_data:
        raise ValueError("PDF data cannot be empty")

    lib = get_lib()
    result_bytes = call_bytes_result(
        lib.GetPageInfo,
        pdf_data,
        len(pdf_data)
    )
    
    return json.loads(result_bytes)

def extract_text_positions(pdf_data: bytes, page_num: int) -> list:
    """
    Extract text positions from a specific page.

    Args:
        pdf_data: The PDF content as bytes
        page_num: Page number (1-based)

    Returns:
        list: List of dictionaries containing text and coordinates
    """
    if not pdf_data:
        raise ValueError("PDF data cannot be empty")

    lib = get_lib()
    result_bytes = call_bytes_result(
        lib.ExtractTextPositions,
        pdf_data,
        len(pdf_data),
        page_num
    )
    
    return json.loads(result_bytes)

def apply_redactions(pdf_data: bytes, redactions: list[dict]) -> bytes:
    """
    Apply visual redaction rectangles to the PDF.

    Args:
        pdf_data: The PDF content as bytes
        redactions: List of dictionaries with keys: pageNum, x, y, width, height

    Returns:
        bytes: The redacted PDF content
    """
    if not pdf_data:
        raise ValueError("PDF data cannot be empty")
    
    if not redactions:
        return pdf_data

    redactions_json = json.dumps(redactions).encode("utf-8")
    
    lib = get_lib()
    return call_bytes_result(
        lib.ApplyRedactions,
        pdf_data,
        len(pdf_data),
        redactions_json
    )

def find_text_occurrences(pdf_data: bytes, text: str) -> list[dict]:
    """
    Find text occurrences and return candidate redaction rectangles.

    Args:
        pdf_data: The PDF content as bytes
        text: Search text

    Returns:
        list[dict]: Redaction rectangles (pageNum, x, y, width, height)
    """
    if not pdf_data:
        raise ValueError("PDF data cannot be empty")

    lib = get_lib()
    result_bytes = call_bytes_result(
        lib.FindTextOccurrences,
        pdf_data,
        len(pdf_data),
        text.encode("utf-8"),
    )

    return json.loads(result_bytes)

def apply_redactions_advanced(pdf_data: bytes, options: dict) -> bytes:
    """
    Apply a unified redaction request.

    Args:
        pdf_data: The PDF content as bytes
        options: Dictionary with optional keys:
            - blocks: list[dict]
            - textSearch: list[dict] e.g. [{"text": "secret"}]
            - mode: "visual_allowed" or "secure_required"
            - password: string

    Returns:
        bytes: The redacted PDF content
    """
    if not pdf_data:
        raise ValueError("PDF data cannot be empty")

    options_json = json.dumps(options).encode("utf-8")

    lib = get_lib()
    return call_bytes_result(
        lib.ApplyRedactionsAdvanced,
        pdf_data,
        len(pdf_data),
        options_json,
    )
