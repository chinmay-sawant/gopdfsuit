"""
PDF form filling functionality.
"""

from ._bindings import get_lib, call_bytes_result


def fill_pdf_with_xfdf(pdf_data: bytes, xfdf_data: bytes) -> bytes:
    """
    Fill a PDF form with data from XFDF (XML Forms Data Format).

    XFDF is an XML-based format for representing form data and annotations
    in PDF documents.

    Args:
        pdf_data: The PDF form content as bytes
        xfdf_data: The XFDF data as bytes

    Returns:
        bytes: The filled PDF content

    Raises:
        GoPDFSuitError: If form filling fails
        ValueError: If pdf_data or xfdf_data is empty

    Example:
        >>> from pypdfsuit import fill_pdf_with_xfdf
        >>> with open("form.pdf", "rb") as f:
        ...     pdf_data = f.read()
        >>> with open("data.xfdf", "rb") as f:
        ...     xfdf_data = f.read()
        >>> filled = fill_pdf_with_xfdf(pdf_data, xfdf_data)
        >>> with open("filled.pdf", "wb") as f:
        ...     f.write(filled)
    """
    if not pdf_data:
        raise ValueError("PDF data cannot be empty")
    if not xfdf_data:
        raise ValueError("XFDF data cannot be empty")

    lib = get_lib()

    return call_bytes_result(
        lib.FillPDFWithXFDF,
        pdf_data,
        len(pdf_data),
        xfdf_data,
        len(xfdf_data),
    )
