package app

import (
	"fmt"
	"image/color"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	goals := sectionStyle.Render("ACTIVE GOALS") + "\n" + valueStyle.Render(fmt.Sprintf("%d in progress", m.activeGoals))
	body += "\n\n" + panel(goals, width)
	return m.shell("Dashboard", "A clear view of today and your financial momentum.", body, help("q", "back", "esc", "quit"))
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
	return m.shell("Tasks", fmt.Sprintf("%d remaining today", m.remainingTasks()), panel(strings.Join(rows, "\n"), width), help("↑/↓", "navigate", "space", "toggle", "q", "back", "esc", "quit"))
}

func (m Model) renderHabitsView() string {
	width := m.contentWidth()
	body := panel(sectionStyle.Render("WEEKLY COMPLETION")+"\n\n"+valueStyle.Render(fmt.Sprintf("%d%%", m.habitCompletion))+"\n"+mutedStyle.Render("Detailed habit tracking is coming next."), width)
	return m.shell("Habits", "Consistency compounds.", body, help("q", "back", "esc", "quit"))
}

func (m Model) renderFinanceList() string {
	width := m.contentWidth()
	summary := metricCard("Net worth", formatMoney(m.netWorth()), fmt.Sprintf("Across %d accounts", len(m.accounts)), min(44, width))
	rows := make([]string, 0, len(m.accounts))
	for i, account := range m.accounts {
		line := fmt.Sprintf("%-28s %14s", account.Name, formatMoney(account.Balance))
		if i == m.financeCursor {
			rows = append(rows, selectedRowStyle.Width(width-10).Render("› "+line))
		} else {
			rows = append(rows, rowStyle.Width(width-10).Render("  "+line))
		}
	}
	if len(rows) == 0 {
		rows = append(rows, mutedStyle.Render("No accounts yet. Press a to add one."))
	}
	body := summary + "\n\n" + sectionStyle.Render("ACCOUNTS") + "\n" + panel(strings.Join(rows, "\n"), width)
	return m.shell("Finances", "Balances you manage manually.", body, help("↑/↓", "navigate", "a", "add", "e", "edit", "d", "delete", "q", "back"))
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
	if len(m.payrollStatements) == 0 {
		empty := panel(sectionStyle.Render("NO STATEMENTS YET")+"\n\n"+mutedStyle.Render("Use import-payroll to add your first pay statement."), width)
		return m.shell("Payroll History", "A private, local record of your income.", empty, help("q", "back", "esc", "quit"))
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

	body := summary + "\n\n"
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
	return m.shell("Payroll History", "Income imported locally from your pay statements.", body, help("↑/↓", "navigate", "g/G", "first/last", "d", "delete", "q", "back", "esc", "quit"))
}

func (m Model) renderStatsView() string {
	width := m.contentWidth()
	body := panel(sectionStyle.Render("INSIGHTS")+"\n\n"+mutedStyle.Render("Your cross-category trends will live here."), width)
	return m.shell("Stats", "See how your habits, work, and money change over time.", body, help("q", "back", "esc", "quit"))
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

	view := tea.NewView(content)
	view.AltScreen = true
	view.BackgroundColor = colorBackground
	view.ForegroundColor = colorText
	view.WindowTitle = "Life"
	return view
}
