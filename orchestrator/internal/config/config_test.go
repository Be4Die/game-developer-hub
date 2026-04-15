package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfig_Validate_Success(t *testing.T) {
	cfg := &Config{
		Env:           EnvLocal,
		HTTP:          HTTPConfig{Port: 8080, ReadTimeout: 10e9, WriteTimeout: 30e9, IdleTimeout: 60e9},
		DB:            DBConfig{Host: "localhost", Port: 5432, User: "postgres", Name: "orchestrator", SSLMode: "disable", MaxConns: 25},
		KV:            KVConfig{Addr: "localhost:6379", KeyTTL: 45e9},
		Storage:       StorageConfig{BuildsPath: "./data/builds"},
		GRPCClient:    GRPCClientConfig{Timeout: 30e9, ConnectTimeout: 10e9, KeepAliveTime: 30e9, KeepAliveTimeout: 10e9, MaxMessageSize: 16777216},
		NodeHeartbeat: NodeHeartbeatCfg{CheckInterval: 15e9, InactivityTimeout: 60e9},
		Limits:        LimitsConfig{MaxBuildsPerGame: 10, MaxInstancesPerGame: 50, MaxLogTailLines: 5000, MaxBuildSizeBytes: 2147483648},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfig_Validate_Errors(t *testing.T) {
	tests := []struct {
		name      string
		cfg       Config
		wantError string
	}{
		{
			name:      "empty env",
			cfg:       Config{Env: ""},
			wantError: "env is required",
		},
		{
			name:      "invalid env",
			cfg:       Config{Env: "staging"},
			wantError: "env must be one of",
		},
		{
			name:      "http port zero",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 0}},
			wantError: "http.port must be between 1 and 65535",
		},
		{
			name:      "http port negative",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: -1}},
			wantError: "http.port must be between 1 and 65535",
		},
		{
			name:      "http port too large",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 70000}},
			wantError: "http.port must be between 1 and 65535",
		},
		{
			name:      "http read timeout zero",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 0}},
			wantError: "http.read_timeout must be positive",
		},
		{
			name:      "http write timeout zero",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 0}},
			wantError: "http.write_timeout must be positive",
		},
		{
			name:      "http idle timeout zero",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 0}},
			wantError: "http.idle_timeout must be positive",
		},
		{
			name:      "db host empty",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: ""}},
			wantError: "db.host is required",
		},
		{
			name:      "db port zero",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 0}},
			wantError: "db.port must be between 1 and 65535",
		},
		{
			name:      "db user empty",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 5432, User: ""}},
			wantError: "db.user is required",
		},
		{
			name:      "db name empty",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 5432, User: "postgres", Name: ""}},
			wantError: "db.name is required",
		},
		{
			name:      "db invalid ssl mode",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 5432, User: "postgres", Name: "orchestrator", SSLMode: "invalid", MaxConns: 25}},
			wantError: "db.ssl_mode must be one of",
		},
		{
			name:      "db max conns zero",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 5432, User: "postgres", Name: "orchestrator", SSLMode: "disable", MaxConns: 0}},
			wantError: "db.max_conns must be positive",
		},
		{
			name:      "storage builds path empty",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 5432, User: "postgres", Name: "orchestrator", SSLMode: "disable", MaxConns: 25}, KV: KVConfig{Addr: "localhost:6379", KeyTTL: 45e9}},
			wantError: "storage.builds_path is required",
		},
		{
			name:      "kv addr empty",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 5432, User: "postgres", Name: "orchestrator", SSLMode: "disable", MaxConns: 25}, Storage: StorageConfig{BuildsPath: "./builds"}},
			wantError: "kv.addr is required",
		},
		{
			name:      "kv key ttl zero",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 5432, User: "postgres", Name: "orchestrator", SSLMode: "disable", MaxConns: 25}, Storage: StorageConfig{BuildsPath: "./builds"}, KV: KVConfig{Addr: "localhost:6379", KeyTTL: 0}},
			wantError: "kv.key_ttl must be positive",
		},
		{
			name:      "grpc timeout zero",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 5432, User: "postgres", Name: "orchestrator", SSLMode: "disable", MaxConns: 25}, Storage: StorageConfig{BuildsPath: "./builds"}, KV: KVConfig{Addr: "localhost:6379", KeyTTL: 45e9}, GRPCClient: GRPCClientConfig{Timeout: 0}},
			wantError: "grpc_client.timeout must be positive",
		},
		{
			name:      "grpc connect timeout zero",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 5432, User: "postgres", Name: "orchestrator", SSLMode: "disable", MaxConns: 25}, Storage: StorageConfig{BuildsPath: "./builds"}, KV: KVConfig{Addr: "localhost:6379", KeyTTL: 45e9}, GRPCClient: GRPCClientConfig{Timeout: 1e9, ConnectTimeout: 0}},
			wantError: "grpc_client.connect_timeout must be positive",
		},
		{
			name:      "grpc keepalive negative",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 5432, User: "postgres", Name: "orchestrator", SSLMode: "disable", MaxConns: 25}, Storage: StorageConfig{BuildsPath: "./builds"}, KV: KVConfig{Addr: "localhost:6379", KeyTTL: 45e9}, GRPCClient: GRPCClientConfig{Timeout: 1e9, ConnectTimeout: 1e9, KeepAliveTime: -1}},
			wantError: "grpc_client.keepalive_time must be non-negative",
		},
		{
			name:      "grpc max message size too small",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 5432, User: "postgres", Name: "orchestrator", SSLMode: "disable", MaxConns: 25}, Storage: StorageConfig{BuildsPath: "./builds"}, KV: KVConfig{Addr: "localhost:6379", KeyTTL: 45e9}, GRPCClient: GRPCClientConfig{Timeout: 1e9, ConnectTimeout: 1e9, MaxMessageSize: 512}},
			wantError: "grpc_client.max_message_size must be at least 1024 bytes",
		},
		{
			name:      "heartbeat check interval zero",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 5432, User: "postgres", Name: "orchestrator", SSLMode: "disable", MaxConns: 25}, Storage: StorageConfig{BuildsPath: "./builds"}, KV: KVConfig{Addr: "localhost:6379", KeyTTL: 45e9}, GRPCClient: GRPCClientConfig{Timeout: 1e9, ConnectTimeout: 1e9, MaxMessageSize: 16777216}},
			wantError: "node_heartbeat.check_interval must be positive",
		},
		{
			name:      "heartbeat inactivity timeout too small",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 5432, User: "postgres", Name: "orchestrator", SSLMode: "disable", MaxConns: 25}, Storage: StorageConfig{BuildsPath: "./builds"}, KV: KVConfig{Addr: "localhost:6379", KeyTTL: 45e9}, GRPCClient: GRPCClientConfig{Timeout: 1e9, ConnectTimeout: 1e9, MaxMessageSize: 16777216}, NodeHeartbeat: NodeHeartbeatCfg{CheckInterval: 15e9, InactivityTimeout: 30e9}},
			wantError: "node_heartbeat.inactivity_timeout must be at least 3x check_interval",
		},
		{
			name:      "max builds per game zero",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 5432, User: "postgres", Name: "orchestrator", SSLMode: "disable", MaxConns: 25}, Storage: StorageConfig{BuildsPath: "./builds"}, KV: KVConfig{Addr: "localhost:6379", KeyTTL: 45e9}, GRPCClient: GRPCClientConfig{Timeout: 1e9, ConnectTimeout: 1e9, MaxMessageSize: 16777216}, NodeHeartbeat: NodeHeartbeatCfg{CheckInterval: 15e9, InactivityTimeout: 60e9}, Limits: LimitsConfig{MaxBuildsPerGame: 0}},
			wantError: "limits.max_builds_per_game must be at least 1",
		},
		{
			name:      "max instances per game zero",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 5432, User: "postgres", Name: "orchestrator", SSLMode: "disable", MaxConns: 25}, Storage: StorageConfig{BuildsPath: "./builds"}, KV: KVConfig{Addr: "localhost:6379", KeyTTL: 45e9}, GRPCClient: GRPCClientConfig{Timeout: 1e9, ConnectTimeout: 1e9, MaxMessageSize: 16777216}, NodeHeartbeat: NodeHeartbeatCfg{CheckInterval: 15e9, InactivityTimeout: 60e9}, Limits: LimitsConfig{MaxBuildsPerGame: 1, MaxInstancesPerGame: 0}},
			wantError: "limits.max_instances_per_game must be at least 1",
		},
		{
			name:      "max log tail lines too small",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 5432, User: "postgres", Name: "orchestrator", SSLMode: "disable", MaxConns: 25}, Storage: StorageConfig{BuildsPath: "./builds"}, KV: KVConfig{Addr: "localhost:6379", KeyTTL: 45e9}, GRPCClient: GRPCClientConfig{Timeout: 1e9, ConnectTimeout: 1e9, MaxMessageSize: 16777216}, NodeHeartbeat: NodeHeartbeatCfg{CheckInterval: 15e9, InactivityTimeout: 60e9}, Limits: LimitsConfig{MaxBuildsPerGame: 1, MaxInstancesPerGame: 1, MaxLogTailLines: 50}},
			wantError: "limits.max_log_tail_lines must be at least 100",
		},
		{
			name:      "max build size too small",
			cfg:       Config{Env: EnvLocal, HTTP: HTTPConfig{Port: 8080, ReadTimeout: 1e9, WriteTimeout: 1e9, IdleTimeout: 1e9}, DB: DBConfig{Host: "localhost", Port: 5432, User: "postgres", Name: "orchestrator", SSLMode: "disable", MaxConns: 25}, Storage: StorageConfig{BuildsPath: "./builds"}, KV: KVConfig{Addr: "localhost:6379", KeyTTL: 45e9}, GRPCClient: GRPCClientConfig{Timeout: 1e9, ConnectTimeout: 1e9, MaxMessageSize: 16777216}, NodeHeartbeat: NodeHeartbeatCfg{CheckInterval: 15e9, InactivityTimeout: 60e9}, Limits: LimitsConfig{MaxBuildsPerGame: 1, MaxInstancesPerGame: 1, MaxLogTailLines: 100, MaxBuildSizeBytes: 512}},
			wantError: "limits.max_build_size must be at least 1024 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("expected error containing %q, got %q", tt.wantError, err.Error())
			}
		})
	}
}

