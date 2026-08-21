package config

import (
	"errors"
	"fmt"
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

var ErrNilConfig = errors.New("config is nil")

var defaultServerConfig = Server{
	Host:      IP(net.ParseIP("0.0.0.0")),
	Broadcast: IP(net.ParseIP("255.255.255.255")),
	Port:      8080,
	TLSConfig: nil,
}

const defaultSSHPort = 22

func validateRequiredValues(cfg *GatewayConfig) error {
	if cfg == nil {
		return ErrNilConfig
	}
	var missing []string

	if cfg.Server.Secret == "" {
		missing = append(missing, "server.secret")
	}

	if cfg.Server.ARPInterfaceName == "" {
		missing = append(missing, "server.arp-interface")
	}

	for key, value := range cfg.Machines {
		if value == nil {
			cfg.Machines[key] = &Machine{}
			value = cfg.Machines[key]
		}

		prefix := fmt.Sprintf("machines.%s", key)

		if value.Address == nil {
			missing = append(missing, prefix+".address")
		}
		if value.MAC == nil {
			missing = append(missing, prefix+".mac")
		}
		if value.IdentityFilePath == "" {
			missing = append(missing, prefix+".identity-file")
		}
	}

	if missing != nil {
		return MissingConfigValuesErr(missing)
	}
	return nil
}

func validateTLSConfig(cfg *GatewayConfig) error {
	if cfg.Server.TLSConfig == nil {
		return nil
	}

	if cfg.Server.TLSConfig.CertificateFilePath == "" && cfg.Server.TLSConfig.KeyFilePath != "" {
		return IncompleteTLSConfigErr{Type: CertificatePathMissing}
	} else if cfg.Server.TLSConfig.CertificateFilePath != "" && cfg.Server.TLSConfig.KeyFilePath == "" {
		return IncompleteTLSConfigErr{Type: PrivateKeyPathMissing}
	}
	return nil
}

func normalizeConfig(cfg *GatewayConfig) {
	if cfg.Server.TLSConfig == nil {
		return
	}
	if cfg.Server.TLSConfig.CertificateFilePath == "" && cfg.Server.TLSConfig.KeyFilePath == "" {
		cfg.Server.TLSConfig = nil
	}
}

func setDefaultMachineValues(cfg *GatewayConfig) {
	for key := range cfg.Machines {
		if cfg.Machines[key].SSHPort == 0 {
			cfg.Machines[key].SSHPort = defaultSSHPort
		}
	}
}

func LoadGateway(path string) (*GatewayConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := GatewayConfig{
		Server: defaultServerConfig,
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	setDefaultMachineValues(&cfg)
	err = validateRequiredValues(&cfg)
	if err != nil {
		return nil, err
	}

	err = validateTLSConfig(&cfg)
	if err != nil {
		return nil, err
	}

	normalizeConfig(&cfg)

	return &cfg, nil
}

func LoadClient(path string) (*ClientConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := ClientConfig{}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
