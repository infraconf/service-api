package mtls

import "log/slog"

type ServerConfig struct {
	Address         string              `yaml:"address"`
	AuthorizedPeers AuthorizationConfig `yaml:"authorized_peers"`
}

type AuthorizationConfig struct {
	AllowAny     bool     `yaml:"allow_any"`
	SPIFFEIDs    []string `yaml:"spiffe_ids"`
	TrustDomains []string `yaml:"trust_domains"`
	Logger       *slog.Logger
}
