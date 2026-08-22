package app

import (
	"crypto/subtle"
	"fmt"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
	"github.com/stepan41k/p-manager/internal/config"
	"github.com/stepan41k/p-manager/internal/crypto"

	ss3 "github.com/stepan41k/p-manager/internal/storage/s3"
)

type vaultLoadedMsg []list.Item
type vaultErrorMsg error
type otpEmailSentMsg struct{}
type deviceRegisteredMsg struct{}
type checkInactivityMsg struct{}
type hidePasswordMsg struct{}
type accessDeniedMsg struct{}
type checkPasswordMsg struct{}
type clearClipboardMsg struct {
	copiedPassword string
}
type setupFinishedMsg struct {
	Storage VaultStorage
	Config  config.Config
	Meta    ss3.Metadata
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.vaultList.SetSize(msg.Width, msg.Height)
		m.help.SetWidth(msg.Width)
		m.lastActivity = time.Now()

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			m.WipeSecrets()
			m.WipeMeta()
			return m, tea.Quit
		}
		m.lastActivity = time.Now()

	case checkInactivityMsg:
		if m.state != authState && m.state != setupState {
			if time.Since(m.lastActivity) >= 5*time.Minute {
				m.WipeSecrets()
				m.vaultList.SetItems([]list.Item{})
				m.state = authState
				m.setupAuthInput()
				m.errorMessage = "Vault locked due to inactivity"
				return m, nil
			}
		}
		cmds = append(cmds, checkInactivityTicker(10*time.Second))

	case accessDeniedMsg:
		m.WipeSecrets()
		m.WipeMeta()
		m.errorMessage = "too many attempts"
		time.Sleep(5 * time.Second)
		return m, tea.Quit

	case checkPasswordMsg:
		if (m.state == createState || m.state == editState) && m.focusIndex == 3 {
			if time.Now().After(m.maskPasswordAt) {
				m.inputs[3].EchoMode = textinput.EchoPassword
			}
		}
		return m, nil

	case clearClipboardMsg:
		currentContent, err := clipboard.ReadAll()

		if err == nil && currentContent == msg.copiedPassword {
			_ = clipboard.WriteAll("")
			m.errorMessage = "Clipboard cleared for security"
		}
		return m, nil

	case hidePasswordMsg:
		now := time.Now()

		if m.showPassword && now.After(m.hidePasswordAt) {
			m.showPassword = false
		}

		return m, hidePasswordTicker(5 * time.Second)

	case setupFinishedMsg:
		m.storage = msg.Storage
		m.config = &msg.Config
		m.meta = msg.Meta
		m.state = authState
		m.setupAuthInput()
		m.errorMessage = "Setup finished! Enter master password."
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
		m.lastActivity = time.Now()
		m.log.Info("start ticker")
		cmds = append(cmds, checkInactivityTicker(10*time.Second))

	case vaultErrorMsg:
		m.errorMessage = "Error: " + msg.Error()
		return m, nil
	}

	var subCmd tea.Cmd

	switch m.state {
	case setupState:
		_, subCmd = m.updateSetup(msg)
	case authState:
		_, subCmd = m.updateAuth(msg)
	case otpState:
		_, subCmd = m.updateOTP(msg)
	case vaultState:
		_, subCmd = m.updateVault(msg)
	case detailsState:
		_, subCmd = m.updateDetails(msg)
	case createState:
		_, subCmd = m.updateCreate(msg)
	case editState:
		_, subCmd = m.updateEdit(msg)
	case deleteState:
		_, subCmd = m.updateDelete(msg)
	}

	if subCmd != nil {
		cmds = append(cmds, subCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) updateSetup(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Setup.Enter):
			if m.focusIndex == len(m.inputs)-1 {
				m.errorMessage = "Saving configuration..."
				return m, m.runSetupCmd()
			}

			m.inputs[m.focusIndex].Blur()
			m.focusIndex++
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
			password := m.inputs[0].Value()

			authKey, vaultKey, err := crypto.DeriveMasterKeys(password, m.meta.Salt)
			if err != nil {
				m.errorMessage = "failed to derive key"
				m.inputs[0].SetValue("")
				return m, nil
			}

			if err := crypto.VerifyMasterKey(authKey, m.meta.Verifier); err != nil {
				if m.authAttempts >= 5 {
					return m, accessDeniedCmd()
				}
				m.errorMessage = "invalid master password"
				m.authAttempts++
				m.inputs[0].SetValue("")
				return m, nil
			}

			m.authKey = authKey
			m.vaultKey = vaultKey

			if m.checkIsTrustedDevice() {
				m.state = vaultState
				m.errorMessage = "Устройство распознано. Загрузка..."
				return m, m.fetchVaultCmd()
			}

			code, _ := crypto.GenerateOTP()
			m.expectedOTPHash = crypto.HashOTP(code)

			m.otpExpiresAt = time.Now().Add(5 * time.Minute)
			m.otpAttempts = 0

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

	m.inputs[0], cmd = m.inputs[0].Update(msg)
	return m, cmd
}

