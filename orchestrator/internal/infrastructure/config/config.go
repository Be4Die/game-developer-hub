// Package config загружает и валидирует конфигурацию оркестратора.
package config

import (
	"errors"
	"flag"
	"fmt"
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

// Config хранит конфигурацию оркестратора.
type Config struct {
	Env           string           `yaml:"env" env-required:"true"`
	GRPC          GRPCConfig       `yaml:"grpc"`
	JWT           JWTConfig        `yaml:"jwt"`
	DB            DBConfig         `yaml:"db"`
	KV            KVConfig         `yaml:"kv"`
	Storage       StorageConfig    `yaml:"storage"`
	GRPCClient    GRPCClientConfig `yaml:"grpc_client"`
	NodeHeartbeat NodeHeartbeatCfg `yaml:"node_heartbeat"`
	Limits        LimitsConfig     `yaml:"limits"`
}

// JWTConfig хранит параметры валидации JWT-токенов.
type JWTConfig struct {
	Secret string `yaml:"secret" env:"ORCHESTRATOR_JWT_SECRET" env-required:"true"`
	Issuer string `yaml:"issuer" env-default:"sso"`
}

// GRPCConfig хранит настройки gRPC-сервера.
type GRPCConfig struct {
	Port int `yaml:"port" env-default:"50052"`
}

// DBConfig хранит настройки подключения к PostgreSQL.
type DBConfig struct {
	Host     string `yaml:"host"     env-default:"localhost"`
	Port     int    `yaml:"port"     env-default:"5432"`
	User     string `yaml:"user"     env-default:"postgres"`
	Password string `yaml:"password" env-default:"postgres"`
	Name     string `yaml:"name"     env-default:"orchestrator"`
	SSLMode  string `yaml:"ssl_mode" env-default:"disable"`
	MaxConns int    `yaml:"max_conns" env-default:"25"`
}

// DSN возвращает строку подключения к PostgreSQL.
func (db DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.User,
		db.Password,
		db.Host,
		db.Port,
		db.Name,
		db.SSLMode,
	)
}

// KVConfig хранит настройки подключения к KV-хранилищу (Valkey/Redis).
type KVConfig struct {
	Addr     string        `yaml:"addr"     env-default:"localhost:6379"`
	Password string        `yaml:"password" env-default:""`
	DB       int           `yaml:"db"       env-default:"0"`
	KeyTTL   time.Duration `yaml:"key_ttl"  env-default:"45s"`
}

// StorageConfig хранит настройки файлового хранилища билдов.
type StorageConfig struct {
	BuildsPath string `yaml:"builds_path" env-default:"./data/builds"`
}

// GRPCClientConfig хранит настройки gRPC-клиента для подключения к нодам.
type GRPCClientConfig struct {
	// Timeout — таймаут на один gRPC-вызов к ноде.
	Timeout time.Duration `yaml:"timeout" env-default:"30s"`

	// ConnectTimeout — таймаут на установление gRPC-соединения с нодой.
	ConnectTimeout time.Duration `yaml:"connect_timeout" env-default:"10s"`

	// KeepAliveTime — интервал отправки keepalive-пингов к ноде.
	// Ноль отключает keepalive.
	KeepAliveTime time.Duration `yaml:"keepalive_time" env-default:"30s"`

	// KeepAliveTimeout — время ожидания ответа на keepalive-пинг.
	KeepAliveTimeout time.Duration `yaml:"keepalive_timeout" env-default:"10s"`

	// MaxMessageSize — максимальный размер gRPC-сообщения в байтах.
	// Используется при загрузке Docker-образов (чанки по 64 КБ).
	MaxMessageSize int `yaml:"max_message_size" env-default:"16777216"` // 16 MB

	// EnableCompression включает gzip-сжатие для gRPC-вызовов.
	EnableCompression bool `yaml:"enable_compression" env-default:"true"`
}

