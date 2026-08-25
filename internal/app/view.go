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
		v := tea.NewView(m.vaultView())
		v.AltScreen = true
		return v
	case customizeKeymapsState:
		content = m.keymapView()
		footer = m.help.View(m.keys.KeyMapConfig)
	case genConfigState:
		content = m.genConfigView()
		footer = m.help.View(m.keys.GenConfig)
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
	case settingsState:
		content = m.settingsView()
		footer = m.help.View(m.keys.Settings)

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
		labels = []string{"Service:", "Email:", "Username:", "Password:", "Note:"}
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
			Render("\n" + m.errorMessage)
	}

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Bold(true)

	row := fmt.Sprintf("%s %s", labelStyle.Render("Master Password:"), m.inputs[0].View())
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		row,
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
	)

	if m.selectedItem.Note != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			drawRow("Note:", m.selectedItem.Note, s.DetailValue),
		)
	}

	return s.Card.Render(content)
}

func (m *Model) deleteView() string {
	s := m.styles

	header := s.Title.
		Foreground(lipgloss.Color("205")).
		MarginBottom(1).
		Render("── " + "DELETE ENTRY" + " ──")

	errStr := ""
	if m.errorMessage != "" {
		errStr = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("\n" + m.errorMessage)
	}

	question := fmt.Sprintf("Are you sure you want to remove the password for %s?",
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render(m.selectedItem.Resource))

	formContent := lipgloss.JoinVertical(lipgloss.Left, header, question, errStr)

	return s.Card.Render(formContent)
}

func (m *Model) keymapView() string {
	s := m.styles
	header := s.Title.Foreground(lipgloss.Color("205")).Render("── " + " KEYBINDINGS CONFIGURATION" + " ──")

	var rows []string
	currentCategory := ""

	for i, item := range m.bindList {
		if item.Category != currentCategory {
			currentCategory = item.Category
			catHeader := lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true).
				MarginTop(1).
				Render(fmt.Sprintf("[%s]", currentCategory))
			rows = append(rows, catHeader)
		}

		prefix := "  "
		nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Width(20)
		keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)

		if i == m.keymapIndex {
			prefix = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("> ")
			nameStyle = nameStyle.Foreground(lipgloss.Color("255")).Bold(true)
		}

		currentKey := "none"
		if item.Binding != nil && len(item.Binding.Keys()) > 0 {
			currentKey = item.Binding.Keys()[0]
		}

		row := fmt.Sprintf("%s%s : %s", prefix, nameStyle.Render(item.Name), keyStyle.Render(currentKey))
		rows = append(rows, row)
	}

	table := lipgloss.JoinVertical(lipgloss.Left, rows...)

	var statusText string

	if m.isRebinding {
		statusText = fmt.Sprintf("Press new key for '%s' (esc to cancel)...", m.bindList[m.keymapIndex].Name)
	} else if m.errorMessage != "" {
		statusText = m.errorMessage
	} else {
		statusText = ""
	}

	help := lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("205")).MarginTop(1).Render(statusText)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		table,
		help,
	)

	return s.Card.Render(content)
}

func (m *Model) genConfigView() string {
	s := m.styles
	header := s.Title.Foreground(lipgloss.Color("205")).Render("── PASSWORD GENERATOR CUSTOMIZATION ──")

	renderCheckbox := func(label string, checked bool, isSelected bool) string {
		box := "[ ]"
		if checked {
			box = "[x]"
		}

		prefix := "  "
		if isSelected {
			prefix = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("> ")
		}

		return fmt.Sprintf("%s%s %s", prefix, lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Render(box), label)
	}

	lengthStr := fmt.Sprintf("Password Length: < %d >", m.genOpts.Length)
	if m.genOptIndex == 0 {
		lengthStr = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("> " + lengthStr)
	} else {
		lengthStr = "  " + lengthStr
	}

	rows := []string{
		header,
		"",
		lengthStr,
		renderCheckbox("Lowercase letters (a-z)", m.genOpts.UseLower, m.genOptIndex == 1),
		renderCheckbox("Capital letters (A-Z)", m.genOpts.UseUpper, m.genOptIndex == 2),
		renderCheckbox("Numbers (0-9)", m.genOpts.UseDigits, m.genOptIndex == 3),
		renderCheckbox("Special characters (!@#$)", m.genOpts.UseSymbols, m.genOptIndex == 4),
		"",
		s.DetailLabel.Render("Preview: ") + s.DetailKey.Render(m.previewPass),
		"",
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return s.Card.Render(content)
}

func (m *Model) settingsView() string {
	s := m.styles
	header := s.Title.Foreground(lipgloss.Color("205")).Render("── APP CONFIGURATION ──")

	labels := []string{
		"S3 Region:", "S3 Endpoint:", "S3 Bucket:", "Access Key:", "Secret Key:",
		"SMTP Host:", "SMTP Port:", "Sender Email:", "SMTP Pass:", "Target Email:",
	}

	settingLabelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Bold(true).
			Width(15)
	
	var inputViews []string
	for i := range m.inputs {
		prefix := "  "
		if i == m.focusIndex {
			prefix = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("> ")
		}

		label := settingLabelStyle.Render(labels[i])
		row := fmt.Sprintf("%s%s %s", prefix, label, m.inputs[i].View())
		inputViews = append(inputViews, row)
	}

	inputs := lipgloss.JoinVertical(lipgloss.Left, inputViews...)

	errStr := ""
	if m.errorMessage != "" {
		errStr = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).MarginTop(1).Render(m.errorMessage)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", inputs, errStr)
	return s.Card.Render(content)
}
