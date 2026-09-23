package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"life/internal/models"
	payrollparser "life/internal/payroll"
	stockparser "life/internal/stock"
	"life/internal/storage"
)

type importParsedMsg struct {
	kind       importKind
	path       string
	payroll    *models.PayrollStatement
	stockVests []models.StockVest
	err        error
}

func cleanImportPath(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}
	}
	return filepath.Clean(strings.TrimSpace(value))
}

func parseImport(kind importKind, path string) tea.Cmd {
	return func() tea.Msg {
		path = cleanImportPath(path)
		message := importParsedMsg{kind: kind, path: path}
		if path == "." || path == "" {
			message.err = errors.New("choose a file to import")
			return message
		}

		switch kind {
		case importPayroll:
			statement, err := payrollparser.ParsePDF(path)
			if err != nil {
				message.err = err
				return message
			}
			message.payroll = &statement
		case importStock:
			file, err := os.Open(path)
			if err != nil {
				message.err = fmt.Errorf("open stock file: %w", err)
				return message
			}
			vests, parseErr := stockparser.ParseTSV(file)
			closeErr := file.Close()
			if parseErr != nil {
				message.err = parseErr
				return message
			}
			if closeErr != nil {
				message.err = fmt.Errorf("close stock file: %w", closeErr)
				return message
			}
			if len(vests) == 0 {
				message.err = errors.New("no restricted-stock rows found")
				return message
			}
			message.stockVests = vests
		}
		return message
	}
}

func (m *Model) beginImport() tea.Cmd {
	m.importMode = importChooseType
	m.importKind = importPayroll
	m.importError = ""
	m.importMessage = ""
	m.pendingPayroll = nil
	m.pendingStockVests = nil
	m.importPath.SetValue("")
	m.importPath.Blur()
	return nil
}

func (m *Model) cancelImport() {
	m.importMode = importNone
	m.importError = ""
	m.importMessage = ""
	m.pendingPayroll = nil
	m.pendingStockVests = nil
	m.importPath.SetValue("")
	m.importPath.Blur()
}

func (m Model) updateImportKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch m.importMode {
	case importChooseType:
		switch msg.String() {
		case "up", "down", "left", "right", "w", "s", "h", "l":
			if m.importKind == importPayroll {
				m.importKind = importStock
			} else {
				m.importKind = importPayroll
			}
		case "enter":
			m.importMode = importEnterPath
			m.importError = ""
			cmd := m.importPath.Focus()
			return m, cmd
		case "q":
			m.cancelImport()
		}

	case importEnterPath:
		if msg.String() == "enter" {
			path := cleanImportPath(m.importPath.Value())
			if path == "." || path == "" {
				m.importError = "Enter the full path to a file."
				return m, nil
			}
			m.importPath.SetValue(path)
			m.importPath.Blur()
			m.importError = ""
			m.importMode = importLoading
			return m, parseImport(m.importKind, path)
		}
		var cmd tea.Cmd
		m.importPath, cmd = m.importPath.Update(msg)
		return m, cmd

	case importConfirm:
		switch msg.String() {
		case "y", "enter":
			m.savePendingImport()
			return m, nil
		case "n", "q":
			m.cancelImport()
		}

	case importResult:
		switch msg.String() {
		case "enter", "q":
			m.cancelImport()
		}
	}
	return m, nil
}

func (m *Model) handleImportParsed(message importParsedMsg) {
	if m.importMode != importLoading || message.kind != m.importKind {
		return
	}
	if message.err != nil {
		m.importMode = importEnterPath
		m.importError = message.err.Error()
		m.importPath.Focus()
		return
	}
	m.pendingPayroll = message.payroll
	m.pendingStockVests = message.stockVests
	m.importMode = importConfirm
	m.importError = ""
}

func (m *Model) savePendingImport() {
	switch m.importKind {
	case importPayroll:
		if m.pendingPayroll == nil {
			m.importError = "No payroll statement is ready to import."
			return
		}
		updated := false
		if _, err := m.db.AddPayrollStatement(*m.pendingPayroll); err != nil {
			if !errors.Is(err, storage.ErrPayrollAlreadyImported) {
				m.importError = fmt.Sprintf("Could not import payroll: %v", err)
				return
			}
			if err := m.db.UpdatePayrollStatement(*m.pendingPayroll); err != nil {
				m.importError = fmt.Sprintf("Could not update payroll: %v", err)
				return
			}
			updated = true
		}
		statements, err := m.db.GetPayrollStatements()
		if err != nil {
			m.importError = fmt.Sprintf("Payroll saved, but refresh failed: %v", err)
			return
		}
		m.payrollStatements = statements
		m.payrollCursor = 0
		now := time.Now()
		m.monthlyNetIncomeCents, err = m.db.NetPayrollForMonth(now.Year(), int(now.Month()))
		if err != nil {
			m.importError = fmt.Sprintf("Payroll saved, but monthly total refresh failed: %v", err)
			return
		}
		if updated {
			m.importMessage = "Existing payroll statement updated."
		} else {
			m.importMessage = "Payroll statement imported."
		}

	case importStock:
		imported, skipped := 0, 0
		for _, vest := range m.pendingStockVests {
			if _, err := m.db.AddStockVest(vest); err != nil {
				if errors.Is(err, storage.ErrStockVestAlreadyImported) {
					skipped++
					continue
				}
				m.importError = fmt.Sprintf("Could not import stock vest: %v", err)
				return
			}
			imported++
		}
		vests, err := m.db.GetStockVests()
		if err != nil {
			m.importError = fmt.Sprintf("Stock data saved, but refresh failed: %v", err)
			return
		}
		m.stockVests = vests
		m.importMessage = fmt.Sprintf("Imported %d stock vest events; skipped %d duplicates.", imported, skipped)
	}
	m.importError = ""
	m.importMode = importResult
}
