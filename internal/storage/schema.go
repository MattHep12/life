package storage

func (d *Database) CreateTables() error {
	const profilesTable = `
	CREATE TABLE IF NOT EXISTS profiles (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		name TEXT NOT NULL,
		birth_date TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	const accountsTable = `
	CREATE TABLE IF NOT EXISTS accounts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		balance REAL NOT NULL DEFAULT 0,
		category TEXT NOT NULL DEFAULT 'cash'
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
			base_pay_cents INTEGER NOT NULL DEFAULT 0,
			overtime_cents INTEGER NOT NULL DEFAULT 0,
			bonus_cents INTEGER NOT NULL DEFAULT 0,
			payroll_stock_cents INTEGER NOT NULL DEFAULT 0,
			other_earnings_cents INTEGER NOT NULL DEFAULT 0,
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
	const spendingTable = `
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
	`
	const recurringExpensesTable = `
		CREATE TABLE IF NOT EXISTS recurring_expenses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			amount_cents INTEGER NOT NULL,
			category TEXT NOT NULL
		);
	`
	const projectionSettingsTable = `
		CREATE TABLE IF NOT EXISTS projection_settings (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			monthly_spending_cents INTEGER NOT NULL,
			base_salary_cents INTEGER NOT NULL,
			bonus_rate REAL NOT NULL,
			real_income_growth_rate REAL NOT NULL,
			real_return_rate REAL NOT NULL,
			cash_real_return_rate REAL NOT NULL DEFAULT 0.015,
			net_pay_rate REAL NOT NULL,
			surplus_invested_rate REAL NOT NULL DEFAULT 1.0,
			monthly_net_stock_cents INTEGER NOT NULL,
			monthly_goal_stock_cents INTEGER NOT NULL DEFAULT 0,
			employee_401k_annual_cents INTEGER NOT NULL,
			employer_match_rate REAL NOT NULL,
			ira_annual_cents INTEGER NOT NULL,
			retirement_age INTEGER NOT NULL DEFAULT 65
		);
	`
	const financialGoalsTable = `
		CREATE TABLE IF NOT EXISTS financial_goals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			target_cents INTEGER NOT NULL,
			saved_cents INTEGER NOT NULL DEFAULT 0,
			planned_monthly_cents INTEGER NOT NULL DEFAULT 0,
			target_date TEXT NOT NULL
		);
	`

	if _, err := d.DB.Exec(profilesTable + accountsTable + payrollTable + stockVestsTable + spendingTable + recurringExpensesTable + projectionSettingsTable + financialGoalsTable); err != nil {
		return err
	}
	if err := d.ensureColumn("accounts", "category", "TEXT NOT NULL DEFAULT 'uncategorized'"); err != nil {
		return err
	}
	if _, err := d.DB.Exec(`UPDATE accounts SET category = CASE
		WHEN lower(name) LIKE '%401%' OR lower(name) LIKE '%roth%' OR lower(name) LIKE '%ira%' THEN 'retirement'
		WHEN lower(name) LIKE '%investment%' OR lower(name) LIKE '%brokerage%' OR lower(name) LIKE '%individual%' OR lower(name) LIKE '%equity%' OR lower(name) LIKE '%award%' OR lower(name) LIKE '%taxable%' OR lower(name) LIKE '%stock%' THEN 'investment'
		ELSE 'cash' END
		WHERE category = 'uncategorized'`); err != nil {
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
	if err := d.ensureColumn("payroll_statements", "employer_401k_ytd_cents", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	for _, column := range []string{"base_pay_cents", "overtime_cents", "bonus_cents", "payroll_stock_cents", "other_earnings_cents"} {
		if err := d.ensureColumn("payroll_statements", column, "INTEGER NOT NULL DEFAULT 0"); err != nil {
			return err
		}
	}
	if err := d.ensureColumn("projection_settings", "cash_real_return_rate", "REAL NOT NULL DEFAULT 0.015"); err != nil {
		return err
	}
	if err := d.ensureColumn("projection_settings", "retirement_age", "INTEGER NOT NULL DEFAULT 65"); err != nil {
		return err
	}
	if err := d.ensureColumn("projection_settings", "surplus_invested_rate", "REAL NOT NULL DEFAULT 1.0"); err != nil {
		return err
	}
	if err := d.ensureColumn("projection_settings", "monthly_goal_stock_cents", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	return nil
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
