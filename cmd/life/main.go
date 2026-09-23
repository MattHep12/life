package main

import (
	tea "charm.land/bubbletea/v2"
	"time"

	"life/internal/app"
	"life/internal/storage"
)

func main() {
	databasePath, err := storage.DefaultPath()
	if err != nil {
		panic(err)
	}
	db, err := storage.Open(databasePath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.CreateTables(); err != nil {
		panic(err)
	}
	profile, err := db.GetProfile()
	if err != nil {
		panic(err)
	}

	accounts, err := db.GetAccounts()
	if err != nil {
		panic(err)
	}
	now := time.Now()
	monthlyNetIncomeCents, err := db.NetPayrollForMonth(now.Year(), int(now.Month()))
	if err != nil {
		panic(err)
	}
	payrollStatements, err := db.GetPayrollStatements()
	if err != nil {
		panic(err)
	}
	stockVests, err := db.GetStockVests()
	if err != nil {
		panic(err)
	}
	spendingTransactions, err := db.GetSpendingTransactions()
	if err != nil {
		panic(err)
	}
	recurringExpenses, err := db.GetRecurringExpenses()
	if err != nil {
		panic(err)
	}
	projectionSettings, err := db.GetProjectionSettings()
	if err != nil {
		panic(err)
	}
	if profile != nil {
		projectionSettings.BirthDate = profile.BirthDate
	}
	financialGoals, err := db.GetFinancialGoals()
	if err != nil {
		panic(err)
	}

	m := app.NewModel(db, profile, accounts, payrollStatements, stockVests, spendingTransactions, recurringExpenses, projectionSettings, financialGoals, monthlyNetIncomeCents)

	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
