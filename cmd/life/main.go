package main

import (
	tea "charm.land/bubbletea/v2"
	"time"

	"life/internal/app"
	"life/internal/storage"
)

func main() {
	db, err := storage.Open("life.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.CreateTables(); err != nil {
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

	m := app.NewModel(db, accounts, payrollStatements, monthlyNetIncomeCents)

	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
