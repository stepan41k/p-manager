package crypto

import "crypto/rand"

func GeneratePassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+"
	
	b := make([]byte, length)
	rand.Read(b)
	
	for i := range b {
		b[i] = charset[int(b[i]) % len(charset)]
	}
	
	return string(b)
}