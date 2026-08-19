package storage

import "life/internal/models"

func (d *Database) GetRecurringExpenses() ([]models.RecurringExpense, error) {
	rows, err := d.DB.Query(`SELECT id, name, amount_cents, category
		FROM recurring_expenses ORDER BY amount_cents DESC, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.RecurringExpense
	for rows.Next() {
		var expense models.RecurringExpense
		if err := rows.Scan(&expense.ID, &expense.Name, &expense.AmountCents, &expense.Category); err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}
	return expenses, rows.Err()
}

func (d *Database) UpsertRecurringExpense(expense models.RecurringExpense) error {
	_, err := d.DB.Exec(`INSERT INTO recurring_expenses (name, amount_cents, category)
		VALUES (?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET amount_cents=excluded.amount_cents,
		category=excluded.category`, expense.Name, expense.AmountCents, expense.Category)
	return err
}
