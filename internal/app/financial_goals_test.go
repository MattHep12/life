package app

import (
	"testing"
	"time"

	"life/internal/models"
)

func TestCombinedGoalMonthlyRequirementSharesContributionsAcrossDeadlines(t *testing.T) {
	now := time.Date(2026, time.August, 18, 0, 0, 0, 0, time.Local)
	goals := []models.FinancialGoal{
		{
			Name:        "House",
			TargetCents: 7000000,
			SavedCents:  1650000,
			TargetDate:  time.Date(2027, time.August, 18, 0, 0, 0, 0, time.Local),
		},
		{
			Name:        "Car",
			TargetCents: 3000000,
			TargetDate:  time.Date(2028, time.March, 31, 0, 0, 0, 0, time.Local),
		},
	}

	got := combinedGoalMonthlyRequirement(goals, now)
	if got < 445000 || got > 447000 {
		t.Fatalf("combined monthly requirement = %d cents, want about $4,458", got)
	}
}
