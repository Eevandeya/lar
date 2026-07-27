package config

type Config struct {
	Gateway   Gateway         `yaml:"gateway"`
	Hosts     map[string]Host `yaml:"hosts"`
	Broadcast string          `yaml:"broadcast"`
}

type Gateway struct {
	User             string `yaml:"user"`
	Address          string `yaml:"address"`
	IdentityFilePath string `yaml:"identity-file"`
}

type Host struct {
	User             string `yaml:"user"`
	Address          string `yaml:"address"`
	IdentityFilePath string `yaml:"identity-file"`
	MAC              string `yaml:"mac"`
}
