package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"life/internal/models"
)

var ErrPayrollAlreadyImported = errors.New("payroll statement already imported")

func (d *Database) AddPayrollStatement(statement models.PayrollStatement) (int64, error) {
	result, err := d.DB.Exec(
		`INSERT INTO payroll_statements (
			document_number, period_start, period_end, pay_date,
			gross_cents, taxable_cents, taxes_cents, deductions_cents,
			net_cents, employee_401k_cents, employer_401k_cents,
			employee_401k_ytd_cents, employer_401k_ytd_cents, source_filename
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		statement.DocumentNumber,
		statement.PeriodStart.Format("2006-01-02"),
		statement.PeriodEnd.Format("2006-01-02"),
		statement.PayDate.Format("2006-01-02"),
		statement.GrossCents,
		statement.TaxableCents,
		statement.TaxesCents,
		statement.DeductionsCents,
		statement.NetCents,
		statement.Employee401KCents,
		statement.Employer401KCents,
		statement.Employee401KYTDCents,
		statement.Employer401KYTDCents,
		statement.SourceFilename,
	)
	if err != nil {
		if isUniqueConstraint(err) {
			return 0, ErrPayrollAlreadyImported
		}
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (d *Database) UpdatePayrollStatement(statement models.PayrollStatement) error {
	result, err := d.DB.Exec(`
		UPDATE payroll_statements SET
			period_start = ?, period_end = ?, pay_date = ?, gross_cents = ?,
			taxable_cents = ?, taxes_cents = ?, deductions_cents = ?, net_cents = ?,
			employee_401k_cents = ?, employer_401k_cents = ?,
			employee_401k_ytd_cents = ?, employer_401k_ytd_cents = ?, source_filename = ?
		WHERE document_number = ?`,
		statement.PeriodStart.Format("2006-01-02"),
		statement.PeriodEnd.Format("2006-01-02"),
		statement.PayDate.Format("2006-01-02"),
		statement.GrossCents,
		statement.TaxableCents,
		statement.TaxesCents,
		statement.DeductionsCents,
		statement.NetCents,
		statement.Employee401KCents,
		statement.Employer401KCents,
		statement.Employee401KYTDCents,
		statement.Employer401KYTDCents,
		statement.SourceFilename,
		statement.DocumentNumber,
	)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("payroll statement %s does not exist", statement.DocumentNumber)
	}
	return nil
}

func (d *Database) GetPayrollStatements() ([]models.PayrollStatement, error) {
	rows, err := d.DB.Query(`
		SELECT id, document_number, period_start, period_end, pay_date,
			gross_cents, taxable_cents, taxes_cents, deductions_cents,
			net_cents, employee_401k_cents, employer_401k_cents,
			employee_401k_ytd_cents, employer_401k_ytd_cents,
			source_filename, imported_at
		FROM payroll_statements
		ORDER BY pay_date DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statements []models.PayrollStatement
	for rows.Next() {
		var statement models.PayrollStatement
		var periodStart, periodEnd, payDate, importedAt string

		if err := rows.Scan(
			&statement.ID,
			&statement.DocumentNumber,
			&periodStart,
			&periodEnd,
			&payDate,
			&statement.GrossCents,
			&statement.TaxableCents,
			&statement.TaxesCents,
			&statement.DeductionsCents,
			&statement.NetCents,
			&statement.Employee401KCents,
			&statement.Employer401KCents,
			&statement.Employee401KYTDCents,
			&statement.Employer401KYTDCents,
			&statement.SourceFilename,
			&importedAt,
		); err != nil {
			return nil, err
		}

		statement.PeriodStart, err = time.Parse("2006-01-02", periodStart)
		if err != nil {
			return nil, fmt.Errorf("parse stored period start: %w", err)
		}
		statement.PeriodEnd, err = time.Parse("2006-01-02", periodEnd)
		if err != nil {
			return nil, fmt.Errorf("parse stored period end: %w", err)
		}
		statement.PayDate, err = time.Parse("2006-01-02", payDate)
		if err != nil {
			return nil, fmt.Errorf("parse stored pay date: %w", err)
		}
		statement.ImportedAt, err = time.Parse("2006-01-02 15:04:05", importedAt)
		if err != nil {
			return nil, fmt.Errorf("parse stored import time: %w", err)
		}

		statements = append(statements, statement)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return statements, nil
}

func (d *Database) DeletePayrollStatement(id int64) error {
	result, err := d.DB.Exec(`DELETE FROM payroll_statements WHERE id = ?`, id)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("payroll statement %d does not exist", id)
	}
	return nil
}

func (d *Database) NetPayrollForMonth(year int, month int) (int64, error) {
	monthStart := fmt.Sprintf("%04d-%02d-01", year, month)
	nextYear, nextMonth := year, month+1
	if nextMonth == 13 {
		nextYear++
		nextMonth = 1
	}
	nextMonthStart := fmt.Sprintf("%04d-%02d-01", nextYear, nextMonth)

	var total sql.NullInt64
	err := d.DB.QueryRow(
		`SELECT SUM(net_cents) FROM payroll_statements WHERE pay_date >= ? AND pay_date < ?`,
		monthStart,
		nextMonthStart,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total.Int64, nil
}

func isUniqueConstraint(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
