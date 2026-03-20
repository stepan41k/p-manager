package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func (m Model) View() tea.View {
	if m.choice != "" {
		return tea.NewView(m.styles.quitText.Render(fmt.Sprintf("%s? Sounds good to me.", m.choice)))
	}
	if m.quitting {
		return tea.NewView(m.styles.quitText.Render("Not hungry? That’s cool."))
	}
	return tea.NewView("\n" + m.list.View())
}