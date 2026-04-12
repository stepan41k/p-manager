package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/zalando/go-keyring"
)

type Config struct {
	S3Config S3Config
}

type S3Config struct {
	AccessKey string
	SecretKey string
	Region    string	`env:"REGION"`
	Endpoint  string	`env:"ENDPOINT"`
	Bucket    string	`env:"BUCKET"`
}

func MustLoad() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	var cfg Config

	accessKey, secretKey, err := GetS3Credentials()
	if err != nil {
		return nil, fmt.Errorf("failed to load keys: %w", err)
	}

	cfg.S3Config.AccessKey = accessKey
	cfg.S3Config.SecretKey = secretKey

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	
	if cfg.S3Config.AccessKey == "" || cfg.S3Config.Bucket == "" || cfg.S3Config.Endpoint == "" || cfg.S3Config.Region == "" || cfg.S3Config.SecretKey == "" {
		return nil, fmt.Errorf("empty field in config")
	}

	return &cfg, nil
}

func GetS3Credentials() (accessKey, secretKey string, err error) {
	service := "vault-app"

	accessKey, err = keyring.Get(service, "access_key")
	if err != nil {
		return "", "", fmt.Errorf("access key not found in system: %w", err)
	}

	secretKey, err = keyring.Get(service, "secret_key")
	if err != nil {
		return "", "", fmt.Errorf("secret key not found in system: %w", err)
	}

	return accessKey, secretKey, nil
}
