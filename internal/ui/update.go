package ui

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

type vaultLoadedMsg []list.Item 

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {	
	if key, ok := msg.(tea.KeyPressMsg); ok {
		if key.String() == "ctrl+c" {
			return m, tea.Quit
		}
	} 

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.vaultList.SetSize(msg.Width, msg.Height)
	case tea.KeyPressMsg:
        if msg.String() == "ctrl+c" {
            return m, tea.Quit
        }
	case vaultLoadedMsg:
		m.vaultList.SetItems(msg)
		m.vaultList.Title = "My passwords"
		m.state = vaultState
		return m, nil
	}

	switch m.state {
	case authState:
		newModel, cmd := m.updateAuth(msg)
		return newModel, cmd
	case vaultState:
		return m.updateVault(msg)
	case entryState:
		return m.updateEntry(msg)
	}

	return m, nil

}

func (m *Model) updateAuth(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			password := m.passInput.Value()

			//crypto

			if password == "secret" {
				m.state = vaultState

				return m, loadVaultCmd(password)


			} else {
				m.errorMessage = "Неверный пароль!"
				m.passInput.SetValue("")
				return m, nil
			}
		}
	}

	m.passInput, cmd = m.passInput.Update(msg)
	return m, cmd
}

func (m *Model) updateVault(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.vaultList, cmd = m.vaultList.Update(msg)

	return m, cmd
}

func (m *Model) updateEntry(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	return m, cmd
}


func loadVaultCmd(password string) tea.Cmd {
    return func() tea.Msg {

        items := []list.Item{
            VaultItem{resource: "Github", email: "example@gmail.com", username: "stepan", password: "123456789"},
            VaultItem{resource: "Google", email: "admin@gmail.com", username: "nickname", password: "55112233"},
        }

        return vaultLoadedMsg(items)
    }
}

// func loadEntryCmd() tea.Cmd {
// 	return func() tea.Msg {
// 		items := []list.Item{
// 			Item
// 		}
// 	}
// }