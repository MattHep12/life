package app

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"life/internal/models"
)

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		inputWidth := min(48, max(20, msg.Width-12))
		m.accountName.SetWidth(inputWidth)
		m.accountBalance.SetWidth(inputWidth)
		m.profileName.SetWidth(inputWidth)
		m.profileBirthDate.SetWidth(inputWidth)
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		if msg.String() == "esc" {
			if m.currentView == financesView && m.financeMode != financeListMode {
				m.financeMode = financeListMode
				m.financeError = ""
				m.resetAccountInputs()
				return m, nil
			}
			if m.currentView == payrollView && m.payrollDeleteMode {
				m.payrollDeleteMode = false
				m.payrollError = ""
				return m, nil
			}

			return m, tea.Quit
		}

		switch m.currentView {
		case profileSetupView:
			switch m.profileSetupStep {
			case 0:
				if msg.String() == "enter" {
					name := strings.TrimSpace(m.profileName.Value())
					if name == "" {
						m.profileError = "Enter your name to continue."
						return m, nil
					}
					m.profileName.SetValue(name)
					m.profileName.Blur()
					m.profileSetupStep = 1
					m.profileError = ""
					cmd := m.profileBirthDate.Focus()
					return m, cmd
				}
				var cmd tea.Cmd
				m.profileName, cmd = m.profileName.Update(msg)
				return m, cmd
			case 1:
				if msg.String() == "enter" {
					birthDate, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(m.profileBirthDate.Value()), time.Local)
					if err != nil || birthDate.After(time.Now()) {
						m.profileError = "Enter a valid birth date as YYYY-MM-DD."
						return m, nil
					}
					profile := models.Profile{Name: m.profileName.Value(), BirthDate: birthDate}
					if err := m.db.SaveProfile(profile); err != nil {
						m.profileError = fmt.Sprintf("Could not save profile: %v", err)
						return m, nil
					}
					profile.ID = 1
					m.profile = &profile
					m.projectionSettings.BirthDate = birthDate
					m.profileBirthDate.Blur()
					m.profileError = ""
					m.currentView = dashboardView
					return m, nil
				}
				var cmd tea.Cmd
				m.profileBirthDate, cmd = m.profileBirthDate.Update(msg)
				return m, cmd
			}

		case menuView:
			switch msg.String() {
			case "enter":
				m.currentView = m.choices[m.cursor].view

			case "up", "w":
				if m.cursor > 0 {
					m.cursor--
				}

			case "down", "s":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			}

		case dashboardView:
			switch msg.String() {
			case "enter":
				if len(m.choices) > 0 {
					m.currentView = m.choices[m.cursor].view
				}
			case "up", "w":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "s":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			}

		case tasksView:
			switch msg.String() {
			case "q":
				m.currentView = dashboardView
			case "tab", "shift+tab":
				m.currentView = habitsView

			case "enter", "space":
				if len(m.tasks) > 0 {
					m.tasks[m.taskCursor].Completed =
						!m.tasks[m.taskCursor].Completed
				}

			case "up", "w":
				if m.taskCursor > 0 {
					m.taskCursor--
				}

			case "down", "s":
				if m.taskCursor < len(m.tasks)-1 {
					m.taskCursor++
				}
			}

		case habitsView:
			switch msg.String() {
			case "q":
				m.currentView = dashboardView
			case "tab", "shift+tab":
				m.currentView = tasksView
			}

		case financesView:
			switch m.financeMode {

			case financeListMode:
				switch msg.String() {
				case "q":
					m.currentView = dashboardView
				case "tab":
					m.currentView = payrollView
				case "shift+tab":
					m.currentView = spendingView

				case "a":
					m.financeError = ""
					m.resetAccountInputs()
					m.accountCategoryCursor = 0
					m.financeMode = financeAddCategoryMode
					return m, nil

				case "e":
					if len(m.accounts) > 0 {
						m.financeError = ""
						selected := m.accounts[m.financeCursor]

						m.accountBalance.SetValue(
							fmt.Sprintf("%.2f", selected.Balance),
						)

						m.financeMode = financeEditBalanceMode

						cmd := m.accountBalance.Focus()
						return m, cmd
					}

				case "d":
					if len(m.accounts) > 0 {
						m.financeError = ""
						m.financeMode = financeDeleteConfirmMode
					}
				case "f":
					m.financeError = ""
					m.recurringExpenseCursor = min(m.recurringExpenseCursor, max(0, len(m.recurringExpenses)-1))
					m.financeMode = financeFixedBillsMode

				case "up", "w":
					if m.financeCursor > 0 {
						m.financeCursor--
					}

				case "down", "s":
					if m.financeCursor < len(m.accounts)-1 {
						m.financeCursor++
					}
				}

			case financeAddCategoryMode:
				switch msg.String() {
				case "up", "left", "w", "h":
					if m.accountCategoryCursor > 0 {
						m.accountCategoryCursor--
					}
				case "down", "right", "s", "l":
					if m.accountCategoryCursor < len(financeCategories)-1 {
						m.accountCategoryCursor++
					}
				case "enter":
					m.financeMode = financeAddNameMode
					cmd := m.accountName.Focus()
					return m, cmd
				}

			case financeAddNameMode:
				switch msg.String() {
				case "enter":
					name := strings.TrimSpace(m.accountName.Value())
					if name != "" {
						m.accountName.SetValue(name)
						m.financeError = ""
						m.financeMode = financeAddBalanceMode
						m.accountName.Blur()

						cmd := m.accountBalance.Focus()
						return m, cmd
					}
					m.financeError = "Account name cannot be empty."
					return m, nil
				}

				var cmd tea.Cmd
				m.accountName, cmd = m.accountName.Update(msg)
				return m, cmd

			case financeAddBalanceMode:
				switch msg.String() {
				case "enter":
					balanceText := strings.TrimSpace(m.accountBalance.Value())

					balance, err := strconv.ParseFloat(balanceText, 64)
					if err != nil || math.IsNaN(balance) || math.IsInf(balance, 0) {
						m.financeError = "Enter a valid, finite balance."
						return m, nil
					}

					account := models.Account{
						Name:     strings.TrimSpace(m.accountName.Value()),
						Balance:  balance,
						Category: string(financeCategories[m.accountCategoryCursor]),
					}

					id, err := m.db.AddAccount(account)
					if err != nil {
						m.financeError = fmt.Sprintf("Could not save account: %v", err)
						return m, nil
					}

					account.ID = id

					m.accounts = append(m.accounts, account)
					sortAccountsByCategory(m.accounts)
					for i := range m.accounts {
						if m.accounts[i].ID == id {
							m.financeCursor = i
							break
						}
					}

					m.financeError = ""
					m.resetAccountInputs()
					m.financeMode = financeListMode

					return m, nil
				}

				var cmd tea.Cmd
				m.accountBalance, cmd = m.accountBalance.Update(msg)
				return m, cmd

			case financeEditBalanceMode:
				switch msg.String() {
				case "enter":
					if len(m.accounts) == 0 {
						return m, nil
					}

					balanceText := strings.TrimSpace(
						m.accountBalance.Value(),
					)

					balance, err := strconv.ParseFloat(balanceText, 64)
					if err != nil || math.IsNaN(balance) || math.IsInf(balance, 0) {
						m.financeError = "Enter a valid, finite balance."
						return m, nil
					}

					account := &m.accounts[m.financeCursor]

					if err := m.db.UpdateAccountBalance(
						account.ID,
						balance,
					); err != nil {
						m.financeError = fmt.Sprintf("Could not update account: %v", err)
						return m, nil
					}

					account.Balance = balance

					m.financeError = ""
					m.resetAccountInputs()
					m.financeMode = financeListMode

					return m, nil
				}

				var cmd tea.Cmd
				m.accountBalance, cmd = m.accountBalance.Update(msg)
				return m, cmd

			case financeDeleteConfirmMode:
				switch msg.String() {
				case "n", "q":
					m.financeMode = financeListMode
					return m, nil

				case "y":
					if len(m.accounts) == 0 {
						m.financeMode = financeListMode
						m.financeError = ""
						return m, nil
					}

					account := m.accounts[m.financeCursor]

					if err := m.db.DeleteAccount(account.ID); err != nil {
						m.financeError = fmt.Sprintf("Could not delete account: %v", err)
						return m, nil
					}

					m.accounts = append(
						m.accounts[:m.financeCursor],
						m.accounts[m.financeCursor+1:]...,
					)

					if m.financeCursor >= len(m.accounts) && m.financeCursor > 0 {
						m.financeCursor--
					}

					m.financeError = ""
					m.financeMode = financeListMode
					return m, nil
				}

			case financeFixedBillsMode:
				switch msg.String() {
				case "q":
					m.financeMode = financeListMode
				case "up", "w":
					if m.recurringExpenseCursor > 0 {
						m.recurringExpenseCursor--
					}
				case "down", "s":
					if m.recurringExpenseCursor < len(m.recurringExpenses)-1 {
						m.recurringExpenseCursor++
					}
				case "enter", "e":
					if len(m.recurringExpenses) > 0 {
						expense := m.recurringExpenses[m.recurringExpenseCursor]
						m.accountBalance.SetValue(fmt.Sprintf("%.2f", float64(expense.AmountCents)/100))
						m.financeMode = financeEditFixedBillMode
						cmd := m.accountBalance.Focus()
						return m, cmd
					}
				}

			case financeEditFixedBillMode:
				switch msg.String() {
				case "enter":
					if len(m.recurringExpenses) == 0 {
						m.financeMode = financeFixedBillsMode
						return m, nil
					}
					amount, err := strconv.ParseFloat(strings.TrimSpace(m.accountBalance.Value()), 64)
					if err != nil || math.IsNaN(amount) || math.IsInf(amount, 0) || amount < 0 {
						m.financeError = "Enter a valid amount of zero or more."
						return m, nil
					}
					expense := &m.recurringExpenses[m.recurringExpenseCursor]
					expense.AmountCents = int64(math.Round(amount * 100))
					if err := m.db.UpsertRecurringExpense(*expense); err != nil {
						m.financeError = fmt.Sprintf("Could not save fixed bill: %v", err)
						return m, nil
					}
					m.accountBalance.Blur()
					m.financeError = ""
					m.financeMode = financeFixedBillsMode
					return m, nil
				}
				var cmd tea.Cmd
				m.accountBalance, cmd = m.accountBalance.Update(msg)
				return m, cmd
			}

		case payrollView:
			if m.payrollDeleteMode {
				switch msg.String() {
				case "n", "q":
					m.payrollDeleteMode = false
					m.payrollError = ""
					return m, nil

				case "y":
					if len(m.payrollStatements) == 0 || m.payrollCursor >= len(m.payrollStatements) {
						m.payrollDeleteMode = false
						return m, nil
					}

					statement := m.payrollStatements[m.payrollCursor]
					if err := m.db.DeletePayrollStatement(statement.ID); err != nil {
						m.payrollError = fmt.Sprintf("Could not delete statement: %v", err)
						return m, nil
					}

					m.payrollStatements = append(
						m.payrollStatements[:m.payrollCursor],
						m.payrollStatements[m.payrollCursor+1:]...,
					)
					if m.payrollCursor >= len(m.payrollStatements) && m.payrollCursor > 0 {
						m.payrollCursor--
					}

					now := time.Now()
					m.monthlyNetIncomeCents = 0
					for _, payroll := range m.payrollStatements {
						if payroll.PayDate.Year() == now.Year() && payroll.PayDate.Month() == now.Month() {
							m.monthlyNetIncomeCents += payroll.NetCents
						}
					}
					m.payrollDeleteMode = false
					m.payrollError = ""
					return m, nil
				}
			}

			switch msg.String() {
			case "q":
				m.currentView = dashboardView
			case "tab":
				m.currentView = spendingView
			case "shift+tab":
				m.currentView = financesView

			case "d":
				if len(m.payrollStatements) > 0 {
					m.payrollDeleteMode = true
					m.payrollError = ""
				}

			case "up", "w":
				if m.payrollCursor > 0 {
					m.payrollCursor--
				}

			case "down", "s":
				if m.payrollCursor < len(m.payrollStatements)-1 {
					m.payrollCursor++
				}

			case "home", "g":
				m.payrollCursor = 0

			case "end", "G":
				if len(m.payrollStatements) > 0 {
					m.payrollCursor = len(m.payrollStatements) - 1
				}
			}

		case spendingView:
			switch msg.String() {
			case "q":
				m.currentView = dashboardView
			case "tab":
				m.currentView = financesView
			case "shift+tab":
				m.currentView = payrollView
			case "up", "w":
				if m.spendingCursor > 0 {
					m.spendingCursor--
				}
			case "down", "s":
				if m.spendingCursor < len(m.spendingTransactions)-1 {
					m.spendingCursor++
				}
			case "g":
				m.spendingCursor = 0
			case "G":
				if len(m.spendingTransactions) > 0 {
					m.spendingCursor = len(m.spendingTransactions) - 1
				}
			}

		case projectionView:
			switch msg.String() {
			case "q":
				m.currentView = dashboardView
			case "tab", "shift+tab":
				m.currentView = financialGoalsView
			case "up", "w":
				if m.projectionCursor > 0 {
					m.projectionCursor--
				}
			case "down", "s":
				if m.projectionCursor < 12 {
					m.projectionCursor++
				}
			case "left", "h":
				m.adjustProjectionSetting(-1)
			case "right", "l":
				m.adjustProjectionSetting(1)
			case "r":
				m.projectionSettings = models.DefaultProjectionSettings()
				if m.profile != nil {
					m.projectionSettings.BirthDate = m.profile.BirthDate
				}
				m.projectionError = ""
				if err := m.db.SaveProjectionSettings(m.projectionSettings); err != nil {
					m.projectionError = fmt.Sprintf("Could not save assumptions: %v", err)
				}
			}

		case financialGoalsView:
			switch msg.String() {
			case "q":
				m.currentView = dashboardView
			case "tab", "shift+tab":
				m.currentView = projectionView
			case "up", "w":
				if m.financialGoalCursor > 0 {
					m.financialGoalCursor--
				}
			case "down", "s":
				if m.financialGoalCursor < len(m.financialGoals)-1 {
					m.financialGoalCursor++
				}
			case "left", "h":
				m.adjustFinancialGoal(0, -10000)
			case "right", "l":
				m.adjustFinancialGoal(0, 10000)
			case "[":
				m.adjustFinancialGoal(-50000, 0)
			case "]":
				m.adjustFinancialGoal(50000, 0)
			case "v":
				m.adjustGoalStockFunding(-10000)
			case "V":
				m.adjustGoalStockFunding(10000)
			}

		}
	}

	return m, nil
}

