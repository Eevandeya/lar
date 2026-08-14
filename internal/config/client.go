package config

type ClientConfig struct {
	Gateway Gateway `yaml:"gateway"`
}

type Gateway struct {
	Address string `yaml:"address"`
	Secret  string `yaml:"secret"`
}
