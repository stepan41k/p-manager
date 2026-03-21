package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) View() tea.View {
	if m.Choice != "" {
		return tea.NewView(m.Styles.QuitText.Render(fmt.Sprintf("%s? Sounds good to me.", m.Choice)))
	}
	if m.Quitting {
		return tea.NewView(m.Styles.QuitText.Render("Not hungry? That’s cool."))
	}
	return tea.NewView("\n" + m.List.View())
}