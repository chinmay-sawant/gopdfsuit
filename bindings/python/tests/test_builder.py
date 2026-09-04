"""Tests for the pypdfsuit builder overlay (Phase 3)."""

import base64

import pytest

from pypdfsuit import (
    Config,
    Footer,
    Title,
)
from pypdfsuit.builder import (
    TableBuilder,
    TemplateBuilder,
    add_bracket_text,
    font_from_path,
    header_cell,
    image_from_path,
    make_props,
    math_cell,
    new_cell,
    set_bracket_font,
    set_cell_alignment,
    set_cell_bg_color,
    set_cell_borders,
    set_cell_color,
    set_cell_font,
    set_cell_text_color,
    set_row_color,
    set_table_colors,
)
from pypdfsuit.generator import generate_pdf_from_dict


def _lib_available() -> bool:
    try:
        from pypdfsuit._bindings import get_lib

        get_lib()
        return True
    except Exception:
        return False


requires_lib = pytest.mark.skipif(not _lib_available(), reason="libgopdfsuit.so missing")


class TestMakeProps:
    def test_regular(self):
        assert make_props("Helvetica", 12, False, False, False, "left", (1, 1, 1, 1)) == "Helvetica:12:000:left:1:1:1:1"

    def test_bold_center(self):
        assert make_props("Helvetica", 18, True, False, False, "center", (1, 1, 1, 1)) == "Helvetica:18:100:center:1:1:1:1"

    def test_italic_underline_bits(self):
        assert make_props("Helvetica", 12, False, True, True, "right", (0, 0, 0, 0)) == "Helvetica:12:011:right:0:0:0:0"

    def test_int_borders(self):
        assert make_props("Helvetica", 12, False, False, False, "left", 0) == "Helvetica:12:000:left:0:0:0:0"


class TestCells:
    def test_new_cell(self):
        c = new_cell("hi", make_props())
        assert c.text == "hi"
        assert c.props == "Helvetica:12:000:left:1:1:1:1"

    def test_header_cell_is_bold_center(self):
        c = header_cell("H")
        assert c.props == "Helvetica:12:100:center:1:1:1:1"

    def test_math_cell_flag(self):
        c = math_cell("x^2")
        assert c.math_enabled is True

    def test_set_cell_font_preserves_align_borders(self):
        c = new_cell("t", "Helvetica:12:000:right:1:0:1:0")
        set_cell_font(c, "Helvetica", 18, bold=True)
        assert c.props == "Helvetica:18:100:right:1:0:1:0"

    def test_set_cell_alignment(self):
        c = new_cell("t", make_props(align="left"))
        set_cell_alignment(c, "right")
        assert ":right:" in c.props

    def test_set_cell_borders(self):
        c = new_cell("t", make_props())
        set_cell_borders(c, (0, 0, 0, 0))
        assert c.props.endswith(":0:0:0:0")

    def test_right_cell_color(self):
        c = new_cell("right", make_props(align="right"))
        set_cell_text_color(c, "#B00020")
        assert c.text_color == "#B00020"
        assert c.to_dict()["textcolor"] == "#B00020"

    def test_set_cell_color_both(self):
        c = new_cell("t")
        set_cell_color(c, bg="#FF0000", text="#00FF00")
        assert (c.bg_color, c.text_color) == ("#FF0000", "#00FF00")

    def test_set_cell_bg_color(self):
        c = new_cell("t")
        set_cell_bg_color(c, "#123456")
        assert c.bg_color == "#123456"

    def test_set_row_color(self):
        cells = [new_cell("a"), new_cell("b")]
        set_row_color(cells, bg="#111111")
        assert all(c.bg_color == "#111111" for c in cells)

    def test_set_table_colors(self):
        from pypdfsuit import Table, Row

        t = Table(max_columns=2, rows=[Row(row=[new_cell("a"), new_cell("b")])])
        set_table_colors(t, text="#B00020")
        assert all(c.text_color == "#B00020" for c in t.rows[0].row)

    def test_bracket(self):
        c = new_cell("clause")
        add_bracket_text(c, "[", "]")
        assert c.text == "[clause]"
        set_bracket_font(c, "Helvetica", 12)
        assert c.props.startswith("Helvetica:12:")


