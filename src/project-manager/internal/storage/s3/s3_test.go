package s3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveMimeAndEncoding(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		expectedType     string
		expectedEncoding string
	}{
		{
			name:             "unity wasm br",
			path:             "Build/game.wasm.br",
			expectedType:     "application/wasm",
			expectedEncoding: "br",
		},
		{
			name:             "unity data br",
			path:             "Build/game.data.br",
			expectedType:     "application/octet-stream",
			expectedEncoding: "br",
		},
		{
			name:             "unity framework js br",
			path:             "Build/game.framework.js.br",
			expectedType:     "application/javascript",
			expectedEncoding: "br",
		},
		{
			name:             "unity wasm gz",
			path:             "Build/game.wasm.gz",
			expectedType:     "application/wasm",
			expectedEncoding: "gzip",
		},
		{
			name:             "standard index html",
			path:             "index.html",
			expectedType:     "text/html; charset=utf-8",
			expectedEncoding: "",
		},
		{
			name:             "javascript file",
			path:             "TemplateData/script.js",
			expectedType:     "application/javascript",
			expectedEncoding: "",
		},
		{
			name:             "css file",
			path:             "TemplateData/style.css",
			expectedType:     "text/css",
			expectedEncoding: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct, ce := resolveMimeAndEncoding(tt.path)
			assert.Equal(t, tt.expectedType, ct)
			assert.Equal(t, tt.expectedEncoding, ce)
		})
	}
}

func TestBuildStorage_ObjectKey(t *testing.T) {
	bs := NewBuildStorage(nil, "builds")
	key := bs.objectKey(42, "1.0.0")
	assert.Equal(t, "projects/42/1.0.0.zip", key)
	assert.Equal(t, "projects/42/1.0.0.zip", bs.GetArchivePath(42, "1.0.0"))
}
