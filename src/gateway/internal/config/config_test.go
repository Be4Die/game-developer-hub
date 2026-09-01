package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Load(t *testing.T) {
	_ = os.Setenv("GATEWAY_HTTPADDR", ":9090")
	defer func() { _ = os.Unsetenv("GATEWAY_HTTPADDR") }()

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, ":9090", cfg.HTTPAddr)
}
