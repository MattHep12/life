# Life

A private, local-first life and financial dashboard built with Go, Bubble Tea,
Lip Gloss, and SQLite.

Life stores profiles and financial data in `life.db`. The database, PDFs, CSVs,
and other personal source documents are excluded from Git. A new installation
starts with an empty database and asks the user to create a local profile.

## Run on macOS or Linux

Install Go 1.25 or newer, clone the repository, and run:

```sh
go run ./cmd/life
```

To build a reusable executable:

```sh
go build -o life ./cmd/life
./life
```

## Run on Windows

Install Go 1.25 or newer and use Windows Terminal or PowerShell:

```powershell
git clone https://github.com/MattHep12/life.git
cd life
go build -o life.exe ./cmd/life
.\life.exe
```

The first launch creates a new local `life.db` and opens Profile Setup. Each
user should keep their own database; do not copy another person's `life.db`.

## Personalize a fresh profile

After profile setup, use the app to add accounts and adjust the assumptions in
Goals & Projections. Payroll and card-statement formats vary by employer and
financial institution, so the included importers may need separate parsing
rules for a different user's files.

Before sharing a sample document for parser work, redact names, addresses,
account numbers, employee IDs, and other identifying information.

## Import a payroll statement

Payroll PDFs are parsed locally by a separate command. The importer previews the
approved values and asks for confirmation before writing to SQLite:

```sh
go run ./cmd/import-payroll --db life.db /path/to/pay-statement.pdf
```

Only the document number, pay-period dates, pay date, summary amounts, source
filename, and import timestamp are stored. Names, addresses, employee IDs, SSNs,
and bank details are discarded in memory and are never written to the database.

The document number has a unique constraint, so importing the same statement
twice is rejected. Use `--yes` to skip confirmation in a trusted automation.

## Import Schwab stock vesting data

Save copied Schwab vesting history as a tab-separated file, preserving its
column headers, then run:

```sh
go run ./cmd/import-stock --db life.db /path/to/schwab-vests.tsv
```

Restricted-stock rows sharing a vest date are aggregated into one event.
Dividend reinvestments are excluded. Award IDs, award dates, market-value
snapshots, and holding-period fields are not stored. Re-importing an existing
symbol and vest date is safely skipped.
