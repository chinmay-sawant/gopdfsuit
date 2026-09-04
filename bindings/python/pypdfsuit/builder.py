"""Builder overlay for programmatic PDF templates.

Pure-Python overlay that emits the current Props grammar
(Font:Size:Style:Align:L:R:T:B) and bgcolor/textcolor hex strings.
The sink stays generate_pdf(PDFTemplate); no engine changes.

Mirrors the Go builder names in snake_case.
"""

import base64
from pathlib import Path
from typing import List, Optional, Sequence, Tuple, Union

from .types import (
    Cell,
    Config,
    CustomFontConfig,
    Element,
    Image,
    PDFTemplate,
    Row,
    Spacer,
    Table,
    Title,
)

Borders = Tuple[int, int, int, int]

_DEFAULT_TITLE_PROPS = "Helvetica:18:100:center:0:0:0:0"


def _style_code(bold: bool = False, italic: bool = False, underline: bool = False) -> str:
    return f"{1 if bold else 0}{1 if italic else 0}{1 if underline else 0}"


def _normalize_borders(borders: Union[int, Sequence[int], None]) -> Borders:
    if borders is None:
        return (1, 1, 1, 1)
    if isinstance(borders, int):
        return (borders, borders, borders, borders)
    vals = list(borders) + [0, 0, 0, 0]
    return (int(vals[0]), int(vals[1]), int(vals[2]), int(vals[3]))


def make_props(
    font: str = "Helvetica",
    size: int = 12,
    bold: bool = False,
    italic: bool = False,
    underline: bool = False,
    align: str = "left",
    borders: Union[int, Sequence[int], None] = (1, 1, 1, 1),
) -> str:
    """Build a Props string: Font:Size:Style:Align:L:R:T:B."""
    b = _normalize_borders(borders)
    return f"{font}:{int(size)}:{_style_code(bold, italic, underline)}:{align}:{b[0]}:{b[1]}:{b[2]}:{b[3]}"


def _split_props(props: str) -> List[str]:
    parts = props.split(":")
    while len(parts) < 8:
        parts.append("")
    return parts


class Font:
    """Fluent builder emitting Props strings via make_props."""

    def __init__(self, name: str = "Helvetica") -> None:
        self._name = name
        self._size = 12
        self._bold = False
        self._italic = False
        self._underline = False
        self._align = "left"
        self._borders: Borders = (1, 1, 1, 1)

    def size(self, n: int) -> "Font":
        """Set the font size."""
        self._size = int(n)
        return self

    def bold(self, value: bool = True) -> "Font":
        """Enable or disable bold."""
        self._bold = bool(value)
        return self

    def italic(self, value: bool = True) -> "Font":
        """Enable or disable italic."""
        self._italic = bool(value)
        return self

    def underline(self, value: bool = True) -> "Font":
        """Enable or disable underline."""
        self._underline = bool(value)
        return self

    def left(self) -> "Font":
        """Left-align text."""
        self._align = "left"
        return self

    def center(self) -> "Font":
        """Center-align text."""
        self._align = "center"
        return self

    def right(self) -> "Font":
        """Right-align text."""
        self._align = "right"
        return self

    def borders(self, l: int, r: int, t: int, b: int) -> "Font":
        """Set L:R:T:B border flags."""
        self._borders = (int(l), int(r), int(t), int(b))
        return self

    def bordered(self) -> "Font":
        """Enable all borders."""
        self._borders = (1, 1, 1, 1)
        return self

    def borderless(self) -> "Font":
        """Disable all borders."""
        self._borders = (0, 0, 0, 0)
        return self

    def props(self) -> str:
        """Emit the Props string for this builder."""
        return make_props(
            self._name,
            self._size,
            self._bold,
            self._italic,
            self._underline,
            self._align,
            self._borders,
        )

    def cell(self, text: str = "", **kwargs) -> Cell:
        """Build a Cell with this font's props."""
        return Cell(props=self.props(), text=text, **kwargs)


