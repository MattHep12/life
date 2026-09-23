package app

import (
	"path/filepath"
	"testing"
	"time"

	"life/internal/models"
	"life/internal/storage"
)

func TestCleanImportPathAcceptsQuotedWindowsPath(t *testing.T) {
	got := cleanImportPath(`"C:\Users\Example User\Downloads\pay stub.pdf"`)
	want := filepath.Clean(`C:\Users\Example User\Downloads\pay stub.pdf`)
	if got != want {
		t.Fatalf("cleanImportPath() = %q, want %q", got, want)
	}
}

func TestSavePendingPayrollImportRefreshesModel(t *testing.T) {
	db, err := storage.Open(filepath.Join(t.TempDir(), "life.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.CreateTables(); err != nil {
		t.Fatal(err)
	}

	statement := models.PayrollStatement{
		DocumentNumber: "12345",
		PeriodStart:    time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:      time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC),
		PayDate:        time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC),
		GrossCents:     500000,
		NetCents:       320000,
	}
	model := Model{db: db, importKind: importPayroll, importMode: importConfirm, pendingPayroll: &statement}
	model.savePendingImport()

	if model.importMode != importResult || len(model.payrollStatements) != 1 {
		t.Fatalf("save result: mode=%v statements=%d error=%q", model.importMode, len(model.payrollStatements), model.importError)
	}
	if model.payrollStatements[0].DocumentNumber != statement.DocumentNumber {
		t.Fatalf("saved document = %q, want %q", model.payrollStatements[0].DocumentNumber, statement.DocumentNumber)
	}
}

func TestSavePendingStockImportSkipsDuplicates(t *testing.T) {
	db, err := storage.Open(filepath.Join(t.TempDir(), "life.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.CreateTables(); err != nil {
		t.Fatal(err)
	}

	vest := models.StockVest{
		Symbol:                "GOOG",
		VestDate:              time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC),
		DepositDate:           time.Date(2026, 9, 29, 0, 0, 0, 0, time.UTC),
		GrossSharesMicros:     2 * models.ShareScale,
		NetSharesMicros:       models.ShareScale,
		WithheldSharesMicros:  models.ShareScale,
		GrossValueCents:       60000,
		NetValueCents:         30000,
		AcquisitionPriceCents: 30000,
	}
	model := Model{db: db, importKind: importStock, importMode: importConfirm, pendingStockVests: []models.StockVest{vest}}
	model.savePendingImport()
	model.importMode = importConfirm
	model.savePendingImport()

	if model.importMode != importResult || len(model.stockVests) != 1 {
		t.Fatalf("save result: mode=%v vests=%d error=%q", model.importMode, len(model.stockVests), model.importError)
	}
	if model.importMessage != "Imported 0 stock vest events; skipped 1 duplicates." {
		t.Fatalf("duplicate message = %q", model.importMessage)
	}
}
