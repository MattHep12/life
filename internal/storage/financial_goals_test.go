package storage

import (
	"path/filepath"
	"testing"
)

func TestFinancialGoalsStartEmptyAndPersistProgress(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "life.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.CreateTables(); err != nil {
		t.Fatal(err)
	}
	goals, err := db.GetFinancialGoals()
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 0 {
		t.Fatalf("fresh database should have no goals, got %d", len(goals))
	}
	if _, err := db.DB.Exec(`INSERT INTO financial_goals
		(name, target_cents, saved_cents, planned_monthly_cents, target_date)
		VALUES ('Example goal', 1000000, 0, 0, '2030-01-01')`); err != nil {
		t.Fatal(err)
	}
	goals, err = db.GetFinancialGoals()
	if err != nil {
		t.Fatal(err)
	}
	if len(goals) != 1 {
		t.Fatalf("goal count = %d, want 1", len(goals))
	}
	if err := db.UpdateFinancialGoalProgress(goals[0].ID, 500000, 100000); err != nil {
		t.Fatal(err)
	}
	goals, err = db.GetFinancialGoals()
	if err != nil {
		t.Fatal(err)
	}
	if goals[0].SavedCents != 500000 || goals[0].PlannedMonthlyCents != 100000 {
		t.Fatalf("updated goal = %#v", goals[0])
	}
}
