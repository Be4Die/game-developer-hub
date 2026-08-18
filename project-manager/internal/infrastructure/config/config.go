// Package config загружает конфигурацию из YAML + env.
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

// Config структура всей конфигурации сервиса.
type Config struct {
	Env        string `yaml:"env" env:"ENV" env-default:"local"`
	GRPC       GRPCConfig
	DB         DBConfig
	Valkey     ValkeyConfig
	Storage    StorageConfig
	Deployment DeploymentConfig
	JWT        JWTConfig
}

// ValkeyConfig настройки подключения к Valkey/Redis для распределенных блокировок.
type ValkeyConfig struct {
	Addr     string `yaml:"addr" env:"VALKEY_ADDR" env-default:"localhost:6379"`
	Password string `yaml:"password" env:"VALKEY_PASSWORD" env-default:""`
	DB       int    `yaml:"db" env:"VALKEY_DB" env-default:"2"`
}

// GRPCConfig настройки gRPC-сервера.
type GRPCConfig struct {
	Port int `yaml:"port" env:"GRPC_PORT" env-default:"50053"`
}

// DBConfig настройки PostgreSQL.
type DBConfig struct {
	Host     string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"DB_PORT" env-default:"5432"`
	User     string `yaml:"user" env:"DB_USER" env-default:"postgres"`
	Password string `yaml:"password" env:"DB_PASSWORD" env-default:"postgres"`
	Database string `yaml:"database" env:"DB_NAME" env-default:"project_manager"`
	SSLMode  string `yaml:"ssl_mode" env:"DB_SSL_MODE" env-default:"disable"`
}

// DSN возвращает строку подключения к PostgreSQL.
func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// StorageConfig настройки файлового хранилища.
type StorageConfig struct {
	ProjectsPath     string `yaml:"projects_path" env:"STORAGE_PROJECTS_PATH" env-default:"./data/projects"`
	MaxBuildVersions int    `yaml:"max_build_versions" env:"STORAGE_MAX_BUILD_VERSIONS" env-default:"5"`
}

// DeploymentConfig настройки подсистемы развертывания.
type DeploymentConfig struct {
	Mode          string `yaml:"mode" env:"DEPLOYMENT_MODE" env-default:"local"` // local | agent
	GamesBasePath string `yaml:"games_base_path" env:"DEPLOYMENT_GAMES_PATH" env-default:"./data/games"`
	URLPrefix     string `yaml:"url_prefix" env:"DEPLOYMENT_URL_PREFIX" env-default:"/games"`
	AgentEndpoint string `yaml:"agent_endpoint" env:"DEPLOYMENT_AGENT_ENDPOINT" env-default:""`
	AgentAPIKey   string `yaml:"agent_api_key" env:"DEPLOYMENT_AGENT_API_KEY" env-default:""`
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
