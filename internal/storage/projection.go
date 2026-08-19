package storage

import (
	"database/sql"
	"errors"

	"life/internal/models"
)

func (d *Database) GetProjectionSettings() (models.ProjectionSettings, error) {
	settings := models.DefaultProjectionSettings()
	err := d.DB.QueryRow(`SELECT monthly_spending_cents, base_salary_cents,
		bonus_rate, real_income_growth_rate, real_return_rate, cash_real_return_rate, net_pay_rate, surplus_invested_rate,
		monthly_net_stock_cents, employee_401k_annual_cents,
		monthly_goal_stock_cents,
		employer_match_rate, ira_annual_cents, retirement_age
		FROM projection_settings WHERE id = 1`).Scan(
		&settings.MonthlySpendingCents,
		&settings.BaseSalaryCents,
		&settings.BonusRate,
		&settings.RealIncomeGrowthRate,
		&settings.RealReturnRate,
		&settings.CashRealReturnRate,
		&settings.NetPayRate,
		&settings.SurplusInvestedRate,
		&settings.MonthlyNetStockCents,
		&settings.Employee401KAnnualCents,
		&settings.MonthlyGoalStockCents,
		&settings.EmployerMatchRate,
		&settings.IRAAnnualCents,
		&settings.RetirementAge,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return settings, d.SaveProjectionSettings(settings)
	}
	return settings, err
}

func (d *Database) SaveProjectionSettings(settings models.ProjectionSettings) error {
	_, err := d.DB.Exec(`INSERT INTO projection_settings (
		id, monthly_spending_cents, base_salary_cents, bonus_rate,
		real_income_growth_rate, real_return_rate, cash_real_return_rate, net_pay_rate, surplus_invested_rate,
		monthly_net_stock_cents, employee_401k_annual_cents, monthly_goal_stock_cents,
		employer_match_rate, ira_annual_cents, retirement_age
	) VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		monthly_spending_cents=excluded.monthly_spending_cents,
		base_salary_cents=excluded.base_salary_cents,
		bonus_rate=excluded.bonus_rate,
		real_income_growth_rate=excluded.real_income_growth_rate,
		real_return_rate=excluded.real_return_rate,
		cash_real_return_rate=excluded.cash_real_return_rate,
		net_pay_rate=excluded.net_pay_rate,
		surplus_invested_rate=excluded.surplus_invested_rate,
		monthly_net_stock_cents=excluded.monthly_net_stock_cents,
		employee_401k_annual_cents=excluded.employee_401k_annual_cents,
		monthly_goal_stock_cents=excluded.monthly_goal_stock_cents,
		employer_match_rate=excluded.employer_match_rate,
		ira_annual_cents=excluded.ira_annual_cents,
		retirement_age=excluded.retirement_age`,
		settings.MonthlySpendingCents,
		settings.BaseSalaryCents,
		settings.BonusRate,
		settings.RealIncomeGrowthRate,
		settings.RealReturnRate,
		settings.CashRealReturnRate,
		settings.NetPayRate,
		settings.SurplusInvestedRate,
		settings.MonthlyNetStockCents,
		settings.Employee401KAnnualCents,
		settings.MonthlyGoalStockCents,
		settings.EmployerMatchRate,
		settings.IRAAnnualCents,
		settings.RetirementAge,
	)
	return err
}
