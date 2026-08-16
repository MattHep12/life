# Life

Personal life dashboard built with Go and Bubble Tea.

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
