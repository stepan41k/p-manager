package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

var (
	ErrNotExists = errors.New("configuration file not exists")
)

const (
	appName = "p-manager"
	fileName = "config.json"
)

type Config struct {
	SMTPConfig SMTPConfig `json:"user_config"`
	S3Config   S3Config   `json:"s3_config"`
}

type S3Config struct {
	Region   string `json:"region"`
	Endpoint string `json:"endpoint"`
	Bucket   string `json:"bucket"`
}

type SMTPConfig struct {
    Email        string `json:"email"`
    SMTPHost     string `json:"smtp_host"`
    SMTPPort     string `json:"smtp_port"`
    SMTPSender   string `json:"smtp_sender"`
}

func GetConfigPath() string {
	dir, _ := os.UserConfigDir()
	
	return filepath.Join(dir, appName, fileName)
}

func SaveConfig(cfg Config) error {
	path := GetConfigPath()
	
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	
	data, _ := json.MarshalIndent(cfg, "", "  ")
	
	return os.WriteFile(path, data, 0600)
}

func MustLoad() (*Config, error) {
	path := GetConfigPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, ErrNotExists
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err = json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, err
}
