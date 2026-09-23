package app

import (
	"fmt"
	"image/color"
	"math"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/linechart"
	legacyLipgloss "github.com/charmbracelet/lipgloss"

	"life/internal/models"
	"life/internal/projection"
)

func (m Model) contentWidth() int {
	width := m.width - 6
	if m.width == 0 {
		width = 92
	}
	return max(38, width)
}

func (m Model) shell(title, subtitle, body, helpText string) string {
	width := m.contentWidth()
	header := lipgloss.JoinHorizontal(lipgloss.Center, logoStyle.Render("LIFE"), "  ", titleStyle.Render(title))
	if subtitle != "" {
		header += "\n" + subtitleStyle.Render(subtitle)
	}
	content := header + "\n\n" + body
	if helpText != "" {
		content += "\n" + helpStyle.Render(helpText)
	}
	return lipgloss.NewStyle().Width(width).Margin(1, 2).Render(content)
}

func help(keys ...string) string {
	parts := make([]string, 0, len(keys)/2)
	for i := 0; i+1 < len(keys); i += 2 {
		parts = append(parts, keyStyle.Render(keys[i])+" "+keys[i+1])
	}
	return strings.Join(parts, mutedStyle.Render("  •  "))
}

func panel(content string, width int) string {
	return panelStyle.Width(max(20, width-6)).Render(content)
}

func metricCard(label, value, note string, width int) string {
	return metricCardAccent(label, value, note, width, colorPrimary)
}

func metricCardAccent(label, value, note string, width int, accent color.Color) string {
	content := subtitleStyle.Render(strings.ToUpper(label)) + "\n" +
		lipgloss.NewStyle().Bold(true).Foreground(accent).Render(value)
	if note != "" {
		content += "\n" + mutedStyle.Render(note)
	}
	return panelStyle.BorderForeground(accent).Width(max(14, width-6)).Render(content)
}

func (m Model) renderProfileSetupView() string {
	width := m.contentWidth()
	label := "YOUR NAME"
	input := m.profileName.View()
	instructions := "Create a private local profile. Your financial data stays in a local database on this computer and is never committed to Git."
	if m.profileSetupStep == 1 {
		label = "BIRTH DATE"
		input = m.profileBirthDate.View()
		instructions = "Your birth date is used only for age-based financial projections. Use YYYY-MM-DD."
	}
	content := sectionStyle.Foreground(colorSecondary).Render("WELCOME TO LIFE") + "\n\n" +
		mutedStyle.Render(instructions) + "\n\n" +
		sectionStyle.Render(label) + "\n" + inputBoxStyle.Render(input)
	if m.profileError != "" {
		content += "\n\n" + errorStyle.Render("! "+m.profileError)
	}
	return m.shell("Profile Setup", "Start with a clean, private workspace.", activePanelStyle.Width(max(48, min(86, width-6))).Render(content), help("enter", "continue", "esc", "quit"))
}

func (m Model) renderMenu() string {
	width := m.contentWidth()
	rows := make([]string, 0, len(m.choices))
	for i, item := range m.choices {
		line := fmt.Sprintf("%s  %-18s %s", item.icon, item.label, item.description)
		if i == m.cursor {
			rows = append(rows, selectedRowStyle.Width(width-10).Render("› "+line))
		} else {
			rows = append(rows, rowStyle.Width(width-10).Render("  "+line))
		}
	}
	return m.shell("Command Center", "Everything important, in one place.", panel(strings.Join(rows, "\n"), width), help("↑/↓", "navigate", "enter", "open", "esc", "quit"))
}

func (m Model) renderDashboardView() string {
	width := m.contentWidth()
	cards := []string{
		metricCard("Net worth", formatMoney(m.netWorth()), fmt.Sprintf("%d accounts", len(m.accounts)), 27),
		metricCard("Net income", formatCents(m.monthlyNetIncomeCents), "this month", 27),
		metricCard("Tasks left", fmt.Sprintf("%d", m.remainingTasks()), "today", 27),
		metricCard("Habit score", fmt.Sprintf("%d%%", m.habitCompletion), "this week", 27),
	}
	var body string
	if width >= 100 {
		body = lipgloss.JoinHorizontal(lipgloss.Top, cards...)
	} else if width >= 64 {
		cardWidth := width / 2
		body = lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Top,
				metricCard("Net worth", formatMoney(m.netWorth()), fmt.Sprintf("%d accounts", len(m.accounts)), cardWidth),
				metricCard("Net income", formatCents(m.monthlyNetIncomeCents), "this month", cardWidth),
			),
			lipgloss.JoinHorizontal(lipgloss.Top,
				metricCard("Tasks left", fmt.Sprintf("%d", m.remainingTasks()), "today", cardWidth),
				metricCard("Habit score", fmt.Sprintf("%d%%", m.habitCompletion), "this week", cardWidth),
			),
		)
	} else {
		body = lipgloss.JoinVertical(lipgloss.Left, cards...)
	}
	activeGoals := 0
	type dashboardGoal struct {
		name      string
		progress  string
		allocated string
		remaining string
		deadline  string
	}
	dashboardGoals := make([]dashboardGoal, 0, len(m.financialGoals))
	for _, goal := range m.financialGoals {
		if goal.SavedCents < goal.TargetCents {
			activeGoals++
			percent := 0.0
			if goal.TargetCents > 0 {
				percent = float64(goal.SavedCents) / float64(goal.TargetCents) * 100
			}
			dashboardGoals = append(dashboardGoals, dashboardGoal{
				name:      goal.Name,
				progress:  fmt.Sprintf("%.0f%%", percent),
				allocated: formatCents(goal.SavedCents),
				remaining: formatCents(goal.TargetCents - goal.SavedCents),
				deadline:  goal.TargetDate.Format("Jan 2006"),
			})
		}
	}
	goalLines := []string{sectionStyle.Background(colorSurface).Width(width - 12).Render(fmt.Sprintf("ACTIVE GOALS  •  %d in progress", activeGoals))}
	if activeGoals == 0 {
		goalLines = append(goalLines, "", valueStyle.Background(colorSurface).Width(width-12).Render("All goals are funded."))
	} else {
		columns := "%-26s %10s %16s %16s %12s"
		goalLines = append(goalLines, "", tableHeaderStyle.Width(width-12).Render(fmt.Sprintf(columns, "GOAL", "PROGRESS", "ALLOCATED", "REMAINING", "TARGET")))
		for _, goal := range dashboardGoals {
			goalLines = append(goalLines, payrollRowStyle.Width(width-12).Render(fmt.Sprintf(columns,
				goal.name,
				goal.progress,
				goal.allocated,
				goal.remaining,
				goal.deadline,
			)))
		}
	}
	goals := strings.Join(goalLines, "\n")
	body += "\n\n" + panel(goals, width)
	rows := make([]string, 0, len(m.choices))
	for i, item := range m.choices {
		line := fmt.Sprintf("%s  %-22s %s", item.icon, item.label, item.description)
		if i == m.cursor {
			rows = append(rows, selectedRowStyle.Width(width-12).Render("› "+line))
		} else {
			rows = append(rows, rowStyle.Background(colorSurface).Width(width-12).Render("  "+line))
		}
	}
	body += "\n\n" + panel(sectionStyle.Foreground(colorInfo).Render("EXPLORE LIFE")+"\n\n"+strings.Join(rows, "\n"), width)
	title := "Dashboard"
	if m.profile != nil && strings.TrimSpace(m.profile.Name) != "" {
		title = m.profile.Name + "'s Dashboard"
	}
	return m.shell(title, "A clear view of today and your financial momentum.", body, help("↑/↓", "navigate", "enter", "open", "esc", "quit"))
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

func renderProductivityTabs(active view) string {
	tasksTab := mutedStyle.Render("  TASKS  ")
	habitsTab := mutedStyle.Render("  HABITS  ")
	selected := lipgloss.NewStyle().Bold(true).Foreground(colorBackground).Background(colorSecondary)
	if active == tasksView {
		tasksTab = selected.Render("  TASKS  ")
	} else {
		habitsTab = selected.Render("  HABITS  ")
	}
	return tasksTab + "  " + habitsTab
}

