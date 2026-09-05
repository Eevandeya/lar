// This config loading logic should probably be refactored before we add more stuff to it.
//
// It kinda grew organically during development, without much of an overall plan,
// so some parts may feel inconsistent or a bit random.
//
// Either refactor it properly with the bigger picture in mind, or just use
// something like Viper.

package config

import (
	"errors"
	"fmt"
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

var ErrNilConfig = errors.New("config is nil")
var ErrGatewayNotConfigured = errors.New("no gateway in client config")

var defaultServerConfig = Server{
	Host:      IP(net.ParseIP("0.0.0.0")),
	Broadcast: IP(net.ParseIP("255.255.255.255")),
	Port:      8080,
	TLSConfig: nil,
}

const defaultSSHPort = 22

func validateRequiredValues(cfg *GatewayConfig) error {
	if cfg == nil {
		return &Error{ErrNilConfig}
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

func validateTLSConfig(tls *TLSConfig) error {
	if tls == nil {
		return nil
	}

	if tls.CertificateFilePath == "" && tls.KeyFilePath != "" {
		return IncompleteTLSConfigErr{Type: CertificatePathMissing}
	} else if tls.CertificateFilePath != "" && tls.KeyFilePath == "" {
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

func setDefaultMachineValues(machines map[string]*Machine) {
	for key := range machines {
		if machines[key].SSHPort == 0 {
			machines[key].SSHPort = defaultSSHPort
		}
	}
}

func LoadGateway(path string) (*GatewayConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &Error{err}
	}

	cfg := GatewayConfig{
		Server: defaultServerConfig,
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, &Error{err}
	}

	setDefaultMachineValues(cfg.Machines)
	err = validateRequiredValues(&cfg)
	if err != nil {
		return nil, &Error{err}
	}

	err = validateTLSConfig(cfg.Server.TLSConfig)
	if err != nil {
		return nil, &Error{err}
	}

	normalizeConfig(&cfg)

	return &cfg, nil
}

func validateClientConfig(cfg ClientConfig) error {
	if cfg.Gateway == nil {
		return ErrGatewayNotConfigured
	}
	return nil
}

func LoadClient(path string) (*ClientConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &Error{err}
	}

	cfg := ClientConfig{}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, &Error{&ParseError{err}}
	}

	err = validateClientConfig(cfg)
	if err != nil {
		return nil, &Error{err}
	}

	return &cfg, nil
}
