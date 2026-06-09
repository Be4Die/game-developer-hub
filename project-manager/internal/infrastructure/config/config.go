// Package config загружает конфигурацию из YAML + env.
package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config структура всей конфигурации сервиса.
type Config struct {
	Env     string `yaml:"env" env:"ENV" env-default:"local"`
	GRPC    GRPCConfig
	DB      DBConfig
	Storage StorageConfig
	JWT     JWTConfig
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

// JWTConfig настройки JWT-валидации.
type JWTConfig struct {
	Secret string `yaml:"secret" env:"JWT_SECRET" env-required:"true"`
	Issuer string `yaml:"issuer" env:"JWT_ISSUER" env-default:"gdh-sso"`
}

// MustLoad загружает конфигурацию из файла, указанного в CONFIG_PATH.
func MustLoad(path string) *Config {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to load config: " + err.Error())
	}
	return &cfg
}