// NodeHeartbeatCfg хранит настройки мониторинга жизнеспособности нод.
type NodeHeartbeatCfg struct {
	// CheckInterval — период опроса нод через Heartbeat RPC.
	CheckInterval time.Duration `yaml:"check_interval" env-default:"15s"`

	// InactivityTimeout — время без ответов от ноды до перевода в offline.
	// Должно быть больше CheckInterval минимум в 3 раза.
	InactivityTimeout time.Duration `yaml:"inactivity_timeout" env-default:"60s"`
}

// LimitsConfig хранит лимиты на ресурсы и сущности.
type LimitsConfig struct {
	// MaxBuildsPerGame — максимальное количество серверных билдов на одну игру.
	MaxBuildsPerGame int `yaml:"max_builds_per_game" env-default:"10"`

	// MaxInstancesPerGame — максимальное количество одновременных инстансов на игру.
	MaxInstancesPerGame int `yaml:"max_instances_per_game" env-default:"50"`

	// MaxLogTailLines — максимальное количество строк лога в запросе tail.
	// Запросы с большим значением урезаются до этого лимита.
	MaxLogTailLines uint32 `yaml:"max_log_tail_lines" env-default:"5000"`

	// MaxBuildSizeBytes — максимальный размер загружаемого билда в байтах.
	MaxBuildSizeBytes int64 `yaml:"max_build_size" env-default:"2147483648"` // 2 GB
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

	if c.JWT.Secret == "" {
		return errors.New("jwt.secret is required")
	}

	if c.GRPC.Port <= 0 || c.GRPC.Port > 65535 {
		return errors.New("grpc.port must be between 1 and 65535")
	}

	if c.DB.Host == "" {
		return errors.New("db.host is required")
	}
	if c.DB.Port <= 0 || c.DB.Port > 65535 {
		return errors.New("db.port must be between 1 and 65535")
	}
	if c.DB.User == "" {
		return errors.New("db.user is required")
	}
	if c.DB.Name == "" {
		return errors.New("db.name is required")
	}
	switch c.DB.SSLMode {
	case "disable", "require", "verify-ca", "verify-full":
		// ok
	default:
		return errors.New("db.ssl_mode must be one of: disable, require, verify-ca, verify-full")
	}
	if c.DB.MaxConns < 1 {
		return errors.New("db.max_conns must be positive")
	}

	if c.Storage.BuildsPath == "" {
		return errors.New("storage.builds_path is required")
	}

	if c.KV.Addr == "" {
		return errors.New("kv.addr is required")
	}
	if c.KV.KeyTTL <= 0 {
		return errors.New("kv.key_ttl must be positive")
	}

	if c.GRPCClient.Timeout <= 0 {
		return errors.New("grpc_client.timeout must be positive")
	}
	if c.GRPCClient.ConnectTimeout <= 0 {
		return errors.New("grpc_client.connect_timeout must be positive")
	}
	if c.GRPCClient.KeepAliveTime < 0 {
		return errors.New("grpc_client.keepalive_time must be non-negative")
	}
	if c.GRPCClient.KeepAliveTimeout <= 0 && c.GRPCClient.KeepAliveTime > 0 {
		return errors.New("grpc_client.keepalive_timeout must be positive when keepalive_time is set")
	}
	if c.GRPCClient.MaxMessageSize < 1024 {
		return errors.New("grpc_client.max_message_size must be at least 1024 bytes")
	}

	if c.NodeHeartbeat.CheckInterval <= 0 {
		return errors.New("node_heartbeat.check_interval must be positive")
	}
	if c.NodeHeartbeat.InactivityTimeout <= c.NodeHeartbeat.CheckInterval*3 {
		return errors.New("node_heartbeat.inactivity_timeout must be at least 3x check_interval")
	}

	if c.Limits.MaxBuildsPerGame < 1 {
		return errors.New("limits.max_builds_per_game must be at least 1")
	}
	if c.Limits.MaxInstancesPerGame < 1 {
		return errors.New("limits.max_instances_per_game must be at least 1")
	}
	if c.Limits.MaxLogTailLines < 100 {
		return errors.New("limits.max_log_tail_lines must be at least 100")
	}
	if c.Limits.MaxBuildSizeBytes < 1024 {
		return errors.New("limits.max_build_size must be at least 1024 bytes")
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
