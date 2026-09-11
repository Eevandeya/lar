package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/eevandeya/lar/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func dummyRunE(_ *cobra.Command, _ []string) error {
	return nil
}

func createTestConfigByContent(t *testing.T, content string) {
	t.Helper()

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	err := os.MkdirAll(filepath.Join(tmpDir, ".config", "lar"), 0755)
	require.NoError(t, err)
	configPath := filepath.Join(tmpDir, ".config", "lar", "config.yml")

	file, err := os.Create(configPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = file.Close()
	})

	_, err = file.WriteString(content)
	require.NoError(t, err)
}

func TestNewRootCommand_Imperative(t *testing.T) {
	t.Run("Invalid config", func(t *testing.T) {
		cfg := ""
		var buf bytes.Buffer
		cmd := NewRootCommand(&buf)

		createTestConfigByContent(t, cfg)

		cmd.RunE = dummyRunE
		err := cmd.Execute()

		var configErr *config.Error
		require.ErrorAs(t, err, &configErr)
		require.ErrorIs(t, err, config.ErrGatewayNotConfigured)
	})

	t.Run("No config at config flag path", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := NewRootCommand(&buf)

		cmd.SetArgs([]string{"--config", "invalid/path/to/config"})
		cmd.RunE = dummyRunE
		err := cmd.Execute()

		var configErr *config.Error
		var pathErr *os.PathError
		require.ErrorAs(t, err, &configErr)
		require.ErrorIs(t, err, os.ErrNotExist)
		require.ErrorAs(t, err, &pathErr)
	})

	t.Run("No config flag and no config", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := NewRootCommand(&buf)

		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)

		cmd.RunE = dummyRunE
		err := cmd.Execute()

		require.ErrorIs(t, err, config.ErrNoConfig)
	})

	t.Run("Debug set by flag", func(t *testing.T) {
		t.Cleanup(func() {
			debug = false
		})

		cfg := "gateway:\n  address: https://lar.example.com\n  secret: secret\n"
		var buf bytes.Buffer
		cmd := NewRootCommand(&buf)

		createTestConfigByContent(t, cfg)

		cmd.RunE = dummyRunE
		cmd.SetArgs([]string{"--debug"})
		err := cmd.Execute()
		require.NoError(t, err)

		require.True(t, debug)
	})

	t.Run("Successful config with flag", func(t *testing.T) {
		cfg := "gateway:\n  address: https://lar.example.com\n  secret: secret\n"
		var buf bytes.Buffer
		cmd := NewRootCommand(&buf)

		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yml")

		file, err := os.Create(configPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			_ = file.Close()
		})

		_, err = file.WriteString(cfg)
		require.NoError(t, err)

		cmd.SetArgs([]string{"--config", configPath})
		cmd.RunE = dummyRunE
		err = cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("Successful config without flag", func(t *testing.T) {
		cfg := "gateway:\n  address: https://lar.example.com\n  secret: secret\n"
		var buf bytes.Buffer
		cmd := NewRootCommand(&buf)

		createTestConfigByContent(t, cfg)

		cmd.RunE = dummyRunE
		err := cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("Subcommands are registered", func(t *testing.T) {
		cmd := NewRootCommand(io.Discard)

		tests := []struct {
			name string
		}{
			{name: "wake"},
			{name: "status"},
			{name: "shutdown"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				subcommand, _, err := cmd.Find([]string{tt.name})

				require.NoError(t, err)
				require.Equal(t, tt.name, subcommand.Name())
			})
		}
	})
}
