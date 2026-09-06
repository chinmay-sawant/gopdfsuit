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

__version__ = "7.0.0"

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

from .generator import (
    generate_pdf,
    generate_pdf_from_dict,
    get_available_fonts,
    serialize_template,
)
from ._bindings import (
    GoPDFSuitError,
    InvalidInputError,
    LimitExceededError,
    UpstreamError,
    InternalError,
)
from .merge import merge_pdfs
from .split import split_pdf, parse_page_spec
from .fill import fill_pdf_with_xfdf
from .compress import compress_pdf
from .builder import (
    Font,
    Text,
    make_props,
    new_cell,
    header_cell,
    math_cell,
    set_cell_font,
    set_cell_alignment,
    set_cell_borders,
    set_cell_color,
    set_cell_text_color,
    set_cell_bg_color,
    set_row_color,
    set_table_colors,
    add_bracket_text,
    set_bracket_font,
    image_from_path,
    font_from_path,
    TableBuilder,
    TemplateBuilder,
)
from .html import convert_html_to_pdf, convert_html_to_image
from .redact import (
    get_page_info,
    extract_text_positions,
    apply_redactions,
    find_text_occurrences,
    apply_redactions_advanced,
)

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
    "generate_pdf_from_dict",
    "serialize_template",
    "get_available_fonts",
    "merge_pdfs",
    "split_pdf",
    "parse_page_spec",
    "fill_pdf_with_xfdf",
    "compress_pdf",
    "Font",
    "Text",
    "make_props",
    "new_cell",
    "header_cell",
    "math_cell",
    "set_cell_font",
    "set_cell_alignment",
    "set_cell_borders",
    "set_cell_color",
    "set_cell_text_color",
    "set_cell_bg_color",
    "set_row_color",
    "set_table_colors",
    "add_bracket_text",
    "set_bracket_font",
    "image_from_path",
    "font_from_path",
    "TableBuilder",
    "TemplateBuilder",
    "convert_html_to_pdf",
    "convert_html_to_image",
    "get_page_info",
    "extract_text_positions",
    "apply_redactions",
    "find_text_occurrences",
    "apply_redactions_advanced",
    # Errors
    "GoPDFSuitError",
    "InvalidInputError",
    "LimitExceededError",
    "UpstreamError",
    "InternalError",
]
