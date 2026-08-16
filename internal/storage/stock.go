package storage

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"life/internal/models"
)

var ErrStockVestAlreadyImported = errors.New("stock vest already imported")

func (d *Database) AddStockVest(vest models.StockVest) (int64, error) {
	result, err := d.DB.Exec(`
		INSERT INTO stock_vests (
			symbol, vest_date, deposit_date, acquisition_price_cents,
			gross_shares_micros, net_shares_micros, withheld_shares_micros,
			gross_value_cents, net_value_cents
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, vest.Symbol, vest.VestDate.Format("2006-01-02"), vest.DepositDate.Format("2006-01-02"), vest.AcquisitionPriceCents,
		vest.GrossSharesMicros, vest.NetSharesMicros, vest.WithheldSharesMicros, vest.GrossValueCents, vest.NetValueCents)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return 0, ErrStockVestAlreadyImported
		}
		return 0, err
	}
	return result.LastInsertId()
}

func (d *Database) GetStockVests() ([]models.StockVest, error) {
	rows, err := d.DB.Query(`
		SELECT id, symbol, vest_date, deposit_date, acquisition_price_cents,
			gross_shares_micros, net_shares_micros, withheld_shares_micros,
			gross_value_cents, net_value_cents
		FROM stock_vests ORDER BY vest_date DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vests []models.StockVest
	for rows.Next() {
		var vest models.StockVest
		var vestDate, depositDate string
		if err := rows.Scan(&vest.ID, &vest.Symbol, &vestDate, &depositDate, &vest.AcquisitionPriceCents,
			&vest.GrossSharesMicros, &vest.NetSharesMicros, &vest.WithheldSharesMicros,
			&vest.GrossValueCents, &vest.NetValueCents); err != nil {
			return nil, err
		}
		vest.VestDate, err = time.Parse("2006-01-02", vestDate)
		if err != nil {
			return nil, fmt.Errorf("parse stored vest date: %w", err)
		}
		vest.DepositDate, err = time.Parse("2006-01-02", depositDate)
		if err != nil {
			return nil, fmt.Errorf("parse stored deposit date: %w", err)
		}
		vests = append(vests, vest)
	}
	return vests, rows.Err()
}
