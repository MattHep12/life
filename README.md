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