class TestHelpers:
    def test_image_from_path(self, tmp_path):
        png = base64.b64decode(
            "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="
        )
        p = tmp_path / "tiny.png"
        p.write_bytes(png)
        img = image_from_path(p, width=100, height=80)
        assert img.image_name == "tiny.png"
        assert img.width == 100 and img.height == 80
        assert base64.b64decode(img.image_data) == png

    def test_font_from_path(self, tmp_path):
        p = tmp_path / "Fake.ttf"
        p.write_bytes(b"\x00\x01\x00\x00fake")
        cfg = font_from_path("Fake", p)
        assert cfg.name == "Fake"
        assert base64.b64decode(cfg.font_data) == b"\x00\x01\x00\x00fake"

    def test_title_textprops_footer_props_config_standard_fonts(self):
        t = Title(props="Helvetica:18:100:center:0:0:0:0", text="T", textprops="Helvetica:12:000:left:0:0:0:0")
        assert t.to_dict()["textprops"] == "Helvetica:12:000:left:0:0:0:0"
        f = Footer(font="Helvetica:10", text="ft", props="Helvetica:10:000:center:0:0:0:0")
        assert f.to_dict()["props"] == "Helvetica:10:000:center:0:0:0:0"
        c = Config(page="A4", page_alignment=1, embed_standard_fonts=True)
        assert c.to_dict()["embedStandardFonts"] is True


class TestTemplateBuilder:
    def test_build_dict_shape(self):
        b = TemplateBuilder("A4", True)
        b.add_title("Document Title")
        tb = b.add_table(3, 1, 2, 1)
        row = tb.add_row(
            new_cell("", "Helvetica:12:000:left:1:1:1:1"),
            new_cell("Document Title", "Helvetica:18:100:center:1:1:1:1"),
            new_cell("", "Helvetica:12:000:right:1:1:1:1"),
        )
        set_cell_text_color(row[2], "#B00020")
        set_cell_font(row[1], "Helvetica", 18, bold=True)
        c = new_cell("clause", make_props("Helvetica", 12, False, False, False, "left", (1, 1, 1, 1)))
        add_bracket_text(c, "[", "]")
        tb2_row_holder: list = []

        b.add_spacer(20)
        template = b.build()
        d = template.to_dict()
        assert d["config"]["page"] == "A4"
        assert d["title"]["text"] == "Document Title"
        assert d["elements"][0]["type"] == "table"
        table = d["elements"][0]["table"]
        assert table["maxcolumns"] == 3
        assert table["columnwidths"] == [1, 2, 1]
        assert table["rows"][0]["row"][2]["textcolor"] == "#B00020"
        assert d["elements"][1] == {"type": "spacer", "spacer": {"height": 20}}
        assert c.text == "[clause]"
        assert tb2_row_holder == []

    def test_table_builder_add_row_returns_cells(self):
        tb = TableBuilder(2, [1.0, 1.0])
        row = tb.add_row(new_cell("a"), new_cell("b"))
        assert [c.text for c in row] == ["a", "b"]
        assert tb.build().max_columns == 2

    @requires_lib
    def test_generate(self):
        b = TemplateBuilder("A4", True)
        b.add_title("Document Title")
        tb = b.add_table(3, 1, 2, 1)
        row = tb.add_row(
            new_cell("", "Helvetica:12:000:left:1:1:1:1"),
            new_cell("Document Title", "Helvetica:18:100:center:1:1:1:1"),
            new_cell("", "Helvetica:12:000:right:1:1:1:1"),
        )
        set_cell_text_color(row[2], "#B00020")
        out = b.generate()
        assert out.startswith(b"%PDF-")

    @requires_lib
    def test_generate_pdf_from_dict(self):
        b = TemplateBuilder()
        b.add_title("dict path")
        out = generate_pdf_from_dict(b.build().to_dict())
        assert out.startswith(b"%PDF-")
