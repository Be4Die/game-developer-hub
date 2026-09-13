package config

import (
	"os"
	"strconv"
	"time"
)

// Config конфигурация ingress-proxy.
type Config struct {
	Port           int
	ValkeyAddr     string
	ValkeyPassword string
	ValkeyDB       int
	DialTimeout    time.Duration
}

// Load загружает параметры конфигурации из переменных окружения.
func Load() *Config {
	port := getEnvInt("PORT", 8085)
	valkeyAddr := getEnv("VALKEY_ADDR", "localhost:6379")
	valkeyPassword := getEnv("VALKEY_PASSWORD", "")
	valkeyDB := getEnvInt("VALKEY_DB", 0)
	dialTimeoutSec := getEnvInt("DIAL_TIMEOUT_SEC", 5)

	return &Config{
		Port:           port,
		ValkeyAddr:     valkeyAddr,
		ValkeyPassword: valkeyPassword,
		ValkeyDB:       valkeyDB,
		DialTimeout:    time.Duration(dialTimeoutSec) * time.Second,
	}
}

func getEnv(key, defVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defVal
}

func getEnvInt(key string, defVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defVal
}
