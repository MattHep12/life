package models

type RecurringExpense struct {
	ID          int64
	Name        string
	AmountCents int64
	Category    string
}
