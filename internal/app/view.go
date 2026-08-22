package app

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *Model) View() tea.View {
	var content string
	var footer string

	if m.width == 0 || m.height == 0 {
		return tea.NewView("")
	}

	switch m.state {
	case setupState:
		content = m.renderForm("INITIAL SETUP")
		footer = m.help.View(m.keys.Setup)
	case otpState:
		content = m.otpView()
		footer = m.help.View(m.keys.OTP)
	case authState:
		content = m.authView()
		footer = m.help.View(m.keys.Auth)
	case vaultState:
		content = m.vaultView()

		v := tea.NewView(lipgloss.Place(
			m.width, m.height,
			lipgloss.Left, lipgloss.Top,
			content,
			lipgloss.WithWhitespaceChars(" "),
		))

		v.AltScreen = true

		return v
	case detailsState:
		content = m.detailsView()
		footer = m.help.View(m.keys.Details)
	case createState:
		content = m.createView()
		footer = m.help.View(m.keys.Create)
	case editState:
		content = m.editView()
		footer = m.help.View(m.keys.Edit)
	case deleteState:
		content = m.deleteView()
		footer = m.help.View(m.keys.Delete)
	default:
		content = "Unknown State"
	}

	fullDisplay := lipgloss.JoinVertical(
		lipgloss.Center,
		content,
		"\n",
		footer,
	)

	v := tea.NewView(lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		fullDisplay,
		lipgloss.WithWhitespaceChars(" "),
	))

	v.AltScreen = true

	return v
}

func (m *Model) vaultView() string  { return m.vaultList.View() }
func (m *Model) createView() string { return m.renderForm("NEW ENTRY") }
func (m *Model) editView() string   { return m.renderForm("EDITING") }

func (m *Model) renderForm(title string) string {
	s := m.styles

	header := s.Title.
		Foreground(lipgloss.Color("205")).
		MarginBottom(1).
		Render("── " + title + " ──")

	errStr := ""
	if m.errorMessage != "" {
		errStr = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			MarginTop(1).
			MarginLeft(1).
			Render(m.errorMessage)
	}

	var labels []string
	if m.state == createState || m.state == editState {
		labels = []string{"Service:", "Email:", "Username:", "Password:"}
	}

	var inputViews []string
	for i := range m.inputs {
		prefix := "  "

		if i == m.focusIndex {
			prefix = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("> ")
		}

		label := ""
		if i < len(labels) {
			label = s.DetailLabel.Render(labels[i]) + " "
		}

		row := fmt.Sprintf("%s%s %s", prefix, label, m.inputs[i].View())
		inputViews = append(inputViews, row)
	}
	inputs := lipgloss.JoinVertical(lipgloss.Left, inputViews...)

	formContent := lipgloss.JoinVertical(lipgloss.Left, header, inputs, errStr)

	return s.Card.Render(formContent)
}

func (m *Model) authView() string {
	s := m.styles

	header := s.Title.Foreground(lipgloss.Color("205")).Render("── PASSWORD MANAGER ──")

	errStr := ""
	if m.errorMessage != "" {
		errStr = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render(m.errorMessage)
	}

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Bold(true)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		labelStyle.Render("Master Password:"),
		m.inputs[0].View(),
		errStr,
	)

	return s.Card.Render(content)
}

func (m *Model) otpView() string {
	s := m.styles
	header := s.Title.Render("📧 EMAIL VERIFICATION")

	errStr := ""
	if m.errorMessage != "" {
		errStr = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("\n" + m.errorMessage)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		"",
		"A confirmation code has been sent to your email.",
		"Enter the 6-digit code:",
		"",
		m.inputs[0].View(),
		errStr,
	)

	return s.Card.Render(content)
}

func (m *Model) detailsView() string {
	s := m.styles

	drawRow := func(label, value string, valueStyle lipgloss.Style) string {
		return fmt.Sprintf("%s %s", s.DetailLabel.Render(label), valueStyle.Render(value))
	}

	passDisplay := strings.Repeat("*", len([]rune(m.selectedItem.Password)))
	if m.showPassword {
		passDisplay = m.selectedItem.Password
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		s.Title.
			Foreground(lipgloss.Color("205")).
			MarginBottom(1).
			Render("── "+"ACCOUNT DETAILS"+" ──"),
		drawRow("Service:", m.selectedItem.Resource, s.DetailValue),
		drawRow("Login:", m.selectedItem.Username, s.DetailValue),
		drawRow("Email:", m.selectedItem.Email, s.DetailValue),
		drawRow("Password:", passDisplay, s.DetailKey),
		"",
	)

	return s.Card.Render(content)
}

func (m *Model) deleteView() string {
	s := m.styles

	header := s.Title.
		Foreground(lipgloss.Color("205")).
		MarginBottom(1).
		Render("── " + "DELETE ENTRY" + " ──")

	question := fmt.Sprintf("Are you sure you want to remove the password for %s?",
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render(m.selectedItem.Resource))

	formContent := lipgloss.JoinVertical(lipgloss.Left, header, question)

	return s.Card.Render(formContent)
}
