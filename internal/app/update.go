package app

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
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
type devicesRevokedMsg struct{}
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
		if key.Matches(msg, m.keys.Common.Quit) {
			m.WipeSecrets()
			m.WipeAll()
			return m, tea.Quit
		}
		m.lastActivity = time.Now()

	case checkInactivityMsg:
		if m.state != authState && m.state != setupState {
			if time.Since(m.lastActivity) >= 5*time.Minute {
				m.WipeSecrets()
				m.vaultList.SetItems([]list.Item{})
				m.SetState(authState)
				m.setupAuthInput()
				m.errorMessage = "Vault locked due to inactivity"
				return m, nil
			}
		}
		cmds = append(cmds, checkInactivityTicker(10*time.Second))

	case accessDeniedMsg:
		m.WipeAll()
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
		m.SetState(authState)
		m.setupAuthInput()
		m.errorMessage = "Setup finished! Enter master password."
		return m, textinput.Blink

	case deviceRegisteredMsg:
		m.SetState(vaultState)
		m.errorMessage = "New device registered..."
		return m, m.fetchVaultCmd()

	case otpEmailSentMsg:
		m.SetState(otpState)
		m.errorMessage = "OTP sent to email..."
		m.setupOTPInput()
		return m, textinput.Blink

	case vaultLoadedMsg:
		m.vaultList.SetItems(checkAndMarkDuplicates(msg))
		m.vaultList.Title = "My passwords"
		m.vaultList.Styles.Title = m.styles.Title
		m.SetState(vaultState)
		m.errorMessage = "Loading..."
		m.lastActivity = time.Now()
		m.log.Info("start ticker")
		cmds = append(cmds, checkInactivityTicker(10*time.Second))

	case devicesRevokedMsg:
		m.errorMessage = "All trusted devices revoked! 2FA required on next login."
		return m, nil

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
	case customizeKeymapsState:
		_, subCmd = m.updateKeymaps(msg)
	case genConfigState:
		_, subCmd = m.updateGenConfig(msg)
	case detailsState:
		_, subCmd = m.updateDetails(msg)
	case createState:
		_, subCmd = m.updateCreate(msg)
	case editState:
		_, subCmd = m.updateEdit(msg)
	case deleteState:
		_, subCmd = m.updateDelete(msg)
	case settingsState:
		_, subCmd = m.updateSettings(msg)
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
		case key.Matches(msg, m.keys.Common.Submit):
			if m.focusIndex == len(m.inputs)-1 {
				m.errorMessage = "Saving configuration..."
				return m, m.runSetupCmd()
			}

			m.inputs[m.focusIndex].Blur()
			m.focusIndex++
			return m, m.inputs[m.focusIndex].Focus()

		case key.Matches(msg, m.keys.Common.Next):
			m.inputs[m.focusIndex].Blur()
			m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
			m.inputs[m.focusIndex].Focus()
			return m, nil
		case key.Matches(msg, m.keys.Common.Previous):
			m.inputs[m.focusIndex].Blur()
			if m.focusIndex-1 < 0 {
				m.focusIndex = len(m.inputs) - 1
			} else {
				m.focusIndex = (m.focusIndex - 1) % len(m.inputs)
			}
			m.inputs[m.focusIndex].Focus()
			return m, nil

		case key.Matches(msg, m.keys.Common.Quit):
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
		case key.Matches(msg, m.keys.Common.Submit):
			if isLocked, remaining := m.lockout.IsLockedOut(); isLocked {
				m.errorMessage = fmt.Sprintf("Locked out (%d fails)! Try again in %s",
					m.lockout.GetFailCount(),
					remaining.Round(time.Second),
				)
				m.inputs[0].SetValue("")
				return m, nil
			}

			password := m.inputs[0].Value()
			authKey, vaultKey, _ := crypto.DeriveMasterKeys(password, m.meta.Salt)

			if err := crypto.VerifyMasterKey(authKey, m.meta.Verifier); err != nil {
				delay, count := m.lockout.RecordAuthFailure()
				m.inputs[0].SetValue("")

				if delay > 0 {
					m.errorMessage = fmt.Sprintf("Locked out (%d fails)! Try again in %s", count, delay.Round(time.Second))

					m.meta.LockedUntil = time.Now().Add(delay)
					m.meta.AuthFailCount = count
					go func() {
						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()
						metaData, _ := json.Marshal(m.meta)
						_ = m.storage.Upload(ctx, "meta.json", bytes.NewReader(metaData))
					}()
				} else {
					m.errorMessage = fmt.Sprintf("Invalid password! Fail count: %d", count)
				}
				return m, nil
			}

			m.inputs[0].SetValue("")

			m.authAttempts = 0
			m.authKey = authKey
			m.vaultKey = vaultKey

			m.lockout.RecordSuccess()
			
			if m.checkIsTrustedDevice() {
				m.SetState(vaultState)
				m.errorMessage = "Device recognized. Loading..."
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
		case key.Matches(msg, m.keys.Common.Submit):
			if time.Now().After(m.otpExpiresAt) {
				m.WipeSecrets()
				m.SetState(authState)
				m.errorMessage = "Code expired! Please enter master password again."
				return m, nil
			}

			m.otpAttempts++
			if m.otpAttempts > 5 {
				m.WipeSecrets()
				m.SetState(authState)
				m.errorMessage = "Too many failed attempts! Access blocked."
				return m, nil
			}

			inputHash := crypto.HashOTP(m.inputs[0].Value())

			isEqual := subtle.ConstantTimeCompare(inputHash[:], m.expectedOTPHash[:]) == 1

			if isEqual {
				m.SetState(vaultState)
				m.expectedOTPHash = [32]byte{}
				m.errorMessage = "Code verified! Registering device..."

				return m, m.registerCurrentDeviceCmd()
			} else {
				attemptsLeft := 3 - m.otpAttempts
				m.errorMessage = fmt.Sprintf("Invalid 2FA code! Attempts remaining: %d", attemptsLeft)
				m.inputs[0].SetValue("")
				return m, nil
			}
		case key.Matches(msg, m.keys.Common.Quit):
			m.SetState(authState)
			return m, nil
		}
	}

	m.inputs[0], cmd = m.inputs[0].Update(msg)
	return m, cmd
}

