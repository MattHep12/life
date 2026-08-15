package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type model struct {
	cursor      int
	choices     []string
	currentView string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "w":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "s":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter":
			m.currentView = m.choices[m.cursor]
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Current View: %s\n\n", m.currentView)

	for i, choice := range m.choices {
		prefix := "  "
		if m.cursor == i {
			prefix = "> "
		}

		fmt.Fprintf(&sb, "%s%s\n", prefix, choice)
	}

	return tea.View{
		Content: sb.String(),
	}
}

func main() {
	m := model{
		cursor:      0,
		choices:     []string{"Dashboard", "Tasks", "Habits", "Finances", "Stats"},
		currentView: "Dashboard",
	}
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
