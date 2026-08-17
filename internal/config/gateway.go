package config

type GatewayConfig struct {
	Hosts         map[string]*Host `yaml:"hosts"`
	Broadcast     IP               `yaml:"broadcast"`
	Secret        string           `yaml:"secret"`
	InterfaceName string           `yaml:"interface"`
	Port          uint16           `yaml:"port"`
	SshPort       uint16           `yaml:"ssh-port"`
}

type Host struct {
	User             string  `yaml:"user"`
	Address          IP      `yaml:"address"`
	IdentityFilePath string  `yaml:"identity-file"`
	MAC              MACAddr `yaml:"mac"`
}
