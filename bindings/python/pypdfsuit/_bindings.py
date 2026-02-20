"""
Low-level ctypes bindings to the Go shared library.

This module handles loading the shared library and defining C function signatures.
"""

import ctypes
from ctypes import c_char, c_char_p, c_int, c_void_p, POINTER, Structure, cast
import platform
import os
from pathlib import Path


class ByteResult(Structure):
    """Result structure for functions returning bytes.
    
    Note: We use c_void_p for data instead of c_char_p because c_char_p
    treats data as null-terminated strings, which truncates binary data
    (like PDFs) that contain null bytes.
    """

    _fields_ = [
        ("data", c_void_p),  # Use void* to avoid null-termination issues
        ("length", c_int),
        ("error", c_char_p),
    ]


class ByteArrayResult(Structure):
    """Result structure for functions returning multiple byte arrays."""

    _fields_ = [
        ("data", POINTER(c_void_p)),  # Array of void* pointers
        ("lengths", POINTER(c_int)),
        ("count", c_int),
        ("error", c_char_p),
    ]


class GoPDFSuitError(Exception):
    """Exception raised when the Go library returns an error."""

    pass


def _find_library() -> str:
    """Find the shared library for the current platform."""
    system = platform.system()
    lib_dir = Path(__file__).parent / "lib"

    if system == "Linux":
        lib_name = "libgopdfsuit.so"
    elif system == "Darwin":
        lib_name = "libgopdfsuit.dylib"
    elif system == "Windows":
        lib_name = "gopdfsuit.dll"
    else:
        raise OSError(f"Unsupported platform: {system}")

    lib_path = lib_dir / lib_name

    # Also check parent directories for development
    if not lib_path.exists():
        # Try relative to the package
        alt_paths = [
            Path(__file__).parent.parent / "pypdfsuit" / "lib" / lib_name,
            Path(__file__).parent.parent / lib_name,
        ]
        for alt in alt_paths:
            if alt.exists():
                lib_path = alt
                break

    if not lib_path.exists():
        raise FileNotFoundError(
            f"Could not find {lib_name}. Make sure to build the shared library first.\n"
            f"Run: cd bindings/python && ./build.sh\n"
            f"Searched in: {lib_dir}"
        )

    return str(lib_path)


def _load_library():
    """Load the shared library and configure function signatures."""
    lib_path = _find_library()
    lib = ctypes.CDLL(lib_path)

    # GeneratePDF
    lib.GeneratePDF.argtypes = [c_char_p]
    lib.GeneratePDF.restype = ByteResult

    # MergePDFs
    lib.MergePDFs.argtypes = [POINTER(c_char_p), POINTER(c_int), c_int]
    lib.MergePDFs.restype = ByteResult

    # SplitPDF
    lib.SplitPDF.argtypes = [c_char_p, c_int, c_char_p]
    lib.SplitPDF.restype = ByteArrayResult

    # ParsePageSpec
    lib.ParsePageSpec.argtypes = [c_char_p, c_int]
    lib.ParsePageSpec.restype = ByteResult

    # FillPDFWithXFDF
    lib.FillPDFWithXFDF.argtypes = [c_char_p, c_int, c_char_p, c_int]
    lib.FillPDFWithXFDF.restype = ByteResult

    # ConvertHTMLToPDF
    lib.ConvertHTMLToPDF.argtypes = [c_char_p]
    lib.ConvertHTMLToPDF.restype = ByteResult

    # ConvertHTMLToImage
    lib.ConvertHTMLToImage.argtypes = [c_char_p]
    lib.ConvertHTMLToImage.restype = ByteResult

    # GetAvailableFonts
    lib.GetAvailableFonts.argtypes = []
    lib.GetAvailableFonts.restype = ByteResult

    # GetPageInfo
    lib.GetPageInfo.argtypes = [c_char_p, c_int]
    lib.GetPageInfo.restype = ByteResult

    # ExtractTextPositions
    lib.ExtractTextPositions.argtypes = [c_char_p, c_int, c_int]
    lib.ExtractTextPositions.restype = ByteResult

    # FindTextOccurrences
    lib.FindTextOccurrences.argtypes = [c_char_p, c_int, c_char_p]
    lib.FindTextOccurrences.restype = ByteResult

    # ApplyRedactions
    lib.ApplyRedactions.argtypes = [c_char_p, c_int, c_char_p]
    lib.ApplyRedactions.restype = ByteResult

    # ApplyRedactionsAdvanced
    lib.ApplyRedactionsAdvanced.argtypes = [c_char_p, c_int, c_char_p]
    lib.ApplyRedactionsAdvanced.restype = ByteResult

    # FreeBytesResult
    lib.FreeBytesResult.argtypes = [ByteResult]
    lib.FreeBytesResult.restype = None

    # FreeBytesArrayResult
    lib.FreeBytesArrayResult.argtypes = [ByteArrayResult]
    lib.FreeBytesArrayResult.restype = None

    return lib


# Lazy loading of the library
_lib = None


def get_lib():
    """Get the loaded library instance, loading it if necessary."""
    global _lib
    if _lib is None:
        _lib = _load_library()
    return _lib


def call_bytes_result(func, *args) -> bytes:
    """
    Call a function that returns ByteResult and handle memory management.

    Args:
        func: The C function to call
        *args: Arguments to pass to the function

    Returns:
        bytes: The result data

    Raises:
        GoPDFSuitError: If the function returns an error
    """
    lib = get_lib()
    result = func(*args)

    try:
        if result.error:
            error_msg = result.error.decode("utf-8")
            raise GoPDFSuitError(error_msg)

        if result.data is None or result.data == 0 or result.length <= 0:
            return b""

        # Cast void* to char* and read the exact number of bytes
        # This avoids null-termination issues with binary data
        data = ctypes.string_at(result.data, result.length)
        return data
    finally:
        lib.FreeBytesResult(result)


def call_bytes_array_result(func, *args) -> list:
    """
    Call a function that returns ByteArrayResult and handle memory management.

    Args:
        func: The C function to call
        *args: Arguments to pass to the function

    Returns:
        list[bytes]: List of result byte arrays

    Raises:
        GoPDFSuitError: If the function returns an error
    """
    lib = get_lib()
    result = func(*args)

    try:
        if result.error:
            error_msg = result.error.decode("utf-8")
            raise GoPDFSuitError(error_msg)

        if result.count <= 0:
            return []

        # Copy all data before freeing
        parts = []
        for i in range(result.count):
            length = result.lengths[i]
            ptr = result.data[i]
            if ptr and length > 0:
                data = ctypes.string_at(ptr, length)
                parts.append(data)

        return parts
    finally:
        lib.FreeBytesArrayResult(result)
