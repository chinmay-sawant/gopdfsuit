#!/usr/bin/env python3

import concurrent.futures
import os
import random
import statistics
import sys
import threading
import time
import platform
import subprocess
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(REPO_ROOT / "bindings" / "python"))

from pypdfsuit import (  # noqa: E402
    Bookmark,
    Cell,
    Config,
    Element,
    Footer,
    PDFTemplate,
    Row,
    SignatureConfig,
    Spacer,
    Table,
    Title,
    TitleTable,
    generate_pdf,
)


SYMBOLS = [
    "RELIANCE", "TCS", "INFY", "HDFCBANK", "TATASTEEL",
    "ICICIBANK", "SBIN", "WIPRO", "BHARTIARTL", "LT",
    "NIFTY24FEB22000CE", "NIFTY24FEB22000PE",
    "BANKNIFTY24FEB46000CE", "BANKNIFTY24FEB46000PE",
    "AXISBANK", "KOTAKBANK", "MARUTI", "TITAN", "ADANIENT", "BAJFINANCE",
]
ACTIONS = ["BUY", "SELL"]


def read_text(path: Path) -> str:
    return path.read_text()


def read_chain() -> list[str]:
    parts = (REPO_ROOT / "certs" / "chain.pem").read_text().strip().split("-----END CERTIFICATE-----")
    chain = []
    for part in parts:
        stripped = part.strip()
        if stripped:
            chain.append(f"{stripped}\n-----END CERTIFICATE-----")
    return chain


def retail_signature() -> SignatureConfig:
    return SignatureConfig(
        enabled=True,
        visible=True,
        name="Zerodha Compliance",
        reason="I am the author of this document",
        location="Mumbai, India",
        contact_info="compliance@brokerage.com",
        private_key_pem=read_text(REPO_ROOT / "certs" / "leaf.key"),
        certificate_pem=read_text(REPO_ROOT / "certs" / "leaf.pem"),
        certificate_chain=read_chain(),
    )


def generate_trades(count: int, seed: int) -> list[dict]:
    rng = random.Random(seed)
    trades = []
    hour, minute, second = 9, 15, 0
    for index in range(count):
        symbol = rng.choice(SYMBOLS)
        action = rng.choice(ACTIONS)
        qty = (rng.randint(1, 50)) * 10
        price = round(100.0 + rng.random() * 3400.0, 2)
        total = round(qty * price, 2)
        trades.append(
            {
                "id": index + 1,
                "time": f"{hour:02d}:{minute:02d}:{second:02d}",
                "symbol": symbol,
                "isin": "INE000000000",
                "action": action,
                "qty": qty,
                "price": price,
                "total": total,
            }
        )
        second += 1
        if second >= 60:
            second = 0
            minute += 1
        if minute >= 60:
            minute = 0
            hour += 1
    return trades


def bookmark(title: str, page: int, dest: str | None = None, children: list[Bookmark] | None = None) -> Bookmark:
    return Bookmark(title=title, page=page, dest=dest, children=children)


def get_machine_info() -> dict[str, str]:
    info = {
        "kernel": platform.platform(),
        "machine": platform.machine(),
        "processor": platform.processor() or "unknown",
        "python": sys.version.split()[0],
        "cpu_count": str(os.cpu_count()),
    }

    try:
        lscpu_output = subprocess.run(
            ["lscpu"],
            check=True,
            capture_output=True,
            text=True,
        ).stdout.splitlines()
        wanted_fields = {
            "Model name": "model_name",
            "Thread(s) per core": "threads_per_core",
            "Core(s) per socket": "cores_per_socket",
            "Socket(s)": "sockets",
        }
        for line in lscpu_output:
            if ":" not in line:
                continue
            key, value = [part.strip() for part in line.split(":", 1)]
            if key in wanted_fields:
                info[wanted_fields[key]] = value
    except (OSError, subprocess.CalledProcessError):
        pass

    try:
        free_output = subprocess.run(
            ["free", "-h"],
            check=True,
            capture_output=True,
            text=True,
        ).stdout.splitlines()
        if len(free_output) >= 2:
            memory_parts = free_output[1].split()
            if len(memory_parts) >= 2:
                info["memory_total"] = memory_parts[1]
    except (OSError, subprocess.CalledProcessError):
        pass

    return info


