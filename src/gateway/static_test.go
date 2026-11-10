package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleProjectBuildDownload(t *testing.T) {
	tmpDir := t.TempDir()
	archiveDir := filepath.Join(tmpDir, "archives", "42")
	require.NoError(t, os.MkdirAll(archiveDir, 0o755))
	archivePath := filepath.Join(archiveDir, "1.0.0.zip")
	require.NoError(t, os.WriteFile(archivePath, []byte("fake-zip-data"), 0o644))

	const secret = "0123456789abcdef0123456789abcdef"
	verifier := &tokenVerifier{secret: []byte(secret), issuer: "sso"}
	handler := handleProjectBuildDownload(verifier, nil, tmpDir)

	sign := func(role int) string {
		tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user-1", "role": role, "iss": "sso", "exp": time.Now().Add(time.Hour).Unix(),
		}).SignedString([]byte(secret))
		require.NoError(t, err)
		return tok
	}

	request := func(version, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/42/builds/"+version+"/download", nil)
		req.SetPathValue("project_id", "42")
		req.SetPathValue("version", version)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		handler(rec, req)
		return rec
	}

	t.Run("successful download by staff", func(t *testing.T) {
		rec := request("1.0.0", sign(2))

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/zip", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Header().Get("Content-Disposition"), "build_project_42_v1.0.0.zip")
		assert.Equal(t, "fake-zip-data", rec.Body.String())
	})

	t.Run("anonymous download is rejected", func(t *testing.T) {
		rec := request("1.0.0", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.NotContains(t, rec.Body.String(), "fake-zip-data")
	})

	t.Run("forged token is rejected", func(t *testing.T) {
		rec := request("1.0.0", "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ4Iiwicm9sZSI6M30.forged")
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("developer without project access is rejected", func(t *testing.T) {
		// Клиент project-manager не задан: доступ подтвердить нечем, значит он не выдаётся.
		rec := request("1.0.0", sign(1))
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		rec := request("9.9.9", sign(2))
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("unsafe version is rejected", func(t *testing.T) {
		rec := request("..%2F..%2Fetc", sign(3))
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
