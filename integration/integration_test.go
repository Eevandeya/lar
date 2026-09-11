package integration

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/eevandeya/lar/internal/cli"
	"github.com/eevandeya/lar/internal/config"
	"github.com/eevandeya/lar/internal/server"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {

	_, path, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get current file path")
	}

	fixturesDir := filepath.Join(filepath.Dir(path), "fixtures")
	gatewayConfigPath := filepath.Join(fixturesDir, "gateway.yml")

	gatewayConfig, err := config.LoadGateway(gatewayConfigPath)
	require.NoError(t, err)

	deps := server.Dependencies{
		InterfaceByName: func(name string) (*net.Interface, error) {
			return &net.Interface{}, nil
		},
		Wake: func(macAddr net.HardwareAddr, ip net.IP) error {
			return nil
		},
		ArpProbe: func(macAddr net.HardwareAddr, machineIP net.IP, ifi *net.Interface) (bool, error) {
			return true, nil
		},
		Shutdown: func(user, keyPath string, machineIP net.IP, sshPort uint16) error {
			return nil
		},
	}

	listener, err := net.Listen("tcp",
		net.JoinHostPort(net.IP(gatewayConfig.Server.Host).String(),
			strconv.FormatUint(uint64(gatewayConfig.Server.Port), 10)))
	require.NoError(t, err)

	srv := server.NewServer(gatewayConfig, deps)

	serverErrCh := make(chan error)

	go func() {
		serverErrCh <- srv.Serve(listener)
	}()

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		require.NoError(t, srv.Shutdown(ctx))
		require.ErrorIs(t, <-serverErrCh, http.ErrServerClosed)
	}()

	address := "http://" + listener.Addr().String()

	tmpDir := t.TempDir()
	clientConfigPath := filepath.Join(tmpDir, "client.yml")
	file, err := os.Create(clientConfigPath)
	require.NoError(t, err)
	defer file.Close()

	rawClientConfig := fmt.Sprintf("gateway:\n  address: %s\n  secret: secret\n", address)

	_, err = file.WriteString(rawClientConfig)
	require.NoError(t, err)

	t.Run("Wake", func(t *testing.T) {
		var buf bytes.Buffer
		rootCmd := cli.NewRootCommand(&buf)
		rootCmd.SetArgs([]string{"wake", "pc1", "--config", clientConfigPath})

		err = rootCmd.Execute()
		require.NoError(t, err)

		expectedOutput := fmt.Sprintf("pc1: %swake request sent%s\n", cli.Blue, cli.Reset)
		require.Equal(t, []byte(expectedOutput), buf.Bytes())
	})

	t.Run("Status", func(t *testing.T) {
		var buf bytes.Buffer
		rootCmd := cli.NewRootCommand(&buf)
		rootCmd.SetArgs([]string{"status", "pc1", "--config", clientConfigPath})

		err = rootCmd.Execute()
		require.NoError(t, err)

		expectedOutput := fmt.Sprintf("pc1: %sonline%s\n", cli.Green, cli.Reset)
		require.Equal(t, []byte(expectedOutput), buf.Bytes())
	})

	t.Run("Shutdown", func(t *testing.T) {
		var buf bytes.Buffer
		rootCmd := cli.NewRootCommand(&buf)
		rootCmd.SetArgs([]string{"shutdown", "pc1", "--config", clientConfigPath})

		err = rootCmd.Execute()
		require.NoError(t, err)

		expectedOutput := fmt.Sprintf("pc1: %sshutdown successful%s\n", cli.Cyan, cli.Reset)
		require.Equal(t, []byte(expectedOutput), buf.Bytes())
	})
}
