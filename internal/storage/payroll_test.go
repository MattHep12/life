package storage

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"life/internal/models"
)

func TestPayrollImportAndMonthlyTotal(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "life.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.CreateTables(); err != nil {
		t.Fatal(err)
	}

	statement := models.PayrollStatement{
		DocumentNumber:       "12345678",
		PeriodStart:          time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC),
		PeriodEnd:            time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC),
		PayDate:              time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC),
		GrossCents:           477732,
		TaxableCents:         477142,
		TaxesCents:           126888,
		DeductionsCents:      1205,
		NetCents:             349639,
		Employee401KCents:    47304,
		Employer401KCents:    23652,
		Employee401KYTDCents: 976198,
		Employer401KYTDCents: 488097,
		SourceFilename:       "pay-statement.pdf",
	}
	id, err := db.AddPayrollStatement(statement)
	if err != nil {
		t.Fatal(err)
	}
	statement.ID = id
	statements, err := db.GetPayrollStatements()
	if err != nil {
		t.Fatal(err)
	}
	if len(statements) != 1 || statements[0].DocumentNumber != statement.DocumentNumber || statements[0].NetCents != statement.NetCents || statements[0].Employee401KCents != statement.Employee401KCents || statements[0].Employee401KYTDCents != statement.Employee401KYTDCents {
		t.Fatalf("unexpected stored payroll statements: %#v", statements)
	}
	if _, err := db.AddPayrollStatement(statement); !errors.Is(err, ErrPayrollAlreadyImported) {
		t.Fatalf("duplicate import error = %v", err)
	}
	statement.Employee401KCents = 50000
	if err := db.UpdatePayrollStatement(statement); err != nil {
		t.Fatal(err)
	}
	statements, err = db.GetPayrollStatements()
	if err != nil {
		t.Fatal(err)
	}
	if statements[0].Employee401KCents != 50000 {
		t.Fatalf("updated employee 401(k) = %d", statements[0].Employee401KCents)
	}

	total, err := db.NetPayrollForMonth(2026, 8)
	if err != nil {
		t.Fatal(err)
	}
	if total != statement.NetCents {
		t.Fatalf("monthly net = %d, want %d", total, statement.NetCents)
	}

	if err := db.DeletePayrollStatement(statement.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.DeletePayrollStatement(statement.ID); err == nil {
		t.Fatal("deleting a missing payroll statement should fail")
	}
}
