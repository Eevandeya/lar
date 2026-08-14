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

func fileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func LocateClientConfig() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".config")

	for _, relativeConfigPath := range defaultConfigNames {
		configPath := filepath.Join(configDir, relativeConfigPath)

		ok, err := fileExists(configPath)
		if err != nil {
			return "", err
		}

		if ok {
			return configPath, nil
		}
	}

	return "", ErrNoConfig
}
