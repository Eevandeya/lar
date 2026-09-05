package config

import (
	"cmp"
	"errors"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
)

func equalUnordered[T cmp.Ordered](s1 []T, s2 []T) bool {
	sortedS1 := append([]T(nil), s1...)
	sortedS2 := append([]T(nil), s2...)
	slices.Sort(sortedS1)
	slices.Sort(sortedS2)
	return slices.Equal(sortedS1, sortedS2)

}

func getTestMac() net.HardwareAddr {
	mac, _ := net.ParseMAC("00:1A:2B:3C:4D:5E")
	return mac
}

func createTestConfigByContent(t *testing.T, config string) string {
	t.Helper()

	testDir := t.TempDir()
	configPath := filepath.Join(testDir, "config.yml")
	t.Cleanup(func() {
		err := os.Remove(configPath)
		if err != nil {
			t.Fatalf("error removing file: %v", err)
		}
	})

	file, err := os.Create(configPath)
	if err != nil {
		t.Fatalf("error creating file: %v", err)
	}
	_, err = file.WriteString(config)
	if err != nil {
		t.Fatalf("error writing to file: %v", err)
	}
	_ = file.Close()

	return configPath
}

func TestValidateRequiredValues(t *testing.T) {
	tests := []struct {
		name              string
		input             *GatewayConfig
		wantErr           error
		wantMissingValues []string
	}{
		{
			name: "Valid config without machines",
			input: &GatewayConfig{
				Server: Server{
					Secret:           "secret",
					ARPInterfaceName: "eth0",
				},
			},
		},
		{
			name:    "Nil config",
			input:   nil,
			wantErr: ErrNilConfig,
		},
		{
			name:  "Empty config",
			input: &GatewayConfig{},
			wantMissingValues: []string{
				"server.secret",
				"server.arp-interface",
			},
		},
		{
			name: "Nil machines",
			input: &GatewayConfig{
				Server: Server{
					Secret:           "secret",
					ARPInterfaceName: "eth0",
				},
				Machines: nil,
			},
		},
		{
			name: "Empty machine",
			input: &GatewayConfig{
				Server: Server{
					Secret:           "secret",
					ARPInterfaceName: "eth0",
				},
				Machines: map[string]*Machine{
					"pc1": {},
				},
			},
			wantMissingValues: []string{
				"machines.pc1.address",
				"machines.pc1.mac",
				"machines.pc1.identity-file",
			},
		},
		{
			name: "Nil machine",
			input: &GatewayConfig{
				Server: Server{
					Secret:           "secret",
					ARPInterfaceName: "eth0",
				},
				Machines: map[string]*Machine{
					"pc1": nil,
				},
			},
			wantMissingValues: []string{
				"machines.pc1.address",
				"machines.pc1.mac",
				"machines.pc1.identity-file",
			},
		},
		{
			name: "Multiple empty machines",
			input: &GatewayConfig{
				Server: Server{
					Secret:           "secret",
					ARPInterfaceName: "eth0",
				},
				Machines: map[string]*Machine{
					"pc1": {},
					"pc2": {},
					"pc3": {},
				},
			},
			wantMissingValues: []string{
				"machines.pc1.address",
				"machines.pc1.mac",
				"machines.pc1.identity-file",
				"machines.pc2.address",
				"machines.pc2.mac",
				"machines.pc2.identity-file",
				"machines.pc3.address",
				"machines.pc3.mac",
				"machines.pc3.identity-file",
			},
		},
		{
			name: "Multiple incomplete machines",
			input: &GatewayConfig{
				Server: Server{
					Secret:           "secret",
					ARPInterfaceName: "eth0",
				},
				Machines: map[string]*Machine{
					"pc1": {
						MAC: MACAddr(getTestMac()),
					},
					"pc2": {
						Address:          IP(net.ParseIP("192.168.1.23")),
						IdentityFilePath: "path/to/file",
					},
				},
			},
			wantMissingValues: []string{
				"machines.pc1.address",
				"machines.pc1.identity-file",
				"machines.pc2.mac",
			},
		},
		{
			name: "Full config",
			input: &GatewayConfig{
				Server: Server{
					Host:             IP(net.ParseIP("0.0.0.0")),
					Port:             8080,
					Broadcast:        IP(net.ParseIP("192.168.1.255")),
					Secret:           "super",
					ARPInterfaceName: "eth0",
					TLSConfig: &TLSConfig{
						CertificateFilePath: "path/to/cert",
						KeyFilePath:         "path/to/key",
					},
				},
				Machines: map[string]*Machine{
					"pc1": {
						User:             "user",
						Address:          IP(net.ParseIP("192.168.1.23")),
						IdentityFilePath: "path/to/file",
						MAC:              MACAddr(getTestMac()),
						SSHPort:          22,
					},
					"pc2": {
						User:             "dog",
						Address:          IP(net.ParseIP("192.168.1.25")),
						IdentityFilePath: "path/to/file",
						MAC:              MACAddr(getTestMac()),
						SSHPort:          22,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequiredValues(tt.input)

			if tt.input != nil {
				// NOTE: validateRequiredValues mutates cfg.Machines
				// though it could be not so obvious by function name.
				// Anyway, we check this mutation here.
				for _, value := range tt.input.Machines {
					if value == nil {
						t.Fatal("always expecting not nil machine value, but got nil")
					}
				}
			}

			if tt.wantMissingValues != nil {
				if missingErr, ok := errors.AsType[MissingConfigValuesErr](err); ok {
					if !equalUnordered(tt.wantMissingValues, missingErr) {
						t.Fatalf("expected %v missing values, but got %v", tt.wantMissingValues, []string(missingErr))
					}
				} else {
					t.Fatalf("expected MissingConfigValuesErr, but got %v", err)
				}
				return
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, but got %v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateTLSConfig(t *testing.T) {
	tests := []struct {
		name    string
		input   *TLSConfig
		wantErr error
	}{
		{
			name:  "Nil config",
			input: nil,
		},
		{
			name: "Valid config",
			input: &TLSConfig{
				CertificateFilePath: "path/to/cert",
				KeyFilePath:         "path/to/key",
			},
		},
		{
			name: "No certificate path",
			input: &TLSConfig{
				KeyFilePath: "path/to/key",
			},
			wantErr: IncompleteTLSConfigErr{Type: CertificatePathMissing},
		},
		{
			name: "No private key path",
			input: &TLSConfig{
				CertificateFilePath: "path/to/cert",
			},
			wantErr: IncompleteTLSConfigErr{Type: PrivateKeyPathMissing},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTLSConfig(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, but got %v", tt.wantErr, err)
			}
		})
	}
}

func TestNormalizeConfig(t *testing.T) {
	tests := []struct {
		name          string
		input         *GatewayConfig
		wantTLSConfig *TLSConfig
	}{
		{
			name: "Nil tls config",
			input: &GatewayConfig{
				Server: Server{
					TLSConfig: nil,
				},
			},
			wantTLSConfig: nil,
		},
		{
			name: "Unempty tls config",
			input: &GatewayConfig{
				Server: Server{
					TLSConfig: &TLSConfig{
						CertificateFilePath: "path/to/cert",
						KeyFilePath:         "path/to/key",
					},
				},
			},
			wantTLSConfig: &TLSConfig{
				CertificateFilePath: "path/to/cert",
				KeyFilePath:         "path/to/key",
			},
		},
		{
			name: "Empty tls config",
			input: &GatewayConfig{
				Server: Server{
					TLSConfig: &TLSConfig{
						CertificateFilePath: "",
						KeyFilePath:         "",
					},
				},
			},
			wantTLSConfig: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizeConfig(tt.input)
			if tt.wantTLSConfig == nil {
				if tt.input.Server.TLSConfig != nil {
					t.Fatalf("expected tls config to be nil")
				}
				return
			}
			if *tt.input.Server.TLSConfig != *tt.wantTLSConfig {
				t.Fatalf("expected %v, but got %v", tt.wantTLSConfig, tt.input)
			}
		})
	}
}

func TestSetDefaultMachineValues(t *testing.T) {
	t.Run("No machines", func(t *testing.T) {
		var machines map[string]*Machine
		setDefaultMachineValues(machines)
		if machines != nil {
			t.Fatalf("expected %v to be nil", machines)
		}
	})

	t.Run("Machine with zero port", func(t *testing.T) {
		machines := map[string]*Machine{
			"pc1": {
				SSHPort: 0,
			},
		}
		setDefaultMachineValues(machines)
		if machines["pc1"].SSHPort != defaultSSHPort {
			t.Fatalf("expected ssh port to be %d", defaultSSHPort)
		}
	})

	t.Run("Machine with non zero port", func(t *testing.T) {
		port := uint16(2222)
		machines := map[string]*Machine{
			"pc1": {
				SSHPort: port,
			},
		}
		setDefaultMachineValues(machines)
		if machines["pc1"].SSHPort != port {
			t.Fatalf("expected ssh port to be %d", port)
		}
	})

	t.Run("Multiple machines", func(t *testing.T) {
		machines := map[string]*Machine{
			"pc1": {SSHPort: 0},
			"pc2": {SSHPort: 2222},
			"pc3": {SSHPort: 0},
		}

		setDefaultMachineValues(machines)

		if machines["pc1"].SSHPort != defaultSSHPort {
			t.Errorf("expected pc1 port to be %d", defaultSSHPort)
		}
		if machines["pc2"].SSHPort != 2222 {
			t.Errorf("expected pc2 port to remain %d", 2222)
		}
		if machines["pc3"].SSHPort != defaultSSHPort {
			t.Errorf("expected pc3 port to be %d", defaultSSHPort)
		}
	})
}

func TestLoadGateway(t *testing.T) {
	t.Run("File does not exist", func(t *testing.T) {
		testDir := t.TempDir()
		unexistingConfigPath := filepath.Join(testDir, "config.yml")
		_, err := LoadGateway(unexistingConfigPath)
		if configErr, ok := errors.AsType[*Error](err); !ok || !os.IsNotExist(configErr.Unwrap()) {
			t.Fatalf("expected %v to be NotExist error", configErr.Unwrap())
		}
	})

	t.Run("YAML type mismatch", func(t *testing.T) {
		config := "server: foo"
		configPath := createTestConfigByContent(t, config)
		_, err := LoadGateway(configPath)
		if _, ok := errors.AsType[*yaml.TypeError](err); !ok {
			t.Fatalf("expected %v to be *yaml.TypeError", err)
		}
	})

	t.Run("Default server values", func(t *testing.T) {
		const config = `
server:
  arp-interface: eth0
  secret: secret
`
		configPath := createTestConfigByContent(t, config)
		cfg, err := LoadGateway(configPath)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}

		if net.IP(cfg.Server.Host).String() != net.IP(defaultServerConfig.Host).String() {
			t.Fatalf("expected %s, but got %s", net.IP(defaultServerConfig.Host).String(), net.IP(cfg.Server.Host).String())
		}

		if net.IP(cfg.Server.Broadcast).String() != net.IP(defaultServerConfig.Broadcast).String() {
			t.Fatalf("expected %s, but got %s", net.IP(defaultServerConfig.Broadcast).String(), net.IP(cfg.Server.Broadcast).String())
		}

		if cfg.Server.Port != defaultServerConfig.Port {
			t.Fatalf("expected %d, but got %d", defaultServerConfig.Port, cfg.Server.Port)
		}

		if cfg.Server.TLSConfig != defaultServerConfig.TLSConfig {
			t.Fatalf("expected %v, but got %v", defaultServerConfig.TLSConfig, cfg.Server.TLSConfig)
		}
	})

	t.Run("Default machine values", func(t *testing.T) {
		const config = `
server:
  arp-interface: eth0
  secret: secret
machines:
  pc1:
    address: 192.168.1.55
    mac: 00:1A:2B:3C:4D:5E
    identity-file: /path/to/file
`
		configPath := createTestConfigByContent(t, config)
		cfg, err := LoadGateway(configPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Machines["pc1"].SSHPort != defaultSSHPort {
			t.Fatalf("expected %d, but got %d", defaultSSHPort, cfg.Machines["pc1"].SSHPort)
		}
	})

	t.Run("Missing required values", func(t *testing.T) {
		const config = `
server:
  secret: secret
machines:
  pc1:
    address: 192.168.1.55
    identity-file: /path/to/file
`
		configPath := createTestConfigByContent(t, config)
		_, err := LoadGateway(configPath)
		missingErr, ok := errors.AsType[MissingConfigValuesErr](err)
		if !ok {
			t.Fatalf("expected MissingConfigValuesErr, but got %v", err)
		}

		expectedErrPayload := []string{
			"server.arp-interface",
			"machines.pc1.mac",
		}
		if !equalUnordered(expectedErrPayload, missingErr) {
			t.Fatalf("expected %v missing values, but got %v", expectedErrPayload, missingErr)
		}
	})

	t.Run("Invalid TLS config", func(t *testing.T) {
		const config = `
server:
  arp-interface: eth0
  secret: secret
  tls:
    cert-file: /path/to/file
`
		configPath := createTestConfigByContent(t, config)
		_, err := LoadGateway(configPath)
		if _, ok := errors.AsType[IncompleteTLSConfigErr](err); !ok {
			t.Fatalf("expected IncompleteTLSConfigErr, but got %v", err)
		}
	})

	t.Run("Normalizes empty TLS config", func(t *testing.T) {
		const config = `
server:
  arp-interface: eth0
  secret: secret
  tls:
    cert-file:
`
		configPath := createTestConfigByContent(t, config)
		cfg, err := LoadGateway(configPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Server.TLSConfig != nil {
			t.Fatalf("expected TLSConfig to be nil, but got %v", cfg.Server.TLSConfig)
		}
	})

	t.Run("Load valid config", func(t *testing.T) {
		const config = `
server:
  host: 0.0.0.0
  port: 8080
  broadcast: 192.168.1.255
  secret: super
  arp-interface: eth0
  tls:
    cert-file: path/to/cert
    key-file: path/to/key

machines:
  pc1:
    user: user
    address: 192.168.1.23
    identity-file: path/to/file
    mac: 00:1A:2B:3C:4D:5E
    ssh-port: 22

  pc2:
    user: dog
    address: 192.168.1.25
    identity-file: path/to/file
    mac: 00:1A:2B:3C:4D:5E
    ssh-port: 22
`
		configPath := createTestConfigByContent(t, config)
		cfg, err := LoadGateway(configPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantServer := Server{
			Host:             IP(net.ParseIP("0.0.0.0")),
			Port:             8080,
			Broadcast:        IP(net.ParseIP("192.168.1.255")),
			Secret:           "super",
			ARPInterfaceName: "eth0",
			TLSConfig: &TLSConfig{
				CertificateFilePath: "path/to/cert",
				KeyFilePath:         "path/to/key",
			},
		}
		wantPC1 := &Machine{
			User:             "user",
			Address:          IP(net.ParseIP("192.168.1.23")),
			IdentityFilePath: "path/to/file",
			MAC:              MACAddr(getTestMac()),
			SSHPort:          22,
		}
		wantPC2 := &Machine{
			User:             "dog",
			Address:          IP(net.ParseIP("192.168.1.25")),
			IdentityFilePath: "path/to/file",
			MAC:              MACAddr(getTestMac()),
			SSHPort:          22,
		}
		if !reflect.DeepEqual(cfg.Server, wantServer) {
			t.Fatalf("expected server %v, but got %v", wantServer, cfg.Server)
		}
		if !reflect.DeepEqual(cfg.Machines["pc1"], wantPC1) {
			t.Fatalf("expected pc1 %v, but got %v", wantPC1, cfg.Machines["pc1"])
		}
		if !reflect.DeepEqual(cfg.Machines["pc2"], wantPC2) {
			t.Fatalf("expected pc2 %v, but got %v", wantPC2, cfg.Machines["pc2"])
		}
	})
}

func TestLoadClient(t *testing.T) {
	t.Run("File does not exist", func(t *testing.T) {
		testDir := t.TempDir()
		unexistingConfigPath := filepath.Join(testDir, "config.yml")
		_, err := LoadClient(unexistingConfigPath)
		if configErr, ok := errors.AsType[*Error](err); !ok || !os.IsNotExist(configErr.Unwrap()) {
			t.Fatalf("expected %v to be NotExist error", configErr.Unwrap())
		}
	})

	t.Run("YAML type mismatch", func(t *testing.T) {
		config := "gateway: foo"
		configPath := createTestConfigByContent(t, config)
		_, err := LoadClient(configPath)
		if _, ok := errors.AsType[*yaml.TypeError](err); !ok {
			t.Fatalf("expected %v to be *yaml.TypeError", err)
		}
	})

	t.Run("No gateway in config", func(t *testing.T) {
		config := ""
		configPath := createTestConfigByContent(t, config)
		_, err := LoadClient(configPath)
		if !errors.Is(err, ErrGatewayNotConfigured) {
			t.Fatalf("expected %v to be ErrGatewayNotConfigured", err)
		}
	})
	
	t.Run("Load valid config", func(t *testing.T) {
		config := `
gateway:
  address: https://lar.example.com
  secret: secret
`
		configPath := createTestConfigByContent(t, config)
		cfg, err := LoadClient(configPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantClient := ClientConfig{
			Gateway: &Gateway{
				Address: "https://lar.example.com",
				Secret:  "secret",
			},
		}

		if !reflect.DeepEqual(*cfg, wantClient) {
			t.Fatalf("expected client config %v, but got %v", wantClient, *cfg)
		}
	})
}
