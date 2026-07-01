package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *Model) View() tea.View {
	var content string

	if m.width == 0 || m.height == 0 {
		return tea.NewView("")
	}

	switch m.state {
	case authState:
		content = m.authView()
	case vaultState:
		content = m.vaultView()
	case stateDetails:
		content = m.detailsView()
	case createState:
		content = m.createView()
	case editState:
		content = m.editView()
	default:
		content = "Unknown State"
	}

	v := tea.NewView(lipgloss.Place(
		m.width, m.height,
		lipgloss.Left, lipgloss.Top,
		content,
		lipgloss.WithWhitespaceChars(" "),
	))

	v.AltScreen = true

	return v
}

func (m *Model) authView() string {
	titleText := " PASSWORD MANAGER "
	
	header := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(lipgloss.Color("62")).
		Render(strings.Repeat("─", 5) + titleText + strings.Repeat("─", 5))

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		"",
		"Enter master-password:",
		m.passInput.View(),
	)

	if m.errorMessage != "" {
		content = lipgloss.JoinVertical(lipgloss.Center, content, "", lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(m.errorMessage))
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *Model) vaultView() string {
	return m.vaultList.View()
}

func (m *Model) detailsView() string {
	s := m.styles

	drawRow := func(label, value string, valueStyle lipgloss.Style) string {
		return fmt.Sprintf("%s %s", s.DetailLabel.Render(label), valueStyle.Render(value))
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		s.Title.Render("ACCOUNT DETAILS"),
		"",
		drawRow("Service:", m.selectedItem.Resource, s.DetailValue),
		drawRow("Login:", m.selectedItem.Username, s.DetailValue),
		drawRow("Email:", m.selectedItem.Email, s.DetailValue),
		drawRow("Password:", m.selectedItem.Password, s.DetailKey),
		"",
		lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("240")).Render("← esc для возврата | c для копирования"),
	)
	return s.Card.Render(content)
}

func (m *Model) createView() string {
	var s string

	s += "Добавление нового сервиса"
	s += "\n\n"

	for i := range m.inputs {
		s += m.inputs[i].View() + "\n"
	}

	s += "\n"
	s += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).
		Render("(стрелки: переход | ctrl+g: генерация | enter: сохранить | esc: отмена)")

	if m.errorMessage != "" {
		s += "\n" + m.errorMessage
	}

	return s
}

func (m *Model) editView() string {
	s := m.styles.Title.Bold(true).Foreground(lipgloss.Color("205")).Render("Обновление сервиса")
	s += "\n\n"

	for i := range m.inputs {
		s += m.inputs[i].View() + "\n"
	}

	s += "\n"
	s += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).
		Render("(стрелки: переход | ctrl+g: генерация | enter: сохранить | esc: отмена)")

	if m.errorMessage != "" {
		s += "\n" + m.errorMessage
	}

	return s
}
