package app

import (
	"testing"

	"life/internal/models"
)

func TestAccountCategoryUsesStoredCategoryNotName(t *testing.T) {
	account := models.Account{Name: "Anything at all", Category: "investment"}
	if got := accountCategory(account); got != financeInvestment {
		t.Fatalf("accountCategory(%q) = %q, want %q", account.Name, got, financeInvestment)
	}
}

func TestSortAccountsByCategory(t *testing.T) {
	accounts := []models.Account{
		{ID: 1, Name: "Retirement", Category: "retirement"},
		{ID: 2, Name: "Cash", Category: "cash"},
		{ID: 3, Name: "Investment", Category: "investment"},
	}
	sortAccountsByCategory(accounts)
	if accounts[0].ID != 2 || accounts[1].ID != 3 || accounts[2].ID != 1 {
		t.Fatalf("unexpected category order: %#v", accounts)
	}
}
