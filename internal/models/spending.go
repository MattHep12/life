package models

import "time"

type SpendingTransaction struct {
	ID              int64
	TransactionDate time.Time
	PostDate        time.Time
	Description     string
	AmountCents     int64
	Category        string
	SourceFilename  string
	SourceIndex     int
	ImportedAt      time.Time
}

func (t SpendingTransaction) IsPayment() bool {
	return t.Category == "Payments and Credits"
}
