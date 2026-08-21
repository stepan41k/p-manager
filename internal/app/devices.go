package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stepan41k/p-manager/internal/crypto"
	"github.com/zalando/go-keyring"
)

func (m *Model) checkIsTrustedDevice() bool {
	deviceToken, err := keyring.Get("p-manager", "device_token")
	if err != nil || deviceToken == "" {
		return false
	}

	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(deviceToken)))

	if c := slices.Contains(m.meta.TrustedDeviceHashes, tokenHash); c {
		return true
	}

	return false
}

func (m *Model) registerCurrentDeviceCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		rawToken, _ := crypto.GenerateSalt(32)
		tokenStr := fmt.Sprintf("%x", rawToken)
		tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(tokenStr)))

		_ = keyring.Set("p-manager", "device_token", tokenStr)

		m.meta.TrustedDeviceHashes = append(m.meta.TrustedDeviceHashes, tokenHash)

		metaData, _ := json.Marshal(m.meta)
		if err := m.storage.Upload(ctx, "meta.json", bytes.NewReader(metaData)); err != nil {
			return vaultErrorMsg(err)
		}

		return deviceRegisteredMsg{}
	}
}

func (m *Model) revokeAllDevicesCmd() tea.Cmd {
	return func() tea.Msg {
		m.meta.TrustedDeviceHashes = []string{}

		metaData, _ := json.Marshal(m.meta)
		_ = m.storage.Upload(context.Background(), "meta.json", bytes.NewReader(metaData))

		return nil
	}
}
