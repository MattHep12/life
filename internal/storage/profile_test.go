package storage

import (
	"path/filepath"
	"testing"
	"time"

	"life/internal/models"
)

func TestProfileStartsEmptyAndRoundTrips(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "life.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.CreateTables(); err != nil {
		t.Fatal(err)
	}

	profile, err := db.GetProfile()
	if err != nil {
		t.Fatal(err)
	}
	if profile != nil {
		t.Fatalf("fresh database profile = %#v, want nil", profile)
	}

	birthDate := time.Date(1990, time.May, 20, 0, 0, 0, 0, time.Local)
	if err := db.SaveProfile(models.Profile{Name: "Example User", BirthDate: birthDate}); err != nil {
		t.Fatal(err)
	}
	profile, err = db.GetProfile()
	if err != nil {
		t.Fatal(err)
	}
	if profile == nil || profile.Name != "Example User" || !profile.BirthDate.Equal(birthDate) {
		t.Fatalf("stored profile = %#v", profile)
	}
}
