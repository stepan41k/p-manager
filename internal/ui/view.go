package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *Model) View() tea.View {
	var content string

	switch m.state {
	case authState:
		content = m.authView()
	case vaultState:
		content = m.vaultView()
	case entryState:
		content = m.entryView()
	}

	return tea.NewView(content)
	// if m.choice != "" {
	// 	return tea.NewView(m.styles.QuitText.Render(fmt.Sprintf("%s? Sounds good to me.", m.choice)))
	// }
	// if m.quitting {
	// 	return tea.NewView(m.styles.QuitText.Render("Not hungry? That’s cool."))
	// }
	// return tea.NewView("\n" + m.list.View())
}

func (m *Model) authView() string {
	title := lipgloss.NewStyle().Bold(true).Render("Password Manager")

	errStr := ""
	if m.errorMessage != "" {
		errStr += lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("\n" + m.errorMessage)
	}

	return fmt.Sprintf(
		"\n%s\n\n%s\n%s\n\n(нажмите Enter для входа)",
		title,
		m.passInput.View(),
		errStr,
	)
}

func (m *Model) vaultView() string {

	return m.vaultList.View()
}

func (m *Model) entryView() string {
	return m.vaultList.View()
}