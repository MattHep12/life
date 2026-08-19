package models

import "time"

type FinancialGoal struct {
	ID                  int64
	Name                string
	TargetCents         int64
	SavedCents          int64
	PlannedMonthlyCents int64
	TargetDate          time.Time
}