func (m *Model) updateOTP(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.OTP.Enter):
			if time.Now().After(m.otpExpiresAt) {
				m.errorMessage = "Code expired! Please enter master password again."
				m.WipeSecrets()
				m.state = authState
				return m, nil
			}

			m.otpAttempts++
			if m.otpAttempts > 5 {
				m.errorMessage = "Too many failed attempts! Access blocked."
				m.WipeSecrets()
				m.state = authState
				return m, nil
			}

			inputHash := crypto.HashOTP(m.inputs[0].Value())

			isEqual := subtle.ConstantTimeCompare(inputHash[:], m.expectedOTPHash[:]) == 1

			if isEqual {
				m.state = vaultState
				m.expectedOTPHash = [32]byte{}
				m.errorMessage = "Code verified! Registering device..."

				return m, m.registerCurrentDeviceCmd()
			} else {
				attemptsLeft := 3 - m.otpAttempts
				m.errorMessage = fmt.Sprintf("Invalid 2FA code! Attempts remaining: %d", attemptsLeft)
				m.inputs[0].SetValue("")
				return m, nil
			}
		case key.Matches(msg, m.keys.OTP.Quit):
			m.state = authState
			return m, nil
		}
	}

	m.inputs[0], cmd = m.inputs[0].Update(msg)
	return m, cmd
}

func (m *Model) updateVault(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Vault.Create):
			m.state = createState
			m.setupFormInputs(nil)
			return m, nil
		case key.Matches(msg, m.keys.Vault.Edit):
			selected := m.vaultList.SelectedItem()

			if item, ok := selected.(VaultItem); ok {
				m.selectedItem = item
				m.state = editState
			}

			m.setupFormInputs([]string{
				m.selectedItem.Resource,
				m.selectedItem.Email,
				m.selectedItem.Username,
				m.selectedItem.Password,
			})

			return m, nil
		case key.Matches(msg, m.keys.Vault.Details):
			selected := m.vaultList.SelectedItem()
			m.showPassword = false

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
			err := clipboard.WriteAll(m.selectedItem.Password)
			if err != nil {
				m.errorMessage = "failed to copy to clipboard"
				return m, nil
			}

			m.errorMessage = "Password copied! Clearing in 30s..."

			return m, clearClipboardCmd(m.selectedItem.Password, 30*time.Second)
		case key.Matches(msg, m.keys.Details.View):
			m.showPassword = !m.showPassword

			if m.showPassword {
				m.hidePasswordAt = time.Now().Add(10 * time.Second)
			}

			return m, hidePasswordTicker(5 * time.Second)
		}
	}

	return m, nil
}

