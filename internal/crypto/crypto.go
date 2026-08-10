package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"

	"golang.org/x/crypto/argon2"
)

func HashOTP(code string) [32]byte {
	return sha256.Sum256([]byte(code))
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

func DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
}