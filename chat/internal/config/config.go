package config

import (
	"os"
)

type Config struct {
	Database DatabaseConfig
	GRPC     GRPCConfig
	SSO      SSOConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type GRPCConfig struct {
	Addr string
}

type SSOConfig struct {
	Addr   string
	APIKey string
}

func (d *DatabaseConfig) DSN() string {
	return "postgres://" + d.User + ":" + d.Password + "@" + d.Host + ":" + d.Port + "/" + d.DBName + "?sslmode=disable"
}

func Load() (*Config, error) {
	return &Config{
		Database: DatabaseConfig{
			Host:     envOr("CHAT_DB_HOST", "localhost"),
			Port:     envOr("CHAT_DB_PORT", "5432"),
			User:     envOr("CHAT_DB_USER", "postgres"),
			Password: envOr("CHAT_DB_PASSWORD", "postgres"),
			DBName:   envOr("CHAT_DB_NAME", "chat"),
		},
		GRPC: GRPCConfig{
			Addr: envOr("CHAT_GRPC_ADDR", ":9090"),
		},
		SSO: SSOConfig{
			Addr:   envOr("SSO_GRPC_ADDR", "localhost:9090"),
			APIKey: envOr("CHAT_SSO_API_KEY", "dev-api-key"),
		},
	}, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}