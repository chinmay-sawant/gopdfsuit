"""Parity tests for the fluent Font/Text builders (Phase 3)."""

from pypdfsuit import Font, Text, make_props, new_cell
from pypdfsuit.builder import Font as BuilderFont
from pypdfsuit.builder import Text as BuilderText


def test_exports():
    assert Font is BuilderFont
    assert Text is BuilderText


def test_font_defaults_match_make_props():
    assert Font("Helvetica").props() == make_props()
    assert Font().props() == make_props()


def test_font_size_style_align():
    assert Font("Helvetica").size(18).bold().center().props() == make_props(
        "Helvetica", 18, True, False, False, "center", (1, 1, 1, 1)
    )
    assert Font("Helvetica").italic().underline().right().props() == make_props(
        "Helvetica", 12, False, True, True, "right", (1, 1, 1, 1)
    )
    assert Font("Helvetica").size(10).bold().italic().underline().left().props() == make_props(
        "Helvetica", 10, True, True, True, "left", (1, 1, 1, 1)
    )


def test_font_borders():
    assert Font("Helvetica").borders(1, 0, 1, 0).props() == make_props(
        "Helvetica", 12, False, False, False, "left", (1, 0, 1, 0)
    )
    assert Font("Helvetica").borderless().props() == make_props(
        "Helvetica", 12, False, False, False, "left", (0, 0, 0, 0)
    )
    assert Font("Helvetica").borderless().bordered().props() == make_props(
        "Helvetica", 12, False, False, False, "left", (1, 1, 1, 1)
    )


def test_font_cell_terminal():
    c = Font("Helvetica").size(18).bold().center().cell("hi")
    assert c.text == "hi"
    assert c.props == make_props("Helvetica", 18, True, False, False, "center", (1, 1, 1, 1))


def test_text_build_defaults():
    c = Text("hi").build()
    assert c.text == "hi"
    assert c.props == make_props()
    assert c.math_enabled is None


def test_text_with_font_colors_math():
    f = Font("Helvetica").size(12).bold()
    c = Text("x^2").font(f).bg("#FF0000").fg("#00FF00").math().build()
    assert c.props == f.props()
    assert c.bg_color == "#FF0000"
    assert c.text_color == "#00FF00"
    assert c.math_enabled is True
    assert c.to_dict()["bgcolor"] == "#FF0000"
    assert c.to_dict()["textcolor"] == "#00FF00"
    assert c.to_dict()["mathEnabled"] is True


def test_text_explicit_props_win_over_font():
    f = Font("Helvetica").size(18).bold()
    explicit = make_props("Courier", 10)
    c = Text("t").font(f).props(explicit).build()
    assert c.props == explicit


def test_new_cell_font_kwarg():
    f = Font("Helvetica").size(18).bold().center()
    c = new_cell("hi", font=f)
    assert c.props == f.props() == make_props("Helvetica", 18, True, False, False, "center", (1, 1, 1, 1))


def test_new_cell_props_wins_over_font():
    f = Font("Helvetica").size(18).bold()
    explicit = make_props("Courier", 10)
    c = new_cell("hi", explicit, font=f)
    assert c.props == explicit
    c2 = new_cell("hi", props=explicit, font=f)
    assert c2.props == explicit


def test_new_cell_default_unchanged():
    c = new_cell("hi")
    assert c.props == "Helvetica:12:000:left:1:1:1:1"
