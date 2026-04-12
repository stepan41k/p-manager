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
	case createState:
		content = m.createView()
	default:
		content = "Неизвестное состояние"
	}

	return tea.NewView(content)
}

func (m *Model) authView() string {
	title := m.styles.Title.Bold(true).Render("------------------------------------------------- Password Manager -------------------------------------------------")

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
    s := m.styles

    drawRow := func(label, value string, valueStyle lipgloss.Style) string {
        return fmt.Sprintf("%s %s", s.DetailLabel.Render(label), valueStyle.Render(value))
    }

    content := lipgloss.JoinVertical(
        lipgloss.Left,
        s.Title.Render("ДЕТАЛИ АККАУНТА"),
        "",
        drawRow("Сервис:", m.selectedItem.Resource, s.DetailValue),
        drawRow("Логин:", m.selectedItem.Username, s.DetailValue),
        drawRow("Email:", m.selectedItem.Email, s.DetailValue),
        drawRow("Пароль:", m.selectedItem.Password, s.DetailKey),
        "",
        lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("240")).Render("← esc для возврата | c для копирования"),
    )

    // Оборачиваем весь контент в рамку карточки
    return s.Card.Render(content)
}

func (m *Model) createView() string {
	var s string
	
	s += m.styles.Title.Bold(true).Foreground(lipgloss.Color("205")).Render("Добавление нового сервиса")
	s += "\n\n"

	for i := range m.inputs {
		s += m.inputs[i].View() + "\n"
	}

	s += "\n"
	s += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).
		Render("(стрелки: переход | g: генерация | enter: сохранить | esc: отмена)")

	if m.errorMessage != "" {
		s += "\n" + m.errorMessage
	}

	return s
}
