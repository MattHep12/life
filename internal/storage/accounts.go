package storage

import (
	"fmt"

	"life/internal/models"
)

func (d *Database) AddAccount(account models.Account) (int64, error) {
	result, err := d.DB.Exec(
		`INSERT INTO accounts (name, balance, category) VALUES (?, ?, ?)`,
		account.Name,
		account.Balance,
		account.Category,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *Database) GetAccounts() ([]models.Account, error) {
	rows, err := d.DB.Query(
		`SELECT id, name, balance, category FROM accounts ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account

	for rows.Next() {
		var account models.Account

		err := rows.Scan(
			&account.ID,
			&account.Name,
			&account.Balance,
			&account.Category,
		)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (d *Database) UpdateAccountBalance(id int64, balance float64) error {
	result, err := d.DB.Exec(
		`UPDATE accounts SET balance = ? WHERE id = ?`,
		balance,
		id,
	)

	if err != nil {
		return err
	}

	return requireAffectedAccount(result, id)
}

func (d *Database) DeleteAccount(id int64) error {
	result, err := d.DB.Exec(
		`DELETE FROM accounts WHERE id = ?`,
		id,
	)

	if err != nil {
		return err
	}

	return requireAffectedAccount(result, id)
}

func requireAffectedAccount(result interface{ RowsAffected() (int64, error) }, id int64) error {
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("account %d does not exist", id)
	}

	return nil
}
