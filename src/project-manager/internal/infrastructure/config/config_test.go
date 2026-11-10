package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validProdConfig возвращает минимальную корректную конфигурацию продуктивной среды.
func validProdConfig() Config {
	return Config{
		Env:     EnvProd,
		DB:      DBConfig{Password: "strong-db-password"},
		Valkey:  ValkeyConfig{Addr: "valkey:6379"},
		Storage: StorageConfig{Driver: StorageDriverS3, S3: S3Config{
			Endpoint:  "seaweedfs:8333",
			AccessKey: "gdh-access",
			SecretKey: "gdh-secret-key",
		}},
		Deployment: DeploymentConfig{
			Mode:      DeploymentModeS3,
			URLPrefix: "https://games.example.com/games",
		},
		Moderation:       ModerationConfig{Addr: "moderation:50054"},
		PurchaseService:  PurchaseServiceConfig{Driver: DriverDisabled},
		AnalyticsService: AnalyticsServiceConfig{Driver: DriverDisabled},
		JWT:              JWTConfig{Secret: "0123456789abcdef0123456789abcdef"},
		Internal:         InternalConfig{APIKey: "fedcba9876543210fedcba9876543210"},
	}
}

func TestConfig_ValidateProd(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(c *Config)
		wantErr string
	}{
		{name: "valid", mutate: func(*Config) {}},
		{
			name:    "games on the same origin",
			mutate:  func(c *Config) { c.Deployment.URLPrefix = "/games" },
			wantErr: "DEPLOYMENT_URL_PREFIX",
		},
		{
			name:    "games over plain http",
			mutate:  func(c *Config) { c.Deployment.URLPrefix = "http://games.example.com/games" },
			wantErr: "DEPLOYMENT_URL_PREFIX",
		},
		{
			name:    "s3 without credentials",
			mutate:  func(c *Config) { c.Storage.S3.SecretKey = "" },
			wantErr: "S3_SECRET_KEY",
		},
		{
			name:    "development jwt secret",
			mutate:  func(c *Config) { c.JWT.Secret = "dev-jwt-secret-for-local-testing" },
			wantErr: "JWT_SECRET",
		},
		{
			name:    "missing internal key",
			mutate:  func(c *Config) { c.Internal.APIKey = "" },
			wantErr: "INTERNAL_API_KEY",
		},
		{
			name:    "default db password",
			mutate:  func(c *Config) { c.DB.Password = "postgres" },
			wantErr: "DB_PASSWORD",
		},
		{
			name:    "stub purchase service",
			mutate:  func(c *Config) { c.PurchaseService.Driver = DriverStub },
			wantErr: "purchase_service",
		},
		{
			name:    "stub analytics service",
			mutate:  func(c *Config) { c.AnalyticsService.Driver = DriverStub },
			wantErr: "analytics_service",
		},
		{
			name:    "moderation client not configured",
			mutate:  func(c *Config) { c.Moderation.Addr = "" },
			wantErr: "moderation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validProdConfig()
			tt.mutate(&cfg)

			err := cfg.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestConfig_ValidateRejectsUnknownModes(t *testing.T) {
	cfg := validProdConfig()
	cfg.Env = EnvLocal
	cfg.Storage.Driver = "memory"

	assert.Error(t, cfg.Validate())

	cfg = validProdConfig()
	cfg.Env = EnvLocal
	cfg.PurchaseService.Driver = DriverHTTP

	assert.Error(t, cfg.Validate(), "http driver without url must be rejected")
}
