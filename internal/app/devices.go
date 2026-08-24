package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stepan41k/p-manager/internal/crypto"
	"github.com/stepan41k/p-manager/internal/storage/s3"
	"github.com/zalando/go-keyring"
)

func (m *Model) checkIsTrustedDevice() bool {
	deviceToken, err := keyring.Get("p-manager", "device_token")
	if err != nil || deviceToken == "" {
		return false
	}

	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(deviceToken)))
	now := time.Now()

	for _, dev := range m.meta.TrustedDevices {
		if dev.Hash == tokenHash {
			if now.After(dev.ExpiresAt) {
				m.log.Info("Device token expired", "device", dev.Name)
				return false
			}
			return true
		}
	}

	return false
}

func (m *Model) registerCurrentDeviceCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		rawToken, _ := crypto.GenerateSalt(32)
		tokenStr := fmt.Sprintf("%x", rawToken)
		tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(tokenStr)))

		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = "unknown-device"
		}

		err := keyring.Set("p-manager", "device_token", tokenStr)
		if err != nil {
			return vaultErrorMsg(err)
		}

		now := time.Now()
		newDevice := s3.TrustedDevice{
			Hash:      tokenHash,
			Name:      hostname,
			CreatedAt: now,
			ExpiresAt: now.Add(30 * 24 * time.Hour), // Токен действителен 30 дней!
		}

		var updatedDevices []s3.TrustedDevice
		for _, dev := range m.meta.TrustedDevices {
			if now.Before(dev.ExpiresAt) && dev.Name != hostname {
				updatedDevices = append(updatedDevices, dev)
			}
		}
		updatedDevices = append(updatedDevices, newDevice)
		m.meta.TrustedDevices = updatedDevices

		metaData, _ := json.Marshal(m.meta)
		if err := m.storage.Upload(ctx, "meta.json", bytes.NewReader(metaData)); err != nil {
			return vaultErrorMsg(err)
		}

		return deviceRegisteredMsg{}
	}
}

func (m *Model) revokeAllDevicesCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		m.meta.TrustedDevices = []s3.TrustedDevice{}

		err := keyring.Delete("p-manager", "device_token")
		if err != nil {
			return vaultErrorMsg(err)
		}

		metaData, err := json.Marshal(m.meta)
		if err != nil {
			return vaultErrorMsg(err)
		}

		if err := m.storage.Upload(ctx, "meta.json", bytes.NewReader(metaData)); err != nil {
			return vaultErrorMsg(err)
		}

		return devicesRevokedMsg{}
	}
}
