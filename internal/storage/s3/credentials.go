package s3

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const serviceName = "p-manager"

func GetSecrets() (accessKey string, secretKey string, err error) {
	accessKey, err = keyring.Get(serviceName, "access_key")
	if err != nil {
		return "", "", fmt.Errorf("access key not found in system: %w", err)
	}

	secretKey, err = keyring.Get(serviceName, "secret_key")
	if err != nil {
		return "", "", fmt.Errorf("secret key not found in system: %w", err)
	}

	return accessKey, secretKey, nil
}
