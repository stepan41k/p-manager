package ui

import (
	"log/slog"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/stepan41k/p-manager/internal/config"
	"github.com/stepan41k/p-manager/internal/crypto"

	ss3 "github.com/stepan41k/p-manager/internal/storage/s3"
)

type vaultLoadedMsg []list.Item
type vaultErrorMsg error
type otpEmailSentMsg struct{}
type deviceRegisteredMsg struct{}
type setupFinishedMsg struct {
	Storage VaultStorage
	Config  config.Config
	Meta    ss3.Metadata
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.vaultList.SetSize(msg.Width, msg.Height)

		return m, nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case setupFinishedMsg:
		m.storage = msg.Storage
		m.config = &msg.Config
		m.meta = msg.Meta

		m.state = authState
		m.errorMessage = "Setup finished! Enter master password."
		m.passInput.SetValue("")
		m.passInput.Focus()

		return m, textinput.Blink

	case deviceRegisteredMsg:
		m.state = vaultState
		m.errorMessage = ""
		return m, m.fetchVaultCmd()

	case otpEmailSentMsg:
		m.state = otpState
		m.errorMessage = ""
		m.setupOTPInput()
		return m, textinput.Blink

	case vaultLoadedMsg:
		m.vaultList.SetItems(msg)
		m.vaultList.Title = "My passwords"
		m.vaultList.Styles.Title = m.styles.Title
		m.state = vaultState
		m.errorMessage = ""
		return m, nil

	case vaultErrorMsg:
		m.errorMessage = "Error: " + msg.Error()
		return m, nil
	}

	switch m.state {
	case setupState:
		return m.updateSetup(msg)
	case authState:
		return m.updateAuth(msg)
	case otpState:
		return m.updateOTP(msg)
	case vaultState:
		return m.updateVault(msg)
	case detailsState:
		return m.updateDetails(msg)
	case createState:
		return m.updateCreate(msg)
	case editState:
		return m.updateEdit(msg)
	case deleteState:
		return m.updateDelete(msg)
	}

	return m, nil

}

func (m *Model) updateSetup(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Setup.Enter):
			if m.focusIndex == len(m.inputs)-1 {
				m.errorMessage = "Saving configuration..."
				m.log.Info("call setup cmd", slog.Int("ind", m.focusIndex))
				return m, m.runSetupCmd()
			}

			m.inputs[m.focusIndex].Blur()
			m.focusIndex++
			m.log.Info("index", slog.Int("ind", m.focusIndex))
			return m, m.inputs[m.focusIndex].Focus()

		case key.Matches(msg, m.keys.Setup.Next):
			m.inputs[m.focusIndex].Blur()
			m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
			m.inputs[m.focusIndex].Focus()
			return m, nil
		case key.Matches(msg, m.keys.Setup.Previous):
			m.inputs[m.focusIndex].Blur()
			if m.focusIndex-1 < 0 {
				m.focusIndex = len(m.inputs) - 1
			} else {
				m.focusIndex = (m.focusIndex - 1) % len(m.inputs)
			}
			m.inputs[m.focusIndex].Focus()
			return m, nil

		case key.Matches(msg, m.keys.Setup.Quit):
			return m, tea.Quit
		}
	}

	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return m, cmd
}

func (m *Model) updateAuth(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Auth.Enter):
			password := m.passInput.Value()

			derivedKey := crypto.DeriveKey(password, m.meta.Salt)

			if err := crypto.VerifyMasterKey(derivedKey, m.meta.Verifier); err != nil {
				m.log.Error("Master key verification failed", "error", err, "salt_len", len(m.meta.Salt), "verifier_len", len(m.meta.Verifier))
				m.errorMessage = "invalid master password"
				m.passInput.SetValue("")
				return m, nil
			}

			m.masterKey = derivedKey

			if m.checkIsTrustedDevice() {
				m.state = vaultState
				m.errorMessage = "Устройство распознано. Загрузка..."
				return m, m.fetchVaultCmd()
			}

			code, _ := crypto.GenerateOTP()
			m.expectedOTPHash = crypto.HashOTP(code)
			m.errorMessage = "Sending code to email..."

			return m, func() tea.Msg {
				sendErr := m.sendOTPEmail(code)
				if sendErr != nil {
					return vaultErrorMsg(sendErr)
				}
				return otpEmailSentMsg{}
			}
		}
	}

	m.passInput, cmd = m.passInput.Update(msg)
	return m, cmd
}

