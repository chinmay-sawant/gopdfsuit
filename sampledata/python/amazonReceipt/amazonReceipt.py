from __future__ import annotations

import argparse
import base64
from dataclasses import dataclass
from decimal import Decimal, ROUND_HALF_UP
import json
from pathlib import Path
import struct
import sys
import zlib

from pypdfsuit import Cell, Config, Element, Footer, Image, PDFTemplate, Row, Spacer, Table, Title
from pypdfsuit._bindings import call_bytes_result, get_lib


MONEY_PLACES = Decimal("0.01")


def money(value: Decimal) -> str:
    return f"${value.quantize(MONEY_PLACES, rounding=ROUND_HALF_UP)}"


@dataclass(frozen=True)
class ReceiptItem:
    description: str
    quantity: int
    unit_price: Decimal

    @property
    def amount(self) -> Decimal:
        return (self.unit_price * self.quantity).quantize(MONEY_PLACES, rounding=ROUND_HALF_UP)


@dataclass(frozen=True)
class ReceiptData:
    merchant_name: str
    merchant_email: str
    merchant_phone: str
    customer_name: str
    customer_email: str
    order_id: str
    receipt_number: str
    transaction_id: str
    authorization_code: str
    purchase_date: str
    payment_method: str
    payment_status: str
    billing_address: str
    shipping_address: str
    delivery_date: str
    tracking_number: str
    items: list[ReceiptItem]
    shipping_fee: Decimal
    tax: Decimal
    promo_discount: Decimal
    notes: str
    support_contact: str
    support_url: str

    @property
    def subtotal(self) -> Decimal:
        total = sum((item.amount for item in self.items), Decimal("0.00"))
        return total.quantize(MONEY_PLACES, rounding=ROUND_HALF_UP)

    @property
    def total(self) -> Decimal:
        total = self.subtotal + self.shipping_fee + self.tax - self.promo_discount
        return total.quantize(MONEY_PLACES, rounding=ROUND_HALF_UP)


def section_label(text: str) -> Cell:
    return Cell(
        props="Helvetica:10:100:left:1:1:1:1",
        text=text,
        bg_color="#f7f9fb",
        text_color="#37475a",
    )


def info_value(text: str) -> Cell:
    return Cell(
        props="Helvetica:10:000:left:1:1:1:1",
        text=text,
        wrap=True,
        text_color="#111111",
    )


def section_header(text: str) -> Table:
    return Table(
        max_columns=1,
        column_widths=[1.0],
        rows=[
            Row(
                row=[
                    Cell(
                        props="Helvetica:11:100:left:1:1:1:1",
                        text=text,
                        bg_color="#232f3e",
                        text_color="#ffffff",
                    )
                ]
            )
        ],
    )


def _png_chunk(chunk_type: bytes, payload: bytes) -> bytes:
    return (
        struct.pack(">I", len(payload))
        + chunk_type
        + payload
        + struct.pack(">I", zlib.crc32(chunk_type + payload) & 0xFFFFFFFF)
    )