def build_retail_template() -> PDFTemplate:
    return PDFTemplate(
        config=Config(
            page_border="0:0:0:0",
            page="A4",
            page_alignment=1,
            pdf_title="Contract Note - CN2024001",
            pdfa_compliant=True,
            arlington_compatible=True,
            embed_fonts=True,
            signature=retail_signature(),
        ),
        title=Title(
            props="Helvetica:24:100:center:0:0:0:0",
            text="CONTRACT NOTE",
            table=TitleTable(
                max_columns=2,
                column_widths=[1.5, 2.5],
                rows=[
                    Row(
                        row=[
                            Cell(props="Helvetica:20:100:center:0:0:0:0", text="CONTRACT NOTE", bg_color="#154360", text_color="#FFFFFF", height=45),
                            Cell(props="Helvetica:11:000:right:0:0:0:0", text="CN2024001 | 2024-02-12", bg_color="#154360", text_color="#AED6F1", height=45),
                        ]
                    )
                ],
            ),
        ),
        elements=[
            Element(type="table", table=Table(max_columns=4, column_widths=[1, 1, 1, 1], rows=[Row(row=[
                Cell(props="Helvetica:8:000:center:0:0:0:1", text="Go to Client Info", text_color="#2E86C1", link="#client-info"),
                Cell(props="Helvetica:8:000:center:0:0:0:1", text="Go to Trades", text_color="#2E86C1", link="#trade-details"),
                Cell(props="Helvetica:8:000:center:0:0:0:1", text="Go to Financials", text_color="#2E86C1", link="#financial-summary"),
                Cell(props="Helvetica:8:000:center:0:0:0:1", text=""),
            ])])),
            Element(type="table", table=Table(max_columns=1, column_widths=[1], rows=[Row(row=[Cell(props="Helvetica:11:100:left:1:1:1:1", text="SECTION A: CLIENT INFORMATION", bg_color="#21618C", text_color="#FFFFFF", dest="client-info")])])),
            Element(type="table", table=Table(max_columns=4, column_widths=[1.2, 2, 1.2, 2], rows=[
                Row(row=[
                    Cell(props="Helvetica:9:100:left:1:0:0:1", text="Client Name:", bg_color="#EBF5FB"),
                    Cell(props="Helvetica:9:000:left:0:0:0:1", text="Rahul Sharma", bg_color="#EBF5FB"),
                    Cell(props="Helvetica:9:100:left:0:0:0:1", text="Client Code:", bg_color="#EBF5FB"),
                    Cell(props="Helvetica:9:000:left:0:1:0:1", text="RS9988", bg_color="#EBF5FB"),
                ]),
                Row(row=[
                    Cell(props="Helvetica:9:100:left:1:0:0:1", text="PAN:"),
                    Cell(props="Helvetica:9:000:left:0:0:0:1", text="ABCDE1234F"),
                    Cell(props="Helvetica:9:100:left:0:0:0:1", text="Trade Date:"),
                    Cell(props="Helvetica:9:000:left:0:1:0:1", text="2024-02-12"),
                ]),
            ])),
            Element(type="table", table=Table(max_columns=1, column_widths=[1], rows=[Row(row=[Cell(props="Helvetica:11:100:left:1:1:1:1", text="SECTION B: TRADE DETAILS", bg_color="#21618C", text_color="#FFFFFF", dest="trade-details")])])),
            Element(type="table", table=Table(max_columns=6, column_widths=[2, 1.5, 1, 1, 1.5, 1.5], rows=[
                Row(row=[
                    Cell(props="Helvetica:9:100:center:1:0:1:1", text="Symbol", bg_color="#D4E6F1"),
                    Cell(props="Helvetica:9:100:center:0:0:1:1", text="ISIN", bg_color="#D4E6F1"),
                    Cell(props="Helvetica:9:100:center:0:0:1:1", text="Action", bg_color="#D4E6F1"),
                    Cell(props="Helvetica:9:100:center:0:0:1:1", text="Qty", bg_color="#D4E6F1"),
                    Cell(props="Helvetica:9:100:right:0:0:1:1", text="Price", bg_color="#D4E6F1"),
                    Cell(props="Helvetica:9:100:right:0:1:1:1", text="Total", bg_color="#D4E6F1"),
                ]),
                Row(row=[
                    Cell(props="Helvetica:9:000:center:1:0:0:1", text="TATASTEEL"),
                    Cell(props="Helvetica:9:000:center:0:0:0:1", text="INE081A01012"),
                    Cell(props="Helvetica:9:000:center:0:0:0:1", text="BUY", text_color="#27AE60"),
                    Cell(props="Helvetica:9:000:center:0:0:0:1", text="10"),
                    Cell(props="Helvetica:9:000:right:0:0:0:1", text="₹145.50"),
                    Cell(props="Helvetica:9:000:right:0:1:0:1", text="₹1,455.00"),
                ]),
                Row(row=[
                    Cell(props="Helvetica:9:000:center:1:0:0:1", text="INFY", bg_color="#F8F9F9"),
                    Cell(props="Helvetica:9:000:center:0:0:0:1", text="INE009A01021", bg_color="#F8F9F9"),
                    Cell(props="Helvetica:9:000:center:0:0:0:1", text="SELL", text_color="#E74C3C", bg_color="#F8F9F9"),
                    Cell(props="Helvetica:9:000:center:0:0:0:1", text="5", bg_color="#F8F9F9"),
                    Cell(props="Helvetica:9:000:right:0:0:0:1", text="₹1,650.00", bg_color="#F8F9F9"),
                    Cell(props="Helvetica:9:000:right:0:1:0:1", text="₹8,250.00", bg_color="#F8F9F9"),
                ]),
            ])),
            Element(type="table", table=Table(max_columns=1, column_widths=[1], rows=[Row(row=[Cell(props="Helvetica:11:100:left:1:1:1:1", text="SECTION C: FINANCIAL SUMMARY", bg_color="#21618C", text_color="#FFFFFF", dest="financial-summary")])])),
            Element(type="table", table=Table(max_columns=2, column_widths=[2, 1], rows=[
                Row(row=[Cell(props="Helvetica:9:000:left:1:0:0:1", text="Net Obligation"), Cell(props="Helvetica:9:100:right:0:1:0:1", text="₹6,795.00")]),
                Row(row=[Cell(props="Helvetica:9:000:left:1:0:0:1", text="STT Tax", bg_color="#F8F9F9"), Cell(props="Helvetica:9:000:right:0:1:0:1", text="₹12.50", bg_color="#F8F9F9")]),
                Row(row=[Cell(props="Helvetica:10:100:left:1:0:1:1", text="Total Payable", bg_color="#A9CCE3"), Cell(props="Helvetica:10:100:right:0:1:1:1", text="₹6,807.50", bg_color="#A9CCE3")]),
            ])),
        ],
        footer=Footer(font="Helvetica:7:000:center", text="ZERODHA BROKING LTD | CONTRACT NOTE | CONFIDENTIAL"),
        bookmarks=[bookmark("Contract Note - CN2024001", 1, children=[
            bookmark("Client Information", 1, "client-info"),
            bookmark("Trade Details", 1, "trade-details"),
            bookmark("Financial Summary", 1, "financial-summary"),
        ])],
    )


