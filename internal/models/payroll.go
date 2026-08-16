package models

import "time"

type PayrollStatement struct {
	ID              int64
	DocumentNumber  string
	PeriodStart     time.Time
	PeriodEnd       time.Time
	PayDate         time.Time
	GrossCents      int64
	TaxableCents    int64
	TaxesCents      int64
	DeductionsCents int64
	NetCents        int64
	SourceFilename  string
	ImportedAt      time.Time
}
