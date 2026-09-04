"""PDF compression via the Go shared library."""

from ._bindings import get_lib, call_bytes_result, json_payload, pdf_args

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
    if level is None:
        level = ""
    level = str(level).lower()
    if level != "" and level not in _VALID_LEVELS:
        raise ValueError(f"level must be one of {list(_VALID_LEVELS)}, got {level!r}")

    lib = get_lib()
    return call_bytes_result(lib.CompressPDF, *pdf_args(src, "src PDF bytes"), json_payload({"level": level}))