func (m *Model) updateCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.focusIndex == 3 {
			m.inputs[3].EchoMode = textinput.EchoNormal
			m.maskPasswordAt = time.Now().Add(2 * time.Second)

			cmds = append(cmds, checkPasswordTicker(2*time.Second))
		}

		switch {
		case key.Matches(msg, m.keys.Create.Cancel):
			m.state = vaultState
			return m, nil
		case key.Matches(msg, m.keys.Create.Next):
			if m.focusIndex == 3 {
				m.inputs[3].EchoMode = textinput.EchoPassword
			}
			m.inputs[m.focusIndex].Blur()
			m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
			m.inputs[m.focusIndex].Focus()
			return m, nil
		case key.Matches(msg, m.keys.Create.Previous):
			if m.focusIndex == 3 {
				m.inputs[3].EchoMode = textinput.EchoPassword
			}
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
				password, err := crypto.GeneratePassword(32)
				if err != nil {
					m.errorMessage = "failed to generate password"
					return m, nil
				}
				m.inputs[3].SetValue(password)

				m.maskPasswordAt = time.Now().Add(2 * time.Second)
				return m, checkPasswordTicker(2 * time.Second)
			}

		case key.Matches(msg, m.keys.Create.Submit):
			for i := range m.inputs {
				if len(m.inputs[i].Value()) == 0 {
					m.errorMessage = fmt.Sprintf("Empty Field: %s", m.inputs[i].Placeholder)
					return m, nil
				}
			}
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

	var inputCmd tea.Cmd
	m.inputs[m.focusIndex], inputCmd = m.inputs[m.focusIndex].Update(msg)
	cmds = append(cmds, inputCmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) updateEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.focusIndex == 3 {
			m.inputs[3].EchoMode = textinput.EchoNormal
			m.maskPasswordAt = time.Now().Add(2 * time.Second)

			cmds = append(cmds, checkPasswordTicker(2*time.Second))
		}

		switch {
		case key.Matches(msg, m.keys.Edit.Cancel):
			m.inputs[3].EchoMode = textinput.EchoPassword
			m.state = vaultState
			return m, nil

		case key.Matches(msg, m.keys.Edit.Next):
			if m.focusIndex == 3 {
				m.inputs[3].EchoMode = textinput.EchoPassword
			}
			m.inputs[m.focusIndex].Blur()
			m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
			m.inputs[m.focusIndex].Focus()
			return m, nil

		case key.Matches(msg, m.keys.Edit.Previous):
			if m.focusIndex == 3 {
				m.inputs[3].EchoMode = textinput.EchoPassword
			}
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
				password, err := crypto.GeneratePassword(32)
				if err != nil {
					m.errorMessage = "failed to generate password"
					return m, nil
				}
				m.inputs[3].SetValue(password)

				m.maskPasswordAt = time.Now().Add(2 * time.Second)
				return m, checkPasswordTicker(2 * time.Second)
			}

		case key.Matches(msg, m.keys.Edit.Submit):
			for i := range m.inputs {
				if len(m.inputs[i].Value()) == 0 {
					m.errorMessage = fmt.Sprintf("Empty Field: %s", m.inputs[i].Placeholder)
					return m, nil
				}
			}

			m.inputs[3].EchoMode = textinput.EchoPassword

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

	var inputCmd tea.Cmd
	m.inputs[m.focusIndex], inputCmd = m.inputs[m.focusIndex].Update(msg)
	if inputCmd != nil {
		cmds = append(cmds, inputCmd)
	}

	return m, tea.Batch(cmds...)
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

func checkInactivityTicker(delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return checkInactivityMsg{}
	})
}

func hidePasswordTicker(delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return hidePasswordMsg{}
	})
}

func clearClipboardCmd(copiedPassword string, delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return clearClipboardMsg{copiedPassword: copiedPassword}
	})
}

func checkPasswordTicker(delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return checkPasswordMsg{}
	})
}

func accessDeniedCmd() tea.Cmd {
	return func() tea.Msg {
		return accessDeniedMsg{}
	}
}
