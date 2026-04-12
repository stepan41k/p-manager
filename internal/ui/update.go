package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/stepan41k/p-manager/internal/crypto"
	"github.com/stepan41k/p-manager/internal/lib/logger/sl"
)

type vaultLoadedMsg []list.Item
type vaultErrorMsg error

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
		m.errorMessage = ""
		return m, nil

	case vaultErrorMsg:
		m.errorMessage = "Ошибка: " + msg.Error()
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

			// TODO:crypto

			if password == "secret" {
				m.state = vaultState
				m.masterKey = password
				return m, m.fetchVaultCmd()

			} else {
				m.errorMessage = "Incorrent password!"
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
		case "up":
			m.inputs[m.focusIndex].Blur()
			if m.focusIndex-1 < 0 {
				m.focusIndex = len(m.inputs) - 1
			} else {
				m.focusIndex = (m.focusIndex - 1) % len(m.inputs)
			}
			m.inputs[m.focusIndex].Focus()
			return m, nil
		case "g":
			if m.focusIndex == 3 {
				m.inputs[3].SetValue(crypto.GeneratePassword(32))
				return m, nil
			}

		case "enter":
			if m.focusIndex == len(m.inputs)-1 {
				newEntry := VaultItem{
					Resource: m.inputs[0].Value(),
					Email:    m.inputs[1].Value(),
					Username: m.inputs[2].Value(),
					Password: m.inputs[3].Value(),
				}

				return m, m.saveAndUploadCmd(newEntry)
			}

			m.inputs[m.focusIndex].Blur()
			m.focusIndex++
			return m, m.inputs[m.focusIndex].Focus()
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return m, cmd
}

func (m *Model) saveAndUploadCmd(entry VaultItem) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		currentItems := m.vaultList.Items()
		allEntries := make([]VaultItem, 0, len(currentItems)+1)

		for _, it := range currentItems {
			if p, ok := it.(VaultItem); ok {
				allEntries = append(allEntries, p)
			}
		}

		allEntries = append(allEntries, entry)

		jsonData, err := json.Marshal(allEntries)
		if err != nil {
			m.log.Error("failed to marshal data: %w", sl.Err(err))
			return vaultErrorMsg(err)
		}

		encryptedData, err := crypto.Encrypt(jsonData, m.masterKey)
		if err != nil {
			m.log.Error("failed to encrypt data: %w", sl.Err(err))
			return vaultErrorMsg(err)
		}

		bodyReader := bytes.NewReader(encryptedData)

		err = m.storage.Upload(ctx, "vault.enc", bodyReader)
		if err != nil {
			m.log.Error("error with uploading to S3: %w", sl.Err(err))
			return vaultErrorMsg(err)
		}

		updatedItems := make([]list.Item, len(allEntries))
		for i, v := range allEntries {
			updatedItems[i] = v
		}

		return vaultLoadedMsg(updatedItems)
	}
}

func (m *Model) fetchVaultCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		body, err := m.storage.Download(ctx, "vault.enc")
		if err != nil {
			m.log.Error("error retrieving data from s3 storage: ", sl.Err(err))
			return vaultLoadedMsg([]list.Item{})
		}

		defer body.Close()

		encryptedData, err := io.ReadAll(body)
		if err != nil {
			m.log.Error("error reading body: ", sl.Err(err))
			return vaultErrorMsg(err)
		}

		decryptedData, err := crypto.Decrypt(encryptedData, m.masterKey)
		if err != nil {
			m.log.Error("error decrypting data: ", sl.Err(err))
			return vaultErrorMsg(fmt.Errorf("error decrypting data: check password"))
		}

		var entries []VaultItem
		if err := json.Unmarshal(decryptedData, &entries); err != nil {
			m.log.Error("failed to unmarshal data: ", sl.Err(err))
			return vaultErrorMsg(err)
		}

		items := make([]list.Item, len(entries))
		for i, v := range entries {
			items[i] = v
		}

		return vaultLoadedMsg(items)
	}
}
