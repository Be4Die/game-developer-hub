// Package config загружает конфигурацию из YAML + env.
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

// Config структура всей конфигурации сервиса.
type Config struct {
	Env              string `yaml:"env" env:"ENV" env-default:"local"`
	GRPC             GRPCConfig
	DB               DBConfig
	Valkey           ValkeyConfig
	Storage          StorageConfig
	Deployment       DeploymentConfig
	Moderation       ModerationConfig
	PurchaseService  PurchaseServiceConfig  `yaml:"purchase_service"`
	AnalyticsService AnalyticsServiceConfig `yaml:"analytics_service"`
	JWT              JWTConfig
	Internal         InternalConfig
}

// InternalConfig настройки межсервисного взаимодействия.
type InternalConfig struct {
	// APIKey — общий секрет сервисов платформы (x-internal-token). Обязателен в продуктивной среде.
	APIKey string `yaml:"api_key" env:"INTERNAL_API_KEY" env-default:""`
}

// AnalyticsServiceConfig настройки подключения к сервису аналитической статистики (PlayerGameAnalytics / Gateway).
type AnalyticsServiceConfig struct {
	Driver  string        `yaml:"driver" env:"ANALYTICS_SERVICE_DRIVER" env-default:"stub"` // stub | http | disabled
	URL     string        `yaml:"url" env:"ANALYTICS_SERVICE_URL" env-default:""`
	Timeout time.Duration `yaml:"timeout" env:"ANALYTICS_SERVICE_TIMEOUT" env-default:"10s"`
	APIKey  string        `yaml:"api_key" env:"ANALYTICS_API_KEY" env-default:""`
}

// PurchaseServiceConfig настройки подключения к сервису внутриигровых покупок (IAP).
type PurchaseServiceConfig struct {
	Driver  string        `yaml:"driver" env:"PURCHASE_SERVICE_DRIVER" env-default:"stub"` // stub | http | disabled
	URL     string        `yaml:"url" env:"PURCHASE_SERVICE_URL" env-default:""`
	Timeout time.Duration `yaml:"timeout" env:"PURCHASE_SERVICE_TIMEOUT" env-default:"5s"`
}

// ModerationConfig настройки подключения к сервису модерации.
type ModerationConfig struct {
	Addr string `yaml:"addr" env:"MODERATION_GRPC_ADDR" env-default:""`
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

// S3Config настройки подключения к SeaweedFS S3.
type S3Config struct {
	Endpoint     string `yaml:"endpoint" env:"S3_ENDPOINT" env-default:""`
	AccessKey    string `yaml:"access_key" env:"S3_ACCESS_KEY" env-default:""`
	SecretKey    string `yaml:"secret_key" env:"S3_SECRET_KEY" env-default:""`
	UseSSL       bool   `yaml:"use_ssl" env:"S3_USE_SSL" env-default:"false"`
	Region       string `yaml:"region" env:"S3_REGION" env-default:"us-east-1"`
	GamesBucket  string `yaml:"games_bucket" env:"S3_GAMES_BUCKET" env-default:"games"`
	MediaBucket  string `yaml:"media_bucket" env:"S3_MEDIA_BUCKET" env-default:"media"`
	BuildsBucket string `yaml:"builds_bucket" env:"S3_BUILDS_BUCKET" env-default:"builds"`
}

// StorageConfig настройки хранилища (s3 или fs).
type StorageConfig struct {
	Driver           string   `yaml:"driver" env:"STORAGE_DRIVER" env-default:"s3"` // s3 | fs
	ProjectsPath     string   `yaml:"projects_path" env:"STORAGE_PROJECTS_PATH" env-default:"./data/projects"`
	MaxBuildVersions int      `yaml:"max_build_versions" env:"STORAGE_MAX_BUILD_VERSIONS" env-default:"5"`
	S3               S3Config `yaml:"s3"`
}

// DeploymentConfig настройки подсистемы развертывания.
type DeploymentConfig struct {
	Mode               string   `yaml:"mode" env:"DEPLOYMENT_MODE" env-default:"s3"` // s3 | local | agent
	GamesBasePath      string   `yaml:"games_base_path" env:"DEPLOYMENT_GAMES_PATH" env-default:"./data/games"`
	URLPrefix          string   `yaml:"url_prefix" env:"DEPLOYMENT_URL_PREFIX" env-default:"/games"`
	AgentEndpoint      string   `yaml:"agent_endpoint" env:"DEPLOYMENT_AGENT_ENDPOINT" env-default:""`
	AgentAPIKey        string   `yaml:"agent_api_key" env:"DEPLOYMENT_AGENT_API_KEY" env-default:""`
	PlatformProxyHosts []string `yaml:"platform_proxy_hosts" env:"DEPLOYMENT_PLATFORM_PROXY_HOSTS" env-separator:","`
}

// JWTConfig настройки JWT-валидации.
type JWTConfig struct {
	Secret string `yaml:"secret" env:"JWT_SECRET" env-required:"true"`
	Issuer string `yaml:"issuer" env:"JWT_ISSUER" env-default:"sso"`
}

// Допустимые значения режимов работы.
const (
	StorageDriverS3 = "s3"
	StorageDriverFS = "fs"

	DeploymentModeS3    = "s3"
	DeploymentModeLocal = "local"
	DeploymentModeAgent = "agent"

	DriverStub     = "stub"
	DriverHTTP     = "http"
	DriverDisabled = "disabled"

	// EnvDocker — локальный запуск в Docker Compose.
	EnvDocker = "docker"

	minSecretLength = 32
)

// MustLoad загружает конфигурацию из файла (путь из CONFIG_PATH или config/local.yaml).
// Паникует при ошибке чтения или недопустимых значениях.
func MustLoad() *Config {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config/local.yaml"
	}
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to load config: " + err.Error())
	}
	if err := cfg.Validate(); err != nil {
		panic("invalid config: " + err.Error())
	}
	return &cfg
}

