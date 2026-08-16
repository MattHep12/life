package app

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func (m Model) renderMenu() string {
	var sb strings.Builder

	sb.WriteString("LIFE\n\n")

	for i, item := range m.choices {
		prefix := "  "

		if i == m.cursor {
			prefix = "> "
		}

		fmt.Fprintf(&sb, "%s%s\n", prefix, item.label)
	}

	sb.WriteString("\nEnter to Open\n")
	sb.WriteString("Esc to quit\n")

	return sb.String()
}

func (m Model) renderDashboardView() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "LIFE DASHBOARD\n\n")

	fmt.Fprintf(&sb, "FINANCES\n")
	fmt.Fprintf(&sb, "Net Worth: %s\n\n", formatMoney(m.netWorth()))
	fmt.Fprintf(&sb, "THIS MONTH\n")
	fmt.Fprintf(&sb, "Net Payroll Income: %s\n\n", formatMoney(float64(m.monthlyNetIncomeCents)/100))

	fmt.Fprintf(&sb, "TODAY\n")
	fmt.Fprintf(&sb, "Tasks Remaining: %d\n\n", m.remainingTasks())

	fmt.Fprintf(&sb, "HABITS\n")
	fmt.Fprintf(&sb, "Weekly Completion: %d%%\n\n", m.habitCompletion)

	fmt.Fprintf(&sb, "GOALS\n")
	fmt.Fprintf(&sb, "Active Goals: %d\n\n", m.activeGoals)

	fmt.Fprintf(&sb, "q to return • Esc to quit\n")

	return sb.String()
}

func (m Model) remainingTasks() int {
	count := 0

	for _, task := range m.tasks {
		if !task.Completed {
			count++
		}
	}

	return count
}

func (m Model) renderTasksView() string {
	var sb strings.Builder

	sb.WriteString("TASKS\n\n")

	for i, task := range m.tasks {
		prefix := "  "
		checkbox := "[ ]"

		if i == m.taskCursor {
			prefix = "> "
		}

		if task.Completed {
			checkbox = "[x]"
		}

		fmt.Fprintf(
			&sb,
			"%s%s %s\n",
			prefix,
			checkbox,
			task.Title,
		)
	}

	sb.WriteString(
		"\nEnter/Space to toggle • q to return • Esc to quit\n",
	)

	return sb.String()
}

func (m Model) renderHabitsView() string {
	return "HABITS\n\nPress q to return to the menu.\n"
}

func (m Model) renderFinanceList() string {
	var sb strings.Builder

	sb.WriteString("FINANCES\n\n")

	fmt.Fprintf(&sb, "NET WORTH:\n")
	fmt.Fprintf(&sb, "%s\n\n", formatMoney(m.netWorth()))

	sb.WriteString("ACCOUNTS:\n")

	for i, account := range m.accounts {
		prefix := "  "

		if i == m.financeCursor {
			prefix = "> "
		}

		fmt.Fprintf(
			&sb,
			"%s%-25s %s\n",
			prefix,
			account.Name,
			formatMoney(account.Balance),
		)
	}

	sb.WriteString("\na Add • e Edit • d Delete • q Back • Esc Quit\n")
	return sb.String()
}

func (m Model) renderFinanceAddName() string {
	var sb strings.Builder

	sb.WriteString("ADD ACCOUNT\n\n")
	sb.WriteString("Account Name:\n")
	sb.WriteString(m.accountName.View())
	if m.financeError != "" {
		fmt.Fprintf(&sb, "\n\nError: %s", m.financeError)
	}

	sb.WriteString("\n\nEnter Continue • Esc Cancel • Ctrl+C Quit\n")
	return sb.String()
}

func (m Model) renderFinanceAddBalance() string {
	var sb strings.Builder

	sb.WriteString("ADD ACCOUNT\n\n")

	fmt.Fprintf(&sb, "Account: %s\n\n", m.accountName.Value())

	sb.WriteString("Balance:\n")
	sb.WriteString(m.accountBalance.View())
	if m.financeError != "" {
		fmt.Fprintf(&sb, "\n\nError: %s", m.financeError)
	}

	sb.WriteString("\n\nEnter Save • Esc Cancel • Ctrl+C Quit\n")
	return sb.String()
}

func (m Model) renderFinanceEditBalance() string {
	if len(m.accounts) == 0 {
		return ""
	}

	account := m.accounts[m.financeCursor]

	var sb strings.Builder

	sb.WriteString("EDIT ACCOUNT\n\n")

	fmt.Fprintf(&sb, "Account: %s\n\n", account.Name)

	sb.WriteString("Balance:\n")
	sb.WriteString(m.accountBalance.View())
	if m.financeError != "" {
		fmt.Fprintf(&sb, "\n\nError: %s", m.financeError)
	}

	sb.WriteString("\n\nEnter Save • Esc Cancel • Ctrl+C Quit\n")
	return sb.String()
}

