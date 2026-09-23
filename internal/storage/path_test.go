package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPathHonorsEnvironmentOverride(t *testing.T) {
	wanted := filepath.Join(t.TempDir(), "custom.db")
	t.Setenv(databaseEnvironmentVariable, wanted)

	got, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if got != wanted {
		t.Fatalf("DefaultPath() = %q, want %q", got, wanted)
	}
}

func TestDefaultPathKeepsExistingWorkingDirectoryDatabase(t *testing.T) {
	t.Setenv(databaseEnvironmentVariable, "")
	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWorkingDirectory) })

	workingDirectory := t.TempDir()
	if err := os.Chdir(workingDirectory); err != nil {
		t.Fatal(err)
	}
	resolvedWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	wanted := filepath.Join(resolvedWorkingDirectory, "life.db")
	if err := os.WriteFile(wanted, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if got != wanted {
		t.Fatalf("DefaultPath() = %q, want %q", got, wanted)
	}
}
