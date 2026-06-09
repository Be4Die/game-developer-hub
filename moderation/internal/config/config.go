package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

type Config struct {
	Env  string     `yaml:"env" env-required:"true"`
	GRPC GRPCConfig `yaml:"grpc"`
	DB   DBConfig   `yaml:"db"`
}

type GRPCConfig struct {
	Port int `yaml:"port"`
}

type DBConfig struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5432"`
	User     string `yaml:"user" env-default:"postgres"`
	Password string `yaml:"password" env-default:"postgres"`
	Name     string `yaml:"name" env-default:"moderation"`
	SSLMode  string `yaml:"ssl_mode" env-default:"disable"`
}

func (db DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.Name, db.SSLMode,
	)
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if _, err := os.Stat(path); err != nil {
		panic(fmt.Errorf("config file not found: %s: %w", path, err))
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic(fmt.Errorf("failed to read config: %w", err))
	}

	if err := cfg.Validate(); err != nil {
		panic(fmt.Errorf("invalid config: %w", err))
	}

	return &cfg
}

func (c *Config) Validate() error {
	if c.Env == "" {
		return fmt.Errorf("env is required")
	}
	if c.GRPC.Port <= 0 {
		return fmt.Errorf("grpc.port must be positive")
	}
	if c.DB.Host == "" {
		return fmt.Errorf("db.host is required")
	}
	return nil
}

func fetchConfigPath() string {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}
	return "config/local.yaml"
}
