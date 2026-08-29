package server

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/eevandeya/lar/internal/api"
	"github.com/eevandeya/lar/internal/config"
	"github.com/stretchr/testify/require"
)

var testMachineName = "pc1"
var testInterfaceName = "eth0"

func getTestConfigMac() config.MACAddr {
	mac, _ := net.ParseMAC("00:1A:2B:3C:4D:5E")
	return config.MACAddr(mac)
}

var testMACAddr = getTestConfigMac()
var testIPAddr = config.IP(net.ParseIP("192.168.1.67"))
var testBroadcastIP = config.IP(net.ParseIP("255.255.255.255"))

func requireAPIError(t *testing.T, expected *api.Error, actualJSON io.Reader) {
	t.Helper()

	if expected == nil {
		data, err := io.ReadAll(actualJSON)
		require.NoError(t, err)
		require.Zero(t, len(data))
		return
	}

	var actual *api.ErrorResponse
	err := json.NewDecoder(actualJSON).Decode(&actual)
	require.NoError(t, err)
	require.NotNil(t, actual)

	require.NotNil(t, actual.Error)
	require.Equal(t, expected.Code, actual.Error.Code)
	require.Equal(t, expected.Message, actual.Error.Message)
}

func getServerWithMachine(machineName string) *Server {
	return &Server{cfg: &config.GatewayConfig{
		Server: config.Server{
			Broadcast:        testBroadcastIP,
			ARPInterfaceName: testInterfaceName,
		},
		Machines: map[string]*config.Machine{
			machineName: {
				MAC:     testMACAddr,
				Address: testIPAddr,
			},
		},
	}}
}

func getRequestWithMachine(t *testing.T, machineName, method string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(method, "https://example.com", nil)
	require.NoError(t, err)
	q := req.URL.Query()
	q.Set("machine", machineName)
	req.URL.RawQuery = q.Encode()

	return req
}

func TestServerGetMachineOrWriteError(t *testing.T) {
	tests := []struct {
		name            string
		query           url.Values
		wantErr         error
		wantResponseErr *api.Error
		wantStatusCode  int
	}{
		{
			name: "Success",
			query: url.Values{
				"machine": []string{
					testMachineName,
				},
			},
			wantStatusCode: http.StatusOK, // http.Recorder initialized with .Code == 200
		},
		{
			name:  "No machine name",
			query: url.Values{},
			wantResponseErr: &api.Error{
				Code:    api.ErrMissingMachineName,
				Message: "missing machine name query param",
			},
			wantStatusCode: http.StatusBadRequest,
			wantErr:        errors.New("machine name query is missing in query"),
		},
		{
			name: "Invalid machine name",
			query: url.Values{
				"machine": []string{
					"invalid_machine",
				},
			},
			wantResponseErr: &api.Error{
				Code:    api.ErrInvalidMachineName,
				Message: "invalid machine name",
			},
			wantStatusCode: http.StatusBadRequest,
			wantErr:        errors.New("no much for requested machine in config"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			cfg := config.GatewayConfig{
				Machines: map[string]*config.Machine{
					testMachineName: {},
				},
			}
			s := Server{cfg: &cfg}

			machine, err := s.getMachineOrWriteError(rec, tt.query)

			requireAPIError(t, tt.wantResponseErr, rec.Body)
			require.Equal(t, tt.wantStatusCode, rec.Code)

			require.Equal(t, tt.wantErr, err)
			if err == nil {
				require.Equal(t, cfg.Machines[testMachineName], machine)
			}
		})
	}
}

func TestServerStatusHandler(t *testing.T) {
	tests := []struct {
		name                string
		machineName         string
		interfaceByNameFunc func(name string) (*net.Interface, error)
		arpProbeFunc        func(macAddr net.HardwareAddr, machineIP net.IP, ifi *net.Interface) (bool, error)
		wantStatusCode      int
		wantErrResponse     *api.ErrorResponse
		wantSuccessResponse *api.StatusResponse
	}{
		{
			name:           "Invalid machine name",
			machineName:    "invalid_machine",
			wantStatusCode: http.StatusBadRequest,
			wantErrResponse: &api.ErrorResponse{
				Error: &api.Error{
					Code:    api.ErrInvalidMachineName,
					Message: "invalid machine name",
				},
			},
		},
		{
			name:        "Interface unavailable",
			machineName: testMachineName,
			interfaceByNameFunc: func(name string) (*net.Interface, error) {
				require.Equal(t, testInterfaceName, name)
				return nil, errors.New("foobar")
			},
			wantStatusCode: http.StatusInternalServerError,
			wantErrResponse: &api.ErrorResponse{
				Error: &api.Error{
					Code:    api.ErrInterfaceUnavailable,
					Message: "configured interface is unavailable",
				},
			},
		},
		{
			name:        "ARP probing fail",
			machineName: testMachineName,
			interfaceByNameFunc: func(name string) (*net.Interface, error) {
				require.Equal(t, testInterfaceName, name)
				return &net.Interface{}, nil
			},
			arpProbeFunc: func(macAddr net.HardwareAddr, machineIP net.IP, ifi *net.Interface) (bool, error) {
				require.Equal(t, net.HardwareAddr(testMACAddr), macAddr)
				require.Equal(t, net.IP(testIPAddr), machineIP)
				return false, errors.New("foobar")
			},
			wantStatusCode: http.StatusInternalServerError,
			wantErrResponse: &api.ErrorResponse{
				Error: &api.Error{
					Code:    api.ErrARPProbingFailed,
					Message: "arp probing has failed",
				},
			},
		},
		{
			name:        "Success",
			machineName: testMachineName,
			interfaceByNameFunc: func(name string) (*net.Interface, error) {
				return &net.Interface{}, nil
			},
			arpProbeFunc: func(macAddr net.HardwareAddr, machineIP net.IP, ifi *net.Interface) (bool, error) {
				require.Equal(t, net.HardwareAddr(testMACAddr), macAddr)
				require.Equal(t, net.IP(testIPAddr), machineIP)
				return true, nil
			},
			wantStatusCode:      http.StatusOK,
			wantSuccessResponse: &api.StatusResponse{Online: true},
		},
	}

	oldInterfaceByName := interfaceByName
	oldArpProbe := arpProbe
	t.Cleanup(func() {
		interfaceByName = oldInterfaceByName
		arpProbe = oldArpProbe
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interfaceByName = tt.interfaceByNameFunc
			arpProbe = tt.arpProbeFunc

			rec := httptest.NewRecorder()
			s := getServerWithMachine(testMachineName)
			req := getRequestWithMachine(t, tt.machineName, http.MethodGet)

			s.statusHandler(rec, req)

			require.Equal(t, tt.wantStatusCode, rec.Code)

			if tt.wantSuccessResponse != nil {
				var statusResp api.StatusResponse
				err := json.NewDecoder(rec.Body).Decode(&statusResp)
				require.NoError(t, err)
				require.Equal(t, tt.wantSuccessResponse.Online, statusResp.Online)
			}

			if tt.wantErrResponse != nil {
				requireAPIError(t, tt.wantErrResponse.Error, rec.Body)
			}
		})
	}
}

