package storage

func (d *Database) CreateTables() error {
	const accountsTable = `
	CREATE TABLE IF NOT EXISTS accounts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		balance REAL NOT NULL DEFAULT 0
	);
	`
	const payrollTable = `
		CREATE TABLE IF NOT EXISTS payroll_statements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			document_number TEXT NOT NULL UNIQUE,
			period_start TEXT NOT NULL,
			period_end TEXT NOT NULL,
			pay_date TEXT NOT NULL,
			gross_cents INTEGER NOT NULL,
			taxable_cents INTEGER NOT NULL,
			taxes_cents INTEGER NOT NULL,
			deductions_cents INTEGER NOT NULL,
			net_cents INTEGER NOT NULL,
			source_filename TEXT NOT NULL,
			imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_payroll_statements_pay_date
		ON payroll_statements(pay_date);
	`

	_, err := d.DB.Exec(accountsTable + payrollTable)
	return err
}
