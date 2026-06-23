"""
Type definitions for pypdfsuit.

These types mirror the Go types in gopdfsuit/internal/models/models.go.
"""

from dataclasses import dataclass, field, asdict
from functools import lru_cache
from typing import Optional, List, Dict, Any, Tuple


_JSON_KEY_MAPPING = {
    "page_border": "pageBorder",
    "page_alignment": "pageAlignment",
    "pdf_title": "pdfTitle",
    "arlington_compatible": "arlingtonCompatible",
    "embed_fonts": "embedFonts",
    "custom_fonts": "customFonts",
    "pdfa_compliant": "pdfaCompliant",
    "user_password": "userPassword",
    "owner_password": "ownerPassword",
    "allow_printing": "allowPrinting",
    "allow_modifying": "allowModifying",
    "allow_copying": "allowCopying",
    "allow_annotations": "allowAnnotations",
    "allow_form_filling": "allowFormFilling",
    "allow_accessibility": "allowAccessibility",
    "allow_assembly": "allowAssembly",
    "allow_high_quality_print": "allowHighQualityPrint",
    "certificate_pem": "certificatePem",
    "private_key_pem": "privateKeyPem",
    "certificate_chain": "certificateChain",
    "contact_info": "contactInfo",
    "max_columns": "maxcolumns",
    "column_widths": "columnwidths",
    "row_heights": "rowheights",
    "bg_color": "bgcolor",
    "text_color": "textcolor",
    "image_name": "imagename",
    "image_data": "imagedata",
    "form_field": "form_field",
    "group_name": "group_name",
    "file_path": "filePath",
    "font_data": "fontData",
    "output_path": "output_path",
    "page_size": "page_size",
    "margin_top": "margin_top",
    "margin_right": "margin_right",
    "margin_bottom": "margin_bottom",
    "margin_left": "margin_left",
    "low_quality": "low_quality",
    "crop_width": "crop_width",
    "crop_height": "crop_height",
    "crop_x": "crop_x",
    "crop_y": "crop_y",
    "display_name": "displayName",
    "max_per_file": "MaxPerFile",
    "math_enabled": "mathEnabled",
}


@lru_cache(maxsize=None)
def _dataclass_json_fields(cls: type) -> Tuple[Tuple[str, str], ...]:
    return tuple(
        (field_name, _python_to_json_key(field_name))
        for field_name in cls.__dataclass_fields__
    )


def _to_dict(obj: Any, remove_none: bool = True) -> Any:
    """Convert a dataclass to a dictionary, optionally removing None values."""
    if obj is None:
        return None
    if isinstance(obj, (str, int, float, bool)):
        return obj
    if isinstance(obj, list):
        return [_to_dict(item, remove_none) for item in obj]
    if isinstance(obj, dict):
        return {k: _to_dict(v, remove_none) for k, v in obj.items()}
    if hasattr(obj, "__dataclass_fields__"):
        result = {}
        for field_name, json_key in _dataclass_json_fields(type(obj)):
            value = getattr(obj, field_name)
            if value is None and remove_none:
                continue
            result[json_key] = _to_dict(value, remove_none)
        return result
    return obj


def _python_to_json_key(key: str) -> str:
    """Convert Python snake_case to JSON camelCase where needed."""
    return _JSON_KEY_MAPPING.get(key, key)


@dataclass
class SecurityConfig:
    """PDF encryption and permission settings."""

    enabled: bool = False
    user_password: str = ""
    owner_password: str = ""
    allow_printing: bool = True
    allow_modifying: bool = False
    allow_copying: bool = True
    allow_annotations: bool = False
    allow_form_filling: bool = False
    allow_accessibility: bool = False
    allow_assembly: bool = False
    allow_high_quality_print: bool = True

    def to_dict(self) -> Dict[str, Any]:
        return _to_dict(self)


@dataclass
class PDFAConfig:
    """PDF/A compliance settings."""

    enabled: bool = False
    conformance: str = "4"  # PDF/A conformance level: "1b", "2b", "3b", "4", "4f", "4e"
    title: Optional[str] = None
    author: Optional[str] = None
    subject: Optional[str] = None
    creator: Optional[str] = None
    keywords: Optional[str] = None

    def to_dict(self) -> Dict[str, Any]:
        return _to_dict(self)


