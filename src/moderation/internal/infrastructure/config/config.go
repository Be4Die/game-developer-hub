// Package config загружает конфигурацию сервиса moderation из YAML + env.
package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// Окружения сервиса.
const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

// Config структура всей конфигурации сервиса moderation.
type Config struct {
	Env            string `yaml:"env" env:"ENV" env-default:"local"`
	StoragePath    string `yaml:"storage_path" env:"PROJECTS_DATA_PATH" env-default:"/data/projects"`
	GRPC           GRPCConfig
	DB             DBConfig
	ProjectManager PMConfig
	Orchestrator   OrchestratorConfig
	JWT            JWTConfig
}

// OrchestratorConfig настройки подключения к сервису orchestrator.
type OrchestratorConfig struct {
	Addr string `yaml:"addr" env:"ORCHESTRATOR_GRPC_ADDR" env-default:"localhost:50052"`
}

// GRPCConfig настройки gRPC-сервера.
type GRPCConfig struct {
	Port int `yaml:"port" env:"GRPC_PORT" env-default:"50054"`
}

// DBConfig настройки PostgreSQL.
type DBConfig struct {
	Host     string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"DB_PORT" env-default:"5432"`
	User     string `yaml:"user" env:"DB_USER" env-default:"postgres"`
	Password string `yaml:"password" env:"DB_PASSWORD" env-default:"postgres"`
	Database string `yaml:"database" env:"DB_NAME" env-default:"orchestrator"`
	SSLMode  string `yaml:"ssl_mode" env:"DB_SSL_MODE" env-default:"disable"`
}

// DSN возвращает строку подключения к PostgreSQL.
func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// PMConfig настройки подключения к сервису project-manager.
type PMConfig struct {
	Addr string `yaml:"addr" env:"PROJECT_MANAGER_GRPC_ADDR" env-default:"localhost:50053"`
}

// JWTConfig настройки JWT-валидации.
type JWTConfig struct {
	Secret string `yaml:"secret" env:"JWT_SECRET" env-required:"true"`
	Issuer string `yaml:"issuer" env:"JWT_ISSUER" env-default:"gdh-sso"`
}

// MustLoad загружает конфигурацию из файла (путь из CONFIG_PATH или config/local.yaml).
func MustLoad() *Config {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config/local.yaml"
	}
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to load config: " + err.Error())
	}
	return &cfg
}