class Text:
    """Fluent cell builder with optional font, colors, and math flag."""

    def __init__(self, text: str = "") -> None:
        self._text = text
        self._font: Optional[Font] = None
        self._props: Optional[str] = None
        self._bg: Optional[str] = None
        self._fg: Optional[str] = None
        self._math = False

    def font(self, font: Font) -> "Text":
        """Attach a Font builder (explicit props() wins at build)."""
        self._font = font
        return self

    def props(self, props: str) -> "Text":
        """Set an explicit Props string (wins over font at build)."""
        self._props = props
        return self

    def bg(self, color: str) -> "Text":
        """Set the background hex color."""
        self._bg = color
        return self

    def fg(self, color: str) -> "Text":
        """Set the text hex color."""
        self._fg = color
        return self

    def math(self, enabled: bool = True) -> "Text":
        """Enable or disable math rendering."""
        self._math = bool(enabled)
        return self

    def build(self, **kwargs) -> Cell:
        """Build the Cell. Explicit props win over font; kwargs win over stored fields."""
        if "props" in kwargs:
            props = kwargs.pop("props")
        elif self._props is not None:
            props = self._props
        elif self._font is not None:
            props = self._font.props()
        else:
            props = make_props()
        text = kwargs.pop("text", self._text)
        if "bg_color" not in kwargs and self._bg is not None:
            kwargs["bg_color"] = self._bg
        if "text_color" not in kwargs and self._fg is not None:
            kwargs["text_color"] = self._fg
        if "math_enabled" not in kwargs and self._math:
            kwargs["math_enabled"] = True
        return Cell(props=props, text=text, **kwargs)


def new_cell(
    text: str = "",
    props: Optional[str] = None,
    font: Optional[Font] = None,
    **kwargs,
) -> Cell:
    """Create a cell with text and a Props string.

    Precedence: an explicit props string always wins when both props
    and font are given; otherwise font.props() is used when font is
    given; otherwise the default "Helvetica:12:000:left:1:1:1:1" applies.
    """
    if props is None:
        if font is not None:
            if not hasattr(font, "props"):
                raise TypeError("font must be a Font builder")
            props = font.props()
        else:
            props = "Helvetica:12:000:left:1:1:1:1"
    return Cell(props=props, text=text, **kwargs)


def header_cell(
    text: str = "",
    font: str = "Helvetica",
    size: int = 12,
    align: str = "center",
    borders: Union[int, Sequence[int], None] = (1, 1, 1, 1),
    **kwargs,
) -> Cell:
    """Bold centered header cell."""
    return Cell(props=make_props(font, size, True, False, False, align, borders), text=text, **kwargs)


def math_cell(
    text: str = "",
    props: str = "Helvetica:12:000:left:1:1:1:1",
    **kwargs,
) -> Cell:
    """Cell with math rendering enabled."""
    kwargs.setdefault("math_enabled", True)
    return Cell(props=props, text=text, **kwargs)


def set_cell_font(
    cell: Cell,
    font: str,
    size: int,
    bold: bool = False,
    italic: bool = False,
    underline: bool = False,
) -> Cell:
    """Override a cell's font fields, preserving align and borders."""
    parts = _split_props(cell.props)
    align = parts[3] or "left"
    try:
        borders: Borders = (int(parts[4] or 0), int(parts[5] or 0), int(parts[6] or 0), int(parts[7] or 0))
    except ValueError:
        borders = (1, 1, 1, 1)
    cell.props = make_props(font, size, bold, italic, underline, align, borders)
    return cell


def set_cell_alignment(cell: Cell, align: str) -> Cell:
    """Override a cell's alignment, preserving other Props fields."""
    parts = _split_props(cell.props)
    parts[3] = align
    cell.props = ":".join(parts[:8])
    return cell


def set_cell_borders(cell: Cell, borders: Union[int, Sequence[int]]) -> Cell:
    """Override a cell's L:R:T:B border flags."""
    b = _normalize_borders(borders)
    parts = _split_props(cell.props)
    parts[4], parts[5], parts[6], parts[7] = str(b[0]), str(b[1]), str(b[2]), str(b[3])
    cell.props = ":".join(parts[:8])
    return cell


def set_cell_bg_color(cell: Cell, color: str) -> Cell:
    """Set a cell's background hex color."""
    cell.bg_color = color
    return cell


def set_cell_text_color(cell: Cell, color: str) -> Cell:
    """Set a cell's text hex color (e.g. different-color-on-right)."""
    cell.text_color = color
    return cell


def set_cell_color(cell: Cell, bg: Optional[str] = None, text: Optional[str] = None) -> Cell:
    """Set a cell's background and/or text color."""
    if bg is not None:
        cell.bg_color = bg
    if text is not None:
        cell.text_color = text
    return cell


def _row_cells(row: Union[Row, List[Cell]]) -> List[Cell]:
    return row.row if isinstance(row, Row) else row


def set_row_color(row: Union[Row, List[Cell]], bg: Optional[str] = None, text: Optional[str] = None) -> Union[Row, List[Cell]]:
    """Apply background and/or text color to every cell in a row."""
    for cell in _row_cells(row):
        set_cell_color(cell, bg=bg, text=text)
    return row