func (m *Model) updateOTP(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.OTP.Enter):
			inputHash := crypto.HashOTP(m.otpInput.Value())

			if inputHash == m.expectedOTPHash {
				m.state = vaultState
				m.expectedOTPHash = [32]byte{}
				m.errorMessage = "Loading data..."

				return m, m.registerCurrentDeviceCmd()
			} else {
				m.errorMessage = "Invalid code!"
				m.otpInput.SetValue("")
				return m, nil
			}
		case key.Matches(msg, m.keys.OTP.Quit):
			m.state = authState
			return m, nil
		}
	}

	m.otpInput, cmd = m.otpInput.Update(msg)
	return m, cmd
}

func (m *Model) updateVault(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Vault.Create):
			m.state = createState
			m.setupInputs()
			return m, nil
		case key.Matches(msg, m.keys.Vault.Edit):
			selected := m.vaultList.SelectedItem()

			if item, ok := selected.(VaultItem); ok {
				m.selectedItem = item
				m.state = editState
			}

			m.setupEditInputs()
			return m, nil
		case key.Matches(msg, m.keys.Vault.Details):
			selected := m.vaultList.SelectedItem()

			if item, ok := selected.(VaultItem); ok {
				m.selectedItem = item
				m.state = detailsState
			}

			return m, nil

		case key.Matches(msg, m.keys.Vault.Delete):
			if item, ok := m.vaultList.SelectedItem().(VaultItem); ok {
				m.selectedItem = item
				m.state = deleteState
				return m, nil
			}
		}

	}

	var cmd tea.Cmd
	m.vaultList, cmd = m.vaultList.Update(msg)

	return m, cmd
}

func (m *Model) updateDetails(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Details.Back):
			m.state = vaultState
			return m, nil
		case key.Matches(msg, m.keys.Details.Copy):
			//TODO: implement copying into buffer
			return m, nil
		}
	}

	return m, nil
}

func (m *Model) updateCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Create.Cancel):
			m.state = vaultState
			return m, nil
		case key.Matches(msg, m.keys.Create.Next):
			m.inputs[m.focusIndex].Blur()
			m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
			m.inputs[m.focusIndex].Focus()
			return m, nil
		case key.Matches(msg, m.keys.Create.Previous):
			m.inputs[m.focusIndex].Blur()
			if m.focusIndex-1 < 0 {
				m.focusIndex = len(m.inputs) - 1
			} else {
				m.focusIndex = (m.focusIndex - 1) % len(m.inputs)
			}
			m.inputs[m.focusIndex].Focus()
			return m, nil
		case key.Matches(msg, m.keys.Create.Generate):
			if m.focusIndex == 3 {
				m.inputs[3].SetValue(crypto.GeneratePassword(32))
				return m, nil
			}

		case key.Matches(msg, m.keys.Create.Submit):
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

func (m *Model) updateEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Edit.Cancel):
			m.state = vaultState
			return m, nil
		case key.Matches(msg, m.keys.Edit.Next):
			m.inputs[m.focusIndex].Blur()
			m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
			m.inputs[m.focusIndex].Focus()
			return m, nil
		case key.Matches(msg, m.keys.Edit.Previous):
			m.inputs[m.focusIndex].Blur()
			if m.focusIndex-1 < 0 {
				m.focusIndex = len(m.inputs) - 1
			} else {
				m.focusIndex = (m.focusIndex - 1) % len(m.inputs)
			}
			m.inputs[m.focusIndex].Focus()
			return m, nil
		case key.Matches(msg, m.keys.Edit.Generate):
			if m.focusIndex == 3 {
				m.inputs[3].SetValue(crypto.GeneratePassword(32))
				return m, nil
			}
		case key.Matches(msg, m.keys.Edit.Submit):
			updated := VaultItem{
				Resource: m.inputs[0].Value(),
				Email:    m.inputs[1].Value(),
				Username: m.inputs[2].Value(),
				Password: m.inputs[3].Value(),
			}

			m.errorMessage = "Updating at Cloud Storage..."
			return m, m.updateAndUploadCmd(updated)
		}
	}
	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return m, cmd
}

func (m *Model) updateDelete(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Delete.Yes):
			m.errorMessage = "Delete from S3..."

			return m, m.deleteAndUploadCmd()
		case key.Matches(msg, m.keys.Delete.No):
			m.state = vaultState

			return m, nil
		}
	}

	return m, nil
}
