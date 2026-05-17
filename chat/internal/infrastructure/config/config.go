package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

type Config struct {
	Env   string     `yaml:"env" env-required:"true"`
	GRPC  GRPCConfig `yaml:"grpc"`
	DB    DBConfig   `yaml:"db"`
	KV    KVConfig   `yaml:"kv"`
	APIKey string    `yaml:"-" env:"CHAT_API_KEY" env-required:"true"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type DBConfig struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5432"`
	User     string `yaml:"user" env-default:"postgres"`
	Password string `yaml:"password" env-default:"postgres"`
	Name     string `yaml:"name" env-default:"chat"`
	SSLMode  string `yaml:"ssl_mode" env-default:"disable"`
	MaxConns int    `yaml:"max_conns" env-default:"25"`
}

func (db DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.Name, db.SSLMode,
	)
}

type KVConfig struct {
	Addr     string        `yaml:"addr" env-default:"localhost:6379"`
	Password string        `yaml:"password" env-default:""`
	DB       int           `yaml:"db" env-default:"2"`
}

func MustLoad() *Config {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config/local.yaml"
	}

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
