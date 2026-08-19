#!/usr/bin/env python3
"""Import a Discover year-to-date PDF into Life's local SQLite database."""

import argparse
import csv
import re
import sqlite3
from pathlib import Path

import pdfplumber


ROW = re.compile(
    r"^\s*(\d{2}/\d{2}/\d{2})\s+(\d{2}/\d{2}/\d{2})\s+"
    r"(.*?)\s+\$\s*(-?[\d,]+\.\d{2})\s+(.+?)\s*$"
)


def cents(value: str) -> int:
    return round(float(value.replace(",", "")) * 100)


def parse_pdf(path: Path) -> list[tuple[str, str, str, int, str, str, int]]:
    transactions = []
    with pdfplumber.open(path) as document:
        for page in document.pages:
            text = page.extract_text(layout=True) or ""
            for line in text.splitlines():
                if not re.match(r"^\s*\d{2}/\d{2}/\d{2}", line):
                    continue
                match = ROW.match(line)
                if not match:
                    raise ValueError(f"unsupported transaction row on page {page.page_number}")
                transaction_date, post_date, description, amount, category = match.groups()
                if category == "Awards and Rebate":
                    category = "Awards and Rebate Credits"
                index = len(transactions) + 1
                transactions.append(
                    (
                        f"20{transaction_date[6:8]}-{transaction_date[0:2]}-{transaction_date[3:5]}",
                        f"20{post_date[6:8]}-{post_date[0:2]}-{post_date[3:5]}",
                        " ".join(description.split()),
                        cents(amount),
                        category,
                        path.name,
                        index,
                    )
                )
    if not transactions:
        raise ValueError("no Discover transactions found")
    return transactions


def parse_csv(path: Path) -> list[tuple[str, str, str, int, str, str, int]]:
    transactions = []
    with path.open(newline="", encoding="utf-8-sig") as source:
        reader = csv.DictReader(source)
        expected = ["Trans. Date", "Post Date", "Description", "Amount", "Category"]
        if reader.fieldnames != expected:
            raise ValueError(f"unsupported Discover CSV columns: {reader.fieldnames}")
        for index, row in enumerate(reader, start=1):
            transaction_date = row["Trans. Date"]
            post_date = row["Post Date"]
            transactions.append(
                (
                    f"{transaction_date[6:10]}-{transaction_date[0:2]}-{transaction_date[3:5]}",
                    f"{post_date[6:10]}-{post_date[0:2]}-{post_date[3:5]}",
                    " ".join(row["Description"].split()),
                    cents(row["Amount"]),
                    row["Category"].strip(),
                    path.name,
                    index,
                )
            )
    if not transactions:
        raise ValueError("no Discover transactions found")
    return transactions


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("source", type=Path)
    parser.add_argument("--db", type=Path, default=Path("life.db"))
    parser.add_argument("--yes", action="store_true")
    args = parser.parse_args()

    if args.source.suffix.lower() == ".csv":
        transactions = parse_csv(args.source)
    elif args.source.suffix.lower() == ".pdf":
        transactions = parse_pdf(args.source)
    else:
        raise ValueError("Discover source must be a CSV or PDF")
    net_spending = sum(row[3] for row in transactions if row[4] != "Payments and Credits")
    payments = -sum(row[3] for row in transactions if row[4] == "Payments and Credits")
    print("SPENDING IMPORT PREVIEW")
    print(f"Transactions: {len(transactions)}")
    print(f"Date range:   {transactions[0][0]} through {transactions[-1][0]}")
    print(f"Net spending: ${net_spending / 100:,.2f}")
    print(f"Payments:     ${payments / 100:,.2f} (excluded from spending)")
    print("\nOnly dates, merchant descriptions, amounts, and categories will be saved.")
    if not args.yes and input("Import these transactions? [y/N]: ").strip().lower() != "y":
        print("Import canceled; nothing was saved.")
        return

    with sqlite3.connect(args.db) as database:
        database.executescript("""
            CREATE TABLE IF NOT EXISTS spending_transactions (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                transaction_date TEXT NOT NULL,
                post_date TEXT NOT NULL,
                description TEXT NOT NULL,
                amount_cents INTEGER NOT NULL,
                category TEXT NOT NULL,
                source_filename TEXT NOT NULL,
                source_index INTEGER NOT NULL,
                imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
                UNIQUE(source_filename, source_index)
            );
            CREATE INDEX IF NOT EXISTS idx_spending_transaction_date
            ON spending_transactions(transaction_date);
        """)
        replaced = database.execute(
            "SELECT COUNT(*) FROM spending_transactions "
            "WHERE source_filename GLOB 'Discover-*-YearToDateSummary.*'"
        ).fetchone()[0]
        database.execute(
            "DELETE FROM spending_transactions "
            "WHERE source_filename GLOB 'Discover-*-YearToDateSummary.*'"
        )
        before = database.total_changes
        database.executemany("""
            INSERT OR IGNORE INTO spending_transactions (
                transaction_date, post_date, description, amount_cents,
                category, source_filename, source_index
            ) VALUES (?, ?, ?, ?, ?, ?, ?)
        """, transactions)
        inserted = database.total_changes - before
    print(f"Replaced {replaced} older Discover rows with {inserted} current transactions.")


if __name__ == "__main__":
    main()
