package mtls

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	WorkloadAPIPath string       `yaml:"workload_api_path"`
	Server          ServerConfig `yaml:"server"`
	Client          ClientConfig `yaml:"client"`
}

type ServerConfig struct {
	Address         string              `yaml:"address"`
	AuthorizedPeers AuthorizationConfig `yaml:"authorized_peers"`
}

type ClientConfig struct {
	Target          string              `yaml:"target"`
	AuthorizedPeers AuthorizationConfig `yaml:"authorized_peers"`
}

type AuthorizationConfig struct {
	AllowAny     bool     `yaml:"allow_any"`
	SPIFFEIDs    []string `yaml:"spiffe_ids"`
	TrustDomains []string `yaml:"trust_domains"`
}

func LoadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return &cfg, nil
}
