package s3

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var samplePng = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
	0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
	0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
	0x42, 0x60, 0x82,
}

func TestMediaStorage_GetFileName(t *testing.T) {
	storage := NewMediaStorage(nil, "media")

	tests := []struct {
		mediaType    string
		expectedName string
		expectErr    bool
	}{
		{"icon", "icon.png", false},
		{"cover", "cover.png", false},
		{"video", "video.mp4", false},
		{"item:sword_01", "items/sword_01.png", false},
		{"invalid", "", true},
		{"item:", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.mediaType, func(t *testing.T) {
			name, err := storage.getFileName(tt.mediaType)
			if tt.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, domain.ErrInvalidInput)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedName, name)
			}
		})
	}
}

func TestMediaStorage_ValidateMediaMime(t *testing.T) {
	t.Run("valid png image", func(t *testing.T) {
		contentType, buf, err := validateMediaMime(bytes.NewReader(samplePng), "icon")
		require.NoError(t, err)
		assert.Equal(t, "image/png", contentType)
		assert.NotEmpty(t, buf)
	})

	t.Run("invalid text for image", func(t *testing.T) {
		_, _, err := validateMediaMime(bytes.NewReader([]byte("plain text")), "icon")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("empty reader", func(t *testing.T) {
		_, _, err := validateMediaMime(bytes.NewReader([]byte{}), "video")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})
}

func TestMediaStorage_KeyFormat(t *testing.T) {
	storage := NewMediaStorage(nil, "media")
	projectID := int64(100)

	fileName, err := storage.getFileName("video")
	require.NoError(t, err)
	key := fmt.Sprintf("%d/%s", projectID, fileName)

	assert.Equal(t, "100/video.mp4", key)
}
