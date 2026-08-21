package cache

import (
	"os"
	"path/filepath"
)

func getCachePath(filename string) string {
	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, "p-manager", "cache", filename)
}

func SaveLocalCache(filename string, data []byte) error {
	path := getCachePath(filename)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func LoadLocalCache(filename string) ([]byte, error) {
	path := getCachePath(filename)
	return os.ReadFile(path)
}