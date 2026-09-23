# Life

A private, local-first life and financial dashboard built with Go, Bubble Tea,
Lip Gloss, and SQLite.

Life keeps each user's profile and financial data in a local `life.db`. The
database, PDFs, CSVs, and other personal source documents are excluded from
Git. A new installation starts with an empty database and opens Profile Setup.

## Download and run on Windows

1. Open the repository's **Releases** page on GitHub.
2. Download `life-windows-amd64.exe` from the latest release.
3. Open Windows Terminal or PowerShell in the download folder.
4. Run:

```powershell
.\life-windows-amd64.exe
```

Windows may show a security warning because this personal project is not code
signed. Choose **More info**, verify that the file came from this repository,
and select **Run anyway**.

On first launch, Life creates a private database in the current user's standard
application-data directory and asks the user to create a profile. Each user has
their own database; never share or commit `life.db`.

## Download and run on macOS

Download the appropriate file from **Releases**:

- Apple silicon: `life-macos-arm64`
- Intel Mac: `life-macos-amd64`

Then run:

```sh
chmod +x life-macos-arm64
./life-macos-arm64
```

## Build from source

Developers can install Go 1.25 or newer and run:

```sh
git clone https://github.com/MattHep12/life.git
cd life
go run ./cmd/life
```

To build a reusable executable on Windows:

```powershell
go build -o life.exe ./cmd/life
.\life.exe
```

If an existing `life.db` is present in the working directory, Life continues to
use it. This preserves databases created by earlier versions. Advanced users
can set `LIFE_DB_PATH` to select a specific database.

## Personalize a fresh profile

After Profile Setup, add accounts and balances, enter recurring expenses, and
adjust the assumptions in Goals & Projections. Payroll and card-statement
formats vary by employer and financial institution, so an included importer may
need separate parsing rules for another user's files.

Before sharing a sample document for parser work, redact names, addresses,
account numbers, employee IDs, and other identifying information.

## Import a payroll statement

Open **Finances → Payroll**, press `i`, choose **Payroll statement**, and paste
the full path to the PDF. Life parses the document locally, previews approved
values, and asks for confirmation before saving anything.

Developers can perform the same import from the source repository:

```sh
go run ./cmd/import-payroll "/path/to/pay-statement.pdf"
```

Only the document number, pay-period dates, pay date, summary amounts, source
filename, and import timestamp are stored. Names, addresses, employee IDs,
SSNs, and bank details are discarded in memory and never written to the
database.

The document number has a unique constraint, so importing the same statement
again updates that statement instead of creating a duplicate. Use `--yes` to
skip confirmation in a trusted automation.

## Import Schwab stock vesting data

Save copied Schwab vesting history as a tab-separated file with its column
headers. Open **Finances → Payroll**, press `i`, choose **Schwab stock vesting
history**, and paste the full path to the TSV file.

Developers can also run:

```sh
go run ./cmd/import-stock "/path/to/schwab-vests.tsv"
```

Restricted-stock rows sharing a vest date are aggregated into one event.
Dividend reinvestments are excluded. Award IDs, award dates, market-value
snapshots, and holding-period fields are not stored. Re-importing an existing
symbol and vest date safely skips the duplicate.

## Create a release

Pushing a version tag runs the test suite, builds Windows, macOS, and Linux
binaries, and publishes them on GitHub Releases:

```sh
git tag v0.1.0
git push origin v0.1.0
```
