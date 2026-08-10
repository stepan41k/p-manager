package crypto

import (
	"crypto/rand"
	"math/big"
)

const (
	digits = "0123456789"
	defaultLength = 6
)

func GenerateOTP() (string, error) {
	result := make([]byte, defaultLength)

	for i := 0; i < defaultLength; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}

		result[i] = digits[num.Int64()]
	}

	return string(result), nil
}