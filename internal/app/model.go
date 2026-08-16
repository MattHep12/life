package app

import (
	"life/internal/models"
	"life/internal/storage"

	"charm.land/bubbles/v2/textinput"
)

type Model struct {
	cursor      int
	currentView view
	choices     []menuItem

	tasks      []models.Task
	taskCursor int

	accounts      []models.Account
	financeCursor int
	financeMode   financeMode

	accountName    textinput.Model
	accountBalance textinput.Model
	financeError   string

	payrollStatements []models.PayrollStatement
	payrollCursor     int

	db *storage.Database

	habitCompletion       int
	activeGoals           int
	monthlyNetIncomeCents int64
}

func (m *Model) resetAccountInputs() {
	m.accountName.Blur()
	m.accountBalance.Blur()
	m.accountName.SetValue("")
	m.accountBalance.SetValue("")
}

type view int

const (
	menuView view = iota
	dashboardView
	tasksView
	habitsView
	financesView
	payrollView
	statsView
)

type financeMode int

const (
	financeListMode financeMode = iota
	financeAddNameMode
	financeAddBalanceMode
	financeEditBalanceMode
	financeDeleteConfirmMode
)

type menuItem struct {
	label string
	view  view
}

func (m Model) netWorth() float64 {
	total := 0.0

	for _, account := range m.accounts {
		total += account.Balance
	}
	return total
}

func NewModel(
	db *storage.Database,
	accounts []models.Account,
	payrollStatements []models.PayrollStatement,
	monthlyNetIncomeCents int64,
) Model {
	nameInput := textinput.New()
	nameInput.Placeholder = "Account name"

	balanceInput := textinput.New()
	balanceInput.Placeholder = "Balance"

	return Model{
		db:          db,
		cursor:      0,
		currentView: menuView,

		choices: []menuItem{
			{label: "Dashboard View", view: dashboardView},
			{label: "Tasks View", view: tasksView},
			{label: "Habits View", view: habitsView},
			{label: "Finances View", view: financesView},
			{label: "Payroll History", view: payrollView},
			{label: "Stats View", view: statsView},
		},

		tasks: []models.Task{
			{Title: "Go to the gym"},
			{Title: "Eye Appointment 08/17 9:00AM"},
		},

		accounts:          accounts,
		accountName:       nameInput,
		accountBalance:    balanceInput,
		payrollStatements: payrollStatements,

		habitCompletion:       82,
		activeGoals:           3,
		monthlyNetIncomeCents: monthlyNetIncomeCents,
	}
}
