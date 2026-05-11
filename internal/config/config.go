package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Node struct {
	Name string `yaml:"name"`
	IP   string `yaml:"ip"`
	Role string `yaml:"role"`
	User string `yaml:"user"`
}

type Config struct {
	SSHKeyPath string `yaml:"ssh_key_path"`
	Nodes      []Node `yaml:"nodes"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}