func (m *Model) updateVault(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		itemsCount := len(m.vaultList.Items())

		if itemsCount > 0 {
			switch {
			case key.Matches(msg, m.keys.Common.Next):
				if m.vaultList.Index() == itemsCount-1 {
					m.vaultList.Select(0)
					return m, nil
				}

			case key.Matches(msg, m.keys.Common.Previous):
				if m.vaultList.Index() == 0 {
					m.vaultList.Select(itemsCount - 1)
					return m, nil
				}
			}
		}

		switch {

		case key.Matches(msg, m.keys.Vault.Create):
			m.SetState(createState)
			m.setupFormInputs(nil)
			return m, nil
		case key.Matches(msg, m.keys.Vault.Edit):
			selected := m.vaultList.SelectedItem()

			if item, ok := selected.(VaultItem); ok {
				m.selectedItem = item
				m.SetState(editState)
			}

			m.setupFormInputs([]string{
				m.selectedItem.Resource,
				m.selectedItem.Email,
				m.selectedItem.Username,
				m.selectedItem.Password,
				m.selectedItem.Note,
			})

			return m, nil
		case key.Matches(msg, m.keys.Vault.Details):
			selected := m.vaultList.SelectedItem()
			m.showPassword = false

			if item, ok := selected.(VaultItem); ok {
				m.selectedItem = item
				m.SetState(detailsState)
			}

			return m, nil

		case key.Matches(msg, m.keys.Vault.Delete):
			if item, ok := m.vaultList.SelectedItem().(VaultItem); ok {
				m.selectedItem = item
				m.SetState(deleteState)
				return m, nil
			}

		case key.Matches(msg, m.keys.Vault.ConfigKeys):
			m.SetState(customizeKeymapsState)
			m.setupKeymapList()
			return m, nil

		case key.Matches(msg, m.keys.Vault.GenConfig):
			m.SetState(genConfigState)
			m.setupGenConfig()
			return m, nil

		case key.Matches(msg, m.keys.Vault.Unauthorize):
			m.SetState(authState)
			m.WipeSecrets()
			return m, nil
		case key.Matches(msg, m.keys.Vault.Settings):
			m.SetState(settingsState)
			m.setupSettingsInputs()
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
		switch {
		case key.Matches(msg, m.keys.Common.Cancel):
			m.SetState(vaultState)
			return m, nil
		case key.Matches(msg, m.keys.Details.Copy):
			err := clipboard.WriteAll(m.selectedItem.Password)
			if err != nil {
				m.errorMessage = "failed to copy to clipboard"
				return m, nil
			}

			m.errorMessage = "Password copied! Clearing in 30s..."

			return m, clearClipboardTicker(m.selectedItem.Password, 30*time.Second)
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
		case key.Matches(msg, m.keys.Common.Cancel):
			m.SetState(vaultState)
			return m, nil
		case key.Matches(msg, m.keys.Common.Next):
			if m.focusIndex == 3 {
				m.inputs[3].EchoMode = textinput.EchoPassword
			}
			m.inputs[m.focusIndex].Blur()
			m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
			m.inputs[m.focusIndex].Focus()
			return m, nil
		case key.Matches(msg, m.keys.Common.Previous):
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
		case key.Matches(msg, m.keys.Common.Generate):
			if m.focusIndex == 3 {
				opts := m.config.Generator
				if opts.Length == 0 {
					opts = crypto.DefaultGeneratorOptions()
				}

				password, err := crypto.GeneratePasswordWithOptions(opts)
				if err != nil {
					m.errorMessage = "Select at least one character set!"
					return m, nil
				}

				m.inputs[3].SetValue(password)
				m.maskPasswordAt = time.Now().Add(2 * time.Second)
				return m, checkPasswordTicker(2 * time.Second)
			}

		case key.Matches(msg, m.keys.Common.Submit):
			for i := range m.inputs {
				if len(m.inputs[i].Value()) > 1024 {
					m.errorMessage = fmt.Sprintf("%s field is too long", m.inputs[i].Placeholder)
					return m, nil
				}

				if len(m.inputs[i].Value()) == 0 && i < 4 {
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
					Note:     m.inputs[4].Value(),
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
		case key.Matches(msg, m.keys.Common.Cancel):
			m.inputs[3].EchoMode = textinput.EchoPassword
			m.SetState(vaultState)
			return m, nil

		case key.Matches(msg, m.keys.Common.Next):
			if m.focusIndex == 3 {
				m.inputs[3].EchoMode = textinput.EchoPassword
			}
			m.inputs[m.focusIndex].Blur()
			m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
			m.inputs[m.focusIndex].Focus()
			return m, nil

		case key.Matches(msg, m.keys.Common.Previous):
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

		case key.Matches(msg, m.keys.Common.Generate):
			if m.focusIndex == 3 {
				opts := m.config.Generator
				if opts.Length == 0 {
					opts = crypto.DefaultGeneratorOptions()
				}

				password, err := crypto.GeneratePasswordWithOptions(opts)
				if err != nil {
					m.errorMessage = "Select at least one character set!"
					return m, nil
				}

				m.inputs[3].SetValue(password)
				m.maskPasswordAt = time.Now().Add(2 * time.Second)
				return m, checkPasswordTicker(2 * time.Second)
			}

		case key.Matches(msg, m.keys.Common.Submit):
			for i := range m.inputs {
				if len(m.inputs[i].Value()) > 1024 {
					m.errorMessage = fmt.Sprintf("%s field is too long", m.inputs[i].Placeholder)
					return m, nil
				}

				if len(m.inputs[i].Value()) == 0 && i < 4 {
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
				Note:     m.inputs[4].Value(),
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
			m.errorMessage = "Deletion from S3 storage..."

			return m, m.deleteAndUploadCmd()
		case key.Matches(msg, m.keys.Delete.No):
			m.SetState(vaultState)

			return m, nil
		}
	}

	return m, nil
}

func (m *Model) updateKeymaps(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		k := msg.String()

		if m.isRebinding {
			if k == "esc" {
				m.isRebinding = false
				m.errorMessage = ""
				return m, nil
			}

			if errText := m.findKeyConflict(k, m.keymapIndex); errText != "" {
				m.isRebinding = false
				m.errorMessage = "Error: " + errText
				return m, nil
			}

			m.bindList[m.keymapIndex].Binding.SetKeys(k)
			currentDesc := m.bindList[m.keymapIndex].Binding.Help().Desc
			m.bindList[m.keymapIndex].Binding.SetHelp(k, currentDesc)

			m.isRebinding = false
			m.errorMessage = fmt.Sprintf("Key for '%s' set to '%s'", m.bindList[m.keymapIndex].Name, k)

			if m.config != nil {
				m.config.Keymaps = m.exportCustomKeys()
				_ = config.SaveConfig(*m.config)
			}

			return m, nil
		}

		switch {
		case key.Matches(msg, m.keys.Common.Cancel):
			if m.config != nil {
				m.config.Keymaps = m.exportCustomKeys()
				if err := config.SaveConfig(*m.config); err != nil {
					m.log.Error("failed to save custom keymaps", "error", err)
				} else {
					m.log.Info("custom keymaps saved to config.json")
				}
			}
			m.SetState(vaultState)
			m.errorMessage = ""
			return m, nil

		case key.Matches(msg, m.keys.Common.Previous):
			if m.keymapIndex > 0 {
				m.keymapIndex--
			} else {
				m.keymapIndex = len(m.bindList) - 1
			}
			return m, nil

		case key.Matches(msg, m.keys.Common.Next):
			if m.keymapIndex < len(m.bindList)-1 {
				m.keymapIndex++
			} else {
				m.keymapIndex = 0
			}
			return m, nil

		case key.Matches(msg, m.keys.Common.Submit):
			m.isRebinding = true
			m.errorMessage = fmt.Sprintf("Press new key for '%s' (esc to cancel)...", m.bindList[m.keymapIndex].Name)
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) updateGenConfig(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Common.Cancel):
			if m.config != nil {
				m.config.Generator = m.genOpts
				_ = config.SaveConfig(*m.config)
			}
			m.SetState(vaultState)
			return m, nil

		case key.Matches(msg, m.keys.Common.Previous):
			if m.genOptIndex > 0 {
				m.genOptIndex--
			} else {
				m.genOptIndex = 4
			}
			return m, nil

		case key.Matches(msg, m.keys.Common.Next):
			if m.genOptIndex < 4 {
				m.genOptIndex++
			} else {
				m.genOptIndex = 0
			}
			return m, nil

		case key.Matches(msg, m.keys.GenConfig.ReduceLength):
			if m.genOptIndex == 0 && m.genOpts.Length > 8 {
				m.genOpts.Length--
				m.previewPass, _ = crypto.GeneratePasswordWithOptions(m.genOpts)
			}
			return m, nil

		case key.Matches(msg, m.keys.GenConfig.IncreaseLength):
			if m.genOptIndex == 0 && m.genOpts.Length < 64 {
				m.genOpts.Length++
				m.previewPass, _ = crypto.GeneratePasswordWithOptions(m.genOpts)
			}
			return m, nil

		case key.Matches(msg, m.keys.GenConfig.Switch):
			switch m.genOptIndex {
			case 1:
				m.genOpts.UseLower = !m.genOpts.UseLower
			case 2:
				m.genOpts.UseUpper = !m.genOpts.UseUpper
			case 3:
				m.genOpts.UseDigits = !m.genOpts.UseDigits
			case 4:
				m.genOpts.UseSymbols = !m.genOpts.UseSymbols
			}

			m.previewPass, _ = crypto.GeneratePasswordWithOptions(m.genOpts)
			return m, nil

		case key.Matches(msg, m.keys.Common.Generate):
			m.previewPass, _ = crypto.GeneratePasswordWithOptions(m.genOpts)
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Common.Cancel):
			m.SetState(vaultState)
			return m, nil

		case key.Matches(msg, m.keys.Settings.RevokeDevices):
			m.errorMessage = "Revoking all trusted devices..."
			return m, m.revokeAllDevicesCmd()

		case key.Matches(msg, m.keys.Common.Submit):
			if m.focusIndex == len(m.inputs)-1 {
				m.errorMessage = "Saving settings..."
				return m, m.saveSettingsCmd()
			}

			m.inputs[m.focusIndex].Blur()
			m.focusIndex++
			return m, m.inputs[m.focusIndex].Focus()

		case key.Matches(msg, m.keys.Common.Next):
			m.inputs[m.focusIndex].Blur()
			m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
			m.inputs[m.focusIndex].Focus()
			return m, nil
		case key.Matches(msg, m.keys.Common.Previous):
			m.inputs[m.focusIndex].Blur()
			if m.focusIndex-1 < 0 {
				m.focusIndex = len(m.inputs) - 1
			} else {
				m.focusIndex = (m.focusIndex - 1) % len(m.inputs)
			}
			m.inputs[m.focusIndex].Focus()
			return m, nil
		}
	}

	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return m, cmd
}