def set_table_colors(table: Table, bg: Optional[str] = None, text: Optional[str] = None) -> Table:
    """Apply background and/or text color to every cell in a table."""
    for row in table.rows:
        set_row_color(row, bg=bg, text=text)
    return table


def add_bracket_text(cell: Cell, open: str = "[", close: str = "]") -> Cell:
    """Wrap cell text in brackets (v1, no rich-text segments)."""
    cell.text = f"{open}{cell.text}{close}"
    return cell


def set_bracket_font(cell: Cell, font: str, size: int) -> Cell:
    """Override the font of a bracketed cell, preserving style flags and layout."""
    parts = _split_props(cell.props)
    try:
        cur_size = int(parts[1]) if parts[1] else size
    except ValueError:
        cur_size = size
    _ = cur_size
    align = parts[3] or "left"
    try:
        borders: Borders = (int(parts[4] or 0), int(parts[5] or 0), int(parts[6] or 0), int(parts[7] or 0))
    except ValueError:
        borders = (1, 1, 1, 1)
    style = parts[2] if len(parts[2]) == 3 else "000"
    bold, italic, underline = style[0] == "1", style[1] == "1", style[2] == "1"
    cell.props = make_props(font, size, bold, italic, underline, align, borders)
    return cell


def image_from_path(path: Union[str, Path], width: float = 0, height: float = 0, link: Optional[str] = None) -> Image:
    """Read an image file and return an Image with base64 data."""
    p = Path(path)
    data = p.read_bytes()
    return Image(
        image_name=p.name,
        image_data=base64.b64encode(data).decode("ascii"),
        width=width,
        height=height,
        link=link,
    )


def font_from_path(name: str, path: Union[str, Path]) -> CustomFontConfig:
    """Read a TTF/OTF file and return a CustomFontConfig with base64 data."""
    data = Path(path).read_bytes()
    return CustomFontConfig(name=name, font_data=base64.b64encode(data).decode("ascii"))


class TableBuilder:
    """Accumulates rows for a single table element."""

    def __init__(self, max_columns: int, column_widths: Optional[Sequence[float]] = None):
        self.max_columns = max_columns
        self.column_widths: Optional[List[float]] = list(column_widths) if column_widths is not None else None
        self._rows: List[Row] = []

    def add_row(self, *cells: Cell) -> List[Cell]:
        """Append a row and return its cell list for post-mutation (e.g. row[2] color)."""
        row_cells = list(cells)
        self._rows.append(Row(row=row_cells))
        return row_cells

    def build(self) -> Table:
        """Build the Table value."""
        return Table(max_columns=self.max_columns, rows=list(self._rows), column_widths=self.column_widths)


class TemplateBuilder:
    """Fluent builder emitting a PDFTemplate; generate() calls generate_pdf."""

    def __init__(self, page: str = "A4", portrait: bool = True):
        self.config = Config(page=page, page_alignment=1 if portrait else 2)
        self.title = Title(props=_DEFAULT_TITLE_PROPS, text="")
        self._elements: List[Element] = []
        self._tables: List[Tuple[Element, TableBuilder]] = []

    def add_title(
        self,
        text: str,
        font: str = "Helvetica",
        size: int = 18,
        bold: bool = True,
        italic: bool = False,
        underline: bool = False,
        align: str = "center",
        textprops: Optional[str] = None,
    ) -> "TemplateBuilder":
        """Set the document title text and props."""
        self.title.text = text
        self.title.props = make_props(font, size, bold, italic, underline, align, (0, 0, 0, 0))
        if textprops is not None:
            self.title.textprops = textprops
        return self

    def add_table(self, max_columns: int, *column_widths: float) -> TableBuilder:
        """Start a table element; rows are added via the returned TableBuilder."""
        widths = list(column_widths) if column_widths else None
        tb = TableBuilder(max_columns, widths)
        placeholder = Element(type="table", table=None)
        self._tables.append((placeholder, tb))
        self._elements.append(placeholder)
        return tb

    def add_spacer(self, height: float) -> "TemplateBuilder":
        """Append a vertical spacer element."""
        self._elements.append(Element(type="spacer", spacer=Spacer(height=height)))
        return self

    def build(self) -> PDFTemplate:
        """Materialize pending table rows into elements and return the template."""
        for placeholder, tb in self._tables:
            placeholder.table = tb.build()
        return PDFTemplate(config=self.config, title=self.title, elements=list(self._elements))

    def generate(self) -> bytes:
        """Build the template and render it to PDF bytes."""
        from .generator import generate_pdf

        return generate_pdf(self.build())