func TestServerWakeHandler(t *testing.T) {
	tests := []struct {
		name            string
		machineName     string
		wakeFunc        func(macAddr net.HardwareAddr, ip net.IP) error
		wantStatusCode  int
		wantErrResponse *api.ErrorResponse
	}{
		{
			name:           "Invalid machine name",
			machineName:    "invalid_machine",
			wantStatusCode: http.StatusBadRequest,
			wantErrResponse: &api.ErrorResponse{
				Error: &api.Error{
					Code:    api.ErrInvalidMachineName,
					Message: "invalid machine name",
				},
			},
		},
		{
			name:        "WOL fail",
			machineName: testMachineName,
			wakeFunc: func(macAddr net.HardwareAddr, ip net.IP) error {
				return errors.New("foobar")
			},
			wantStatusCode: http.StatusInternalServerError,
			wantErrResponse: &api.ErrorResponse{
				Error: &api.Error{
					Code:    api.ErrWOLFailed,
					Message: "wake-on-lan failed",
				},
			},
		},
		{
			name:        "Success",
			machineName: testMachineName,
			wakeFunc: func(macAddr net.HardwareAddr, ip net.IP) error {
				require.Equal(t, net.HardwareAddr(testMACAddr), macAddr)
				require.Equal(t, net.IP(testBroadcastIP), ip)
				return nil
			},
			wantStatusCode: http.StatusAccepted,
		},
	}

	oldWake := wake
	t.Cleanup(func() {
		wake = oldWake
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wake = tt.wakeFunc

			rec := httptest.NewRecorder()
			s := getServerWithMachine(testMachineName)
			req := getRequestWithMachine(t, tt.machineName, http.MethodPost)

			s.wakeHandler(rec, req)

			require.Equal(t, tt.wantStatusCode, rec.Code)

			if tt.wantErrResponse != nil {
				requireAPIError(t, tt.wantErrResponse.Error, rec.Body)
			}
		})
	}
}

func TestServerShutdownHandler(t *testing.T) {
	tests := []struct {
		name            string
		machineName     string
		shutdownFunc    func(user, keyPath string, machineIP net.IP, sshPort uint16) error
		wantStatusCode  int
		wantErrResponse *api.ErrorResponse
	}{
		{
			name:           "Invalid machine name",
			machineName:    "invalid_machine",
			wantStatusCode: http.StatusBadRequest,
			wantErrResponse: &api.ErrorResponse{
				Error: &api.Error{
					Code:    api.ErrInvalidMachineName,
					Message: "invalid machine name",
				},
			},
		},
		{
			name:        "SSH shutdown fail",
			machineName: testMachineName,
			shutdownFunc: func(user, keyPath string, machineIP net.IP, sshPort uint16) error {
				return errors.New("foobar")
			},
			wantStatusCode: http.StatusInternalServerError,
			wantErrResponse: &api.ErrorResponse{
				Error: &api.Error{
					Code:    api.ErrSSHFailed,
					Message: "ssh to machine failed",
				},
			},
		},
		{
			name:        "Success",
			machineName: testMachineName,
			shutdownFunc: func(user, keyPath string, machineIP net.IP, sshPort uint16) error {
				return nil
			},
			wantStatusCode: http.StatusOK,
		},
	}

	oldShutdown := shutdown
	t.Cleanup(func() {
		shutdown = oldShutdown
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shutdown = tt.shutdownFunc

			rec := httptest.NewRecorder()
			s := getServerWithMachine(testMachineName)
			req := getRequestWithMachine(t, tt.machineName, http.MethodPost)

			s.shutdownHandler(rec, req)

			require.Equal(t, tt.wantStatusCode, rec.Code)

			if tt.wantErrResponse != nil {
				requireAPIError(t, tt.wantErrResponse.Error, rec.Body)
			}
		})
	}
}
