package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleProjectBuildDownload(t *testing.T) {
	tmpDir := t.TempDir()
	archiveDir := filepath.Join(tmpDir, "archives", "42")
	require.NoError(t, os.MkdirAll(archiveDir, 0o755))
	archivePath := filepath.Join(archiveDir, "1.0.0.zip")
	require.NoError(t, os.WriteFile(archivePath, []byte("fake-zip-data"), 0o644))

	handler := handleProjectBuildDownload(tmpDir)

	t.Run("successful download", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/42/builds/1.0.0/download", nil)
		req.SetPathValue("project_id", "42")
		req.SetPathValue("version", "1.0.0")

		rec := httptest.NewRecorder()
		handler(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/zip", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Header().Get("Content-Disposition"), "build_project_42_v1.0.0.zip")
		assert.Equal(t, "fake-zip-data", rec.Body.String())
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/42/builds/9.9.9/download", nil)
		req.SetPathValue("project_id", "42")
		req.SetPathValue("version", "9.9.9")

		rec := httptest.NewRecorder()
		handler(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestHandleProjectMediaServe(t *testing.T) {
	tmpDir := t.TempDir()
	mediaDir := filepath.Join(tmpDir, "media", "42")
	require.NoError(t, os.MkdirAll(mediaDir, 0o755))
	iconPath := filepath.Join(mediaDir, "icon.png")
	require.NoError(t, os.WriteFile(iconPath, []byte("png-image-data"), 0o644))

	handler := handleProjectMediaServe(tmpDir)

	t.Run("serve icon", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/42/media/icon", nil)
		req.SetPathValue("project_id", "42")
		req.SetPathValue("type", "icon")

		rec := httptest.NewRecorder()
		handler(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Cache-Control"), "max-age=3600")
		assert.Equal(t, "png-image-data", rec.Body.String())
	})

	t.Run("path traversal prevented", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/media/../../../etc/passwd", nil)
		req.SetPathValue("path", "../../../etc/passwd")

		rec := httptest.NewRecorder()
		handler(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
