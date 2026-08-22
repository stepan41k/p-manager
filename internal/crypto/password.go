package crypto

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	defaultCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+"
)

func GeneratePassword(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid password length: %d", length)
	}

	charsetLen := big.NewInt(int64(len(defaultCharset)))
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random character: %w", err)
		}
		result[i] = defaultCharset[num.Int64()]
	}

	return string(result), nil
}