def build_active_trader_template() -> PDFTemplate:
    trades = generate_trades(40, 42)
    trade_rows = [
        Row(row=[
            Cell(props="Helvetica:8:100:center:1:0:1:1", text="Symbol", bg_color="#D4E6F1"),
            Cell(props="Helvetica:8:100:center:0:0:1:1", text="Action", bg_color="#D4E6F1"),
            Cell(props="Helvetica:8:100:center:0:0:1:1", text="Qty", bg_color="#D4E6F1"),
            Cell(props="Helvetica:8:100:right:0:0:1:1", text="Price", bg_color="#D4E6F1"),
            Cell(props="Helvetica:8:100:right:0:1:1:1", text="Total", bg_color="#D4E6F1"),
        ])
    ]
    total_turnover = 0.0
    for index, trade in enumerate(trades):
        bg = "#F8F9F9" if index % 2 else None
        total_turnover += trade["total"]
        trade_rows.append(Row(row=[
            Cell(props="Helvetica:8:000:center:1:0:0:1", text=trade["symbol"], bg_color=bg),
            Cell(props="Helvetica:8:000:center:0:0:0:1", text=trade["action"], text_color="#E74C3C" if trade["action"] == "SELL" else "#27AE60", bg_color=bg),
            Cell(props="Helvetica:8:000:center:0:0:0:1", text=str(trade["qty"]), bg_color=bg),
            Cell(props="Helvetica:8:000:right:0:0:0:1", text=f"₹{trade['price']:.2f}", bg_color=bg),
            Cell(props="Helvetica:8:000:right:0:1:0:1", text=f"₹{trade['total']:.2f}", bg_color=bg),
        ]))

    return PDFTemplate(
        config=Config(
            page_border="0:0:0:0",
            page="A4",
            page_alignment=1,
            watermark="CONFIDENTIAL",
            pdf_title="Contract Note - Active Trader",
            pdfa_compliant=True,
            arlington_compatible=True,
            embed_fonts=True,
        ),
        title=Title(
            props="Helvetica:24:100:center:0:0:0:0",
            text="ACTIVE TRADER CONTRACT NOTE",
            table=TitleTable(max_columns=2, column_widths=[1.5, 2.5], rows=[Row(row=[
                Cell(props="Helvetica:18:100:center:0:0:0:0", text="ACTIVE TRADER CONTRACT NOTE", bg_color="#154360", text_color="#FFFFFF", height=45),
                Cell(props="Helvetica:11:000:right:0:0:0:0", text="40 Trades | 2024-02-12", bg_color="#154360", text_color="#AED6F1", height=45),
            ])]),
        ),
        elements=[
            Element(type="table", table=Table(max_columns=1, column_widths=[1], rows=[Row(row=[Cell(props="Helvetica:11:100:left:1:1:1:1", text="SECTION A: CLIENT INFORMATION", bg_color="#21618C", text_color="#FFFFFF", dest="active-client-info")])])),
            Element(type="table", table=Table(max_columns=4, column_widths=[1.2, 2, 1.2, 2], rows=[
                Row(row=[
                    Cell(props="Helvetica:9:100:left:1:0:0:1", text="Client Name:", bg_color="#EBF5FB"),
                    Cell(props="Helvetica:9:000:left:0:0:0:1", text="Priya Venkatesh", bg_color="#EBF5FB"),
                    Cell(props="Helvetica:9:100:left:0:0:0:1", text="Client Code:", bg_color="#EBF5FB"),
                    Cell(props="Helvetica:9:000:left:0:1:0:1", text="PV5544", bg_color="#EBF5FB"),
                ]),
                Row(row=[
                    Cell(props="Helvetica:9:100:left:1:0:0:1", text="PAN:"),
                    Cell(props="Helvetica:9:000:left:0:0:0:1", text="FGHIJ5678K"),
                    Cell(props="Helvetica:9:100:left:0:0:0:1", text="Trade Date:"),
                    Cell(props="Helvetica:9:000:left:0:1:0:1", text="2024-02-12"),
                ]),
                Row(row=[
                    Cell(props="Helvetica:9:000:left:1:0:0:1", text=""),
                    Cell(props="Helvetica:9:000:left:0:0:0:1", text="Go to Trades", text_color="#2E86C1", link="#trades-section"),
                    Cell(props="Helvetica:9:000:left:0:0:0:1", text="Go to Summary", text_color="#2E86C1", link="#summary-section"),
                    Cell(props="Helvetica:9:000:left:0:1:0:1", text=""),
                ]),
            ])),
            Element(type="table", table=Table(max_columns=1, column_widths=[1], rows=[Row(row=[Cell(props="Helvetica:11:100:left:1:1:1:1", text="SECTION B: TRADE DETAILS (40 TRADES)", bg_color="#21618C", text_color="#FFFFFF", dest="trades-section")])])),
            Element(type="table", table=Table(max_columns=5, column_widths=[2.5, 1, 1, 1.5, 1.5], rows=trade_rows)),
            Element(type="table", table=Table(max_columns=1, column_widths=[1], rows=[Row(row=[Cell(props="Helvetica:11:100:left:1:1:1:1", text="SECTION C: SUMMARY", bg_color="#21618C", text_color="#FFFFFF", dest="summary-section")])])),
            Element(type="table", table=Table(max_columns=2, column_widths=[2, 1], rows=[
                Row(row=[Cell(props="Helvetica:9:000:left:1:0:0:1", text="Total Turnover"), Cell(props="Helvetica:9:000:right:0:1:0:1", text=f"₹{total_turnover:.2f}")]),
                Row(row=[Cell(props="Helvetica:9:000:left:1:0:0:1", text="Brokerage", bg_color="#F8F9F9"), Cell(props="Helvetica:9:000:right:0:1:0:1", text="₹20.00", bg_color="#F8F9F9")]),
                Row(row=[Cell(props="Helvetica:9:000:left:1:0:0:1", text="Regulatory Charges"), Cell(props="Helvetica:9:000:right:0:1:0:1", text="₹150.00")]),
                Row(row=[Cell(props="Helvetica:10:100:left:1:0:1:1", text="Net Payable", bg_color="#A9CCE3"), Cell(props="Helvetica:10:100:right:0:1:1:1", text=f"₹{total_turnover + 170:.2f}", bg_color="#A9CCE3")]),
            ])),
        ],
        footer=Footer(font="Helvetica:7:000:center", text="ZERODHA BROKING LTD | ACTIVE TRADER CONTRACT NOTE | CONFIDENTIAL"),
        bookmarks=[bookmark("Active Trader Contract Note", 1, children=[
            bookmark("Client Information", 1, "active-client-info"),
            bookmark("Trade Details (40 Trades)", 1, "trades-section"),
            bookmark("Summary", 2, "summary-section"),
        ])],
    )