// Validate проверяет согласованность конфигурации.
// Неизвестные значения режимов — ошибка, а не тихий переход на режим по умолчанию.
func (c *Config) Validate() error {
	switch c.Env {
	case EnvLocal, EnvDocker, EnvDev, EnvProd:
	default:
		return fmt.Errorf("env must be one of: local, docker, dev, prod")
	}
	switch c.Storage.Driver {
	case StorageDriverS3, StorageDriverFS:
	default:
		return fmt.Errorf("storage.driver must be one of: s3, fs")
	}
	switch c.Deployment.Mode {
	case DeploymentModeS3, DeploymentModeLocal, DeploymentModeAgent:
	default:
		return fmt.Errorf("deployment.mode must be one of: s3, local, agent")
	}
	if (c.Storage.Driver == StorageDriverS3 || c.Deployment.Mode == DeploymentModeS3) && c.Storage.S3.Endpoint == "" {
		return fmt.Errorf("storage.s3.endpoint is required for s3 storage or deployment")
	}
	if c.Deployment.Mode == DeploymentModeAgent && (c.Deployment.AgentEndpoint == "" || c.Deployment.AgentAPIKey == "") {
		return fmt.Errorf("deployment.agent_endpoint and deployment.agent_api_key are required for agent mode")
	}
	for name, svc := range map[string][2]string{
		"purchase_service":  {c.PurchaseService.Driver, c.PurchaseService.URL},
		"analytics_service": {c.AnalyticsService.Driver, c.AnalyticsService.URL},
	} {
		switch svc[0] {
		case DriverStub, DriverDisabled:
		case DriverHTTP:
			if svc[1] == "" {
				return fmt.Errorf("%s.url is required when driver is http", name)
			}
		default:
			return fmt.Errorf("%s.driver must be one of: stub, http, disabled", name)
		}
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret is required")
	}
	if c.Env == EnvProd {
		return c.validateProd()
	}
	return nil
}

// validateProd запрещает запуск продуктивной среды с заглушками и отладочными секретами.
func (c *Config) validateProd() error {
	if len(c.JWT.Secret) < minSecretLength || isDevSecret(c.JWT.Secret) {
		return fmt.Errorf("prod: JWT_SECRET must be a unique random value of at least %d characters", minSecretLength)
	}
	if len(c.Internal.APIKey) < minSecretLength || isDevSecret(c.Internal.APIKey) {
		return fmt.Errorf("prod: INTERNAL_API_KEY must be a unique random value of at least %d characters", minSecretLength)
	}
	if c.DB.Password == "" || c.DB.Password == "postgres" {
		return fmt.Errorf("prod: DB_PASSWORD must not be empty or default")
	}
	if c.Valkey.Addr == "" {
		return fmt.Errorf("prod: valkey.addr is required (distributed project locks)")
	}
	usesS3 := c.Storage.Driver == StorageDriverS3 || c.Deployment.Mode == DeploymentModeS3
	if usesS3 && (c.Storage.S3.AccessKey == "" || c.Storage.S3.SecretKey == "") {
		return fmt.Errorf("prod: S3_ACCESS_KEY and S3_SECRET_KEY are required for s3 storage or deployment")
	}
	if c.Moderation.Addr == "" {
		return fmt.Errorf("prod: moderation.addr is required (stub moderation client is not allowed)")
	}
	if c.PurchaseService.Driver == DriverStub {
		return fmt.Errorf("prod: purchase_service.driver must be http or disabled (stub is not allowed)")
	}
	if c.AnalyticsService.Driver == DriverStub {
		return fmt.Errorf("prod: analytics_service.driver must be http or disabled (stub returns generated statistics)")
	}
	if c.Deployment.Mode == DeploymentModeAgent && isDevSecret(c.Deployment.AgentAPIKey) {
		return fmt.Errorf("prod: DEPLOYMENT_AGENT_API_KEY must not be a development value")
	}
	// Игры — недоверенный код разработчиков. На одном origin с кабинетом они получили бы
	// доступ к токенам пользователей, поэтому в prod требуется отдельный https-домен.
	if !strings.HasPrefix(c.Deployment.URLPrefix, "https://") {
		return fmt.Errorf("prod: DEPLOYMENT_URL_PREFIX must be an absolute https URL on a dedicated games domain")
	}
	return nil
}

func isDevSecret(v string) bool {
	return strings.Contains(v, "dev-") || strings.Contains(v, "local-testing")
}
