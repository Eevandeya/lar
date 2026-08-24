package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eevandeya/lar/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requireValidMachineRequest(t *testing.T, r *http.Request, expectedMethod, expectedTarget string) {
	t.Helper()

	require.Equal(t, "Bearer secret", r.Header.Get("Authorization"))
	require.Equal(t, expectedMethod, r.Method)
	require.Equal(t, expectedTarget, r.URL.Path)
	require.Equal(t, "pc1", r.URL.Query().Get("machine"))
}

func getInvalidMachineServer(t *testing.T, expectedMethod, expectedTarget string) (server *httptest.Server, expectedErr *MachineError) {
	t.Helper()

	expectedErr = &MachineError{
		MachineName: "pc1",
		Err: APIError{
			StatusCode: http.StatusBadRequest,
			Code:       api.ErrInvalidMachineName,
			Message:    "invalid machine name",
		},
	}

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireValidMachineRequest(t, r, expectedMethod, expectedTarget)

		w.Header().Set("Content-Type", api.ApplicationJSON)
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(api.ErrorResponse{
			Error: &api.Error{
				Code:    api.ErrInvalidMachineName,
				Message: "invalid machine name",
			},
		})
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	return server, expectedErr
}

func getClientFromServer(server *httptest.Server) *Client {
	return &Client{
		baseURL: server.URL,
		secret:  "secret",
		client:  server.Client(),
	}
}

func requireMachineError(t *testing.T, err error, expectedErr *MachineError) {
	t.Helper()

	var machineErr *MachineError
	require.ErrorAs(t, err, &machineErr)
	require.Equal(t, expectedErr.MachineName, machineErr.MachineName)

	var apiErr APIError
	require.ErrorAs(t, err, &apiErr)

	var expectedAPIErr APIError
	require.ErrorAs(t, expectedErr.Err, &expectedAPIErr)

	require.Equal(t, expectedAPIErr.StatusCode, apiErr.StatusCode)
	require.Equal(t, expectedAPIErr.Code, apiErr.Code)
	require.Equal(t, expectedAPIErr.Message, apiErr.Message)
}

func TestRequestMachine(t *testing.T) {
	t.Run("Invalid request target", func(t *testing.T) {
		c := New("https://example.com", "secret")
		_, err := c.requestMachine(http.MethodGet, "invalid target", "pc1")
		require.Error(t, err)
	})

	t.Run("Standard API error", func(t *testing.T) {
		server, expectedErr := getInvalidMachineServer(t, http.MethodGet, "/test")

		client := getClientFromServer(server)

		_, err := client.requestMachine(http.MethodGet, "/test", "pc1")
		requireMachineError(t, err, expectedErr)
	})

	t.Run("JSON body unmarshal error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requireValidMachineRequest(t, r, http.MethodGet, "/test")

			w.Header().Set("Content-Type", api.ApplicationJSON)
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(`{"foo":"bar",}`))
			require.NoError(t, err)
		}))
		t.Cleanup(server.Close)

		client := getClientFromServer(server)

		_, err := client.requestMachine(http.MethodGet, "/test", "pc1")

		var syntaxErr *json.SyntaxError
		require.ErrorAs(t, err, &syntaxErr)
	})

	t.Run("Unexpected API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requireValidMachineRequest(t, r, http.MethodGet, "/test")
			w.WriteHeader(http.StatusBadRequest)
		}))
		t.Cleanup(server.Close)

		client := getClientFromServer(server)

		_, err := client.requestMachine(http.MethodGet, "/test", "pc1")
		require.Equal(t, UnexpectedAPIError(http.StatusBadRequest), err)
	})

	t.Run("No errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requireValidMachineRequest(t, r, http.MethodGet, "/test")
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(server.Close)

		client := getClientFromServer(server)

		resp, err := client.requestMachine(http.MethodGet, "/test", "pc1")
		require.NoError(t, err)
		body, err := io.ReadAll(resp.Body)
		defer func() {
			_ = resp.Body.Close()
		}()
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "", string(body))
	})
}

func TestClientShutdown(t *testing.T) {
	t.Run("Failed request machine", func(t *testing.T) {
		server, expectedErr := getInvalidMachineServer(t, http.MethodPost, "/shutdown")

		client := getClientFromServer(server)
		err := client.Shutdown("pc1")
		requireMachineError(t, err, expectedErr)
	})

	t.Run("Successful shutdown", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requireValidMachineRequest(t, r, http.MethodPost, "/shutdown")
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(server.Close)

		client := getClientFromServer(server)

		err := client.Shutdown("pc1")
		require.NoError(t, err)
	})
}

func TestClientStatus(t *testing.T) {
	t.Run("Failed request machine", func(t *testing.T) {
		server, expectedErr := getInvalidMachineServer(t, http.MethodGet, "/status")

		client := getClientFromServer(server)
		_, err := client.Status("pc1")
		requireMachineError(t, err, expectedErr)
	})

	t.Run("JSON unmarshal error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requireValidMachineRequest(t, r, http.MethodGet, "/status")
			w.Header().Set("Content-Type", api.ApplicationJSON)
			_, err := w.Write([]byte(`{"foo":"bar",}`))
			require.NoError(t, err)
		}))
		t.Cleanup(server.Close)

		client := getClientFromServer(server)
		_, err := client.Status("pc1")
		var syntaxErr *json.SyntaxError
		require.ErrorAs(t, err, &syntaxErr)
	})

	t.Run("Successful status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requireValidMachineRequest(t, r, http.MethodGet, "/status")
			w.Header().Set("Content-Type", api.ApplicationJSON)
			err := json.NewEncoder(w).Encode(api.StatusResponse{
				Online: true,
			})
			require.NoError(t, err)
		}))
		t.Cleanup(server.Close)

		client := getClientFromServer(server)
		isOnline, err := client.Status("pc1")
		require.NoError(t, err)

		require.Equal(t, true, isOnline)
	})
}

func TestClientWake(t *testing.T) {
	t.Run("Failed request machine", func(t *testing.T) {
		server, expectedErr := getInvalidMachineServer(t, http.MethodPost, "/wake")

		client := getClientFromServer(server)
		err := client.Wake("pc1")
		requireMachineError(t, err, expectedErr)
	})

	t.Run("Successful shutdown", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requireValidMachineRequest(t, r, http.MethodPost, "/wake")
			w.WriteHeader(http.StatusAccepted)
		}))
		t.Cleanup(server.Close)

		client := getClientFromServer(server)

		err := client.Wake("pc1")
		require.NoError(t, err)
	})
}
