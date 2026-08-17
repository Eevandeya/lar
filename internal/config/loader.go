package config

import (
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadGateway(path string) (*GatewayConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Default values here
	cfg := GatewayConfig{
		Broadcast: IP(net.ParseIP("255.255.255.255")),
		Port:      8080,
		SshPort:   22,
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
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
