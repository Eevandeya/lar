package config

type Config struct {
	Gateway Gateway
	Host    map[string]Host
}

type Gateway struct {
	User             string `yaml:"host"`
	Address          string `yaml:"user"`
	IdentityFilePath string `yaml:"identity-file-path"`
}

type Host struct {
	User             string `yaml:"user"`
	Address          string `yaml:"address"`
	IdentityFilePath string `yaml:"identity-file-path"`
	MAC              string `yaml:"MAC"`
}
