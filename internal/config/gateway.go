package config

type GatewayConfig struct {
	Machines map[string]*Machine `yaml:"machines"`
	Server   Server              `yaml:"server"`
}

type Machine struct {
	User             string  `yaml:"user"`
	Address          IP      `yaml:"address"`
	IdentityFilePath string  `yaml:"identity-file"`
	MAC              MACAddr `yaml:"mac"`
	SSHPort          uint16  `yaml:"ssh-port"`
}

type Server struct {
	Host             IP         `yaml:"host"`
	Port             uint16     `yaml:"port"`
	Broadcast        IP         `yaml:"broadcast"`
	Secret           string     `yaml:"secret"`
	ARPInterfaceName string     `yaml:"apr-interface"`
	TLSConfig        *TLSConfig `yaml:"tls"`
}

type TLSConfig struct {
	CertificateFilePath string `yaml:"cert-file"`
	KeyFilePath         string `yaml:"key-file"`
}
