package app

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"life/internal/models"
)

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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

			return m, tea.Quit
		}

		switch m.currentView {
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
			case "q":
				m.currentView = menuView
			}

		case tasksView:
			switch msg.String() {
			case "q":
				m.currentView = menuView

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
				m.currentView = menuView
			}

		case financesView:
			switch m.financeMode {

			case financeListMode:
				switch msg.String() {
				case "q":
					m.currentView = menuView

				case "a":
					m.financeError = ""
					m.resetAccountInputs()
					m.financeMode = financeAddNameMode
					cmd := m.accountName.Focus()
					return m, cmd

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

				case "up", "w":
					if m.financeCursor > 0 {
						m.financeCursor--
					}

				case "down", "s":
					if m.financeCursor < len(m.accounts)-1 {
						m.financeCursor++
					}
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
						Name:    strings.TrimSpace(m.accountName.Value()),
						Balance: balance,
					}

					id, err := m.db.AddAccount(account)
					if err != nil {
						m.financeError = fmt.Sprintf("Could not save account: %v", err)
						return m, nil
					}

					account.ID = id

					m.accounts = append(m.accounts, account)

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
			}

		case payrollView:
			switch msg.String() {
			case "q":
				m.currentView = menuView

			case "up", "w":
				if m.payrollCursor > 0 {
					m.payrollCursor--
				}

			case "down", "s":
				if m.payrollCursor < len(m.payrollStatements)-1 {
					m.payrollCursor++
				}
			}

		case statsView:
			switch msg.String() {
			case "q":
				m.currentView = menuView
			}
		}
	}

	return m, nil
}
