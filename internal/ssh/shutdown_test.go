package ssh

import (
	"errors"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeSSHClient struct {
	closeFn      func() error
	newSessionFn func() (Session, error)
	closed       bool
}

func (c *fakeSSHClient) Close() error {
	c.closed = true
	return c.closeFn()
}

func (c *fakeSSHClient) NewSession() (Session, error) {
	return c.newSessionFn()
}

type fakeSession struct {
	closeFn func() error
	runFn   func(string) error
	closed  bool
}

func (s *fakeSession) Close() error {
	s.closed = true
	return s.closeFn()
}

func (s *fakeSession) Run(cmd string) error {
	return s.runFn(cmd)
}

func createTestKeyFile(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "key")
	file, err := os.Create(keyPath)
	require.NoError(t, err)
	defer file.Close()

	_, err = file.WriteString(content)
	require.NoError(t, err)

	return keyPath
}

func TestShutdown(t *testing.T) {
	testUser := "user"
	testIP := net.IP{32, 202, 40, 227}
	testPort := uint16(10)
	testKey := "secret"

	expectedAddr := net.JoinHostPort(testIP.String(), strconv.FormatUint(uint64(testPort), 10))

	closeFn := func() error {
		return nil
	}

	t.Run("Read file error", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "ghost")

		err := shutdown(testUser, path, testIP, testPort, nil)
		require.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("Provider error", func(t *testing.T) {
		keyPath := createTestKeyFile(t, testKey)

		provider := ClientProvider(func(addr, user string, key []byte) (Client, error) {
			require.Equal(t, expectedAddr, addr)
			require.Equal(t, testUser, user)
			require.Equal(t,  []byte(testKey), key)
			return nil, errors.New("foo")
		})

		err := shutdown(testUser, keyPath, testIP, testPort, provider)
		require.EqualError(t, err, "foo")
	})

	t.Run("New session error", func(t *testing.T) {
		keyPath := createTestKeyFile(t, testKey)
		client := &fakeSSHClient{
			closeFn: closeFn,
			newSessionFn: func() (Session, error) {
				return nil, errors.New("foo")
			},
		}

		provider := ClientProvider(func(addr, user string, key []byte) (Client, error) {
			require.Equal(t, expectedAddr, addr)
			require.Equal(t, testUser, user)
			require.Equal(t,  []byte(testKey), key)
			return client, nil
		})

		err := shutdown(testUser, keyPath, testIP, testPort, provider)
		require.EqualError(t, err, "foo")
		require.True(t, client.closed)
	})

	t.Run("Session run error", func(t *testing.T) {
		keyPath := createTestKeyFile(t, testKey)
		session := &fakeSession{
			closeFn: closeFn,
			runFn: func(s string) error {
				require.Equal(t, poweroffCmd, s)
				return errors.New("foo")
			},
		}
		client := &fakeSSHClient{
			closeFn: closeFn,
			newSessionFn: func() (Session, error) {
				return session, nil
			},
		}

		provider := ClientProvider(func(addr, user string, key []byte) (Client, error) {
			require.Equal(t, expectedAddr, addr)
			require.Equal(t, testUser, user)
			require.Equal(t,  []byte(testKey), key)
			return client, nil
		})

		err := shutdown(testUser, keyPath, testIP, testPort, provider)
		require.EqualError(t, err, "foo")
		require.True(t, client.closed)
		require.True(t, session.closed)
	})

	t.Run("Success", func(t *testing.T) {
		keyPath := createTestKeyFile(t, testKey)
		session := &fakeSession{
			closeFn: closeFn,
			runFn: func(s string) error {
				require.Equal(t, poweroffCmd, s)
				return nil
			},
		}
		client := &fakeSSHClient{
			closeFn: closeFn,
			newSessionFn: func() (Session, error) {
				return session, nil
			},
		}

		provider := ClientProvider(func(addr, user string, key []byte) (Client, error) {
			require.Equal(t, expectedAddr, addr)
			require.Equal(t, testUser, user)
			require.Equal(t,  []byte(testKey), key)
			return client, nil
		})

		err := shutdown(testUser, keyPath, testIP, testPort, provider)
		require.NoError(t, err)
		require.True(t, client.closed)
		require.True(t, session.closed)
	})
}
