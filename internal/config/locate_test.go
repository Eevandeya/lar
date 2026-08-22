package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func createTestConfigByPath(t *testing.T, path string) {
	t.Helper()
	t.Cleanup(func() {
		err := os.Remove(path)
		if err != nil {
			t.Fatalf("error removing file: %v", err)
		}
	})

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("error creating file: %v", err)
	}
	_ = file.Close()
}

func TestLocateClientConfig(t *testing.T) {
	testHomeDir := t.TempDir()
	t.Setenv("HOME", testHomeDir)
	err := os.MkdirAll(filepath.Join(testHomeDir, ".config", "lar"), 0755)
	if err != nil {
		t.Fatalf("error creating directories: %v", err)
	}

	t.Run("YML", func(t *testing.T) {
		path := filepath.Join(testHomeDir, ".config", "lar", "config.yml")
		createTestConfigByPath(t, path)

		resultPath, err := LocateClientConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resultPath != path {
			t.Fatalf("unexpected config path: %s", resultPath)
		}
	})

	t.Run("YAML", func(t *testing.T) {
		path := filepath.Join(testHomeDir, ".config", "lar", "config.yaml")
		createTestConfigByPath(t, path)

		resultPath, err := LocateClientConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resultPath != path {
			t.Fatalf("unexpected config path: %s", resultPath)
		}
	})

	t.Run("YML and YAML", func(t *testing.T) {
		yamlPath := filepath.Join(testHomeDir, ".config", "lar", "config.yaml")
		ymlPath := filepath.Join(testHomeDir, ".config", "lar", "config.yml")
		createTestConfigByPath(t, yamlPath)
		createTestConfigByPath(t, ymlPath)

		resultPath, err := LocateClientConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resultPath != ymlPath {
			t.Fatalf("unexpected config path: %s", resultPath)
		}
	})

	t.Run("Not exist", func(t *testing.T) {
		_, err := LocateClientConfig()
		if !errors.Is(err, ErrNoConfig) {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("Config is directory", func(t *testing.T) {
		path := filepath.Join(testHomeDir, ".config", "lar", "config.yaml")
		err = os.Mkdir(path, 0755)
		if err != nil {
			t.Fatalf("error creating directory: %v", err)
		}

		_, err := LocateClientConfig()
		if !errors.Is(err, ErrNoConfig) {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
