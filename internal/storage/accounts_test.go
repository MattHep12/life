package storage

import (
	"path/filepath"
	"testing"

	"life/internal/models"
)

func TestAccountCRUD(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "life.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.CreateTables(); err != nil {
		t.Fatal(err)
	}

	id, err := db.AddAccount(models.Account{Name: "Checking", Balance: 100.25})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.UpdateAccountBalance(id, 250.75); err != nil {
		t.Fatal(err)
	}

	accounts, err := db.GetAccounts()
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 || accounts[0].ID != id || accounts[0].Balance != 250.75 {
		t.Fatalf("unexpected accounts after update: %#v", accounts)
	}

	if err := db.DeleteAccount(id); err != nil {
		t.Fatal(err)
	}
	if err := db.DeleteAccount(id); err == nil {
		t.Fatal("deleting a missing account should fail")
	}
}
