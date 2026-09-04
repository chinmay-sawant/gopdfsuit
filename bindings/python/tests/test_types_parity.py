"""Parity: to_dict() keys must match the Go JSON tags in internal/models/models.go.

Catches wire drift between the Python mirror (pypdfsuit/types.py) and the
Go structs handlers decode. Go remains the source of truth; KNOWN_GAPS
documents mirror fields that do not exist yet.
"""

import re
from pathlib import Path

from pypdfsuit.types import (
    Bookmark,
    Cell,
    Config,
    CustomFontConfig,
    FormField,
    HtmlToImageRequest,
    HtmlToPDFRequest,
    Image,
    PDFAConfig,
    SecurityConfig,
    SignatureConfig,
)

MODELS_GO = Path(__file__).parents[3] / "internal" / "models" / "models.go"

# Mirror fields intentionally missing from the Python dataclasses.
KNOWN_GAPS = {
    "Config": {"pageMargin", "taggedPDF"},
}


def go_json_tags(struct_name: str) -> set:
    src = MODELS_GO.read_text()
    m = re.search(r"type %s struct \{(.*?)\n\}" % re.escape(struct_name), src, re.S)
    assert m, "struct %s not found in models.go" % struct_name
    tags = set()
    for tag in re.findall(r'json:"([^"]+)"', m.group(1)):
        key = tag.split(",")[0]
        if key and key != "-":
            tags.add(key)
    return tags


def check(struct_name: str, obj) -> None:
    got = set(obj.to_dict().keys())
    want = go_json_tags(struct_name)
    gaps = KNOWN_GAPS.get(struct_name, set())
    assert gaps <= want, "documented gap %s no longer in Go %s" % (gaps - want, struct_name)
    assert got == want - gaps, "%s keys %s != Go tags %s" % (
        struct_name, sorted(got), sorted(want - gaps))


def test_config_parity() -> None:
    check("Config", Config(
        page_border="1:1:1:1",
        watermark="w",
        pdf_title="t",
        arlington_compatible=True,
        bookmarks=[Bookmark(title="b")],
        security=SecurityConfig(owner_password="o"),
        pdfa=PDFAConfig(),
        signature=SignatureConfig(),
        embed_fonts=True,
        embed_standard_fonts=False,
        custom_fonts=[CustomFontConfig(name="f")],
        pdfa_compliant=True,
    ))


def test_cell_parity() -> None:
    # checkbox attribute must serialize under the historical Go wire key.
    cell = Cell(props="p", checkbox=True)
    assert cell.to_dict()["chequebox"] is True
    assert "checkbox" not in cell.to_dict()
    check("Cell", Cell(
        props="p",
        text="t",
        checkbox=True,
        image=Image(image_name="i", image_data="d"),
        width=1.0,
        height=2.0,
        form_field=FormField(type="checkbox", name="n"),
        bg_color="#ffffff",
        text_color="#000000",
        link="https://example.com",
        wrap=True,
        dest="d",
        math_enabled=True,
    ))


def test_signature_config_parity() -> None:
    check("SignatureConfig", SignatureConfig(
        certificate_pem="c",
        private_key_pem="k",
        certificate_chain=["i"],
        visible=True,
        page=2,
        x=1.0,
        y=2.0,
        width=200.0,
        height=50.0,
        reason="r",
        location="l",
        contact_info="c",
        name="n",
    ))


def test_html_requests_parity() -> None:
    check("HTMLToPDFRequest", HtmlToPDFRequest(
        html="<p>h</p>",
        url="https://example.com",
        output_path="/tmp/x.pdf",
        page_size="A4",
        orientation="Portrait",
        margin_top="10mm",
        margin_right="10mm",
        margin_bottom="10mm",
        margin_left="10mm",
        dpi=300,
        grayscale=True,
        low_quality=True,
        options={"a": "b"},
    ))
    check("HTMLToImageRequest", HtmlToImageRequest(
        html="<p>h</p>",
        url="https://example.com",
        output_path="/tmp/x.png",
        format="png",
        width=800,
        height=600,
        quality=94,
        zoom=1.0,
        crop_width=10,
        crop_height=10,
        crop_x=1,
        crop_y=2,
        options={"a": "b"},
    ))
