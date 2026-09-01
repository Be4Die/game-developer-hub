// Package config предоставляет конфигурацию SSO-сервиса.
package config

import (
	"fmt"
	"os"
	"time"

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

// Config — корневая конфигурация SSO-сервиса.
type Config struct {
	Env   string      `yaml:"env" env-required:"true"`
	GRPC  GRPCConfig  `yaml:"grpc"`
	DB    DBConfig    `yaml:"db"`
	KV    KVConfig    `yaml:"kv"`
	JWT   JWTConfig   `yaml:"jwt"`
	SMTP  SMTPConfig  `yaml:"smtp"`
	Email EmailConfig `yaml:"email"`
	// APIKey читается только из переменной окружения.
	APIKey string `yaml:"-" env:"SSO_API_KEY" env-required:"true"`
	// Admin — конфигурация учётной записи администратора.
	Admin AdminConfig `yaml:"admin"`
}

// AdminConfig — конфигурация учётной записи администратора.
type AdminConfig struct {
	Email       string `yaml:"email" env-default:"admin@welwise.com"`
	Password    string `yaml:"password" env:"ADMIN_PASSWORD" env-default:""`
	DisplayName string `yaml:"display_name" env-default:"Administrator"`
}

const (
	// WelwiseDomain — домен для внутренних пользователей.
	WelwiseDomain = "@welwise.com"
)

// GRPCConfig — конфигурация gRPC сервера.
type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

// DBConfig — конфигурация PostgreSQL.
type DBConfig struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5432"`
	User     string `yaml:"user" env-default:"postgres"`
	Password string `yaml:"password" env-default:"postgres"`
	Name     string `yaml:"name" env-default:"sso"`
	SSLMode  string `yaml:"ssl_mode" env-default:"disable"`
	MaxConns int    `yaml:"max_conns" env-default:"25"`
}

// DSN возвращает строку подключения к PostgreSQL.
func (db DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.Name, db.SSLMode,
	)
}

// KVConfig — конфигурация Valkey/Redis.
type KVConfig struct {
	Addr     string        `yaml:"addr" env-default:"localhost:6379"`
	Password string        `yaml:"password" env-default:""`
	DB       int           `yaml:"db" env-default:"0"`
	KeyTTL   time.Duration `yaml:"key_ttl" env-default:"30m"`
}

// JWTConfig хранит параметры генерации и валидации JWT-токенов.
type JWTConfig struct {
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl" env-default:"15m"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" env-default:"720h"` // 30 дней
	Secret          string        `yaml:"-" env:"JWT_SECRET" env-required:"true"`
	Issuer          string        `yaml:"issuer" env-default:"sso"`
}

// SMTPConfig хранит параметры подключения к SMTP-серверу.
type SMTPConfig struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"25"`
	Username string `yaml:"username" env-default:""`
	Password string `yaml:"password" env-default:""`
	From     string `yaml:"from" env-default:"noreply@welwise.games"`
	TLS      bool   `yaml:"tls" env-default:"false"`
}

// EmailConfig хранит параметры отправки email-уведомлений.
type EmailConfig struct {
	VerificationCodeTTL time.Duration `yaml:"verification_code_ttl" env-default:"24h"`
	ResetTokenTTL       time.Duration `yaml:"reset_token_ttl" env-default:"1h"`
}

// MustLoad загружает конфигурацию из файла и завершает работу при ошибке.
// Паникует, если файл не найден или содержит недопустимые значения.
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

// Validate проверяет обязательные поля конфигурации.
// Возвращает ошибку при отсутствии required-параметров или некорректных значениях.
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
	if c.KV.Addr == "" {
		return fmt.Errorf("kv.addr is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret is required")
	}
	if c.JWT.AccessTokenTTL <= 0 {
		return fmt.Errorf("jwt.access_token_ttl must be positive")
	}
	if c.JWT.RefreshTokenTTL <= 0 {
		return fmt.Errorf("jwt.refresh_token_ttl must be positive")
	}
	return nil
}

func fetchConfigPath() string {
	// First try --config flag (parsed by caller)
	// Then CONFIG_PATH env var
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}
	return "config/local.yaml"
}
