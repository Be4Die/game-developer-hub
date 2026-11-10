// Package config предоставляет конфигурацию SSO-сервиса.
package config

import (
	"fmt"
	"os"
	"strings"
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
	Env   string      `yaml:"env" env:"ENV" env-required:"true"`
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
	Email       string `yaml:"email" env:"ADMIN_EMAIL" env-default:"admin@welwise.com"`
	Password    string `yaml:"password" env:"ADMIN_PASSWORD" env-default:""`
	DisplayName string `yaml:"display_name" env-default:"Admin"`
}

const (
	// EmailDriverStub — письма не отправляются, код пишется в лог (только для разработки).
	EmailDriverStub = "stub"
	// EmailDriverSMTP — письма отправляются через SMTP-сервер.
	EmailDriverSMTP = "smtp"

	minSecretLength        = 32
	minAdminPasswordLength = 12
)

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
	Host     string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"DB_PORT" env-default:"5432"`
	User     string `yaml:"user" env:"DB_USER" env-default:"postgres"`
	Password string `yaml:"password" env:"DB_PASSWORD" env-default:"postgres"`
	Name     string `yaml:"name" env:"DB_NAME" env-default:"sso"`
	SSLMode  string `yaml:"ssl_mode" env:"DB_SSL_MODE" env-default:"disable"`
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
	Addr     string        `yaml:"addr" env:"KV_ADDR" env-default:"localhost:6379"`
	Password string        `yaml:"password" env:"KV_PASSWORD" env-default:""`
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
	Host     string `yaml:"host" env:"SMTP_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"SMTP_PORT" env-default:"25"`
	Username string `yaml:"username" env:"SMTP_USERNAME" env-default:""`
	Password string `yaml:"password" env:"SMTP_PASSWORD" env-default:""`
	From     string `yaml:"from" env:"SMTP_FROM" env-default:"noreply@welwise.games"`
	TLS      bool   `yaml:"tls" env:"SMTP_TLS" env-default:"false"`
}

// EmailConfig хранит параметры отправки email-уведомлений.
type EmailConfig struct {
	// Driver — способ доставки писем: "stub" (код пишется в лог, только для разработки) или "smtp".
	Driver              string        `yaml:"driver" env:"EMAIL_DRIVER" env-default:"stub"`
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
	switch c.Email.Driver {
	case EmailDriverStub, EmailDriverSMTP:
	default:
		return fmt.Errorf("email.driver must be one of: stub, smtp")
	}
	if c.Email.Driver == EmailDriverSMTP && c.SMTP.Host == "" {
		return fmt.Errorf("smtp.host is required when email.driver is smtp")
	}
	if c.Env == EnvProd {
		return c.validateProd()
	}
	return nil
}

// validateProd запрещает запуск продуктивной среды с отладочными значениями и заглушками.
func (c *Config) validateProd() error {
	if len(c.JWT.Secret) < minSecretLength || isDevSecret(c.JWT.Secret) {
		return fmt.Errorf("prod: JWT_SECRET must be a unique random value of at least %d characters", minSecretLength)
	}
	if len(c.APIKey) < minSecretLength || isDevSecret(c.APIKey) {
		return fmt.Errorf("prod: SSO_API_KEY must be a unique random value of at least %d characters", minSecretLength)
	}
	if c.Email.Driver != EmailDriverSMTP {
		return fmt.Errorf("prod: email.driver must be smtp (stub sender only logs codes and makes registration impossible)")
	}
	if !c.SMTP.TLS {
		return fmt.Errorf("prod: smtp.tls must be enabled")
	}
	if c.Admin.Password != "" && (len(c.Admin.Password) < minAdminPasswordLength || c.Admin.Password == "password123") {
		return fmt.Errorf("prod: ADMIN_PASSWORD must be at least %d characters and not a default value", minAdminPasswordLength)
	}
	if c.DB.Password == "" || c.DB.Password == "postgres" {
		return fmt.Errorf("prod: DB_PASSWORD must not be empty or default")
	}
	return nil
}

func isDevSecret(v string) bool {
	return strings.Contains(v, "dev-") || strings.Contains(v, "local-testing")
}

func fetchConfigPath() string {
	// First try --config flag (parsed by caller)
	// Then CONFIG_PATH env var
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}
	return "config/local.yaml"
}
