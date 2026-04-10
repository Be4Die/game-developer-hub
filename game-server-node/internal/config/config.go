// Package config загружает и валидирует конфигурацию приложения.
package config

import (
	"errors"
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Окружения приложения.
const (
	EnvLocal string = "local"
	EnvDev   string = "dev"
	EnvProd  string = "prod"
)

// Config хранит конфигурацию приложения.
type Config struct {
	Env         string     `yaml:"env" env-required:"true"`
	StoragePath string     `yaml:"storage_path" env-required:"true"`
	GRPC        GRPCConfig `yaml:"grpc"`
	Node        NodeConfig `yaml:"node"`
	APIKey      string     `yaml:"-" env:"NODE_API_KEY" env-required:"true"`
}

// GRPCConfig хранит настройки gRPC-сервера.
type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

// NodeConfig хранит информацию об узле.
type NodeConfig struct {
	Region  string `yaml:"region" env-default:"unknown"`
	Version string `yaml:"version" env-default:"0.0.1"`
	EthName string `yaml:"eth_name" env-default:""`
}

// MustLoad загружает конфигурацию из файла или env. Паникует при ошибке.
func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist: " + path)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	if err := cfg.Validate(); err != nil {
		panic("config validation failed: " + err.Error())
	}

	return &cfg
}

// Validate проверяет корректность конфигурации.
func (c *Config) Validate() error {
	if c.Env == "" {
		return errors.New("env is required")
	}

	validEnvs := map[string]bool{EnvLocal: true, EnvDev: true, EnvProd: true}
	if !validEnvs[c.Env] {
		return errors.New("env must be one of: local, dev, prod")
	}

	if c.StoragePath == "" {
		return errors.New("storage_path is required")
	}

	if c.GRPC.Port <= 0 || c.GRPC.Port > 65535 {
		return errors.New("grpc.port must be between 1 and 65535")
	}

	if c.GRPC.Timeout <= 0 {
		return errors.New("grpc.timeout must be positive")
	}

	if c.APIKey == "" {
		return errors.New("NODE_API_KEY is required")
	}

	return nil
}

func fetchConfigPath() string {
	var res string

	// usage: --config="path/to/config.yaml"
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.StringVar(&res, "config", "", "path to config file")
	_ = fs.Parse(os.Args[1:])

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
