package crypto

import (
	"golang.org/x/crypto/argon2"
)

func deriveKey(password string, salt []byte) []byte {
    // time=1, memory=64MB, threads=4, keyLen=32
	return argon2.IDKey([]byte(password), salt, 1, 64 * 1024, 4, 32)
}