@dataclass
class SignatureConfig:
    """Digital signature settings."""

    enabled: bool = False
    certificate_pem: str = ""
    private_key_pem: str = ""
    certificate_chain: Optional[List[str]] = None
    visible: bool = False
    page: int = 1
    x: float = 0.0
    y: float = 0.0
    width: float = 200.0
    height: float = 50.0
    reason: Optional[str] = None
    location: Optional[str] = None
    contact_info: Optional[str] = None
    name: Optional[str] = None

    def to_dict(self) -> Dict[str, Any]:
        result = {
            "enabled": self.enabled,
            "certificatePem": self.certificate_pem,
            "privateKeyPem": self.private_key_pem,
            "visible": self.visible,
            "page": self.page,
            "x": self.x,
            "y": self.y,
            "width": self.width,
            "height": self.height,
        }
        if self.certificate_chain is not None:
            result["certificateChain"] = self.certificate_chain
        if self.reason is not None:
            result["reason"] = self.reason
        if self.location is not None:
            result["location"] = self.location
        if self.contact_info is not None:
            result["contactInfo"] = self.contact_info
        if self.name is not None:
            result["name"] = self.name
        return result


@dataclass
class CustomFontConfig:
    """Custom font configuration for embedding TTF/OTF fonts."""

    name: str  # Reference name used in props (e.g., "MyFont")
    file_path: Optional[str] = None  # Path to TTF/OTF file
    font_data: Optional[str] = None  # Base64-encoded font data

    def to_dict(self) -> Dict[str, Any]:
        result = {"name": self.name}
        if self.file_path is not None:
            result["filePath"] = self.file_path
        if self.font_data is not None:
            result["fontData"] = self.font_data
        return result


@dataclass
class Bookmark:
    """PDF outline entry for document navigation."""

    title: str
    dest: Optional[str] = None  # Named destination
    page: int = 0  # Target page number (1-based)
    y: float = 0.0  # Y position on target page
    children: Optional[List["Bookmark"]] = None
    open: bool = False  # Whether children are expanded by default

    def to_dict(self) -> Dict[str, Any]:
        result = {
            "title": self.title,
            "page": self.page,
            "y": self.y,
            "open": self.open,
        }
        if self.dest is not None:
            result["dest"] = self.dest
        if self.children is not None:
            result["children"] = [child.to_dict() for child in self.children]
        return result


@dataclass
class Config:
    """Page configuration and optional features."""

    page: str = "A4"  # Page size: "A4", "Letter", "Legal", etc.
    page_alignment: int = 1  # 1 = Portrait, 2 = Landscape
    page_border: str = ""
    watermark: Optional[str] = None
    pdf_title: Optional[str] = None
    arlington_compatible: bool = False
    bookmarks: Optional[List[Bookmark]] = None
    security: Optional[SecurityConfig] = None
    pdfa: Optional[PDFAConfig] = None
    signature: Optional[SignatureConfig] = None
    embed_fonts: Optional[bool] = True
    custom_fonts: Optional[List[CustomFontConfig]] = None
    pdfa_compliant: bool = False

    def to_dict(self) -> Dict[str, Any]:
        result = {
            "page": self.page,
            "pageAlignment": self.page_alignment,
            "pageBorder": self.page_border,
            "pdfaCompliant": self.pdfa_compliant,
        }
        if self.embed_fonts is not None:
            result["embedFonts"] = self.embed_fonts
        if self.watermark is not None:
            result["watermark"] = self.watermark
        if self.pdf_title is not None:
            result["pdfTitle"] = self.pdf_title
        result["arlingtonCompatible"] = self.arlington_compatible
        if self.bookmarks is not None:
            result["bookmarks"] = [bookmark.to_dict() for bookmark in self.bookmarks]
        if self.security is not None:
            result["security"] = self.security.to_dict()
        if self.pdfa is not None:
            result["pdfa"] = self.pdfa.to_dict()
        if self.signature is not None:
            result["signature"] = self.signature.to_dict()
        if self.custom_fonts is not None:
            result["customFonts"] = [font.to_dict() for font in self.custom_fonts]
        return result