def build_hft_template() -> PDFTemplate:
    trades = generate_trades(2000, 99)
    trade_rows = [
        Row(row=[
            Cell(props="Helvetica:7:100:center:1:0:1:1", text="ID", bg_color="#D4E6F1"),
            Cell(props="Helvetica:7:100:center:0:0:1:1", text="Time", bg_color="#D4E6F1"),
            Cell(props="Helvetica:7:100:center:0:0:1:1", text="Symbol", bg_color="#D4E6F1"),
            Cell(props="Helvetica:7:100:center:0:0:1:1", text="Action", bg_color="#D4E6F1"),
            Cell(props="Helvetica:7:100:center:0:0:1:1", text="Qty", bg_color="#D4E6F1"),
            Cell(props="Helvetica:7:100:right:0:0:1:1", text="Price", bg_color="#D4E6F1"),
            Cell(props="Helvetica:7:100:right:0:1:1:1", text="Total", bg_color="#D4E6F1"),
        ])
    ]
    for index, trade in enumerate(trades):
        bg = "#F8F9F9" if index % 2 else None
        trade_rows.append(Row(row=[
            Cell(props="Helvetica:7:000:center:1:0:0:1", text=str(trade["id"]), bg_color=bg),
            Cell(props="Helvetica:7:000:center:0:0:0:1", text=trade["time"], bg_color=bg),
            Cell(props="Helvetica:7:000:center:0:0:0:1", text=trade["symbol"], bg_color=bg),
            Cell(props="Helvetica:7:000:center:0:0:0:1", text=trade["action"], text_color="#E74C3C" if trade["action"] == "SELL" else "#27AE60", bg_color=bg),
            Cell(props="Helvetica:7:000:center:0:0:0:1", text=str(trade["qty"]), bg_color=bg),
            Cell(props="Helvetica:7:000:right:0:0:0:1", text=f"₹{trade['price']:.2f}", bg_color=bg),
            Cell(props="Helvetica:7:000:right:0:1:0:1", text=f"₹{trade['total']:.2f}", bg_color=bg),
        ]))

    return PDFTemplate(
        config=Config(
            page_border="0:0:0:0",
            page="A4",
            page_alignment=1,
            pdf_title="Contract Note - HFT Algo Capital LLP",
            pdfa_compliant=True,
            arlington_compatible=True,
            embed_fonts=True,
        ),
        title=Title(
            props="Helvetica:24:100:center:0:0:0:0",
            text="HFT CONTRACT NOTE",
            table=TitleTable(max_columns=2, column_widths=[1.5, 2.5], rows=[Row(row=[
                Cell(props="Helvetica:16:100:center:0:0:0:0", text="HFT CONTRACT NOTE", bg_color="#154360", text_color="#FFFFFF", height=40),
                Cell(props="Helvetica:10:000:right:0:0:0:0", text="ALGO CAPITAL LLP | 2,000 Trades", bg_color="#154360", text_color="#AED6F1", height=40),
            ])]),
        ),
        elements=[
            Element(type="table", table=Table(max_columns=4, column_widths=[1, 1, 1, 1], rows=[Row(row=[
                Cell(props="Helvetica:7:000:center:0:0:0:1", text="Go to Client Info", text_color="#2E86C1", link="#hft-client-info"),
                Cell(props="Helvetica:7:000:center:0:0:0:1", text="Go to Trades", text_color="#2E86C1", link="#hft-trades"),
                Cell(props="Helvetica:7:000:center:0:0:0:1", text="Go to Compliance", text_color="#2E86C1", link="#hft-compliance"),
                Cell(props="Helvetica:7:000:center:0:0:0:1", text=""),
            ])])),
            Element(type="table", table=Table(max_columns=1, column_widths=[1], rows=[Row(row=[Cell(props="Helvetica:10:100:left:1:1:1:1", text="SECTION A: CLIENT INFORMATION", bg_color="#21618C", text_color="#FFFFFF", dest="hft-client-info")])])),
            Element(type="table", table=Table(max_columns=4, column_widths=[1.2, 2, 1.2, 2], rows=[
                Row(row=[
                    Cell(props="Helvetica:8:100:left:1:0:0:1", text="Client Name:", bg_color="#EBF5FB"),
                    Cell(props="Helvetica:8:000:left:0:0:0:1", text="Algo Capital LLP", bg_color="#EBF5FB"),
                    Cell(props="Helvetica:8:100:left:0:0:0:1", text="Client Code:", bg_color="#EBF5FB"),
                    Cell(props="Helvetica:8:000:left:0:1:0:1", text="HFT001", bg_color="#EBF5FB"),
                ]),
                Row(row=[
                    Cell(props="Helvetica:8:100:left:1:0:0:1", text="PAN:"),
                    Cell(props="Helvetica:8:000:left:0:0:0:1", text="ZZZZZ9999Z"),
                    Cell(props="Helvetica:8:100:left:0:0:0:1", text="Mode:"),
                    Cell(props="Helvetica:8:000:left:0:1:0:1", text="BATCH PROCESSING"),
                ]),
            ])),
            Element(type="table", table=Table(max_columns=1, column_widths=[1], rows=[Row(row=[Cell(props="Helvetica:10:100:left:1:1:1:1", text="SECTION B: TRADE DETAILS (2,000 TRADES)", bg_color="#21618C", text_color="#FFFFFF", dest="hft-trades")])])),
            Element(type="table", table=Table(max_columns=7, column_widths=[0.6, 1, 2, 0.8, 0.6, 1.5, 1.5], rows=trade_rows)),
            Element(type="table", table=Table(max_columns=1, column_widths=[1], rows=[Row(row=[Cell(props="Helvetica:10:100:left:1:1:1:1", text="SECTION C: COMPLIANCE AUDIT", bg_color="#21618C", text_color="#FFFFFF", dest="hft-compliance")])])),
            Element(type="table", table=Table(max_columns=2, column_widths=[2, 1], rows=[
                Row(row=[Cell(props="Helvetica:8:100:left:1:0:0:1", text="Audit Timestamp:"), Cell(props="Helvetica:8:000:left:0:1:0:1", text="2024-02-12T17:00:00Z")]),
                Row(row=[Cell(props="Helvetica:8:100:left:1:0:0:1", text="Auditor Signature:", bg_color="#F8F9F9"), Cell(props="Helvetica:8:010:left:0:1:0:1", text="[Placeholder]", bg_color="#F8F9F9")]),
            ])),
        ],
        footer=Footer(font="Helvetica:7:000:center", text="ALGO CAPITAL LLP | HFT CONTRACT NOTE | STRICTLY CONFIDENTIAL"),
        bookmarks=[bookmark("HFT Contract Note - Algo Capital LLP", 1, children=[
            bookmark("Client Information", 1, "hft-client-info"),
            bookmark("Trade Details (2,000 Trades)", 1, "hft-trades"),
            bookmark("Compliance Audit", 1, "hft-compliance"),
        ])],
    )


