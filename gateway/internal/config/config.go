package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
)

// Config содержит конфигурацию API Gateway.
type Config struct {
	HTTPAddr     string       `yaml:"http_addr" env-default:":8080"`
	GRPC         GRPCServices `yaml:"grpc"`
	CORS         CORSConfig   `yaml:"cors"`
	AllowedHosts []string     `yaml:"allowed_hosts"`
}

// GRPCServices содержит адреса gRPC-сервисов.
type GRPCServices struct {
	Orchestrator ServiceAddr `yaml:"orchestrator"`
	SSO          ServiceAddr `yaml:"sso"`
}

// ServiceAddr — адрес одного gRPC-сервиса.
type ServiceAddr struct {
	Address string `yaml:"address" env-required:"true"`
}

// CORSConfig содержит настройки CORS.
type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods" env-default:"GET,POST,PUT,PATCH,DELETE,OPTIONS"`
	AllowedHeaders []string `yaml:"allowed_headers" env-default:"Content-Type,Authorization"`
}

// Load загружает конфигурацию из YAML и переменных окружения.
func Load() (*Config, error) {
	cfgPath := os.Getenv("GATEWAY_CONFIG")
	if cfgPath == "" {
		cfgPath = "config/local.yaml"
	}

	var cfg Config
	if err := envconfig.Process("gateway", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
