"""
PDF merging functionality.
"""

from typing import List

from ._bindings import get_lib, call_bytes_result, merge_args


def merge_pdfs(pdf_files: List[bytes]) -> bytes:
    """
    Merge multiple PDFs into a single document.

    Args:
        pdf_files: List of PDF file contents as bytes

    Returns:
        bytes: The merged PDF file content

    Raises:
        GoPDFSuitError: If merging fails
        ValueError: If no PDF files provided

    Example:
        >>> from pypdfsuit import merge_pdfs
        >>> with open("doc1.pdf", "rb") as f1, open("doc2.pdf", "rb") as f2:
        ...     merged = merge_pdfs([f1.read(), f2.read()])
        >>> with open("merged.pdf", "wb") as f:
        ...     f.write(merged)
    """
    lib = get_lib()
    return call_bytes_result(lib.MergePDFs, *merge_args(pdf_files))