def sample_product_image() -> Image:
    width = 180
    height = 120
    rows: list[bytes] = []

    for y in range(height):
        row = bytearray([0])
        for x in range(width):
            if y < 22:
                pixel = (35, 47, 62)
            elif 18 < y < 102 and 22 < x < 98:
                pixel = (255, 216, 20)
            elif 30 < y < 90 and 108 < x < 162:
                pixel = (89, 110, 129)
            else:
                red = 238 - min(y, 80)
                green = 244 - min(y // 2, 70)
                blue = 248 - min(x // 4, 40)
                pixel = (red, green, blue)
            row.extend(pixel)
        rows.append(bytes(row))

    raw = b"".join(rows)
    png = b"\x89PNG\r\n\x1a\n"
    png += _png_chunk(b"IHDR", struct.pack(">IIBBBBB", width, height, 8, 2, 0, 0, 0))
    png += _png_chunk(b"IDAT", zlib.compress(raw, level=9))
    png += _png_chunk(b"IEND", b"")

    return Image(
        image_name="sample-product.png",
        image_data=base64.b64encode(png).decode("ascii"),
        width=180.0,
        height=120.0,
    )


def build_summary_table(data: ReceiptData) -> Table:
    rows = [
        Row(row=[section_label("Receipt number"), info_value(data.receipt_number)]),
        Row(row=[section_label("Order ID"), info_value(data.order_id)]),
        Row(row=[section_label("Transaction ID"), info_value(data.transaction_id)]),
        Row(row=[section_label("Authorization code"), info_value(data.authorization_code)]),
        Row(row=[section_label("Purchase date"), info_value(data.purchase_date)]),
        Row(row=[section_label("Payment method"), info_value(data.payment_method)]),
        Row(row=[section_label("Payment status"), info_value(data.payment_status)]),
        Row(row=[section_label("Merchant"), info_value(f"{data.merchant_name}\n{data.merchant_email}\n{data.merchant_phone}")]),
        Row(row=[section_label("Customer"), info_value(f"{data.customer_name}\n{data.customer_email}")]),
    ]
    return Table(max_columns=2, column_widths=[1.2, 2.8], rows=rows, bg_color="#ffffff")


def build_fulfillment_table(data: ReceiptData) -> Table:
    rows = [
        Row(row=[section_label("Delivery date"), info_value(data.delivery_date)]),
        Row(row=[section_label("Tracking number"), info_value(data.tracking_number)]),
        Row(row=[section_label("Support"), info_value(f"{data.support_contact}\n{data.support_url}")]),
    ]
    return Table(max_columns=2, column_widths=[1.2, 2.8], rows=rows, bg_color="#ffffff")


def build_address_table(data: ReceiptData) -> Table:
    header_props = "Helvetica:11:100:left:1:1:1:1"
    value_props = "Helvetica:10:000:left:1:1:1:1"
    rows = [
        Row(
            row=[
                Cell(props=header_props, text="Billing address", bg_color="#f3f3f3"),
                Cell(props=header_props, text="Shipping address", bg_color="#f3f3f3"),
            ]
        ),
        Row(
            row=[
                Cell(props=value_props, text=data.billing_address, wrap=True),
                Cell(props=value_props, text=data.shipping_address, wrap=True),
            ]
        ),
    ]
    return Table(max_columns=2, column_widths=[1.0, 1.0], rows=rows)


def build_product_table() -> Table:
    detail_lines = "\n".join(
        [
            "Example product preview",
            "Echo Studio Smart Speaker",
            "Finish: Glacier White",
            "SKU: ES-GL-2048",
            "Coverage: 24-month limited warranty",
            "Package: Speaker, power adapter, quick-start guide",
        ]
    )
    rows = [
        Row(
            row=[
                Cell(
                    props="Helvetica:10:000:center:1:1:1:1",
                    image=sample_product_image(),
                    height=125.0,
                    bg_color="#ffffff",
                ),
                Cell(
                    props="Helvetica:10:000:left:1:1:1:1",
                    text=detail_lines,
                    wrap=True,
                    bg_color="#ffffff",
                ),
            ]
        )
    ]
    return Table(max_columns=2, column_widths=[1.2, 2.4], rows=rows)


def build_items_table(data: ReceiptData) -> Table:
    rows = [
        Row(
            row=[
                Cell(props="Helvetica:11:100:left:1:1:1:1", text="Description", bg_color="#ffd814"),
                Cell(props="Helvetica:11:100:center:1:1:1:1", text="Qty", bg_color="#ffd814"),
                Cell(props="Helvetica:11:100:right:1:1:1:1", text="Unit price", bg_color="#ffd814"),
                Cell(props="Helvetica:11:100:right:1:1:1:1", text="Amount", bg_color="#ffd814"),
            ]
        )
    ]

    for item in data.items:
        rows.append(
            Row(
                row=[
                    Cell(props="Helvetica:10:000:left:1:1:1:1", text=item.description, wrap=True),
                    Cell(props="Helvetica:10:000:center:1:1:1:1", text=str(item.quantity)),
                    Cell(props="Helvetica:10:000:right:1:1:1:1", text=money(item.unit_price)),
                    Cell(props="Helvetica:10:000:right:1:1:1:1", text=money(item.amount)),
                ]
            )
        )

    return Table(max_columns=4, column_widths=[3.5, 0.7, 1.3, 1.3], rows=rows)


def build_totals_table(data: ReceiptData) -> Table:
    rows = [
        Row(row=[Cell(props="Helvetica:10:000:right:1:1:1:1", text="Subtotal"), Cell(props="Helvetica:10:000:right:1:1:1:1", text=money(data.subtotal))]),
        Row(row=[Cell(props="Helvetica:10:000:right:1:1:1:1", text="Shipping"), Cell(props="Helvetica:10:000:right:1:1:1:1", text=money(data.shipping_fee))]),
        Row(row=[Cell(props="Helvetica:10:000:right:1:1:1:1", text="Tax"), Cell(props="Helvetica:10:000:right:1:1:1:1", text=money(data.tax))]),
        Row(row=[Cell(props="Helvetica:10:000:right:1:1:1:1", text="Promo discount"), Cell(props="Helvetica:10:000:right:1:1:1:1", text=f"-{money(data.promo_discount)}")]),
        Row(
            row=[
                Cell(props="Helvetica:12:100:right:1:1:1:1", text="Total charged", bg_color="#232f3e", text_color="#ffffff"),
                Cell(props="Helvetica:12:100:right:1:1:1:1", text=money(data.total), bg_color="#232f3e", text_color="#ffffff"),
            ]
        ),
    ]
    return Table(max_columns=2, column_widths=[2.0, 1.0], rows=rows)


def build_notes_table(data: ReceiptData) -> Table:
    rows = [
        Row(row=[Cell(props="Helvetica:11:100:left:1:1:1:1", text="Payment note", text_color="#37475a", bg_color="#f7f9fb")]),
        Row(row=[Cell(props="Helvetica:10:000:left:1:1:1:1", text=data.notes, wrap=True)]),
    ]
    return Table(max_columns=1, column_widths=[1.0], rows=rows)


def build_title_table() -> Table:
    return Table(
        max_columns=3,
        column_widths=[1.0, 2.0, 1.0],
        rows=[
            Row(
                row=[
                    Cell(
                        props="Helvetica:18:100:center:0:0:1:1",
                        text="",
                        height=76.0,
                        bg_color="#232f3e",
                        text_color="#ffffff",
                    ),
                    Cell(
                        props="Helvetica:18:100:center:0:0:1:1",
                        text="Amazon Pay Receipt",
                        height=76.0,
                        bg_color="#232f3e",
                        text_color="#ffffff",
                    ),
                    Cell(
                        props="Helvetica:18:100:center:0:0:1:1",
                        text="",
                        height=76.0,
                        bg_color="#232f3e",
                        text_color="#ffffff",
                    ),
                ]
            )
        ],
    )


def build_template(data: ReceiptData) -> PDFTemplate:
    elements = [
        Element(type="table", table=section_header("Receipt details")),
        Element(type="table", table=build_summary_table(data)),
        Element(type="spacer", spacer=Spacer(height=8)),
        Element(type="table", table=section_header("Delivery and support")),
        Element(type="table", table=build_fulfillment_table(data)),
        Element(type="spacer", spacer=Spacer(height=8)),
        Element(type="table", table=section_header("Addresses")),
        Element(type="table", table=build_address_table(data)),
        Element(type="spacer", spacer=Spacer(height=8)),
        Element(type="table", table=section_header("Example product image")),
        Element(type="table", table=build_product_table()),
        Element(type="spacer", spacer=Spacer(height=150)),
        Element(type="table", table=section_header("Items purchased")),
        Element(type="table", table=build_items_table(data)),
        Element(type="spacer", spacer=Spacer(height=8)),
        Element(type="table", table=section_header("Payment summary")),
        Element(type="table", table=build_totals_table(data)),
        Element(type="spacer", spacer=Spacer(height=8)),
        Element(type="table", table=section_header("Notes")),
        Element(type="table", table=build_notes_table(data)),
    ]

    return PDFTemplate(
        config=Config(
            page="A4",
            page_alignment=1,
            page_border="0:0:0:0",
            watermark="",
            pdf_title="Amazon Pay Receipt",
            arlington_compatible=True,
            embed_fonts=True,
            pdfa_compliant=True,
        ),
        title=Title(
            props="Helvetica:12:000:left:1:1:1:1",
            text="2",
            table=build_title_table(),
        ),
        elements=elements,
        footer=Footer(font="Helvetica", text="Thank you for using Amazon Pay.")
    )


def build_template_payload(data: ReceiptData) -> dict:
    payload = build_template(data).to_dict()
    payload.setdefault("config", {})["pageMargin"] = "30:30:30:30"
    payload["config"]["page"] = "A4"
    payload["config"]["pageAlignment"] = 1
    payload["config"]["watermark"] = ""
    payload["config"]["pdfTitle"] = "Amazon Pay Receipt"
    payload["config"]["pdfaCompliant"] = True
    payload["config"]["arlingtonCompatible"] = True
    payload["config"]["embedFonts"] = True
    payload["config"]["embedStandardFonts"] = True
    payload["config"]["signature"] = {"enabled": False}
    payload.setdefault("title", {})["textprops"] = "Helvetica:18:100:center:1:1:1:1"
    return payload


def generate_pdf_with_page_margin(template_payload: dict) -> bytes:
    lib = get_lib()
    template_json = json.dumps(template_payload).encode("utf-8")
    return call_bytes_result(lib.GeneratePDF, template_json)


def sample_receipt() -> ReceiptData:
    return ReceiptData(
        merchant_name="Northwind Home Audio",
        merchant_email="payments@northwind-audio.example",
        merchant_phone="+91 20 5550 2048",
        customer_name="Aarav Patel",
        customer_email="aarav.patel@example.com",
        order_id="P01-8429916-4402715",
        receipt_number="AMZPAY-2026-0310-001",
        transaction_id="txn_8F3K4R91M2A7",
        authorization_code="AUTH-204831",
        purchase_date="2026-03-10 19:45 UTC",
        payment_method="Amazon Pay Visa ending in 2048",
        payment_status="Paid in full",
        billing_address="Aarav Patel\n48 Riverstone Avenue\nPune, Maharashtra 411014\nIndia",
        shipping_address="Aarav Patel\n48 Riverstone Avenue\nPune, Maharashtra 411014\nIndia",
        delivery_date="2026-03-13 by 20:00",
        tracking_number="NWHA204800319IN",
        items=[
            ReceiptItem(description="Echo Studio Smart Speaker\nColor: Glacier White | SKU: ES-GL-2048", quantity=1, unit_price=Decimal("189.99")),
            ReceiptItem(description="15W USB-C Fast Charger\nDual-port adapter | SKU: UC-15-FAST", quantity=2, unit_price=Decimal("24.50")),
            ReceiptItem(description="Screen cleaning kit\nMicrofiber cloth + spray bottle | SKU: CLN-101", quantity=1, unit_price=Decimal("11.90")),
        ],
        shipping_fee=Decimal("9.99"),
        tax=Decimal("22.04"),
        promo_discount=Decimal("15.00"),
        notes=(
            "This receipt was generated successfully through Amazon Pay and includes shipment, support, and authorization details "
            "for easy reconciliation. For returns, tax invoices, or merchant disputes, contact Northwind Home Audio within 14 days "
            "of delivery and reference both the order ID and transaction ID above."
        ),
        support_contact="Northwind Home Audio Billing Desk | support@northwind-audio.example | +91 20 5550 2055",
        support_url="https://northwind-audio.example/support/receipts",
    )


def create_receipt(output_path: Path) -> Path:
    try:
        template_payload = build_template_payload(sample_receipt())
        pdf_bytes = generate_pdf_with_page_margin(template_payload)
    except FileNotFoundError as exc:
        raise RuntimeError(
            "pypdfsuit could not find the platform library required for PDF generation. "
            "On Windows, build and install the package from the gopdfsuit source repository so "
            "pypdfsuit/lib/gopdfsuit.dll is present. The package documentation says this requires "
            "Go plus a GCC-compatible compiler."
        ) from exc
    output_path.write_bytes(pdf_bytes)
    return output_path


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate an Amazon Pay style receipt PDF with pypdfsuit.")
    parser.add_argument(
        "-o",
        "--output",
        default="amazon_pay_receipt.pdf",
        help="Output PDF path. Defaults to amazon_pay_receipt.pdf in the current directory.",
    )
    args = parser.parse_args()

    output_path = Path(args.output).resolve()
    try:
        create_receipt(output_path)
    except RuntimeError as exc:
        print(exc, file=sys.stderr)
        print("Expected source-build flow from the published documentation:", file=sys.stderr)
        print("  1. git clone https://github.com/chinmay-sawant/gopdfsuit.git", file=sys.stderr)
        print("  2. cd gopdfsuit/bindings/python", file=sys.stderr)
        print("  3. ./build.sh", file=sys.stderr)
        print("  4. pip install .", file=sys.stderr)
        raise SystemExit(1)
    print(f"Receipt written to {output_path}")


if __name__ == "__main__":
    main()