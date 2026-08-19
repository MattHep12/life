package storage

import (
	"time"

	"life/internal/models"
)

func (d *Database) GetFinancialGoals() ([]models.FinancialGoal, error) {
	rows, err := d.DB.Query(`SELECT id, name, target_cents, saved_cents,
		planned_monthly_cents, target_date FROM financial_goals ORDER BY target_date, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []models.FinancialGoal
	for rows.Next() {
		var goal models.FinancialGoal
		var targetDate string
		if err := rows.Scan(&goal.ID, &goal.Name, &goal.TargetCents, &goal.SavedCents, &goal.PlannedMonthlyCents, &targetDate); err != nil {
			return nil, err
		}
		goal.TargetDate, err = time.ParseInLocation("2006-01-02", targetDate, time.Local)
		if err != nil {
			return nil, err
		}
		goals = append(goals, goal)
	}
	return goals, rows.Err()
}

func (d *Database) UpdateFinancialGoalProgress(id, savedCents, plannedMonthlyCents int64) error {
	_, err := d.DB.Exec(`UPDATE financial_goals SET saved_cents=?, planned_monthly_cents=? WHERE id=?`, savedCents, plannedMonthlyCents, id)
	return err
}