@dataclass
class Image:
    """Image element in the PDF."""

    image_name: str = ""
    image_data: str = ""  # Base64 encoded image data
    width: float = 0.0
    height: float = 0.0
    link: Optional[str] = None

    def to_dict(self) -> Dict[str, Any]:
        result = {
            "imagename": self.image_name,
            "imagedata": self.image_data,
            "width": self.width,
            "height": self.height,
        }
        if self.link is not None:
            result["link"] = self.link
        return result


@dataclass
class FormField:
    """Fillable form field."""

    type: str  # "checkbox", "radio", "text"
    name: str
    value: str = ""  # Export value for radio/checkbox, default value for text
    checked: bool = False
    group_name: Optional[str] = None  # For radio buttons
    shape: Optional[str] = None  # "round" or "square" (for radio)

    def to_dict(self) -> Dict[str, Any]:
        result = {
            "type": self.type,
            "name": self.name,
            "value": self.value,
            "checked": self.checked,
        }
        if self.group_name is not None:
            result["group_name"] = self.group_name
        if self.shape is not None:
            result["shape"] = self.shape
        return result


@dataclass
class Cell:
    """Cell in a table row."""

    props: str  # Font and style properties
    text: str = ""
    checkbox: Optional[bool] = None
    image: Optional[Image] = None
    width: Optional[float] = None
    height: Optional[float] = None
    form_field: Optional[FormField] = None
    bg_color: Optional[str] = None
    text_color: Optional[str] = None
    link: Optional[str] = None
    wrap: Optional[bool] = None
    dest: Optional[str] = None
    math_enabled: Optional[bool] = None

    def to_dict(self) -> Dict[str, Any]:
        result = {
            "props": self.props,
            "text": self.text,
        }
        if self.checkbox is not None:
            result["chequebox"] = self.checkbox
        if self.image is not None:
            result["image"] = self.image.to_dict()
        if self.width is not None:
            result["width"] = self.width
        if self.height is not None:
            result["height"] = self.height
        if self.form_field is not None:
            result["form_field"] = self.form_field.to_dict()
        if self.bg_color is not None:
            result["bgcolor"] = self.bg_color
        if self.text_color is not None:
            result["textcolor"] = self.text_color
        if self.link is not None:
            result["link"] = self.link
        if self.wrap is not None:
            result["wrap"] = self.wrap
        if self.dest is not None:
            result["dest"] = self.dest
        if self.math_enabled is not None:
            result["mathEnabled"] = self.math_enabled
        return result


@dataclass
class Row:
    """Row in a table."""

    row: List[Cell] = field(default_factory=list)

    def to_dict(self) -> Dict[str, Any]:
        return {"row": [cell.to_dict() for cell in self.row]}


@dataclass
class Table:
    """Table element in the PDF."""

    max_columns: int
    rows: List[Row] = field(default_factory=list)
    column_widths: Optional[List[float]] = None
    row_heights: Optional[List[float]] = None
    bg_color: Optional[str] = None
    text_color: Optional[str] = None

    def to_dict(self) -> Dict[str, Any]:
        result = {
            "maxcolumns": self.max_columns,
            "rows": [row.to_dict() for row in self.rows],
        }
        if self.column_widths is not None:
            result["columnwidths"] = self.column_widths
        if self.row_heights is not None:
            result["rowheights"] = self.row_heights
        if self.bg_color is not None:
            result["bgcolor"] = self.bg_color
        if self.text_color is not None:
            result["textcolor"] = self.text_color
        return result


@dataclass
class Spacer:
    """Vertical space between elements."""

    height: float

    def to_dict(self) -> Dict[str, Any]:
        return {"height": self.height}


@dataclass
class Element:
    """Ordered element in the PDF (table, spacer, or image)."""

    type: str  # "table", "spacer", "image"
    index: Optional[int] = None
    table: Optional[Table] = None
    spacer: Optional[Spacer] = None
    image: Optional[Image] = None

    def to_dict(self) -> Dict[str, Any]:
        result = {"type": self.type}
        if self.index is not None:
            result["index"] = self.index
        if self.table is not None:
            result["table"] = self.table.to_dict()
        if self.spacer is not None:
            result["spacer"] = self.spacer.to_dict()
        if self.image is not None:
            result["image"] = self.image.to_dict()
        return result


