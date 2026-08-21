package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/stepan41k/p-manager/internal/config"
	"github.com/stepan41k/p-manager/internal/crypto"
	"github.com/stepan41k/p-manager/internal/lib/logger/sl"
	"github.com/stepan41k/p-manager/internal/storage/s3"
	"github.com/zalando/go-keyring"
)

func (m *Model) runSetupCmd() tea.Cmd {
	return func() tea.Msg {
		reg, endp, buck := m.inputs[0].Value(), m.inputs[1].Value(), m.inputs[2].Value()
		accKey, secKey := m.inputs[3].Value(), m.inputs[4].Value()
		smtpHost, smtpPort := m.inputs[5].Value(), m.inputs[6].Value()
		smtpSender, smtpPass := m.inputs[7].Value(), m.inputs[8].Value()
		email, master := m.inputs[9].Value(), m.inputs[10].Value()

		keyring.Set("p-manager", "access_key", accKey)
		keyring.Set("p-manager", "secret_key", secKey)
		keyring.Set("p-manager", "smtp_password", smtpPass)

		cfg := config.Config{
			SMTPConfig: config.SMTPConfig{
				Email:      email,
				SMTPHost:   smtpHost,
				SMTPPort:   smtpPort,
				SMTPSender: smtpSender,
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

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		storage, err := s3.New(context.Background(), &cfg.S3Config, m.log)
		if err != nil {
			return vaultErrorMsg(err)
		}

		existingMeta, err := storage.DownloadMeta(ctx)

		if err == nil && existingMeta != nil {
			authKey, _, err := crypto.DeriveMasterKeys(master, existingMeta.Salt)
			if err != nil || crypto.VerifyMasterKey(authKey, existingMeta.Verifier) != nil {
				return vaultErrorMsg(fmt.Errorf("неверный мастер-пароль от существующего сейфа"))
			}

			rawToken, _ := crypto.GenerateSalt(32)
			tokenStr := fmt.Sprintf("%x", rawToken)
			tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(tokenStr)))

			_ = keyring.Set("p-manager", "device_token", tokenStr)
			existingMeta.TrustedDeviceHashes = append(existingMeta.TrustedDeviceHashes, tokenHash)

			metaData, _ := json.Marshal(existingMeta)
			if err := storage.Upload(ctx, "meta.json", bytes.NewReader(metaData)); err != nil {
				return vaultErrorMsg(err)
			}

			return setupFinishedMsg{
				Storage: storage,
				Config:  cfg,
				Meta:    *existingMeta,
			}
		}

		salt, _ := crypto.GenerateSalt(16)
		authKey, vaultKey, _ := crypto.DeriveMasterKeys(master, salt)
		verifier, _ := crypto.Encrypt([]byte("OK"), authKey)

		emptyVault, _ := json.Marshal([]VaultItem{})
		encryptedVault, _ := crypto.Encrypt(emptyVault, vaultKey)

		meta := s3.Metadata{
			Salt:                salt,
			Verifier:            verifier,
			TrustedDeviceHashes: []string{},
		}

		metaData, _ := json.Marshal(meta)

		if err := storage.Upload(ctx, "meta.json", bytes.NewReader(metaData)); err != nil {
			return vaultErrorMsg(err)
		}
		if err := storage.Upload(ctx, "vault.enc", bytes.NewReader(encryptedVault)); err != nil {
			return vaultErrorMsg(err)
		}

		return setupFinishedMsg{
			Storage: storage,
			Config:  cfg,
			Meta:    meta,
		}
	}
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
		encrypted, err := crypto.Encrypt(jsonData, m.vaultKey)
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
		encrypted, err := crypto.Encrypt(jsonData, m.vaultKey)
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
			m.log.Warn("failed to marshal data: %w", sl.Err(err))
			return vaultErrorMsg(err)
		}

		encryptedData, err := crypto.Encrypt(jsonData, m.vaultKey)
		if err != nil {
			m.log.Warn("failed to encrypt data: %w", sl.Err(err))
			return vaultErrorMsg(err)
		}

		bodyReader := bytes.NewReader(encryptedData)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = m.storage.Upload(ctx, "vault.enc", bodyReader)
		if err != nil {
			m.log.Warn("error with uploading to S3: %w", sl.Err(err))
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
			m.log.Warn("error retrieving data from s3 storage: ", sl.Err(err))
			return vaultLoadedMsg([]list.Item{})
		}

		defer body.Close()

		encryptedData, err := io.ReadAll(body)
		if err != nil {
			m.log.Warn("error reading body: ", sl.Err(err))
			return vaultErrorMsg(err)
		}

		decryptedData, err := crypto.Decrypt(encryptedData, m.vaultKey)
		if err != nil {
			m.log.Warn("error decrypting data: ", sl.Err(err))
			return vaultErrorMsg(fmt.Errorf("error decrypting data: check password"))
		}

		var entries []VaultItem
		if err := json.Unmarshal(decryptedData, &entries); err != nil {
			m.log.Warn("failed to unmarshal data: ", sl.Err(err))
			return vaultErrorMsg(err)
		}

		items := make([]list.Item, len(entries))
		for i, v := range entries {
			items[i] = v
		}

		return vaultLoadedMsg(items)
	}
}
