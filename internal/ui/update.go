package ui

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/stepan41k/p-manager/internal/crypto"
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
		return m.updateAuth(msg)
	case vaultState:
		return m.updateVault(msg)
	case stateDetails:
		return m.updateDetails(msg)
	case createState:
		return m.updateCreate(msg)
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
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "n":
			m.state = createState
			m.setupInputs()
			return m, nil
		case "enter":
			selected := m.vaultList.SelectedItem()

			if item, ok := selected.(VaultItem); ok {
				m.selectedItem = item
				m.state = stateDetails
			}

			return m, nil
		}
	}

	var cmd tea.Cmd
	m.vaultList, cmd = m.vaultList.Update(msg)

	return m, cmd
}

func (m *Model) updateDetails(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "backspace":
			m.state = vaultState
			return m, nil
		case "c":
			//TODO: implement copying into buffer
			return m, nil
		}
	}

	return m, nil
}

func (m *Model) updateCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			m.state = vaultState
			return m, nil
		case "tab", "down":
			m.inputs[m.focusIndex].Blur()
			m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
			m.inputs[m.focusIndex].Focus()
			return m, nil
		case "g":
			if m.focusIndex == 2 {
				m.inputs[2].SetValue(crypto.GeneratePassword(16))
			}
			return m, nil
		case "enter":
			if m.focusIndex == len(m.inputs)-1 {
				newEntry := VaultItem{
					Resource: m.inputs[0].Value(),
					Username: m.inputs[1].Value(),
					Email:    m.inputs[2].Value(),
					Password: m.inputs[3].Value(),
				}
				return m, m.saveAndUploadCmd(newEntry)
			}
			m.inputs[m.focusIndex].Blur()
			m.focusIndex++
			m.inputs[m.focusIndex].Focus()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return m, cmd
}

func (m *Model) saveAndUploadCmd(entry VaultItem) tea.Cmd {
	return func() tea.Msg {
		// currentItems := m.vaultList.Items()
		// var allEntries []VaultItem

		// 2. Шифруем (заглушка)
		// encryptedData := crypto.Encrypt(allEntries, m.masterKey)

		// 3. Отправляем в S3
		// err := storage.UploadToS3(encryptedData)

		// if err != nil { return errorMsg(err) }
		return vaultLoadedMsg(m.vaultList.Items()) // Возвращаемся в список
	}
}

func loadVaultCmd(password string) tea.Cmd {
	return func() tea.Msg {

		items := []list.Item{
			VaultItem{Resource: "Github", Email: "example@gmail.com", Username: "stepan", Password: "123456789"},
			VaultItem{Resource: "Google", Email: "admin@gmail.com", Username: "nickname", Password: "55112233"},
		}

		return vaultLoadedMsg(items)
	}
}
