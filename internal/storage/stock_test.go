package storage

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"life/internal/models"
)

func TestStockVestStorage(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "life.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.CreateTables(); err != nil {
		t.Fatal(err)
	}

	vest := models.StockVest{Symbol: "GOOG", VestDate: time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC), DepositDate: time.Date(2026, 7, 29, 0, 0, 0, 0, time.UTC), AcquisitionPriceCents: 31909, GrossSharesMicros: 8_056_000, NetSharesMicros: 5_267_000, WithheldSharesMicros: 2_789_000, GrossValueCents: 257059, NetValueCents: 168065}
	if _, err := db.AddStockVest(vest); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AddStockVest(vest); !errors.Is(err, ErrStockVestAlreadyImported) {
		t.Fatalf("duplicate error = %v", err)
	}
	vests, err := db.GetStockVests()
	if err != nil {
		t.Fatal(err)
	}
	if len(vests) != 1 || vests[0].GrossValueCents != vest.GrossValueCents {
		t.Fatalf("unexpected stock vests: %#v", vests)
	}
}
