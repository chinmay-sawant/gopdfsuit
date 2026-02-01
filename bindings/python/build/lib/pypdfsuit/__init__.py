"""
pypdfsuit - Python bindings for gopdfsuit PDF library.

A comprehensive PDF library for generation, merging, splitting, form filling,
and HTML to PDF/Image conversion.

Example usage:

    from pypdfsuit import generate_pdf, PDFTemplate, Config, Title

    template = PDFTemplate(
        config=Config(page="A4", page_alignment=1),
        title=Title(props="Helvetica:18:100:center:0:0:0:0", text="My Document"),
        elements=[]
    )
    pdf_bytes = generate_pdf(template)

    with open("output.pdf", "wb") as f:
        f.write(pdf_bytes)
"""

__version__ = "1.0.0"

from .types import (
    PDFTemplate,
    Config,
    SecurityConfig,
    PDFAConfig,
    SignatureConfig,
    CustomFontConfig,
    Title,
    TitleTable,
    Table,
    Row,
    Cell,
    FormField,
    Image,
    Footer,
    Spacer,
    Element,
    Bookmark,
    FontInfo,
    HtmlToPDFRequest,
    HtmlToImageRequest,
    SplitSpec,
)

from .generator import generate_pdf, get_available_fonts
from .merge import merge_pdfs
from .split import split_pdf, parse_page_spec
from .fill import fill_pdf_with_xfdf
from .html import convert_html_to_pdf, convert_html_to_image

__all__ = [
    # Types
    "PDFTemplate",
    "Config",
    "SecurityConfig",
    "PDFAConfig",
    "SignatureConfig",
    "CustomFontConfig",
    "Title",
    "TitleTable",
    "Table",
    "Row",
    "Cell",
    "FormField",
    "Image",
    "Footer",
    "Spacer",
    "Element",
    "Bookmark",
    "FontInfo",
    "HtmlToPDFRequest",
    "HtmlToImageRequest",
    "SplitSpec",
    # Functions
    "generate_pdf",
    "get_available_fonts",
    "merge_pdfs",
    "split_pdf",
    "parse_page_spec",
    "fill_pdf_with_xfdf",
    "convert_html_to_pdf",
    "convert_html_to_image",
]
