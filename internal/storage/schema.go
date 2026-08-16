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
			employee_401k_cents INTEGER NOT NULL DEFAULT 0,
			employer_401k_cents INTEGER NOT NULL DEFAULT 0,
			employee_401k_ytd_cents INTEGER NOT NULL DEFAULT 0,
			employer_401k_ytd_cents INTEGER NOT NULL DEFAULT 0,
			source_filename TEXT NOT NULL,
			imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_payroll_statements_pay_date
		ON payroll_statements(pay_date);
	`
	const stockVestsTable = `
		CREATE TABLE IF NOT EXISTS stock_vests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			symbol TEXT NOT NULL,
			vest_date TEXT NOT NULL,
			deposit_date TEXT NOT NULL,
			acquisition_price_cents INTEGER NOT NULL,
			gross_shares_micros INTEGER NOT NULL,
			net_shares_micros INTEGER NOT NULL,
			withheld_shares_micros INTEGER NOT NULL,
			gross_value_cents INTEGER NOT NULL,
			net_value_cents INTEGER NOT NULL,
			imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(symbol, vest_date)
		);
	`

	if _, err := d.DB.Exec(accountsTable + payrollTable + stockVestsTable); err != nil {
		return err
	}
	if err := d.ensureColumn("payroll_statements", "employee_401k_cents", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := d.ensureColumn("payroll_statements", "employer_401k_cents", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := d.ensureColumn("payroll_statements", "employee_401k_ytd_cents", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	return d.ensureColumn("payroll_statements", "employer_401k_ytd_cents", "INTEGER NOT NULL DEFAULT 0")
}

func (d *Database) ensureColumn(table, column, definition string) error {
	rows, err := d.DB.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull, primaryKey int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = d.DB.Exec("ALTER TABLE " + table + " ADD COLUMN " + column + " " + definition)
	return err
}
