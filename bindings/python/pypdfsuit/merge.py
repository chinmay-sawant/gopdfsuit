"""
PDF merging functionality.
"""

import ctypes
from ctypes import c_char_p, c_int, POINTER
from typing import List

from ._bindings import get_lib, call_bytes_result


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
    if not pdf_files:
        raise ValueError("At least one PDF file is required")

    lib = get_lib()

    # Create C arrays
    count = len(pdf_files)
    pdf_data_array = (c_char_p * count)()
    pdf_lengths_array = (c_int * count)()

    for i, pdf in enumerate(pdf_files):
        pdf_data_array[i] = ctypes.create_string_buffer(pdf).raw
        pdf_lengths_array[i] = len(pdf)

    return call_bytes_result(
        lib.MergePDFs,
        ctypes.cast(pdf_data_array, POINTER(c_char_p)),
        ctypes.cast(pdf_lengths_array, POINTER(c_int)),
        count,
    )
