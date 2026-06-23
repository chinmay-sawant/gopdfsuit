import json

from pypdfsuit import (
    Bookmark,
    Cell,
    Config,
    Element,
    PDFTemplate,
    Row,
    SignatureConfig,
    Table,
    Title,
    serialize_template,
)


def test_specialized_serializer_preserves_go_json_keys():
    template = PDFTemplate(
        config=Config(
            page="A4",
            page_alignment=1,
            page_border="0:0:0:0",
            pdf_title="Schema Test",
            arlington_compatible=True,
            pdfa_compliant=True,
            embed_fonts=True,
            signature=SignatureConfig(
                enabled=True,
                certificate_pem="cert",
                private_key_pem="key",
                certificate_chain=["chain"],
                contact_info="contact@example.com",
            ),
        ),
        title=Title(props="Helvetica:18:100:center:0:0:0:0", text="Schema Test"),
        elements=[
            Element(
                type="table",
                table=Table(
                    max_columns=1,
                    column_widths=[1],
                    rows=[
                        Row(
                            row=[
                                Cell(
                                    props="Helvetica:9:000:left:1:1:1:1",
                                    text="A",
                                    checkbox=True,
                                    bg_color="#FFFFFF",
                                    text_color="#000000",
                                    math_enabled=True,
                                )
                            ]
                        )
                    ],
                ),
            )
        ],
        bookmarks=[Bookmark(title="Root", page=1, dest="root")],
    )

    payload = json.loads(serialize_template(template).decode("utf-8"))

    assert payload["config"]["pageAlignment"] == 1
    assert payload["config"]["pageBorder"] == "0:0:0:0"
    assert payload["config"]["pdfTitle"] == "Schema Test"
    assert payload["config"]["arlingtonCompatible"] is True
    assert payload["config"]["pdfaCompliant"] is True
    assert payload["config"]["embedFonts"] is True
    assert payload["config"]["signature"]["certificatePem"] == "cert"
    assert payload["config"]["signature"]["privateKeyPem"] == "key"
    assert payload["config"]["signature"]["certificateChain"] == ["chain"]
    assert payload["config"]["signature"]["contactInfo"] == "contact@example.com"
    cell = payload["elements"][0]["table"]["rows"][0]["row"][0]
    assert cell["chequebox"] is True
    assert cell["bgcolor"] == "#FFFFFF"
    assert cell["textcolor"] == "#000000"
    assert cell["mathEnabled"] is True
    assert "checkbox" not in cell
