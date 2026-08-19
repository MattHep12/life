#!/usr/bin/env python3
"""Extract compensation components from local Google pay-statement PDFs."""

import argparse
import re
import sqlite3
from pathlib import Path

import pdfplumber


CATEGORIES = {
    "base": {"Regular Pay", "Holiday", "Sick Pay", "Vacation Pay"},
    "overtime": {"Over Time Wk 1", "Over Time Wk 2", "Holiday OnCall", "Holiday Wk Prem", "Travel Time"},
    "bonus": {"Annual Bonus", "Annl Bonus OT", "On Call Bonus", "Peer Bonus", "Spot Bonus"},
    "stock": {"Goog Stock Unit", "Equity OT"},
    "other": {"Group Term Life", "HSA ER Seed"},
}
KNOWN = {label for labels in CATEGORIES.values() for label in labels}
MONEY = re.compile(r"\$?([\d,]+\.\d{2})(?!\d)")
DOCUMENT = re.compile(r"Document\s+(\d+)", re.I)
SUMMARY = re.compile(
    r"Pay\s+Summary.*?Current\s+\$?([\d,]+\.\d{2})",
    re.I | re.S,
)


def to_cents(value: str) -> int:
    return round(float(value.replace(",", "")) * 100)


def extract(path: Path) -> tuple[str, dict[str, int]]:
    totals = {category: 0 for category in CATEGORIES}
    with pdfplumber.open(path) as document:
        text = "\n".join(page.extract_text() or "" for page in document.pages)
        document_match = DOCUMENT.search(text)
        summary_match = SUMMARY.search(text)
        if not document_match or not summary_match:
            raise ValueError(f"unsupported payroll layout: {path.name}")
        gross = to_cents(summary_match.group(1))
        for page in document.pages:
            for table in page.extract_tables() or []:
                if not table or not table[0] or "Pay Type" not in (table[0][0] or ""):
                    continue
                for row in table:
                    if not row:
                        continue
                    label = " ".join((row[0] or "").split())
                    if label not in KNOWN:
                        continue
                    values = MONEY.findall(" ".join((cell or "") for cell in row[1:]))
                    if not values:
                        continue
                    # Most rows contain pay rate, current, and YTD dollar values.
                    # Equity-only rows omit hours/rate and contain current and YTD.
                    current = to_cents(values[1] if len(values) >= 3 else values[0])
                    for category, labels in CATEGORIES.items():
                        if label in labels:
                            totals[category] += current
                            break
    classified = sum(totals.values())
    if classified != gross:
        raise ValueError(
            f"{path.name}: classified ${classified/100:,.2f}, expected gross ${gross/100:,.2f}"
        )
    return document_match.group(1), totals


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("pdfs", nargs="+", type=Path)
    parser.add_argument("--db", type=Path, default=Path("life.db"))
    args = parser.parse_args()

    enriched = 0
    with sqlite3.connect(args.db) as database:
        existing = {
            row[1] for row in database.execute("PRAGMA table_info(payroll_statements)")
        }
        for column in (
            "base_pay_cents", "overtime_cents", "bonus_cents",
            "payroll_stock_cents", "other_earnings_cents",
        ):
            if column not in existing:
                database.execute(
                    f"ALTER TABLE payroll_statements ADD COLUMN {column} INTEGER NOT NULL DEFAULT 0"
                )
        for path in args.pdfs:
            document_number, totals = extract(path)
            cursor = database.execute(
                """UPDATE payroll_statements SET base_pay_cents=?, overtime_cents=?,
                   bonus_cents=?, payroll_stock_cents=?, other_earnings_cents=?
                   WHERE document_number=?""",
                (
                    totals["base"], totals["overtime"], totals["bonus"],
                    totals["stock"], totals["other"], document_number,
                ),
            )
            enriched += cursor.rowcount
    print(f"Enriched {enriched} stored payroll statements from {len(args.pdfs)} PDFs.")


if __name__ == "__main__":
    main()
