package crypto

// AES-GCM шифрование

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

func Encrypt(plaintext []byte, password []byte) ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	
	key := argon2.IDKey([]byte(password), salt, 1, 64 * 1024, 4, 32)
	
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	
	cipherText := aesgcm.Seal(nil, nonce, plaintext, nil)
	
	res := append(salt, nonce...)
	res = append(res, cipherText...)
	
	return res, nil
}

func Decrypt(data []byte, password []byte) ([]byte, error) {
	if len(data) < 16 + 12 {
		return nil, errors.New("invalid data length")
	}
	
	salt := data[:16]
	nonce := data[16:28]
	cipherText := data[28:]
	
	key := argon2.IDKey([]byte(password), salt, 1, 64 * 1024, 4, 32)
	
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	return aesgcm.Open(nil, nonce, cipherText, nil)
}