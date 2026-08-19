package storage

import (
	"path/filepath"
	"testing"

	"life/internal/models"
)

func TestProjectionSettingsRoundTrip(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "life.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.CreateTables(); err != nil {
		t.Fatal(err)
	}

	settings, err := db.GetProjectionSettings()
	if err != nil {
		t.Fatal(err)
	}
	settings.MonthlySpendingCents = 345000
	settings.RealReturnRate = 0.065
	settings.MonthlyGoalStockCents = 80000
	if err := db.SaveProjectionSettings(settings); err != nil {
		t.Fatal(err)
	}

	stored, err := db.GetProjectionSettings()
	if err != nil {
		t.Fatal(err)
	}
	if stored.MonthlySpendingCents != 345000 || stored.RealReturnRate != 0.065 || stored.MonthlyGoalStockCents != 80000 {
		t.Fatalf("stored settings = %#v", stored)
	}
	if stored.BirthDate.Year() != models.DefaultProjectionSettings().BirthDate.Year() {
		t.Fatalf("birth date was not preserved: %v", stored.BirthDate)
	}
}