def run_benchmark(iterations: int = 5000, workers: int = 48) -> None:
    print("=== PyPDFSuit Zerodha Benchmark ===")
    print("Workload Mix: 80% Retail | 15% Active | 5% HFT")
    print()
    machine_info = get_machine_info()
    print(
        "Machine: "
        f"kernel={machine_info.get('kernel', 'unknown')}, "
        f"cpu={machine_info.get('model_name', machine_info.get('processor', 'unknown'))}, "
        f"cores={machine_info.get('cores_per_socket', 'unknown')}, "
        f"threads_per_core={machine_info.get('threads_per_core', 'unknown')}, "
        f"sockets={machine_info.get('sockets', 'unknown')}, "
        f"logical_cpus={machine_info.get('cpu_count', 'unknown')}, "
        f"memory={machine_info.get('memory_total', 'unknown')}, "
        f"python={machine_info.get('python', 'unknown')}"
    )
    print(f"Running {iterations} iterations using {workers} workers...\n")

    print("Building templates...")
    retail_template = build_retail_template()
    active_template = build_active_trader_template()
    hft_template = build_hft_template()
    print("Templates built.")

    print("Warm-up runs...")
    retail_pdf = generate_pdf(retail_template)
    active_pdf = generate_pdf(active_template)
    hft_pdf = generate_pdf(hft_template)
    print(f"  Retail PDF size:  {len(retail_pdf)} bytes ({len(retail_pdf) / 1024.0:.2f} KB)")
    print(f"  Active PDF size:  {len(active_pdf)} bytes ({len(active_pdf) / 1024.0:.2f} KB)")
    print(f"  HFT PDF size:     {len(hft_pdf)} bytes ({len(hft_pdf) / 1024.0:.2f} KB)")
    print()

    output_dir = Path(__file__).resolve().parent
    (output_dir / "zerodha_retail_output_pypdfsuit.pdf").write_bytes(retail_pdf)
    (output_dir / "zerodha_active_output_pypdfsuit.pdf").write_bytes(active_pdf)
    (output_dir / "zerodha_hft_output_pypdfsuit.pdf").write_bytes(hft_pdf)

    durations = []
    counts = {"retail": 0, "active": 0, "hft": 0}
    lock = threading.Lock()

    def one_run(seed: int) -> float:
        rng = random.Random(seed)
        roll = rng.randint(0, 99)
        if roll < 80:
            kind = "retail"
            template = retail_template
        elif roll < 95:
            kind = "active"
            template = active_template
        else:
            kind = "hft"
            template = hft_template

        start = time.perf_counter()
        generate_pdf(template)
        elapsed_ms = (time.perf_counter() - start) * 1000
        with lock:
            counts[kind] += 1
        return elapsed_ms

    total_start = time.perf_counter()
    with concurrent.futures.ThreadPoolExecutor(max_workers=workers) as executor:
        for elapsed_ms in executor.map(one_run, range(iterations)):
            durations.append(elapsed_ms)
    total_seconds = time.perf_counter() - total_start

    print("=== Performance Summary ===")
    print(f"  Iterations:      {iterations}")
    print(f"  Concurrency:     {workers} workers")
    print(f"  Total time:      {total_seconds:.3f} s")
    print(f"  Throughput:      {iterations / total_seconds:.2f} ops/sec")
    print()
    print(f"  Avg Latency:     {statistics.mean(durations):.3f} ms")
    print(f"  Min Latency:     {min(durations):.3f} ms")
    print(f"  Max Latency:     {max(durations):.3f} ms")
    print()
    print("=== Workload Distribution ===")
    print(f"  Retail  (80%):   {counts['retail']} iterations")
    print(f"  Active  (15%):   {counts['active']} iterations")
    print(f"  HFT      (5%):   {counts['hft']} iterations")
    print()
    print(f"Saved: zerodha_retail_output_pypdfsuit.pdf ({len(retail_pdf)} bytes)")
    print(f"Saved: zerodha_active_output_pypdfsuit.pdf ({len(active_pdf)} bytes)")
    print(f"Saved: zerodha_hft_output_pypdfsuit.pdf ({len(hft_pdf)} bytes)")
    print()
    print("=== Done ===")


if __name__ == "__main__":
    run_benchmark()