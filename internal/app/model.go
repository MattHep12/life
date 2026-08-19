package app

import (
	"sort"

	"life/internal/models"
	"life/internal/storage"

	"charm.land/bubbles/v2/textinput"
)

type Model struct {
	width  int
	height int

	cursor           int
	currentView      view
	choices          []menuItem
	profile          *models.Profile
	profileName      textinput.Model
	profileBirthDate textinput.Model
	profileSetupStep int
	profileError     string

	tasks      []models.Task
	taskCursor int

	accounts      []models.Account
	financeCursor int
	financeMode   financeMode

	accountName            textinput.Model
	accountBalance         textinput.Model
	accountCategoryCursor  int
	recurringExpenseCursor int
	financeError           string

	payrollStatements    []models.PayrollStatement
	payrollCursor        int
	payrollDeleteMode    bool
	payrollError         string
	stockVests           []models.StockVest
	spendingTransactions []models.SpendingTransaction
	spendingCursor       int
	recurringExpenses    []models.RecurringExpense
	projectionSettings   models.ProjectionSettings
	projectionCursor     int
	projectionError      string
	financialGoals       []models.FinancialGoal
	financialGoalCursor  int
	financialGoalError   string

	db *storage.Database

	habitCompletion       int
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
	profileSetupView
	dashboardView
	tasksView
	habitsView
	financesView
	payrollView
	spendingView
	projectionView
	financialGoalsView
)

type financeMode int

const (
	financeListMode financeMode = iota
	financeAddCategoryMode
	financeAddNameMode
	financeAddBalanceMode
	financeEditBalanceMode
	financeDeleteConfirmMode
	financeFixedBillsMode
	financeEditFixedBillMode
)

type menuItem struct {
	label       string
	description string
	icon        string
	view        view
}

func sortAccountsByCategory(accounts []models.Account) {
	sort.SliceStable(accounts, func(i, j int) bool {
		rank := func(account models.Account) int {
			switch accountCategory(account) {
			case financeCash:
				return 0
			case financeInvestment:
				return 1
			case financeRetirement:
				return 2
			default:
				return 3
			}
		}
		return rank(accounts[i]) < rank(accounts[j])
	})
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
	profile *models.Profile,
	accounts []models.Account,
	payrollStatements []models.PayrollStatement,
	stockVests []models.StockVest,
	spendingTransactions []models.SpendingTransaction,
	recurringExpenses []models.RecurringExpense,
	projectionSettings models.ProjectionSettings,
	financialGoals []models.FinancialGoal,
	monthlyNetIncomeCents int64,
) Model {
	sortAccountsByCategory(accounts)

	nameInput := textinput.New()
	nameInput.Placeholder = "Account name"
	nameInput.Prompt = "› "
	nameInput.SetWidth(42)

	balanceInput := textinput.New()
	balanceInput.Placeholder = "0.00"
	balanceInput.Prompt = "$ "
	balanceInput.SetWidth(42)

	profileNameInput := textinput.New()
	profileNameInput.Placeholder = "Your name"
	profileNameInput.Prompt = "› "
	profileNameInput.SetWidth(42)

	profileBirthDateInput := textinput.New()
	profileBirthDateInput.Placeholder = "YYYY-MM-DD"
	profileBirthDateInput.Prompt = "› "
	profileBirthDateInput.SetWidth(42)

	currentView := dashboardView
	if profile == nil {
		currentView = profileSetupView
		profileNameInput.Focus()
	} else {
		projectionSettings.BirthDate = profile.BirthDate
	}

	return Model{
		db:               db,
		cursor:           0,
		currentView:      currentView,
		profile:          profile,
		profileName:      profileNameInput,
		profileBirthDate: profileBirthDateInput,

		choices: []menuItem{
			{label: "Tasks & Habits", description: "Plan your day and build consistent routines", icon: "✓", view: tasksView},
			{label: "Finances", description: "Accounts, payroll, spending, and net worth", icon: "$", view: financesView},
			{label: "Goals & Projections", description: "Major purchases, investing, and future net worth", icon: "◎", view: projectionView},
		},

		tasks: []models.Task{},

		accounts:             accounts,
		accountName:          nameInput,
		accountBalance:       balanceInput,
		payrollStatements:    payrollStatements,
		stockVests:           stockVests,
		spendingTransactions: spendingTransactions,
		recurringExpenses:    recurringExpenses,
		projectionSettings:   projectionSettings,
		financialGoals:       financialGoals,

		habitCompletion:       0,
		monthlyNetIncomeCents: monthlyNetIncomeCents,
	}
}