func (m Model) renderFinanceDeleteConfirm() string {
	if len(m.accounts) == 0 {
		return ""
	}

	account := m.accounts[m.financeCursor]

	var sb strings.Builder
	sb.WriteString("DELETE ACCOUNT\n\n")

	fmt.Fprintf(
		&sb,
		"Delete %s (%s)?\n\n",
		account.Name,
		formatMoney(account.Balance),
	)
	if m.financeError != "" {
		fmt.Fprintf(&sb, "Error: %s\n\n", m.financeError)
	}

	sb.WriteString("y Confirm • n/q/Esc Cancel • Ctrl+C Quit\n")
	return sb.String()
}

func (m Model) renderFinancesView() string {
	switch m.financeMode {
	case financeListMode:
		return m.renderFinanceList()

	case financeAddNameMode:
		return m.renderFinanceAddName()

	case financeAddBalanceMode:
		return m.renderFinanceAddBalance()

	case financeEditBalanceMode:
		return m.renderFinanceEditBalance()

	case financeDeleteConfirmMode:
		return m.renderFinanceDeleteConfirm()

	default:
		return ""
	}
}

func (m Model) renderStatsView() string {
	return "STATS\n\nPress q to return to the menu.\n"
}

func (m Model) renderPayrollView() string {
	var sb strings.Builder
	sb.WriteString("PAYROLL HISTORY\n\n")

	if len(m.payrollStatements) == 0 {
		sb.WriteString("No payroll statements imported yet.\n\n")
		sb.WriteString("Use import-payroll to add a pay statement.\n\n")
		sb.WriteString("q Back • Esc Quit\n")
		return sb.String()
	}

	var totalGross, totalTaxes, totalDeductions, totalNet int64
	for _, statement := range m.payrollStatements {
		totalGross += statement.GrossCents
		totalTaxes += statement.TaxesCents
		totalDeductions += statement.DeductionsCents
		totalNet += statement.NetCents
	}

	fmt.Fprintf(&sb, "TOTALS (%d statements)\n", len(m.payrollStatements))
	fmt.Fprintf(&sb, "Gross: %s • Taxes: %s • Deductions: %s • Net: %s\n\n",
		formatCents(totalGross),
		formatCents(totalTaxes),
		formatCents(totalDeductions),
		formatCents(totalNet),
	)

	sb.WriteString("STATEMENTS\n")
	for i, statement := range m.payrollStatements {
		prefix := "  "
		if i == m.payrollCursor {
			prefix = "> "
		}
		fmt.Fprintf(
			&sb,
			"%s%s  %-23s Net %s\n",
			prefix,
			statement.PayDate.Format("Jan 02, 2006"),
			statement.PeriodStart.Format("Jan 02")+" - "+statement.PeriodEnd.Format("Jan 02, 2006"),
			formatCents(statement.NetCents),
		)
	}

	selected := m.payrollStatements[m.payrollCursor]
	sb.WriteString("\nSELECTED STATEMENT\n")
	fmt.Fprintf(&sb, "Pay period:  %s - %s\n", selected.PeriodStart.Format("Jan 02, 2006"), selected.PeriodEnd.Format("Jan 02, 2006"))
	fmt.Fprintf(&sb, "Pay date:    %s\n", selected.PayDate.Format("Jan 02, 2006"))
	fmt.Fprintf(&sb, "Gross pay:   %s\n", formatCents(selected.GrossCents))
	fmt.Fprintf(&sb, "Taxable pay: %s\n", formatCents(selected.TaxableCents))
	fmt.Fprintf(&sb, "Taxes:       %s\n", formatCents(selected.TaxesCents))
	fmt.Fprintf(&sb, "Deductions:  %s\n", formatCents(selected.DeductionsCents))
	fmt.Fprintf(&sb, "Net pay:     %s\n", formatCents(selected.NetCents))

	sb.WriteString("\n↑/↓ Navigate • q Back • Esc Quit\n")
	return sb.String()
}

func (m Model) View() tea.View {
	var content string

	switch m.currentView {
	case menuView:
		content = m.renderMenu()

	case dashboardView:
		content = m.renderDashboardView()

	case tasksView:
		content = m.renderTasksView()

	case habitsView:
		content = m.renderHabitsView()

	case financesView:
		content = m.renderFinancesView()

	case payrollView:
		content = m.renderPayrollView()

	case statsView:
		content = m.renderStatsView()
	}

	return tea.View{
		Content: content,
	}
}
