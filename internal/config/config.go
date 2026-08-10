package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	UserConfig UserConfig `json:"UserConfig"`
	S3Config   S3Config   `json:"S3Config"`
}

type UserConfig struct {
	Email string `json:"email"`
}

type S3Config struct {
	Region   string `json:"region"`
	Endpoint string `json:"endpoint"`
	Bucket   string `json:"bucket"`
}

func GetConfigPath() string {
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "p-manager", "config.json")
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
		return nil, err
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