func TestHTTPConfig_Addr(t *testing.T) {
	h := HTTPConfig{Port: 8080}
	if got := h.Addr(); got != ":8080" {
		t.Errorf("expected :8080, got %q", got)
	}
}

func TestDBConfig_DSN(t *testing.T) {
	db := DBConfig{
		User:     "admin",
		Password: "secret",
		Host:     "db.example.com",
		Port:     5433,
		Name:     "testdb",
		SSLMode:  "require",
	}
	want := "postgres://admin:secret@db.example.com:5433/testdb?sslmode=require"
	if got := db.DSN(); got != want {
		t.Errorf("DSN = %q, want %q", got, want)
	}
}

func TestMustLoad_PanicOnMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "nonexistent.yaml")
	t.Setenv("CONFIG_PATH", cfgPath)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustLoad did not panic on missing config file")
		}
	}()

	MustLoad()
}

func TestMustLoad_PanicOnInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	os.WriteFile(cfgPath, []byte(":::invalid::yaml:::"), 0600)
	t.Setenv("CONFIG_PATH", cfgPath)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustLoad did not panic on invalid YAML")
		}
	}()

	MustLoad()
}

func TestMustLoad_Success(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	yaml := `env: local
http:
  port: 9090
db:
  host: localhost
  port: 5432
  user: postgres
  name: orchestrator
kv:
  addr: localhost:6379
storage:
  builds_path: ./builds
grpc_client:
  timeout: 30s
  connect_timeout: 10s
node_heartbeat:
  check_interval: 15s
  inactivity_timeout: 60s
limits:
  max_builds_per_game: 10
  max_instances_per_game: 50
  max_log_tail_lines: 5000
  max_build_size: 2147483648
`
	os.WriteFile(cfgPath, []byte(yaml), 0600)
	t.Setenv("CONFIG_PATH", cfgPath)

	cfg := MustLoad()
	if cfg.Env != EnvLocal {
		t.Errorf("env = %q, want %q", cfg.Env, EnvLocal)
	}
	if cfg.HTTP.Port != 9090 {
		t.Errorf("http.port = %d, want 9090", cfg.HTTP.Port)
	}
}

func TestMustLoad_PanicOnValidationError(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	// Invalid: env is missing
	yaml := `http:
  port: 8080
`
	os.WriteFile(cfgPath, []byte(yaml), 0600)
	t.Setenv("CONFIG_PATH", cfgPath)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustLoad did not panic on validation error")
		}
	}()

	MustLoad()
}