func (m Model) renderTasksView() string {
	width := m.contentWidth()
	rows := make([]string, 0, len(m.tasks))
	for i, task := range m.tasks {
		checkbox := "○"
		if task.Completed {
			checkbox = valueStyle.Render("●")
		}
		line := fmt.Sprintf("%s  %s", checkbox, task.Title)
		if i == m.taskCursor {
			rows = append(rows, selectedRowStyle.Width(width-10).Render("› "+line))
		} else {
			rows = append(rows, rowStyle.Width(width-10).Render("  "+line))
		}
	}
	if len(rows) == 0 {
		rows = append(rows, mutedStyle.Render("No tasks for today."))
	}
	body := renderProductivityTabs(tasksView) + "\n\n" + panel(strings.Join(rows, "\n"), width)
	return m.shell("Tasks & Habits", fmt.Sprintf("%d tasks remaining today", m.remainingTasks()), body, help("tab", "habits", "↑/↓", "navigate", "space", "toggle", "q", "back", "esc", "quit"))
}

func (m Model) renderHabitsView() string {
	width := m.contentWidth()
	body := renderProductivityTabs(habitsView) + "\n\n" + panel(sectionStyle.Render("WEEKLY COMPLETION")+"\n\n"+valueStyle.Render(fmt.Sprintf("%d%%", m.habitCompletion))+"\n"+mutedStyle.Render("Detailed habit tracking is coming next."), width)
	return m.shell("Tasks & Habits", "Plan your day and build consistent routines.", body, help("tab", "tasks", "q", "back", "esc", "quit"))
}

type financeCategory string

const (
	financeCash       financeCategory = "cash"
	financeInvestment financeCategory = "investment"
	financeRetirement financeCategory = "retirement"
	financeOther      financeCategory = "other"
)

var financeCategories = []financeCategory{financeCash, financeInvestment, financeRetirement}

func renderFinanceTabs(active view) string {
	accountsTab := mutedStyle.Render("  ACCOUNTS  ")
	payrollTab := mutedStyle.Render("  PAYROLL  ")
	spendingTab := mutedStyle.Render("  SPENDING  ")
	selected := lipgloss.NewStyle().Bold(true).Foreground(colorBackground).Background(colorSecondary)
	switch active {
	case financesView:
		accountsTab = selected.Render("  ACCOUNTS  ")
	case payrollView:
		payrollTab = selected.Render("  PAYROLL  ")
	case spendingView:
		spendingTab = selected.Render("  SPENDING  ")
	}
	return accountsTab + "  " + payrollTab + "  " + spendingTab
}

func accountCategory(account models.Account) financeCategory {
	switch financeCategory(account.Category) {
	case financeRetirement:
		return financeRetirement
	case financeInvestment:
		return financeInvestment
	case financeCash:
		return financeCash
	default:
		return financeCash
	}
}

func (m Model) financeTotals() (cash, investments, retirement float64) {
	for _, account := range m.accounts {
		switch accountCategory(account) {
		case financeCash:
			cash += account.Balance
		case financeInvestment:
			investments += account.Balance
		case financeRetirement:
			retirement += account.Balance
		default:
			cash += account.Balance
		}
	}
	return
}

func (m Model) financeAccountsPanel(title string, category financeCategory, width int, accent color.Color) string {
	rows := make([]string, 0)
	for i, account := range m.accounts {
		accountGroup := accountCategory(account)
		if accountGroup != category && !(category == financeCash && accountGroup == financeOther) {
			continue
		}
		line := fmt.Sprintf("%-27s %14s", account.Name, formatMoney(account.Balance))
		if i == m.financeCursor {
			rows = append(rows, selectedRowStyle.Foreground(colorSecondary).Width(max(38, width-10)).Render("› "+line))
		} else {
			rows = append(rows, payrollRowStyle.Width(max(38, width-10)).Render("  "+line))
		}
	}
	if len(rows) == 0 {
		rows = append(rows, mutedStyle.Render("No accounts in this group."))
	}
	for len(rows) < 2 {
		rows = append(rows, " ")
	}
	return panelStyle.BorderForeground(accent).Width(max(42, width-6)).Render(
		sectionStyle.Foreground(accent).Render(title) + "\n\n" + strings.Join(rows, "\n"),
	)
}

func (m Model) renderFinanceList() string {
	width := m.contentWidth()
	cash, investments, retirement := m.financeTotals()
	cardWidth := width / 4
	summary := lipgloss.JoinHorizontal(lipgloss.Top,
		metricCardAccent("Net worth", formatMoney(m.netWorth()), fmt.Sprintf("%d accounts", len(m.accounts)), cardWidth, colorPrimary),
		metricCardAccent("Cash", formatMoney(cash), "checking + savings", cardWidth, colorInfo),
		metricCardAccent("Investments", formatMoney(investments), "taxable + equity awards", cardWidth, colorSecondary),
		metricCardAccent("Retirement", formatMoney(retirement), "401(k) + Roth IRA", cardWidth, colorGold),
	)

	var accounts string
	if width >= 150 {
		panelWidth := width / 3
		accounts = lipgloss.JoinHorizontal(lipgloss.Top,
			m.financeAccountsPanel("CASH", financeCash, panelWidth, colorInfo),
			m.financeAccountsPanel("INVESTMENTS", financeInvestment, panelWidth, colorSecondary),
			m.financeAccountsPanel("RETIREMENT", financeRetirement, width-panelWidth*2, colorGold),
		)
	} else {
		accounts = lipgloss.JoinVertical(lipgloss.Left,
			m.financeAccountsPanel("CASH", financeCash, width, colorInfo),
			m.financeAccountsPanel("INVESTMENTS", financeInvestment, width, colorSecondary),
			m.financeAccountsPanel("RETIREMENT", financeRetirement, width, colorGold),
		)
	}

	year := time.Now().Year()
	stockValue, _, _ := m.stockVestTotals(year)
	employee401K, employer401K := m.retirementTotals(year)
	insights := panel(
		sectionStyle.Render("YEAR-TO-DATE FLOWS")+"\n\n"+
			fmt.Sprintf("Vested stock income  %s    Employee 401(k)  %s    Employer 401(k)  %s",
				valueStyle.Render(formatCents(stockValue)),
				valueStyle.Render(formatCents(employee401K)),
				valueStyle.Render(formatCents(employer401K)),
			)+"\n"+
			mutedStyle.Render("Flows explain how balances grew; they are not added to net worth again."),
		width,
	)

	body := renderFinanceTabs(financesView) + "\n\n" + summary + "\n\n" + accounts + "\n\n" + insights
	if len(m.spendingTransactions) > 0 {
		body += "\n\n" + m.renderMonthlyCashFlow(width)
	}
	return m.shell("Finances", "Accounts, income, and spending in one place.", body, help("tab", "next section", "↑/↓", "navigate", "a", "add", "e", "edit", "d", "delete", "f", "fixed bills", "q", "back"))
}

func (m Model) financeForm(title, subtitle, label, context, input, action string) string {
	width := m.contentWidth()
	content := ""
	if context != "" {
		content += subtitleStyle.Render(context) + "\n\n"
	}
	content += sectionStyle.Render(label) + "\n" + inputBoxStyle.Render(input)
	if m.financeError != "" {
		content += "\n\n" + errorStyle.Render("! "+m.financeError)
	}
	return m.shell(title, subtitle, activePanelStyle.Width(max(28, width-6)).Render(content), help("enter", action, "esc", "cancel", "ctrl+c", "quit"))
}

func (m Model) renderFinanceAddName() string {
	return m.financeForm("Add account", "Create a manually tracked account.", "ACCOUNT NAME", "", m.accountName.View(), "continue")
}

func (m Model) renderFinanceAddCategory() string {
	width := m.contentWidth()
	rows := make([]string, 0, len(financeCategories))
	for i, category := range financeCategories {
		label := strings.ToUpper(string(category))
		if i == m.accountCategoryCursor {
			rows = append(rows, selectedRowStyle.Width(max(24, width-12)).Render("› "+label))
		} else {
			rows = append(rows, payrollRowStyle.Width(max(24, width-12)).Render("  "+label))
		}
	}
	content := sectionStyle.Render("ACCOUNT CATEGORY") + "\n\n" + strings.Join(rows, "\n")
	return m.shell("Add account", "Choose where this account belongs.", activePanelStyle.Width(max(28, width-6)).Render(content), help("↑/↓", "select", "enter", "continue", "esc", "cancel"))
}

func (m Model) renderFinanceAddBalance() string {
	return m.financeForm("Add account", "Set the account's current balance.", "BALANCE", "Account: "+m.accountName.Value(), m.accountBalance.View(), "save")
}

