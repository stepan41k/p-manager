package ui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// switch msg := msg.(type) {
	// case tea.KeyMsg:
	// 	switch msg.String() {
	// 	case "ctrl+c", "q":
	// 		return m, tea.Quit
	// 	}
	// case tea.WindowSizeMsg:
	// 	m.width, m.height = msg.Width, msg.Height
	// }

	// if m.state == authState {
	// 	return m.updateAuth(msg)
	// }

	// return m.updateList(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			m.Counter++
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}