package config

type GatewayConfig struct {
	Hosts         map[string]Host `yaml:"hosts"`
	Broadcast     IP              `yaml:"broadcast"`
	Secret        string          `yaml:"secret"`
	InterfaceName string          `yaml:"interface"`
}

type Host struct {
	User             string  `yaml:"user"`
	Address          IP      `yaml:"address"`
	IdentityFilePath string  `yaml:"identity-file"`
	MAC              MACAddr `yaml:"mac"`
}
