#!/usr/bin/env python3
"""
Builder-fluent sample for pypdfsuit.

Shows the preferred builder spelling instead of raw colon props strings:

    Font("Helvetica").size(12).bold().center().bordered().cell("Name")

emits the same bytes as:

    Cell(props="Helvetica:12:100:center:1:1:1:1", text="Name")

Run with: python builder_fluent_sample.py
Output: builder_fluent_sample.pdf next to this file.
"""

from pathlib import Path

from pypdfsuit.builder import Font, TemplateBuilder, Text, add_bracket_text, set_cell_text_color


def build_template() -> TemplateBuilder:
    b = TemplateBuilder("A4", True)
    b.add_title("Financial Summary", font="Helvetica", size=24, bold=True)

    header = b.add_table(2, 2.0, 1.0)
    header.add_row(
        Font("Helvetica").size(12).bold().center().bordered().cell("Metric"),
        Font("Helvetica").size(12).bold().center().bordered().cell("Value"),
    )

    body = b.add_table(2, 2.0, 1.0)
    body.add_row(
        Font("Helvetica").size(10).bordered().cell("Total Revenue"),
        Font("Helvetica").size(10).right().bordered().cell("$2,450,000"),
    )
    profit = body.add_row(
        Font("Helvetica").size(10).bold().bordered().cell("Gross Profit"),
        Font("Helvetica").size(10).bold().right().bordered().cell("$1,225,000"),
    )
    set_cell_text_color(profit[1], "#1E8449")

    clause = Text("[clause]").font(Font("Helvetica").size(10).right()).fg("#B00020").build()
    body.add_row(
        Font("Helvetica").size(10).bordered().cell("Note"),
        clause,
    )

    plain = Text("plain cell").font(Font("Helvetica").size(10)).build()
    add_bracket_text(plain, "[", "]")
    body.add_row(
        Font("Helvetica").size(10).borderless().cell("Bracket demo"),
        plain,
    )

    b.add_spacer(20)

    totals = b.add_table(2, 2.0, 1.0)
    totals.add_row(
        Font("Helvetica").size(11).bold().borderless().cell("Net Income"),
        Font("Helvetica").size(11).bold().right().borderless().cell("$125,000"),
    )

    return b


def main() -> int:
    print("=== pypdfsuit fluent-builder sample ===")
    demo = Font("Helvetica").size(12).bold().center().bordered()
    print(f"props demo: {demo.props()}")
    assert demo.props() == "Helvetica:12:100:center:1:1:1:1"

    b = build_template()
    pdf_bytes = b.generate()
    if not pdf_bytes.startswith(b"%PDF-"):
        print("ERROR: output is not a PDF")
        return 1

    out = Path(__file__).parent / "builder_fluent_sample.pdf"
    out.write_bytes(pdf_bytes)
    print(f"wrote {len(pdf_bytes)} bytes to {out}")
    print("=== Done ===")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
