package models

import "time"

const ShareScale int64 = 1_000_000

type StockVest struct {
	ID                    int64
	Symbol                string
	VestDate              time.Time
	DepositDate           time.Time
	AcquisitionPriceCents int64
	GrossSharesMicros     int64
	NetSharesMicros       int64
	WithheldSharesMicros  int64
	GrossValueCents       int64
	NetValueCents         int64
}
