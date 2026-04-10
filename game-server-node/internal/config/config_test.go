package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMustLoad(t *testing.T) {
	tests := []struct {
		name      string
		yaml      string
		wantPanic bool
	}{
		{
			name: "valid config",
			yaml: `
env: local
storage_path: /tmp/storage
grpc:
  port: 50051
  timeout: 5s
node:
  region: us-east
  version: 1.0.0
`,
			wantPanic: false,
		},
		{
			name: "missing env field",
			yaml: `
storage_path: /tmp/storage
grpc:
  port: 50051
  timeout: 5s
`,
			wantPanic: true,
		},
		{
			name: "invalid env value",
			yaml: `
env: staging
storage_path: /tmp/storage
grpc:
  port: 50051
  timeout: 5s
`,
			wantPanic: true,
		},
		{
			name: "invalid grpc port",
			yaml: `
env: dev
storage_path: /tmp/storage
grpc:
  port: 0
  timeout: 5s
`,
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfgPath := filepath.Join(tmpDir, "config.yaml")
			if err := os.WriteFile(cfgPath, []byte(tt.yaml), 0600); err != nil { //nolint:gosec // test file, permissions don't need to be restrictive
				t.Fatalf("failed to write temp config: %v", err)
			}

			t.Setenv("CONFIG_PATH", cfgPath)
			t.Setenv("NODE_API_KEY", "test-api-key-for-config")

			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("MustLoad() did not panic")
					}
				}()
			}

			cfg := MustLoad()
			if !tt.wantPanic {
				if cfg.Env != "local" {
					t.Errorf("expected env 'local', got '%s'", cfg.Env)
				}
				if cfg.GRPC.Port != 50051 {
					t.Errorf("expected grpc port 50051, got %d", cfg.GRPC.Port)
				}
				if cfg.APIKey != "test-api-key-for-config" {
					t.Errorf("expected APIKey 'test-api-key-for-config', got '%s'", cfg.APIKey)
				}
			}
		})
	}
}

func TestMustLoad_EmptyConfigPath(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustLoad() did not panic when config path is empty")
		}
	}()

	t.Setenv("CONFIG_PATH", "")
	MustLoad()
}

func TestMustLoad_NonExistentFile(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustLoad() did not panic when config file does not exist")
		}
	}()

	t.Setenv("CONFIG_PATH", "/nonexistent/path/config.yaml")
	MustLoad()
}

func TestFetchConfigPath_FromEnv(t *testing.T) {
	expected := "/some/config/path.yaml"
	t.Setenv("CONFIG_PATH", expected)

	result := fetchConfigPath()
	if result != expected {
		t.Errorf("fetchConfigPath() = %v, want %v", result, expected)
	}
}

func TestFetchConfigPath_Empty(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")

	result := fetchConfigPath()
	if result != "" {
		t.Errorf("fetchConfigPath() = %v, want empty", result)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: Config{
				Env:         EnvLocal,
				StoragePath: "/tmp/storage",
				GRPC: GRPCConfig{
					Port:    50051,
					Timeout: 5 * time.Second,
				},
				APIKey: "some-api-key",
			},
			wantErr: false,
		},
		{
			name: "valid dev env",
			cfg: Config{
				Env:         EnvDev,
				StoragePath: "/tmp/storage",
				GRPC: GRPCConfig{
					Port:    8080,
					Timeout: time.Second,
				},
				APIKey: "some-api-key",
			},
			wantErr: false,
		},
		{
			name: "valid prod env",
			cfg: Config{
				Env:         EnvProd,
				StoragePath: "/data/storage",
				GRPC: GRPCConfig{
					Port:    443,
					Timeout: 10 * time.Second,
				},
				APIKey: "some-api-key",
			},
			wantErr: false,
		},
		{
			name: "missing env",
			cfg: Config{
				StoragePath: "/tmp/storage",
				GRPC: GRPCConfig{
					Port:    50051,
					Timeout: 5 * time.Second,
				},
				APIKey: "some-api-key",
			},
			wantErr: true,
			errMsg:  "env is required",
		},
		{
			name: "invalid env value",
			cfg: Config{
				Env:         "staging",
				StoragePath: "/tmp/storage",
				GRPC: GRPCConfig{
					Port:    50051,
					Timeout: 5 * time.Second,
				},
				APIKey: "some-api-key",
			},
			wantErr: true,
			errMsg:  "env must be one of: local, dev, prod",
		},
		{
			name: "missing storage path",
			cfg: Config{
				Env: EnvLocal,
				GRPC: GRPCConfig{
					Port:    50051,
					Timeout: 5 * time.Second,
				},
				APIKey: "some-api-key",
			},
			wantErr: true,
			errMsg:  "storage_path is required",
		},
		{
			name: "grpc port zero",
			cfg: Config{
				Env:         EnvLocal,
				StoragePath: "/tmp/storage",
				GRPC: GRPCConfig{
					Port:    0,
					Timeout: 5 * time.Second,
				},
				APIKey: "some-api-key",
			},
			wantErr: true,
			errMsg:  "grpc.port must be between 1 and 65535",
		},
		{
			name: "grpc port too high",
			cfg: Config{
				Env:         EnvLocal,
				StoragePath: "/tmp/storage",
				GRPC: GRPCConfig{
					Port:    65536,
					Timeout: 5 * time.Second,
				},
				APIKey: "some-api-key",
			},
			wantErr: true,
			errMsg:  "grpc.port must be between 1 and 65535",
		},
		{
			name: "grpc port negative",
			cfg: Config{
				Env:         EnvLocal,
				StoragePath: "/tmp/storage",
				GRPC: GRPCConfig{
					Port:    -1,
					Timeout: 5 * time.Second,
				},
				APIKey: "some-api-key",
			},
			wantErr: true,
			errMsg:  "grpc.port must be between 1 and 65535",
		},
		{
			name: "grpc timeout zero",
			cfg: Config{
				Env:         EnvLocal,
				StoragePath: "/tmp/storage",
				GRPC: GRPCConfig{
					Port:    50051,
					Timeout: 0,
				},
				APIKey: "some-api-key",
			},
			wantErr: true,
			errMsg:  "grpc.timeout must be positive",
		},
		{
			name: "grpc timeout negative",
			cfg: Config{
				Env:         EnvLocal,
				StoragePath: "/tmp/storage",
				GRPC: GRPCConfig{
					Port:    50051,
					Timeout: -time.Second,
				},
				APIKey: "some-api-key",
			},
			wantErr: true,
			errMsg:  "grpc.timeout must be positive",
		},
		{
			name: "missing api key",
			cfg: Config{
				Env:         EnvLocal,
				StoragePath: "/tmp/storage",
				GRPC: GRPCConfig{
					Port:    50051,
					Timeout: 5 * time.Second,
				},
			},
			wantErr: true,
			errMsg:  "NODE_API_KEY is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error, got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}
