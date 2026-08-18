package config

import (
	"fmt"
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

var defaultServerConfig = Server{
	Host:      IP(net.ParseIP("0.0.0.0")),
	Broadcast: IP(net.ParseIP("255.255.255.255")),
	Port:      8080,
	TLSConfig: nil,
}

var defaultMachineConfig = Machine{
	SSHPort: 22,
}

func checkMandatoryValues(cfg *GatewayConfig) []string {
	var missing []string

	if cfg.Server.Secret == "" {
		missing = append(missing, "server.secret")
	}

	if cfg.Server.ARPInterfaceName == "" {
		missing = append(missing, "server.arp-interface")
	}

	for key, value := range cfg.Machines {
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

	return missing
}

func setDefaultMachineValues(cfg *GatewayConfig) {
	for key := range cfg.Machines {
		if cfg.Machines[key].SSHPort == 0 {
			cfg.Machines[key].SSHPort = defaultMachineConfig.SSHPort
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
	missing := checkMandatoryValues(&cfg)
	if missing != nil {
		return nil, MissingConfigValuesErr(missing)
	}

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
