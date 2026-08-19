package models

import "time"

type ProjectionSettings struct {
	BirthDate               time.Time
	MonthlySpendingCents    int64
	BaseSalaryCents         int64
	BonusRate               float64
	RealIncomeGrowthRate    float64
	RealReturnRate          float64
	CashRealReturnRate      float64
	NetPayRate              float64
	SurplusInvestedRate     float64
	MonthlyNetStockCents    int64
	MonthlyGoalStockCents   int64
	Employee401KAnnualCents int64
	EmployerMatchRate       float64
	IRAAnnualCents          int64
	RetirementAge           int
}

func DefaultProjectionSettings() ProjectionSettings {
	return ProjectionSettings{
		BirthDate:               time.Date(1990, time.January, 1, 0, 0, 0, 0, time.Local),
		MonthlySpendingCents:    0,
		BaseSalaryCents:         0,
		BonusRate:               0,
		RealIncomeGrowthRate:    0.02,
		RealReturnRate:          0.05,
		CashRealReturnRate:      0.01,
		NetPayRate:              0.75,
		SurplusInvestedRate:     1.00,
		MonthlyNetStockCents:    0,
		MonthlyGoalStockCents:   0,
		Employee401KAnnualCents: 0,
		EmployerMatchRate:       0,
		IRAAnnualCents:          0,
		RetirementAge:           65,
	}
}