func (m *Model) adjustGoalStockFunding(delta int64) {
	s := &m.projectionSettings
	s.MonthlyGoalStockCents = min(s.MonthlyNetStockCents, max(0, s.MonthlyGoalStockCents+delta))
	m.financialGoalError = ""
	if err := m.db.SaveProjectionSettings(*s); err != nil {
		m.financialGoalError = fmt.Sprintf("Could not save stock funding: %v", err)
	}
}

func (m *Model) adjustProjectionSetting(direction int64) {
	s := &m.projectionSettings
	switch m.projectionCursor {
	case 0:
		s.MonthlySpendingCents = max(0, s.MonthlySpendingCents+direction*10000)
	case 1:
		s.BaseSalaryCents = max(0, s.BaseSalaryCents+direction*100000)
	case 2:
		s.BonusRate = min(1, max(0, s.BonusRate+float64(direction)*0.005))
	case 3:
		s.RealIncomeGrowthRate = min(0.20, max(-0.20, s.RealIncomeGrowthRate+float64(direction)*0.005))
	case 4:
		s.RealReturnRate = min(0.20, max(-0.20, s.RealReturnRate+float64(direction)*0.005))
	case 5:
		s.CashRealReturnRate = min(0.20, max(-0.20, s.CashRealReturnRate+float64(direction)*0.005))
	case 6:
		s.MonthlyNetStockCents = max(0, s.MonthlyNetStockCents+direction*10000)
	case 7:
		s.NetPayRate = min(1, max(0, s.NetPayRate+float64(direction)*0.01))
	case 8:
		s.SurplusInvestedRate = min(1, max(0, s.SurplusInvestedRate+float64(direction)*0.05))
	case 9:
		s.Employee401KAnnualCents = max(0, s.Employee401KAnnualCents+direction*50000)
	case 10:
		s.EmployerMatchRate = min(2, max(0, s.EmployerMatchRate+float64(direction)*0.05))
	case 11:
		s.IRAAnnualCents = max(0, s.IRAAnnualCents+direction*50000)
	case 12:
		s.RetirementAge = int(min(80, max(30, int64(s.RetirementAge)+direction)))
	}
	m.projectionError = ""
	if err := m.db.SaveProjectionSettings(*s); err != nil {
		m.projectionError = fmt.Sprintf("Could not save assumptions: %v", err)
	}
}

func (m *Model) adjustFinancialGoal(savedDelta, plannedDelta int64) {
	if len(m.financialGoals) == 0 || m.financialGoalCursor >= len(m.financialGoals) {
		return
	}
	goal := &m.financialGoals[m.financialGoalCursor]
	goal.SavedCents = min(goal.TargetCents, max(0, goal.SavedCents+savedDelta))
	goal.PlannedMonthlyCents = max(0, goal.PlannedMonthlyCents+plannedDelta)
	m.financialGoalError = ""
	if err := m.db.UpdateFinancialGoalProgress(goal.ID, goal.SavedCents, goal.PlannedMonthlyCents); err != nil {
		m.financialGoalError = fmt.Sprintf("Could not save goal: %v", err)
	}
}
