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
	case stateDetails:
		content = m.detailsView()
	}

	return tea.NewView(content)
}

func (m *Model) authView() string {
	title := lipgloss.NewStyle().Bold(true).Render("------------------------------- Password Manager -------------------------------------------------")

	errStr := ""
	if m.errorMessage != "" {
		errStr += lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("\n" + m.errorMessage)
	}

	return fmt.Sprintf(
		"\n%s\n\n%s\n%s\n",
		title,
		m.passInput.View(),
		errStr,
	)
}

func (m *Model) vaultView() string {

	return m.vaultList.View()
}

func (m *Model) detailsView() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Underline(true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	
	s := "\n"
	s += titleStyle.Render("Детали Аккаунта") + "\n\n"
	
	s += labelStyle.Render("Сервис: ") + m.selectedItem.Resource + "\n"
	s += labelStyle.Render("Логин: ") + m.selectedItem.Username + "\n"
	s += labelStyle.Render("Email: ") + m.selectedItem.Email + "\n"
	s += labelStyle.Render("Пароль: ") + m.selectedItem.Password + "\n"
	
	s += lipgloss.NewStyle().Italic(true).Render("Esc для возврата")
	
	return s
}