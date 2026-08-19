package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"life/internal/models"
)

var ErrSpendingAlreadyImported = errors.New("spending report already imported")

func (d *Database) AddSpendingTransactions(transactions []models.SpendingTransaction) (int, error) {
	if len(transactions) == 0 {
		return 0, nil
	}
	tx, err := d.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	inserted := 0
	for _, transaction := range transactions {
		_, err := tx.Exec(`INSERT INTO spending_transactions (
			transaction_date, post_date, description, amount_cents, category,
			source_filename, source_index
		) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			transaction.TransactionDate.Format("2006-01-02"),
			transaction.PostDate.Format("2006-01-02"),
			transaction.Description,
			transaction.AmountCents,
			transaction.Category,
			transaction.SourceFilename,
			transaction.SourceIndex,
		)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				continue
			}
			return 0, err
		}
		inserted++
	}
	if inserted == 0 {
		return 0, ErrSpendingAlreadyImported
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return inserted, nil
}

func (d *Database) GetSpendingTransactions() ([]models.SpendingTransaction, error) {
	rows, err := d.DB.Query(`SELECT id, transaction_date, post_date, description,
		amount_cents, category, source_filename, source_index, imported_at
		FROM spending_transactions ORDER BY transaction_date DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.SpendingTransaction
	for rows.Next() {
		var transaction models.SpendingTransaction
		var transactionDate, postDate, importedAt string
		if err := rows.Scan(&transaction.ID, &transactionDate, &postDate,
			&transaction.Description, &transaction.AmountCents, &transaction.Category,
			&transaction.SourceFilename, &transaction.SourceIndex, &importedAt); err != nil {
			return nil, err
		}
		transaction.TransactionDate, err = time.Parse("2006-01-02", transactionDate)
		if err != nil {
			return nil, fmt.Errorf("parse transaction date: %w", err)
		}
		transaction.PostDate, err = time.Parse("2006-01-02", postDate)
		if err != nil {
			return nil, fmt.Errorf("parse post date: %w", err)
		}
		transaction.ImportedAt, err = time.Parse("2006-01-02 15:04:05", importedAt)
		if err != nil {
			return nil, fmt.Errorf("parse import time: %w", err)
		}
		transactions = append(transactions, transaction)
	}
	return transactions, rows.Err()
}

func (d *Database) SpendingTotalForMonth(year int, month time.Month) (int64, error) {
	start := fmt.Sprintf("%04d-%02d-01", year, month)
	next := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	var total sql.NullInt64
	err := d.DB.QueryRow(`SELECT SUM(amount_cents) FROM spending_transactions
		WHERE transaction_date >= ? AND transaction_date < ?
		AND category <> 'Payments and Credits'`, start, next).Scan(&total)
	return total.Int64, err
}