func (m Model) renderFinanceEditBalance() string {
	if len(m.accounts) == 0 || m.financeCursor >= len(m.accounts) {
		return m.renderFinanceList()
	}
	return m.financeForm("Edit balance", "Update a manually tracked balance.", "BALANCE", "Account: "+m.accounts[m.financeCursor].Name, m.accountBalance.View(), "save")
}

func (m Model) renderFinanceDeleteConfirm() string {
	if len(m.accounts) == 0 || m.financeCursor >= len(m.accounts) {
		return m.renderFinanceList()
	}
	account := m.accounts[m.financeCursor]
	message := warningStyle.Render("Delete "+account.Name+"?") + "\n\n" + subtitleStyle.Render("Current balance: ") + valueStyle.Render(formatMoney(account.Balance)) + "\n\n" + mutedStyle.Render("This removes the account from local storage.")
	if m.financeError != "" {
		message += "\n\n" + errorStyle.Render("! "+m.financeError)
	}
	return m.shell("Confirm deletion", "This action cannot be undone.", activePanelStyle.Width(max(28, m.contentWidth()-6)).Render(message), help("y", "confirm", "n/q/esc", "cancel", "ctrl+c", "quit"))
}

func (m Model) renderFixedBills() string {
	width := m.contentWidth()
	rows := make([]string, 0, len(m.recurringExpenses))
	var total int64
	for i, expense := range m.recurringExpenses {
		total += expense.AmountCents
		line := fmt.Sprintf("%-30s %15s", expense.Name, formatCents(expense.AmountCents))
		if i == m.recurringExpenseCursor {
			rows = append(rows, selectedRowStyle.Width(max(48, width-12)).Render("› "+line))
		} else {
			rows = append(rows, payrollRowStyle.Width(max(48, width-12)).Render("  "+line))
		}
	}
	if len(rows) == 0 {
		rows = append(rows, mutedStyle.Render("No fixed bills configured."))
	}
	content := sectionStyle.Foreground(colorGold).Render("MONTHLY FIXED BILLS") + "\n\n" +
		strings.Join(rows, "\n") + "\n\n" +
		valueStyle.Render(fmt.Sprintf("%-30s %15s", "TOTAL", formatCents(total)))
	return m.shell("Finances", "Edit the recurring bills used by Monthly Cash Flow.", activePanelStyle.Width(max(54, width-6)).Render(content), help("↑/↓", "select", "enter/e", "edit", "q/esc", "back"))
}

func (m Model) renderEditFixedBill() string {
	if len(m.recurringExpenses) == 0 || m.recurringExpenseCursor >= len(m.recurringExpenses) {
		return m.renderFixedBills()
	}
	expense := m.recurringExpenses[m.recurringExpenseCursor]
	return m.financeForm("Edit fixed bill", "Update the monthly amount used by cash flow.", "MONTHLY AMOUNT", "Bill: "+expense.Name, m.accountBalance.View(), "save")
}

func (m Model) renderFinancesView() string {
	switch m.financeMode {
	case financeListMode:
		return m.renderFinanceList()
	case financeAddCategoryMode:
		return m.renderFinanceAddCategory()
	case financeAddNameMode:
		return m.renderFinanceAddName()
	case financeAddBalanceMode:
		return m.renderFinanceAddBalance()
	case financeEditBalanceMode:
		return m.renderFinanceEditBalance()
	case financeDeleteConfirmMode:
		return m.renderFinanceDeleteConfirm()
	case financeFixedBillsMode:
		return m.renderFixedBills()
	case financeEditFixedBillMode:
		return m.renderEditFixedBill()
	default:
		return m.renderFinanceList()
	}
}

func (m Model) payrollTotals() (gross, taxes, deductions, net int64) {
	for _, statement := range m.payrollStatements {
		gross += statement.GrossCents
		taxes += statement.TaxesCents
		deductions += statement.DeductionsCents
		net += statement.NetCents
	}
	return
}

func (m Model) stockVestTotals(year int) (grossValue, grossShares, netShares int64) {
	for _, vest := range m.stockVests {
		if vest.VestDate.Year() != year {
			continue
		}
		grossValue += vest.GrossValueCents
		grossShares += vest.GrossSharesMicros
		netShares += vest.NetSharesMicros
	}
	return
}

func (m Model) retirementTotals(year int) (employee, employer int64) {
	var employeeCurrentTotal, employerCurrentTotal int64
	for _, statement := range m.payrollStatements {
		if statement.PayDate.Year() != year {
			continue
		}
		employeeCurrentTotal += statement.Employee401KCents
		employerCurrentTotal += statement.Employer401KCents
		employee = max(employee, statement.Employee401KYTDCents)
		employer = max(employer, statement.Employer401KYTDCents)
	}
	// Older imported rows may not have YTD fields until they are reprocessed.
	if employee == 0 {
		employee = employeeCurrentTotal
	}
	if employer == 0 {
		employer = employerCurrentTotal
	}
	return
}

