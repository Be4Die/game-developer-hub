// Package config предоставляет конфигурацию сервиса агента развертывания.
package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	// EnvLocal — локальная разработка.
	EnvLocal = "local"
	// EnvDev — staging/development окружение.
	EnvDev = "dev"
	// EnvProd — production окружение.
	EnvProd = "prod"
)

// Config описывает полную конфигурацию сервиса агента развертывания веб-игр.
type Config struct {
	Env        string           `yaml:"env" env:"ENV" env-default:"local"`
	Server     ServerConfig     `yaml:"server"`
	Deployment DeploymentConfig `yaml:"deployment"`
}

// ServerConfig описывает сетевые параметры gRPC сервера и ключ аутентификации.
type ServerConfig struct {
	Port   int    `yaml:"port" env:"PORT" env-default:"50055"`
	APIKey string `yaml:"api_key" env:"AGENT_API_KEY" env-default:"dev-deployment-agent-key"`
}

// DeploymentConfig описывает пути к статике и ограничения распаковки сборок.
type DeploymentConfig struct {
	GamesBasePath     string `yaml:"games_base_path" env:"GAMES_BASE_PATH" env-default:"/data/games"`
	URLPrefix         string `yaml:"url_prefix" env:"URL_PREFIX" env-default:"/games"`
	MaxUnpackedSizeMB int    `yaml:"max_unpacked_size_mb" env:"MAX_UNPACKED_SIZE_MB" env-default:"500"`
	MaxFilesCount     int    `yaml:"max_files_count" env:"MAX_FILES_COUNT" env-default:"50000"`
}

// MustLoad загружает конфигурацию из файла или CONFIG_PATH.
func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/local.yaml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) { //nolint:gosec
		panic(fmt.Sprintf("config file does not exist: %s", configPath))
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(fmt.Sprintf("read config %s: %v", configPath, err))
	}

	return &cfg
}
