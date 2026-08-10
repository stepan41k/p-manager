package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/stepan41k/p-manager/internal/config"
	"github.com/stepan41k/p-manager/internal/crypto"
	"github.com/stepan41k/p-manager/internal/lib/email"
	"github.com/stepan41k/p-manager/internal/lib/logger/sl"

	"github.com/zalando/go-keyring"

	ss3 "github.com/stepan41k/p-manager/internal/storage/s3"
)

type vaultLoadedMsg []list.Item
type vaultErrorMsg error
type otpEmailSentMsg struct{}
type setupFinishedMsg struct {
	Storage  VaultStorage
	Salt     []byte
	Verifier []byte
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
		m.salt = msg.Salt
		m.verifier = msg.Verifier

		m.state = authState

		m.errorMessage = "Setup complete! Please login."
		m.passInput.Focus()

		return m, nil

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

func (m *Model) runSetupCmd() tea.Cmd {
	return func() tea.Msg {
		reg, endp, buck := m.inputs[0].Value(), m.inputs[1].Value(), m.inputs[2].Value()
		accKey, secKey := m.inputs[3].Value(), m.inputs[4].Value()
		email, master := m.inputs[5].Value(), m.inputs[6].Value()

		keyring.Set("p-manager", "access_key", accKey)
		keyring.Set("p-manager", "secret_key", secKey)

		cfg := config.Config{
			UserConfig: config.UserConfig{
				Email: email,
			},
			S3Config: config.S3Config{
				Region:   reg,
				Endpoint: endp,
				Bucket:   buck,
			},
		}

		if err := config.SaveConfig(cfg); err != nil {
			return vaultErrorMsg(err)
		}

		storage, err := ss3.New(context.Background(), &cfg.S3Config, m.log)
		if err != nil {
			return vaultErrorMsg(err)
		}

		salt, _ := crypto.GenerateSalt(16)
		masterKey := crypto.DeriveKey(master, salt)
		verifier, _ := crypto.Encrypt([]byte("OK"), masterKey)

		meta := struct {
			Salt     []byte `json:"salt"`
			Verifier []byte `json:"verifier"`
		}{Salt: salt, Verifier: verifier}

		metaData, _ := json.Marshal(meta)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = storage.Upload(ctx, "meta.json", bytes.NewReader(metaData))
		if err != nil {
			m.log.Warn("failed to upload metadata")
			return vaultErrorMsg(err)
		}

		return setupFinishedMsg{
			Storage:  storage,
			Salt:     salt,
			Verifier: verifier,
		}
	}
}

func (m *Model) updateAuth(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Auth.Enter):
			password := m.passInput.Value()

			derivedKey := crypto.DeriveKey(password, m.salt)

			if err := crypto.VerifyMasterKey(derivedKey, m.verifier); err != nil {
				m.errorMessage = "invalid master password"
				m.passInput.SetValue("")
				return m, nil
			}

			m.masterKey = derivedKey

			code, _ := crypto.GenerateOTP()

			m.expectedOTPHash = crypto.HashOTP(code)

			m.errorMessage = "Sending code to email..."

			return m, func() tea.Msg {
				sendErr := email.SendOTPEmail(code)
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

				return m, m.fetchVaultCmd()
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

func (m *Model) deleteAndUploadCmd() tea.Cmd {
	return func() tea.Msg {
		idx := m.vaultList.Index()
		items := m.vaultList.Items()

		var newEntries []VaultItem
		for i, it := range items {
			if i == idx {
				continue
			}

			newEntries = append(newEntries, it.(VaultItem))
		}

		jsonData, _ := json.Marshal(newEntries)
		encrypted, err := crypto.Encrypt(jsonData, m.masterKey)
		if err != nil {
			return vaultErrorMsg(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		body := bytes.NewReader(encrypted)
		err = m.storage.Upload(ctx, "vault.enc", body)
		if err != nil {
			return vaultErrorMsg(err)
		}

		updatedList := make([]list.Item, len(newEntries))
		for i, v := range newEntries {
			updatedList[i] = v
		}

		return vaultLoadedMsg(updatedList)
	}
}

func (m *Model) updateAndUploadCmd(updatedEntry VaultItem) tea.Cmd {
	return func() tea.Msg {
		items := m.vaultList.Items()
		allEntries := make([]VaultItem, len(items))

		selectedIndex := m.vaultList.Index()

		for i, it := range items {
			if i == selectedIndex {
				allEntries[i] = updatedEntry
			} else {
				allEntries[i] = it.(VaultItem)
			}
		}

		jsonData, _ := json.Marshal(allEntries)
		encrypted, err := crypto.Encrypt(jsonData, m.masterKey)
		if err != nil {
			return vaultErrorMsg(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = m.storage.Upload(ctx, "vault.enc", bytes.NewReader(encrypted))
		if err != nil {
			return vaultErrorMsg(err)
		}

		newItems := make([]list.Item, len(allEntries))
		for i, v := range allEntries {
			newItems[i] = v
		}
		return vaultLoadedMsg(newItems)
	}
}

func (m *Model) saveAndUploadCmd(entry VaultItem) tea.Cmd {
	return func() tea.Msg {
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

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

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
