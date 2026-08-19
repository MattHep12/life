package app

import (
	"path/filepath"
	"testing"
	"time"

	"life/internal/models"
	"life/internal/storage"
)

func TestNewModelStartsInProfileSetupForFreshDatabase(t *testing.T) {
	db, err := storage.Open(filepath.Join(t.TempDir(), "life.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	settings := models.DefaultProjectionSettings()
	model := NewModel(db, nil, nil, nil, nil, nil, nil, settings, nil, 0)
	if model.currentView != profileSetupView {
		t.Fatalf("fresh model view = %v, want profile setup", model.currentView)
	}
}

func TestNewModelUsesProfileBirthDate(t *testing.T) {
	db, err := storage.Open(filepath.Join(t.TempDir(), "life.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	birthDate := time.Date(1988, time.April, 3, 0, 0, 0, 0, time.Local)
	profile := &models.Profile{ID: 1, Name: "Example User", BirthDate: birthDate}
	model := NewModel(db, profile, nil, nil, nil, nil, nil, models.DefaultProjectionSettings(), nil, 0)
	if model.currentView != dashboardView || !model.projectionSettings.BirthDate.Equal(birthDate) {
		t.Fatalf("profile was not applied: view=%v birth=%v", model.currentView, model.projectionSettings.BirthDate)
	}
}
