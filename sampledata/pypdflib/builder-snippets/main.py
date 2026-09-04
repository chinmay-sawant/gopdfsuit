#!/usr/bin/env python3
"""Builder-snippets example using pypdfsuit.

Mirrors sampledata/builder-snippets/snippet.json and the Go sample in
sampledata/gopdflib/builder-snippets/main.go: a 3-column title table with a
bracketed right cell in a different color, a spacer, a placeholder
image, and a 2-column data table with a red right cell.

Run from the repo root (requires the built libgopdfsuit shared library):

    python3 sampledata/pypdflib/builder-snippets/main.py
"""

import sys
from pathlib import Path

from pypdfsuit.builder import (
    TemplateBuilder,
    add_bracket_text,
    header_cell,
    make_props,
    new_cell,
    set_bracket_font,
    set_cell_font,
    set_cell_text_color,
)
from pypdfsuit.types import Element, Footer, Image, Row, TitleTable

# 1x1 PNG so the sample needs no image files.
PLACEHOLDER_PNG = (
    "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
)

OUT = Path(__file__).with_name("builder_snippets.pdf")


def build_template():
    builder = TemplateBuilder(page="A4", portrait=True)
    builder.add_title("Document Title", font="Helvetica", size=18, bold=True)
    builder.config.page_border = "1:1:1:1"
    builder.config.pdf_title = "Copy-clip snippet"

    # Title table is set directly: 3 columns with weights [1, 2, 1].
    title_row = [
        new_cell("", "Helvetica:12:000:left:1:1:1:1"),
        new_cell("Document Title", "Helvetica:18:100:center:1:1:1:1"),
        new_cell("clause", "Helvetica:12:000:right:1:1:1:1"),
    ]
    set_cell_font(title_row[1], "Helvetica", 18, bold=True)
    add_bracket_text(title_row[2], "[", "]")
    set_bracket_font(title_row[2], "Helvetica", 12)
    set_cell_text_color(title_row[2], "#B00020")  # different color on right
    builder.title.table = TitleTable(
        max_columns=3,
        rows=[Row(row=title_row)],
        column_widths=[1, 2, 1],
    )
    builder.title.textprops = make_props(
        "Helvetica", 18, bold=True, align="center", borders=(0, 0, 0, 0)
    )

    builder.add_spacer(20)
    builder._elements.append(
        Element(
            type="image",
            image=Image(
                image_name="placeholder",
                image_data=PLACEHOLDER_PNG,
                width=100,
                height=80,
            ),
        )
    )

    table = builder.add_table(2, 2, 1)
    table.add_row(header_cell("Item"), header_cell("Price"))
    amounts = table.add_row(
        new_cell(
            "Total Revenue",
            make_props("Helvetica", 10, align="left", borders=(1, 1, 1, 1)),
        ),
        new_cell(
            "$2,450,000",
            make_props("Helvetica", 10, align="right", borders=(1, 1, 1, 1)),
        ),
    )
    set_cell_text_color(amounts[1], "#B00020")

    template = builder.build()
    template.footer = Footer(
        font="Helvetica:8:000:center:0:0:0:0",
        text="Copy-clip snippet footer",
        props="Helvetica:8:000:center:0:0:0:0",
    )
    return template


def main() -> int:
    try:
        from pypdfsuit import generate_pdf
    except ImportError as exc:
        print(f"pypdfsuit bindings unavailable: {exc}")
        print("Build libgopdfsuit first, then retry.")
        return 1

    pdf_bytes = generate_pdf(build_template())
    OUT.write_bytes(pdf_bytes)
    print(f"wrote {len(pdf_bytes)} bytes to {OUT}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
