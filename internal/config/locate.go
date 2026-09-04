package config

import (
	"errors"
	"os"
	"path/filepath"
)

var defaultConfigNames = []string{
	"lar/config.yml",
	"lar/config.yaml",
}

var ErrNoConfig = errors.New("config was not found")

func LocateClientConfig() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", &Error{err}
	}
	configDir := filepath.Join(homeDir, ".config")

	for _, relativeConfigPath := range defaultConfigNames {
		configPath := filepath.Join(configDir, relativeConfigPath)

		stat, err := os.Stat(configPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", &Error{err}
		}
		if stat.IsDir() {
			continue
		}

		return configPath, nil
	}

	return "", &Error{ErrNoConfig}
}