func (m Model) spendingByMonth() map[time.Time]int64 {
	totals := make(map[time.Time]int64)
	for _, transaction := range m.spendingTransactions {
		if transaction.IsPayment() {
			continue
		}
		month := time.Date(transaction.TransactionDate.Year(), transaction.TransactionDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		totals[month] += transaction.AmountCents
	}
	return totals
}

func (m Model) renderMonthlyCashFlow(width int) string {
	income := m.monthlyIncomeTotals()
	spending := m.spendingByMonth()
	var fixedExpenses int64
	for _, expense := range m.recurringExpenses {
		fixedExpenses += expense.AmountCents
	}
	lines := []string{
		sectionStyle.Foreground(colorPrimary).Render("MONTHLY CASH FLOW"),
		mutedStyle.Render("Net payroll minus Discover spending and recurring rent/car obligations."),
		"",
		tableHeaderStyle.Render(fmt.Sprintf("%-7s %14s %14s %14s %14s", "MONTH", "NET PAY", "CARD SPEND", "FIXED BILLS", "REMAINING")),
	}
	for _, month := range income {
		cardSpend := spending[month.month]
		lines = append(lines, payrollRowStyle.Render(fmt.Sprintf("%-7s %14s %14s %14s %14s",
			month.month.Format("Jan"), formatCents(month.net), formatCents(cardSpend), formatCents(fixedExpenses), formatCents(month.net-cardSpend-fixedExpenses))))
	}
	expenseNames := make([]string, 0, len(m.recurringExpenses))
	for _, expense := range m.recurringExpenses {
		expenseNames = append(expenseNames, fmt.Sprintf("%s %s", expense.Name, formatCents(expense.AmountCents)))
	}
	if len(expenseNames) > 0 {
		lines = append(lines, "", mutedStyle.Render("Fixed bills: "+strings.Join(expenseNames, "  •  ")))
	}
	return panel(strings.Join(lines, "\n"), width)
}

type monthlyIncomeTotal struct {
	month time.Time
	gross int64
	stock int64
	net   int64
}

func (m Model) monthlyIncomeTotals() []monthlyIncomeTotal {
	byMonth := make(map[time.Time]*monthlyIncomeTotal)
	for _, statement := range m.payrollStatements {
		month := time.Date(statement.PayDate.Year(), statement.PayDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		total := byMonth[month]
		if total == nil {
			total = &monthlyIncomeTotal{month: month}
			byMonth[month] = total
		}
		total.gross += statement.GrossCents
		total.net += statement.NetCents
	}
	for _, vest := range m.stockVests {
		month := time.Date(vest.VestDate.Year(), vest.VestDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		total := byMonth[month]
		if total == nil {
			total = &monthlyIncomeTotal{month: month}
			byMonth[month] = total
		}
		total.stock += vest.GrossValueCents
	}

	months := make([]time.Time, 0, len(byMonth))
	for month := range byMonth {
		months = append(months, month)
	}
	sort.Slice(months, func(i, j int) bool { return months[i].Before(months[j]) })

	totals := make([]monthlyIncomeTotal, 0, len(months))
	for _, month := range months {
		totals = append(totals, *byMonth[month])
	}
	if len(totals) > 8 {
		totals = totals[len(totals)-8:]
	}
	return totals
}

func (m Model) renderMonthlyIncomeChart(width int) string {
	totals := m.monthlyIncomeTotals()
	if len(totals) == 0 {
		return panel(sectionStyle.Foreground(colorGold).Render("COMPENSATION MIX")+"\n\n"+mutedStyle.Render("No income data available."), width)
	}

	lines := []string{
		sectionStyle.Foreground(colorGold).Render("COMPENSATION MIX"),
		mutedStyle.Render("Payroll by pay date  •  equity by vest date"),
		"",
		tableHeaderStyle.Render(fmt.Sprintf("%-5s %11s %11s %11s %11s %6s", "MONTH", "PAYROLL", "STOCK", "TOTAL", "NET CASH", "NET %")),
	}
	now := time.Now()
	var payrollTotal, stockTotal, netTotal int64
	for _, total := range totals {
		payrollTotal += total.gross
		stockTotal += total.stock
		netTotal += total.net
		takeHomeRate := 0.0
		if total.gross > 0 {
			takeHomeRate = float64(total.net) / float64(total.gross) * 100
		}
		month := total.month.Format("Jan")
		if total.month.Year() == now.Year() && total.month.Month() == now.Month() {
			month += "*"
		}
		line := fmt.Sprintf("%-5s %11s %11s %11s %11s %5.1f%%",
			month,
			formatCents(total.gross),
			formatCents(total.stock),
			formatCents(total.gross+total.stock),
			formatCents(total.net),
			takeHomeRate,
		)
		lines = append(lines, payrollRowStyle.Render(line))
	}
	totalTakeHomeRate := 0.0
	if payrollTotal > 0 {
		totalTakeHomeRate = float64(netTotal) / float64(payrollTotal) * 100
	}
	lines = append(lines,
		mutedStyle.Render(strings.Repeat("─", 61)),
		lipgloss.NewStyle().Bold(true).Foreground(colorGold).Background(colorSurface).Render(
			fmt.Sprintf("%-5s %11s %11s %11s %11s %5.1f%%",
				"TOTAL",
				formatCents(payrollTotal),
				formatCents(stockTotal),
				formatCents(payrollTotal+stockTotal),
				formatCents(netTotal),
				totalTakeHomeRate,
			),
		),
	)
	if m.payrollCursor >= 0 && m.payrollCursor < len(m.payrollStatements) {
		selectedStatement := m.payrollStatements[m.payrollCursor]
		selectedMonth := time.Date(selectedStatement.PayDate.Year(), selectedStatement.PayDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		var selectedTotal *monthlyIncomeTotal
		for i := range totals {
			if totals[i].month.Equal(selectedMonth) {
				selectedTotal = &totals[i]
			}
		}
		if selectedTotal != nil {
			var base, overtime, bonuses, payrollStock, other, taxes, deductions int64
			var employee401K, employer401K int64
			payDates := make(map[time.Time]struct{})
			for _, statement := range m.payrollStatements {
				if statement.PayDate.Year() != selectedMonth.Year() || statement.PayDate.Month() != selectedMonth.Month() {
					continue
				}
				payDates[statement.PayDate] = struct{}{}
				base += statement.BasePayCents
				overtime += statement.OvertimeCents
				bonuses += statement.BonusCents
				payrollStock += statement.PayrollStockCents
				other += statement.OtherEarningsCents
				taxes += statement.TaxesCents
				deductions += statement.DeductionsCents
				employee401K += statement.Employee401KCents
				employer401K += statement.Employer401KCents
			}
			otherDeductions := deductions - employee401K
			if otherDeductions < 0 {
				otherDeductions = 0
			}
			const detailColumns = "%-17s %13s  %-20s %12s"

			lines = append(lines,
				"",
				sectionStyle.Foreground(colorPrimary).Render("SELECTED MONTH • "+strings.ToUpper(selectedMonth.Format("January 2006"))),
				mutedStyle.Render(fmt.Sprintf("%d pay dates  •  payroll by pay date  •  equity by vest date", len(payDates))),
				tableHeaderStyle.Render(fmt.Sprintf(detailColumns, "EARNINGS", "AMOUNT", "WITHHOLDINGS / VALUE", "AMOUNT")),
				payrollRowStyle.Render(fmt.Sprintf(detailColumns, "Base pay", formatCents(base), "Taxes", formatCents(taxes))),
				payrollRowStyle.Render(fmt.Sprintf(detailColumns, "OT + travel", formatCents(overtime), "Employee 401(k)", formatCents(employee401K))),
				payrollRowStyle.Render(fmt.Sprintf(detailColumns, "Bonuses", formatCents(bonuses), "Employer 401(k)", formatCents(employer401K))),
				payrollRowStyle.Render(fmt.Sprintf(detailColumns, "Payroll stock", formatCents(payrollStock), "Other deductions", formatCents(otherDeductions))),
				payrollRowStyle.Render(fmt.Sprintf(detailColumns, "Other earnings", formatCents(other), "Net cash", formatCents(selectedTotal.net))),
				payrollRowStyle.Render(fmt.Sprintf(detailColumns, "Payroll gross", formatCents(selectedTotal.gross), "Vested equity", formatCents(selectedTotal.stock))),
			)
		}
	}
	lines = append(lines, "", mutedStyle.Render("* month to date  •  stock is gross vest value"))
	return panel(strings.Join(lines, "\n"), width)
}

func (m Model) visiblePayrollRange() (int, int) {
	count := len(m.payrollStatements)
	limit := 8
	if m.height > 0 {
		if m.height < 30 {
			limit = max(3, min(7, m.height-20))
		} else {
			limit = max(4, min(12, m.height-23))
		}
	}
	if count <= limit {
		return 0, count
	}
	start := max(0, min(m.payrollCursor-limit/2, count-limit))
	return start, start + limit
}

func (m Model) renderPayrollView() string {
	width := m.contentWidth()
	if m.importMode != importNone {
		return m.renderImportView()
	}
	if len(m.payrollStatements) == 0 {
		empty := renderFinanceTabs(payrollView) + "\n\n" + panel(sectionStyle.Render("NO STATEMENTS YET")+"\n\n"+mutedStyle.Render("Press i to import a payroll PDF or stock vesting file."), width)
		return m.shell("Finances", "Accounts, income, and spending in one place.", empty, help("i", "import", "tab", "next section", "q", "back", "esc", "quit"))
	}
	if m.payrollDeleteMode {
		statement := m.payrollStatements[m.payrollCursor]
		message := warningStyle.Render("Delete this payroll statement?") + "\n\n" +
			fmt.Sprintf("%-13s %s\n", "Pay date", statement.PayDate.Format("Jan 02, 2006")) +
			fmt.Sprintf("%-13s %s\n", "Pay period", statement.PeriodStart.Format("Jan 02")+" – "+statement.PeriodEnd.Format("Jan 02, 2006")) +
			fmt.Sprintf("%-13s %s", "Net pay", valueStyle.Render(formatCents(statement.NetCents))) + "\n\n" +
			mutedStyle.Render("This removes the imported record from local SQLite.")
		if m.payrollError != "" {
			message += "\n\n" + errorStyle.Render("! "+m.payrollError)
		}
		body := activePanelStyle.BorderForeground(colorDanger).Width(max(42, min(64, width-6))).Render(message)
		return m.shell("Confirm deletion", "Review the statement before removing it.", body, help("y", "delete", "n/q/esc", "cancel", "ctrl+c", "quit"))
	}

	gross, taxes, deductions, net := m.payrollTotals()
	var summary string
	if width >= 92 {
		cardWidth := width / 4
		summary = lipgloss.JoinHorizontal(lipgloss.Top,
			metricCardAccent("Gross", formatCents(gross), "year to date", cardWidth, colorInfo),
			metricCardAccent("Taxes", formatCents(taxes), "year to date", cardWidth, colorTax),
			metricCardAccent("Deductions", formatCents(deductions), "year to date", cardWidth, colorGold),
			metricCardAccent("Net", formatCents(net), fmt.Sprintf("%d statements", len(m.payrollStatements)), cardWidth, colorSecondary),
		)
	} else {
		summary = panel(
			sectionStyle.Render("YEAR TO DATE")+"\n"+
				fmt.Sprintf("Gross %s   Net %s   %s", formatCents(gross), valueStyle.Render(formatCents(net)), mutedStyle.Render(fmt.Sprintf("%d statements", len(m.payrollStatements)))),
			width,
		)
	}
	stockValue, grossShares, netShares := m.stockVestTotals(time.Now().Year())
	equityContent := ""
	if grossShares > 0 {
		equityContent = sectionStyle.Foreground(colorPrimary).Render("VESTED EQUITY YTD") + "\n" +
			fmt.Sprintf("GOOG  Gross %s  •  %s shares vested  •  %s deposited  •  %s withheld",
				valueStyle.Render(formatCents(stockValue)),
				formatShares(grossShares),
				formatShares(netShares),
				formatShares(grossShares-netShares),
			)
	}
	employee401K, employer401K := m.retirementTotals(time.Now().Year())
	retirementContent := ""
	if employee401K > 0 || employer401K > 0 {
		retirementContent = sectionStyle.Foreground(colorGold).Render("RETIREMENT YTD") + "\n" +
			fmt.Sprintf("Employee %s  •  Employer %s  •  Total %s",
				valueStyle.Render(formatCents(employee401K)),
				valueStyle.Render(formatCents(employer401K)),
				valueStyle.Render(formatCents(employee401K+employer401K)),
			)
	}

	start, end := m.visiblePayrollRange()
	wide := width >= 110
	// Both tables have fairly wide fixed columns. Keep them side by side only
	// when neither panel needs to wrap; narrower terminals get a clean stack.
	chartBeside := width >= 196
	listWidth := width
	chartWidth := width
	if chartBeside {
		// The compensation table needs enough interior room for all six
		// columns after the panel's border and horizontal padding are applied.
		chartWidth = min(82, max(78, width-108))
		listWidth = width - chartWidth - 2
	}
	rows := make([]string, 0, end-start+2)
	if wide {
		rows = append(rows, tableHeaderStyle.Render(fmt.Sprintf("  %-13s  %-23s  %12s  %11s  %12s  %13s", "PAY DATE", "PAY PERIOD", "GROSS", "TAXES", "NET", "YTD NET")))
	} else {
		rows = append(rows, tableHeaderStyle.Render(fmt.Sprintf("  %-13s  %-23s  %12s", "PAY DATE", "PAY PERIOD", "NET")))
	}
	if start > 0 {
		rows = append(rows, mutedStyle.Render(fmt.Sprintf("  ↑ %d newer statements", start)))
	}
	runningNet := make([]int64, len(m.payrollStatements))
	var cumulativeNet int64
	runningYear := 0
	for i := len(m.payrollStatements) - 1; i >= 0; i-- {
		if year := m.payrollStatements[i].PayDate.Year(); year != runningYear {
			cumulativeNet = 0
			runningYear = year
		}
		cumulativeNet += m.payrollStatements[i].NetCents
		runningNet[i] = cumulativeNet
	}
	for i := start; i < end; i++ {
		statement := m.payrollStatements[i]
		period := statement.PeriodStart.Format("Jan 02") + " – " + statement.PeriodEnd.Format("Jan 02, 2006")
		line := fmt.Sprintf("%-13s  %-23s  %12s", statement.PayDate.Format("Jan 02, 2006"), period, formatCents(statement.NetCents))
		if wide {
			line = fmt.Sprintf(
				"%-13s  %-23s  %12s  %11s  %12s  %13s",
				statement.PayDate.Format("Jan 02, 2006"),
				period,
				formatCents(statement.GrossCents),
				formatCents(statement.TaxesCents),
				formatCents(statement.NetCents),
				formatCents(runningNet[i]),
			)
		}
		if i == m.payrollCursor {
			rows = append(rows, selectedRowStyle.Foreground(colorSecondary).Width(max(40, listWidth-12)).Render("› "+line))
		} else {
			rows = append(rows, payrollRowStyle.Width(max(40, listWidth-12)).Render("  "+line))
		}
	}
	if end < len(m.payrollStatements) {
		rows = append(rows, mutedStyle.Render(fmt.Sprintf("  ↓ %d older statements", len(m.payrollStatements)-end)))
	}
	listPanel := panel(sectionStyle.Foreground(colorInfo).Render("STATEMENTS")+"\n\n"+strings.Join(rows, "\n"), listWidth)

	chart := m.renderMonthlyIncomeChart(chartWidth)

	body := renderFinanceTabs(payrollView) + "\n\n" + summary + "\n\n"
	if equityContent != "" && retirementContent != "" && width >= 110 {
		body += lipgloss.JoinHorizontal(lipgloss.Top,
			panel(equityContent, width/2),
			"  ",
			panel(retirementContent, width-width/2-2),
		) + "\n\n"
	} else {
		if equityContent != "" {
			body += panel(equityContent, width) + "\n\n"
		}
		if retirementContent != "" {
			body += panel(retirementContent, width) + "\n\n"
		}
	}
	if chartBeside {
		body += lipgloss.JoinHorizontal(lipgloss.Top, listPanel, "  ", chart)
	} else {
		body += listPanel
		if m.height == 0 || m.height >= 40 {
			body += "\n\n" + chart
		}
	}
	return m.shell("Finances", "Accounts, income, and spending in one place.", body, help("i", "import", "tab", "next section", "↑/↓", "navigate", "g/G", "first/last", "d", "delete", "q", "back", "esc", "quit"))
}

func (m Model) renderImportView() string {
	width := m.contentWidth()
	title := sectionStyle.Foreground(colorInfo).Render("IMPORT FINANCIAL DATA")
	var content, footer string

	switch m.importMode {
	case importChooseType:
		payrollLine := payrollRowStyle.Width(max(42, width-12)).Render("  Payroll statement (PDF)")
		stockLine := payrollRowStyle.Width(max(42, width-12)).Render("  Schwab stock vesting history (TSV)")
		if m.importKind == importPayroll {
			payrollLine = selectedRowStyle.Foreground(colorSecondary).Width(max(42, width-12)).Render("› Payroll statement (PDF)")
		} else {
			stockLine = selectedRowStyle.Foreground(colorSecondary).Width(max(42, width-12)).Render("› Schwab stock vesting history (TSV)")
		}
		content = title + "\n\n" + mutedStyle.Render("Choose the type of local file you want Life to read.") + "\n\n" + payrollLine + "\n" + stockLine
		footer = help("↑/↓", "select", "enter", "continue", "q/esc", "cancel")

	case importEnterPath:
		kind := "payroll PDF"
		if m.importKind == importStock {
			kind = "Schwab TSV"
		}
		content = title + "\n\n" + sectionStyle.Render("FILE PATH") + "\n" + inputBoxStyle.Render(m.importPath.View()) + "\n\n" + mutedStyle.Render("Paste the full path to the "+kind+". Quoted Windows paths are accepted.")
		if m.importError != "" {
			content += "\n\n" + errorStyle.Render("! "+m.importError)
		}
		footer = help("enter", "preview", "esc", "cancel")

	case importLoading:
		content = title + "\n\n" + valueStyle.Render("Reading file…") + "\n" + mutedStyle.Render(m.importPath.Value())
		footer = help("esc", "cancel")

	case importConfirm:
		content = title + "\n\n" + sectionStyle.Foreground(colorGold).Render("IMPORT PREVIEW") + "\n\n"
		if m.importKind == importPayroll && m.pendingPayroll != nil {
			s := m.pendingPayroll
			content += fmt.Sprintf("%-18s %s\n%-18s %s\n%-18s %s\n%-18s %s\n%-18s %s\n%-18s %s",
				"Pay date", s.PayDate.Format("Jan 02, 2006"),
				"Pay period", s.PeriodStart.Format("Jan 02")+" – "+s.PeriodEnd.Format("Jan 02, 2006"),
				"Gross", formatCents(s.GrossCents),
				"Taxes", formatCents(s.TaxesCents),
				"Deductions", formatCents(s.DeductionsCents),
				"Net", valueStyle.Render(formatCents(s.NetCents)))
		} else {
			var grossValue, grossShares, netShares int64
			for _, vest := range m.pendingStockVests {
				grossValue += vest.GrossValueCents
				grossShares += vest.GrossSharesMicros
				netShares += vest.NetSharesMicros
			}
			content += fmt.Sprintf("%-18s %d\n%-18s %s\n%-18s %s\n%-18s %s\n%-18s %s",
				"Vest events", len(m.pendingStockVests),
				"Gross vest value", formatCents(grossValue),
				"Shares vested", formatShares(grossShares),
				"Shares deposited", formatShares(netShares),
				"Shares withheld", formatShares(grossShares-netShares))
		}
		content += "\n\n" + mutedStyle.Render("Only the previewed financial fields are saved to your local database.")
		if m.importError != "" {
			content += "\n\n" + errorStyle.Render("! "+m.importError)
		}
		footer = help("y/enter", "import", "n/q/esc", "cancel")

	case importResult:
		content = title + "\n\n" + valueStyle.Render("✓ "+m.importMessage)
		footer = help("enter/q/esc", "return to payroll")
	}

	body := renderFinanceTabs(payrollView) + "\n\n" + activePanelStyle.Width(max(48, width-6)).Render(content)
	return m.shell("Finances", "Import payroll and vested equity locally.", body, footer)
}

type namedSpendingTotal struct {
	name  string
	total int64
}

func (m Model) spendingCategoryTotals() []namedSpendingTotal {
	byCategory := make(map[string]int64)
	for _, transaction := range m.spendingTransactions {
		if transaction.IsPayment() {
			continue
		}
		byCategory[transaction.Category] += transaction.AmountCents
	}
	totals := make([]namedSpendingTotal, 0, len(byCategory))
	for category, total := range byCategory {
		totals = append(totals, namedSpendingTotal{name: category, total: total})
	}
	sort.Slice(totals, func(i, j int) bool { return totals[i].total > totals[j].total })
	return totals
}

func (m Model) renderSpendingView() string {
	width := m.contentWidth()
	if len(m.spendingTransactions) == 0 {
		body := renderFinanceTabs(spendingView) + "\n\n" + panel(sectionStyle.Render("NO SPENDING DATA")+"\n\n"+mutedStyle.Render("Use scripts/import_discover.py to import a Discover year-to-date summary."), width)
		return m.shell("Finances", "Accounts, income, and spending in one place.", body, help("tab", "next section", "q", "back", "esc", "quit"))
	}

	now := time.Now()
	var netSpending, currentMonth, refunds int64
	var purchaseCount int
	for _, transaction := range m.spendingTransactions {
		if transaction.IsPayment() {
			continue
		}
		netSpending += transaction.AmountCents
		if transaction.AmountCents < 0 {
			refunds -= transaction.AmountCents
		} else {
			purchaseCount++
		}
		if transaction.TransactionDate.Year() == now.Year() && transaction.TransactionDate.Month() == now.Month() {
			currentMonth += transaction.AmountCents
		}
	}
	cardWidth := width / 4
	summary := lipgloss.JoinHorizontal(lipgloss.Top,
		metricCardAccent("Card spending", formatCents(netSpending), "year to date", cardWidth, colorTax),
		metricCardAccent("This month", formatCents(currentMonth), "Discover only", cardWidth, colorGold),
		metricCardAccent("Purchases", fmt.Sprintf("%d", purchaseCount), "payments excluded", cardWidth, colorInfo),
		metricCardAccent("Refunds", formatCents(refunds), "reduces spending", cardWidth, colorSecondary),
	)

	categoryRows := make([]string, 0)
	for _, category := range m.spendingCategoryTotals() {
		categoryRows = append(categoryRows, payrollRowStyle.Render(fmt.Sprintf("%-28s %13s", category.name, formatCents(category.total))))
	}
	categories := panel(sectionStyle.Foreground(colorGold).Render("SPENDING BY CATEGORY")+"\n\n"+strings.Join(categoryRows, "\n"), width/2)

	monthly := m.spendingByMonth()
	months := make([]time.Time, 0, len(monthly))
	for month := range monthly {
		months = append(months, month)
	}
	sort.Slice(months, func(i, j int) bool { return months[i].Before(months[j]) })
	monthRows := []string{tableHeaderStyle.Render(fmt.Sprintf("%-10s %14s", "MONTH", "SPENDING"))}
	for _, month := range months {
		monthRows = append(monthRows, payrollRowStyle.Render(fmt.Sprintf("%-10s %14s", month.Format("January"), formatCents(monthly[month]))))
	}
	monthlyPanel := panel(sectionStyle.Foreground(colorInfo).Render("MONTHLY TOTALS")+"\n\n"+strings.Join(monthRows, "\n"), width-width/2-2)

	limit := 10
	start := max(0, min(m.spendingCursor-limit/2, len(m.spendingTransactions)-limit))
	end := min(len(m.spendingTransactions), start+limit)
	transactionRows := []string{tableHeaderStyle.Render(fmt.Sprintf("  %-12s %-48s %-26s %13s", "DATE", "DESCRIPTION", "CATEGORY", "AMOUNT"))}
	for i := start; i < end; i++ {
		transaction := m.spendingTransactions[i]
		description := transaction.Description
		if len(description) > 46 {
			description = description[:45] + "…"
		}
		line := fmt.Sprintf("%-12s %-48s %-26s %13s", transaction.TransactionDate.Format("Jan 02, 2006"), description, transaction.Category, formatCents(transaction.AmountCents))
		if i == m.spendingCursor {
			transactionRows = append(transactionRows, selectedRowStyle.Foreground(colorSecondary).Width(max(100, width-12)).Render("› "+line))
		} else {
			transactionRows = append(transactionRows, payrollRowStyle.Width(max(100, width-12)).Render("  "+line))
		}
	}
	recent := panel(sectionStyle.Foreground(colorSecondary).Render("TRANSACTIONS")+"\n\n"+strings.Join(transactionRows, "\n"), width)
	body := renderFinanceTabs(spendingView) + "\n\n" + summary + "\n\n" + lipgloss.JoinHorizontal(lipgloss.Top, categories, "  ", monthlyPanel) + "\n\n" + recent
	return m.shell("Finances", "Accounts, income, and spending in one place.", body, help("tab", "next section", "↑/↓", "navigate", "g/G", "first/last", "q", "back", "esc", "quit"))
}

func renderGoalsTabs(active view) string {
	projectionTab := mutedStyle.Render("  NET WORTH PROJECTION  ")
	purchasesTab := mutedStyle.Render("  MAJOR PURCHASES  ")
	selected := lipgloss.NewStyle().Bold(true).Foreground(colorBackground).Background(colorSecondary)
	if active == projectionView {
		projectionTab = selected.Render("  NET WORTH PROJECTION  ")
	} else {
		purchasesTab = selected.Render("  MAJOR PURCHASES  ")
	}
	return projectionTab + "  " + purchasesTab
}

func (m Model) renderProjectionView() string {
	width := m.contentWidth()
	settings := m.projectionSettings
	cash, investments, retirement := m.financeTotals()
	result := projection.Calculate(int64((investments+retirement)*100), int64(cash*100), settings, time.Now(), []int{35, 40, 50, 60, 65})

	milestoneCards := make([]string, 0, len(result.Milestones))
	cardWidth := max(24, width/max(1, len(result.Milestones)))
	for _, milestone := range result.Milestones {
		milestoneCards = append(milestoneCards, metricCardAccent(
			fmt.Sprintf("Age %d", milestone.Age),
			formatCents(milestone.NetWorthCents),
			fmt.Sprintf("in %d • today's dollars", milestone.Year),
			cardWidth,
			colorSecondary,
		))
	}
	var milestones string
	if width >= 125 {
		milestones = lipgloss.JoinHorizontal(lipgloss.Top, milestoneCards...)
	} else {
		milestones = lipgloss.JoinVertical(lipgloss.Left, milestoneCards...)
	}

	type assumptionRow struct {
		label string
		value string
		note  string
	}
	assumptions := []assumptionRow{
		{"Monthly spending", formatCents(settings.MonthlySpendingCents), "$100 steps"},
		{"Base salary", fmt.Sprintf("%s (%s/hr)", formatCents(settings.BaseSalaryCents), formatCents((settings.BaseSalaryCents+1040)/2080)), "$1,000 steps"},
		{"Annual bonus", fmt.Sprintf("%.1f%%", settings.BonusRate*100), "percentage of base"},
		{"Real income growth", fmt.Sprintf("%.1f%%", settings.RealIncomeGrowthRate*100), "after inflation"},
		{"Real return", fmt.Sprintf("%.1f%%", settings.RealReturnRate*100), "after inflation"},
		{"HYSA real return", fmt.Sprintf("%.1f%%", settings.CashRealReturnRate*100), "4% APY less inflation"},
		{"Net stock invested", formatCents(settings.MonthlyNetStockCents) + "/mo", "after tax"},
		{"Net-pay rate", fmt.Sprintf("%.0f%%", settings.NetPayRate*100), "historical estimate"},
		{"Surplus net pay invested", fmt.Sprintf("%.0f%%", settings.SurplusInvestedRate*100), "after spending + IRA"},
		{"Employee 401(k)", formatCents(settings.Employee401KAnnualCents) + "/yr", "annual maximum"},
		{"Employer match", fmt.Sprintf("%.0f%%", settings.EmployerMatchRate*100), "true-up enabled"},
		{"IRA contribution", formatCents(settings.IRAAnnualCents) + "/yr", "annual maximum"},
		{"Retirement age", fmt.Sprintf("%d", settings.RetirementAge), "contributions stop"},
	}
	assumptionLines := []string{tableHeaderStyle.Render(fmt.Sprintf("  %-23s %16s   %-22s", "ASSUMPTION", "VALUE", "DETAIL"))}
	for i, assumption := range assumptions {
		line := fmt.Sprintf("%-23s %16s   %-22s", assumption.label, assumption.value, assumption.note)
		if i == m.projectionCursor {
			assumptionLines = append(assumptionLines, selectedRowStyle.Foreground(colorSecondary).Width(max(58, width/2-12)).Render("› "+line))
		} else {
			assumptionLines = append(assumptionLines, payrollRowStyle.Width(max(58, width/2-12)).Render("  "+line))
		}
	}
	assumptionPanel := panel(sectionStyle.Foreground(colorInfo).Render("ADJUSTABLE ASSUMPTIONS")+"\n\n"+strings.Join(assumptionLines, "\n"), width/2)

	annualLines := []string{
		sectionStyle.Foreground(colorGold).Render("FIRST-YEAR INVESTING PLAN"),
		mutedStyle.Render("How take-home pay becomes savings, then combines with retirement and stock contributions."),
		"",
		tableHeaderStyle.Padding(0).Render("PAYCHECK CASH FLOW"),
		fmt.Sprintf("%-30s %16s", "Gross base + bonus", formatCents(result.AnnualGrossPayCents)),
		fmt.Sprintf("%-30s %16s", fmt.Sprintf("Estimated take-home (%.0f%%)", settings.NetPayRate*100), formatCents(result.AnnualNetPayCents)),
		fmt.Sprintf("%-30s %16s", fmt.Sprintf("Annual spending (%s/mo)", formatCents(settings.MonthlySpendingCents)), "−"+formatCents(result.AnnualSpendingCents)),
		fmt.Sprintf("%-30s %16s", "Roth IRA funded", "−"+formatCents(result.AnnualIRAContribution)),
		mutedStyle.Render(strings.Repeat("─", 48)),
		lipgloss.NewStyle().Bold(true).Foreground(colorSecondary).Render(fmt.Sprintf("%-30s %16s", "CASH SURPLUS", formatCents(result.AnnualPostIRASurplusCents))),
		"",
		tableHeaderStyle.Padding(0).Render("INVESTMENT SOURCES"),
		fmt.Sprintf("%-30s %16s", "Employee 401(k)", formatCents(result.AnnualEmployee401KCents)),
		fmt.Sprintf("%-30s %16s", "Employer match", formatCents(result.AnnualEmployer401KCents)),
		fmt.Sprintf("%-30s %16s", "Roth IRA", formatCents(result.AnnualIRAContribution)),
		fmt.Sprintf("%-30s %16s", "Taxable brokerage", formatCents(result.AnnualTaxableContribution)),
		fmt.Sprintf("%-30s %16s", "Net vested stock", formatCents(result.AnnualStockContribution)),
		mutedStyle.Render(strings.Repeat("─", 48)),
		lipgloss.NewStyle().Bold(true).Foreground(colorGold).Render(fmt.Sprintf("%-30s %16s", "TOTAL INVESTED", formatCents(
			result.AnnualEmployee401KCents+result.AnnualEmployer401KCents+result.AnnualIRAContribution+result.AnnualTaxableContribution+result.AnnualStockContribution,
		))),
	}
	if result.AnnualUninvestedCents > 0 {
		annualLines = append(annualLines, mutedStyle.Render(fmt.Sprintf("Uninvested surplus retained: %s", formatCents(result.AnnualUninvestedCents))))
	}
	annualPanel := panel(strings.Join(annualLines, "\n"), width-width/2-2)

	currentAge := projection.AgeOn(settings.BirthDate, time.Now())
	chartAges := make([]int, 0, 66-currentAge)
	for age := currentAge + 1; age <= 65; age++ {
		chartAges = append(chartAges, age)
	}
	chartResult := projection.Calculate(int64((investments+retirement)*100), int64(cash*100), settings, time.Now(), chartAges)
	startingNetWorth := int64(m.netWorth() * 100)
	chartPoints := append([]projection.Milestone{{Age: currentAge, Year: time.Now().Year(), NetWorthCents: startingNetWorth, ContributionsCents: startingNetWorth}}, chartResult.Milestones...)
	chartPanelWidth := min(width, 132)
	chart := renderNetWorthLineChart(chartPoints, settings.RetirementAge, chartPanelWidth)
	chart = lipgloss.PlaceHorizontal(width, lipgloss.Center, chart)

	body := renderGoalsTabs(projectionView) + "\n\n" + milestones + "\n\n" + lipgloss.JoinHorizontal(lipgloss.Top, assumptionPanel, "  ", annualPanel) + "\n\n" + chart
	if m.projectionError != "" {
		body += "\n\n" + errorStyle.Render("! "+m.projectionError)
	}
	return m.shell("Goals & Projections", "Explore major purchases and your future net worth in today's dollars.", body, help("tab", "major purchases", "↑/↓", "select", "←/→", "adjust", "r", "reset", "q", "back"))
}

func renderNetWorthLineChart(points []projection.Milestone, retirementAge, width int) string {
	if len(points) < 2 {
		return panel(sectionStyle.Foreground(colorSecondary).Render("NET WORTH TRAJECTORY")+"\n\n"+mutedStyle.Render("Not enough projection data."), width)
	}
	chartWidth := max(48, width-14)
	chartHeight := 16
	var maximum float64
	for _, point := range points {
		maximum = math.Max(maximum, float64(point.NetWorthCents)/100)
	}
	if maximum <= 0 {
		maximum = 1
	}
	scale := 100_000.0
	if maximum >= 1_000_000 {
		scale = 1_000_000
	}
	maximum = math.Ceil(maximum/scale) * scale

	axisStyle := legacyLipgloss.NewStyle().Foreground(legacyLipgloss.Color("#59617F"))
	labelStyle := legacyLipgloss.NewStyle().Foreground(legacyLipgloss.Color("#AFB5CA"))
	netWorthStyle := legacyLipgloss.NewStyle().Foreground(legacyLipgloss.Color("#69E9BF"))
	contributionStyle := legacyLipgloss.NewStyle().Foreground(legacyLipgloss.Color("#65CCFF"))
	retirementStyle := legacyLipgloss.NewStyle().Bold(true).Foreground(legacyLipgloss.Color("#FFD166"))

	chart := linechart.New(
		chartWidth,
		chartHeight,
		float64(points[0].Age),
		float64(points[len(points)-1].Age),
		0,
		maximum,
		linechart.WithXYSteps(max(6, chartWidth/6), 3),
		linechart.WithXLabelFormatter(func(_ int, value float64) string {
			return fmt.Sprintf("%.0f", value)
		}),
		linechart.WithYLabelFormatter(func(_ int, value float64) string {
			return compactDollars(value)
		}),
		linechart.WithStyles(axisStyle, labelStyle, netWorthStyle),
	)
	chart.DrawXYAxisAndLabel()
	for i := 1; i < len(points); i++ {
		previous := points[i-1]
		current := points[i]
		chart.DrawBrailleLineWithStyle(
			canvas.Float64Point{X: float64(previous.Age), Y: float64(previous.ContributionsCents) / 100},
			canvas.Float64Point{X: float64(current.Age), Y: float64(current.ContributionsCents) / 100},
			contributionStyle,
		)
		chart.DrawBrailleLineWithStyle(
			canvas.Float64Point{X: float64(previous.Age), Y: float64(previous.NetWorthCents) / 100},
			canvas.Float64Point{X: float64(current.Age), Y: float64(current.NetWorthCents) / 100},
			netWorthStyle,
		)
	}
	for _, point := range points {
		if point.Age == retirementAge {
			chart.DrawRuneWithStyle(
				canvas.Float64Point{X: float64(point.Age), Y: float64(point.NetWorthCents) / 100},
				'◆',
				retirementStyle,
			)
			break
		}
	}

	lines := []string{
		sectionStyle.Foreground(colorSecondary).Render("NET WORTH TRAJECTORY"),
		mutedStyle.Render("Projected annually in today's dollars • X-axis is age"),
		"",
		chart.View(),
		valueStyle.Render("━━ net worth") + "    " + lipgloss.NewStyle().Foreground(colorInfo).Render("━━ net contributions") + "    " + lipgloss.NewStyle().Foreground(colorGold).Render("◆ retirement at age "+fmt.Sprint(retirementAge)),
	}
	return panel(strings.Join(lines, "\n"), width)
}

func compactDollars(dollars float64) string {
	switch {
	case dollars >= 1_000_000:
		return fmt.Sprintf("$%.1fM", dollars/1_000_000)
	case dollars >= 1_000:
		return fmt.Sprintf("$%.0fK", dollars/1_000)
	default:
		return fmt.Sprintf("$%.0f", dollars)
	}
}

func monthsUntil(from, target time.Time) float64 {
	days := target.Sub(from).Hours() / 24
	if days <= 0 {
		return 0
	}
	return days / 30.4375
}

func combinedGoalMonthlyRequirement(goals []models.FinancialGoal, now time.Time) int64 {
	ordered := append([]models.FinancialGoal(nil), goals...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].TargetDate.Before(ordered[j].TargetDate)
	})

	var cumulativeRemaining, requiredMonthly int64
	for _, goal := range ordered {
		cumulativeRemaining += max(0, goal.TargetCents-goal.SavedCents)
		months := monthsUntil(now, goal.TargetDate)
		candidate := cumulativeRemaining
		if months > 0 {
			candidate = int64(math.Ceil(float64(cumulativeRemaining) / months))
		}
		requiredMonthly = max(requiredMonthly, candidate)
	}
	return requiredMonthly
}

func renderGoalProgress(saved, target int64, width int) string {
	width = max(10, width)
	ratio := 0.0
	if target > 0 {
		ratio = min(1, float64(saved)/float64(target))
	}
	filled := int(math.Round(ratio * float64(width)))
	return lipgloss.NewStyle().Foreground(colorSecondary).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(colorBorder).Render(strings.Repeat("░", width-filled))
}

func (m Model) renderFinancialGoalsView() string {
	width := m.contentWidth()
	if len(m.financialGoals) == 0 {
		body := renderGoalsTabs(financialGoalsView) + "\n\n" + panel(mutedStyle.Render("No major-purchase goals yet."), width)
		return m.shell("Goals & Projections", "Explore major purchases and your future net worth in today's dollars.", body, help("tab", "net worth projection", "q", "back"))
	}

	cash, investments, retirement := m.financeTotals()
	projectionSummary := projection.Calculate(int64((investments+retirement)*100), int64(cash*100), m.projectionSettings, time.Now(), nil)
	availableCashMonthly := projectionSummary.AnnualPostIRASurplusCents / 12
	availableStockMonthly := min(m.projectionSettings.MonthlyGoalStockCents, m.projectionSettings.MonthlyNetStockCents)
	totalAvailableMonthly := availableCashMonthly + availableStockMonthly

	now := time.Now()
	goalPanels := make([]string, 0, len(m.financialGoals))
	var totalPlannedMonthly, totalTarget, totalSaved int64
	panelWidth := width
	if width >= 100 {
		panelWidth = width / len(m.financialGoals)
	}
	for i, goal := range m.financialGoals {
		remaining := max(0, goal.TargetCents-goal.SavedCents)
		months := monthsUntil(now, goal.TargetDate)
		requiredMonthly := remaining
		if months > 0 {
			requiredMonthly = int64(math.Ceil(float64(remaining) / months))
		}
		totalPlannedMonthly += goal.PlannedMonthlyCents
		totalTarget += goal.TargetCents
		totalSaved += goal.SavedCents

		status := "Set a monthly contribution"
		statusColor := colorWarning
		if remaining == 0 {
			status = "Fully funded"
			statusColor = colorSecondary
		} else if goal.PlannedMonthlyCents >= requiredMonthly {
			status = "On track"
			statusColor = colorSecondary
		} else if goal.PlannedMonthlyCents > 0 {
			status = fmt.Sprintf("Increase by %s/mo", formatCents(requiredMonthly-goal.PlannedMonthlyCents))
		}
		content := sectionStyle.Foreground(colorInfo).Render(strings.ToUpper(goal.Name)) + "\n\n" +
			renderGoalProgress(goal.SavedCents, goal.TargetCents, max(20, panelWidth-18)) + "\n" +
			fmt.Sprintf("%-22s %14s\n", "Target", formatCents(goal.TargetCents)) +
			fmt.Sprintf("%-22s %14s\n", "Allocated", formatCents(goal.SavedCents)) +
			fmt.Sprintf("%-22s %14s\n", "Remaining", formatCents(remaining)) +
			fmt.Sprintf("%-22s %14s\n", "Target date", goal.TargetDate.Format("Jan 2006")) +
			fmt.Sprintf("%-22s %14s\n", "Time remaining", fmt.Sprintf("%.1f months", months)) +
			fmt.Sprintf("%-22s %14s\n", "Required monthly", formatCents(requiredMonthly)) +
			fmt.Sprintf("%-22s %14s\n", "Planned monthly", formatCents(goal.PlannedMonthlyCents)) +
			lipgloss.NewStyle().Bold(true).Foreground(statusColor).Render(status)
		style := panelStyle.BorderForeground(colorBorder).Width(max(42, panelWidth-6))
		if i == m.financialGoalCursor {
			style = style.BorderForeground(colorSecondary)
		}
		goalPanels = append(goalPanels, style.Render(content))
	}
	totalRequiredMonthly := combinedGoalMonthlyRequirement(m.financialGoals, now)

	goalsLayout := lipgloss.JoinVertical(lipgloss.Left, goalPanels...)
	if width >= 100 {
		goalsLayout = lipgloss.JoinHorizontal(lipgloss.Top, goalPanels...)
	}
	requiredGap := max(0, totalRequiredMonthly-totalAvailableMonthly)
	plannedRemaining := totalAvailableMonthly - totalPlannedMonthly
	comparison := []string{
		sectionStyle.Foreground(colorGold).Render("COMBINED FUNDING CHECK"),
		mutedStyle.Render("Goal funding redirects money that would otherwise remain available for long-term investing."),
		"",
		fmt.Sprintf("%-37s %15s", "Combined targets", formatCents(totalTarget)),
		fmt.Sprintf("%-37s %15s", "Currently allocated", formatCents(totalSaved)),
		fmt.Sprintf("%-37s %15s", "Required monthly for deadlines", formatCents(totalRequiredMonthly)),
		fmt.Sprintf("%-37s %15s", "Available monthly cash surplus", formatCents(availableCashMonthly)),
		fmt.Sprintf("%-37s %15s", "Vested stock available for goals", formatCents(availableStockMonthly)),
		fmt.Sprintf("%-37s %15s", "Monthly funding capacity", formatCents(totalAvailableMonthly)),
		fmt.Sprintf("%-37s %15s", "Current planned contributions", "−"+formatCents(totalPlannedMonthly)),
		mutedStyle.Render(strings.Repeat("─", 54)),
		lipgloss.NewStyle().Bold(true).Foreground(colorSecondary).Render(fmt.Sprintf("%-37s %15s", "Total monthly funding available", formatCents(max(0, plannedRemaining)))),
	}
	if requiredGap > 0 {
		comparison = append(comparison, warningStyle.Render(fmt.Sprintf("Deadline funding gap: %s per month", formatCents(requiredGap))))
	} else {
		comparison = append(comparison, valueStyle.Render("Both deadlines are fundable from the selected cash and stock sources."))
	}
	if plannedRemaining >= 0 {
		comparison = append(comparison, mutedStyle.Render("Available funding is the portion not yet assigned to a goal."))
	} else {
		comparison = append(comparison, warningStyle.Render(fmt.Sprintf("Current plan exceeds surplus by %s/month", formatCents(-plannedRemaining))))
	}

	body := renderGoalsTabs(financialGoalsView) + "\n\n" + goalsLayout + "\n\n" + panel(strings.Join(comparison, "\n"), width)
	if m.financialGoalError != "" {
		body += "\n\n" + errorStyle.Render("! "+m.financialGoalError)
	}
	return m.shell("Goals & Projections", "Explore major purchases and your future net worth in today's dollars.", body, help("tab", "net worth projection", "↑/↓", "select goal", "←/→", "planned ±$100", "[/]", "allocated ±$500", "v/V", "stock ±$100", "q", "back"))
}

func (m Model) View() tea.View {
	var content string
	switch m.currentView {
	case profileSetupView:
		content = m.renderProfileSetupView()
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
	case spendingView:
		content = m.renderSpendingView()
	case projectionView:
		content = m.renderProjectionView()
	case financialGoalsView:
		content = m.renderFinancialGoalsView()
	}

	view := tea.NewView(content)
	view.AltScreen = true
	view.BackgroundColor = colorBackground
	view.ForegroundColor = colorText
	view.WindowTitle = "Life"
	return view
}
