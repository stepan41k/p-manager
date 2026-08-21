package crypto

import (
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"errors"

	"golang.org/x/crypto/argon2"
)

func WipeBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func GenerateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)

	_, err := rand.Read(salt)
	if err != nil {
		return nil, errors.New("failed to generate random salt")
	}

	return salt, nil
}

func VerifyMasterKey(key []byte, encryptedVerifier []byte) error {
	decrypted, err := Decrypt(encryptedVerifier, key)
	if err != nil {
		return errors.New("invalid master password")
	}

	if string(decrypted) != "OK" {
		return errors.New("verifier missmatches")
	}

	return nil
}

func DeriveMasterKeys(password string, salt []byte) ([]byte, []byte, error) {
	masterSecret := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)

	authKey, err := hkdf.Key(sha256.New, masterSecret, salt, "p-manager-auth", 32)
	if err != nil {
		return nil, nil, err
	}

	vaultKey, err := hkdf.Key(sha256.New, masterSecret, salt, "p-manager-vault", 32)
	if err != nil {
		return nil, nil, err
	}

	return authKey, vaultKey, nil
}
