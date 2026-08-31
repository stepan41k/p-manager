package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/zalando/go-keyring"
)

type LockoutState struct {
	AuthFailCount int       `json:"auth_fail_count"`
	OTPFailCount  int       `json:"otp_fail_count"`
	LockedUntil   time.Time `json:"locked_until"`
	Signature     string    `json:"signature"`
}

type LockoutManager struct {
	mu       sync.Mutex
	filePath string
	state    LockoutState
	hmacKey  []byte
}

func NewLockoutManager() *LockoutManager {
	configDir, _ := os.UserConfigDir()
	dir := filepath.Join(configDir, "p-manager")
	_ = os.MkdirAll(dir, 0700)

	lm := &LockoutManager{
		filePath: filepath.Join(dir, "state.json"),
	}

	lm.initHMACKey()
	
	lm.Load()
	return lm
}

func (lm *LockoutManager) initHMACKey() {
	keyHex, err := keyring.Get("p-manager", "lockout_hmac")
	if err != nil || keyHex == "" {
		key := make([]byte, 32)
		_, _ = rand.Read(key)
		keyHex = hex.EncodeToString(key)
		_ = keyring.Set("p-manager", "lockout_hmac", keyHex)
		lm.hmacKey = key
	} else {
		lm.hmacKey, _ = hex.DecodeString(keyHex)
	}
}

func (lm *LockoutManager) computeSignature(authFails, otpFails int, lockedUntil time.Time) string {
	payload := fmt.Sprintf("%d:%d:%s", authFails, otpFails, lockedUntil.Format(time.RFC3339Nano))
	h := hmac.New(sha256.New, lm.hmacKey)
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

func (lm *LockoutManager) Load() {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	data, err := os.ReadFile(lm.filePath)
	if err != nil {
		return
	}

	var loaded LockoutState
	if err := json.Unmarshal(data, &loaded); err != nil {
		lm.enforceTamperLockout()
		return
	}

	expectedSig := lm.computeSignature(loaded.AuthFailCount, loaded.OTPFailCount, loaded.LockedUntil)
	if subtle.ConstantTimeCompare([]byte(loaded.Signature), []byte(expectedSig)) != 1 {
		lm.enforceTamperLockout()
		return
	}

	lm.state = loaded
}

func (lm *LockoutManager) enforceTamperLockout() {
	lm.state.AuthFailCount = 5
	lm.state.OTPFailCount = 5
	lm.state.LockedUntil = time.Now().Add(30 * time.Minute)
	lm.saveUnlocked()
}

func (lm *LockoutManager) Save() {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.saveUnlocked()
}

func (lm *LockoutManager) saveUnlocked() {
	lm.state.Signature = lm.computeSignature(lm.state.AuthFailCount, lm.state.OTPFailCount, lm.state.LockedUntil)
	data, _ := json.MarshalIndent(lm.state, "", "  ")
	_ = os.WriteFile(lm.filePath, data, 0600)
}

func (lm *LockoutManager) IsLockedOut() (bool, time.Duration) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	now := time.Now()
	if now.Before(lm.state.LockedUntil) {
		return true, lm.state.LockedUntil.Sub(now)
	}
	return false, 0
}

func (lm *LockoutManager) RecordAuthFailure() (time.Duration, int) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.state.AuthFailCount++
	delay := lm.calculateBackoff(lm.state.AuthFailCount)
	if delay > 0 {
		lm.state.LockedUntil = time.Now().Add(delay)
	}
	lm.saveUnlocked()
	return delay, lm.state.AuthFailCount
}

func (lm *LockoutManager) RecordOTPFailure() (time.Duration, int) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.state.OTPFailCount++
	delay := lm.calculateBackoff(lm.state.OTPFailCount)
	if delay > 0 {
		lm.state.LockedUntil = time.Now().Add(delay)
	}
	lm.saveUnlocked()
	return delay, lm.state.OTPFailCount
}

func (lm *LockoutManager) RecordSuccess() {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.state.AuthFailCount = 0
	lm.state.OTPFailCount = 0
	lm.state.LockedUntil = time.Time{}
	lm.saveUnlocked()
}

func (lm *LockoutManager) SyncCloudLockout(cloudLockedUntil time.Time, cloudFails int) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if cloudLockedUntil.After(lm.state.LockedUntil) {
		lm.state.LockedUntil = cloudLockedUntil
		if cloudFails > lm.state.AuthFailCount {
			lm.state.AuthFailCount = cloudFails
		}
		lm.saveUnlocked()
	}
}

func (lm *LockoutManager) GetLockedUntil() time.Time {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	return lm.state.LockedUntil
}

func (lm *LockoutManager) calculateBackoff(fails int) time.Duration {
	if fails < 3 {
		return 0
	}
	seconds := 30.0 * math.Pow(2, float64(fails-3))
	if seconds > 1800 {
		seconds = 1800
	}
	return time.Duration(seconds) * time.Second
}