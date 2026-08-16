package app

import (
	"life/internal/models"
	"life/internal/storage"

	"charm.land/bubbles/v2/textinput"
)

type Model struct {
	width  int
	height int

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
	payrollDeleteMode bool
	payrollError      string
	stockVests        []models.StockVest

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
	label       string
	description string
	icon        string
	view        view
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
	stockVests []models.StockVest,
	monthlyNetIncomeCents int64,
) Model {
	nameInput := textinput.New()
	nameInput.Placeholder = "Account name"
	nameInput.Prompt = "› "
	nameInput.SetWidth(42)

	balanceInput := textinput.New()
	balanceInput.Placeholder = "0.00"
	balanceInput.Prompt = "$ "
	balanceInput.SetWidth(42)

	return Model{
		db:          db,
		cursor:      0,
		currentView: menuView,

		choices: []menuItem{
			{label: "Dashboard", description: "Your life at a glance", icon: "◆", view: dashboardView},
			{label: "Tasks", description: "Plan and complete your day", icon: "✓", view: tasksView},
			{label: "Habits", description: "Build consistent routines", icon: "↻", view: habitsView},
			{label: "Finances", description: "Accounts and net worth", icon: "$", view: financesView},
			{label: "Payroll History", description: "Income, taxes, and deductions", icon: "▤", view: payrollView},
			{label: "Stats", description: "Trends across your life", icon: "↗", view: statsView},
		},

		tasks: []models.Task{
			{Title: "Go to the gym"},
			{Title: "Eye Appointment 08/17 9:00AM"},
		},

		accounts:          accounts,
		accountName:       nameInput,
		accountBalance:    balanceInput,
		payrollStatements: payrollStatements,
		stockVests:        stockVests,

		habitCompletion:       82,
		activeGoals:           3,
		monthlyNetIncomeCents: monthlyNetIncomeCents,
	}
}