@dataclass
class TitleTable:
    """Embedded table within the title section."""

    max_columns: int
    rows: List[Row] = field(default_factory=list)
    column_widths: Optional[List[float]] = None

    def to_dict(self) -> Dict[str, Any]:
        result = {
            "maxcolumns": self.max_columns,
            "rows": [row.to_dict() for row in self.rows],
        }
        if self.column_widths is not None:
            result["columnwidths"] = self.column_widths
        return result


@dataclass
class Title:
    """Document title section."""

    props: str
    text: str = ""
    table: Optional[TitleTable] = None
    bg_color: Optional[str] = None
    text_color: Optional[str] = None
    link: Optional[str] = None

    def to_dict(self) -> Dict[str, Any]:
        result = {"props": self.props, "text": self.text}
        if self.table is not None:
            result["table"] = self.table.to_dict()
        if self.bg_color is not None:
            result["bgcolor"] = self.bg_color
        if self.text_color is not None:
            result["textcolor"] = self.text_color
        if self.link is not None:
            result["link"] = self.link
        return result


@dataclass
class Footer:
    """Document footer."""

    font: str
    text: str
    link: Optional[str] = None

    def to_dict(self) -> Dict[str, Any]:
        result = {"font": self.font, "text": self.text}
        if self.link is not None:
            result["link"] = self.link
        return result


@dataclass
class PDFTemplate:
    """Main input structure for PDF generation."""

    config: Config
    title: Title
    table: Optional[List[Table]] = None
    spacer: Optional[List[Spacer]] = None
    image: Optional[List[Image]] = None
    elements: Optional[List[Element]] = None
    footer: Optional[Footer] = None
    bookmarks: Optional[List[Bookmark]] = None

    def to_dict(self) -> Dict[str, Any]:
        result = {
            "config": self.config.to_dict(),
            "title": self.title.to_dict(),
        }
        if self.table is not None:
            result["table"] = [t.to_dict() for t in self.table]
        if self.spacer is not None:
            result["spacer"] = [s.to_dict() for s in self.spacer]
        if self.image is not None:
            result["image"] = [i.to_dict() for i in self.image]
        if self.elements is not None:
            result["elements"] = [e.to_dict() for e in self.elements]
        if self.footer is not None:
            result["footer"] = self.footer.to_dict()
        if self.bookmarks is not None:
            result["bookmarks"] = [b.to_dict() for b in self.bookmarks]
        return result


@dataclass
class FontInfo:
    """Font information."""

    id: str
    name: str
    display_name: str
    reference: str

    def to_dict(self) -> Dict[str, Any]:
        return _to_dict(self)


@dataclass
class HtmlToPDFRequest:
    """Input for HTML to PDF conversion."""

    html: Optional[str] = None
    url: Optional[str] = None
    output_path: Optional[str] = None
    page_size: str = "A4"
    orientation: str = "Portrait"
    margin_top: str = "10mm"
    margin_right: str = "10mm"
    margin_bottom: str = "10mm"
    margin_left: str = "10mm"
    dpi: Optional[int] = None
    grayscale: bool = False
    low_quality: bool = False
    options: Optional[Dict[str, str]] = None

    def to_dict(self) -> Dict[str, Any]:
        return _to_dict(self)


@dataclass
class HtmlToImageRequest:
    """Input for HTML to image conversion."""

    html: Optional[str] = None
    url: Optional[str] = None
    output_path: Optional[str] = None
    format: str = "png"
    width: Optional[int] = None
    height: Optional[int] = None
    quality: Optional[int] = None
    zoom: Optional[float] = None
    crop_width: Optional[int] = None
    crop_height: Optional[int] = None
    crop_x: Optional[int] = None
    crop_y: Optional[int] = None
    options: Optional[Dict[str, str]] = None

    def to_dict(self) -> Dict[str, Any]:
        return _to_dict(self)


@dataclass
class SplitSpec:
    """Split criteria for splitting PDFs."""

    pages: Optional[List[int]] = None  # Explicit pages (1-based)
    ranges: Optional[List[Tuple[int, int]]] = None  # Page ranges
    max_per_file: Optional[int] = None  # Maximum pages per output file

    def to_dict(self) -> Dict[str, Any]:
        result = {}
        if self.pages is not None:
            result["Pages"] = self.pages
        if self.ranges is not None:
            result["Ranges"] = [[r[0], r[1]] for r in self.ranges]
        if self.max_per_file is not None:
            result["MaxPerFile"] = self.max_per_file
        return result
