package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDomain_DeploymentResult(t *testing.T) {
	res := DeploymentResult{
		URL:          "https://play.game.local/42",
		UnpackedPath: "/data/webgames/42/1.0.0",
		Success:      true,
	}

	assert.Equal(t, "https://play.game.local/42", res.URL)
	assert.Equal(t, "/data/webgames/42/1.0.0", res.UnpackedPath)
	assert.True(t, res.Success)

	assert.NotNil(t, ErrInvalidArchive)
	assert.NotNil(t, ErrZipSlip)
	assert.NotNil(t, ErrZipBomb)
	assert.NotNil(t, ErrNoIndexHTML)
	assert.NotNil(t, ErrDisallowedFileType)
	assert.NotNil(t, ErrUnauthorized)
}
