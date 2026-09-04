"""PDF compression via the Go shared library."""

import json

from ._bindings import get_lib, call_bytes_result

_VALID_LEVELS = ("light", "medium", "heavy")


def compress_pdf(src: bytes, level: str = "medium") -> bytes:
    """Compress a PDF at the given tier.

    Args:
        src: Source PDF bytes.
        level: One of "light", "medium", "heavy". Empty string selects
            the engine default (Medium).

    Returns:
        bytes: The compressed PDF content.
    """
    if not src:
        raise ValueError("src PDF bytes cannot be empty")
    if level is None:
        level = ""
    level = str(level).lower()
    if level != "" and level not in _VALID_LEVELS:
        raise ValueError(f"level must be one of {list(_VALID_LEVELS)}, got {level!r}")

    lib = get_lib()
    opts = json.dumps({"level": level}, separators=(",", ":")).encode("utf-8")
    return call_bytes_result(lib.CompressPDF, src, len(src), opts)
