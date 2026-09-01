package service

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func createTestZipBytes(t *testing.T) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	f, err := w.Create("index.html")
	if err != nil {
		t.Fatalf("create zip file: %v", err)
	}
	if _, err := f.Write([]byte("<!DOCTYPE html><html><body>Game Dev</body></html>")); err != nil {
		t.Fatalf("write zip content: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	return buf.Bytes()
}

func TestUnit_DeploymentService_Lifecycle(t *testing.T) {
	tempBase, err := os.MkdirTemp("", "agent-games-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempBase)
	}()

	svc := NewDeploymentService(tempBase, "/games", 100, 1000)

	ctx := context.Background()
	projectID := int64(101)
	version := "v1.0.0"

	// 1. DeployDevStream
	zipBytes := createTestZipBytes(t)
	res, err := svc.DeployDevStream(ctx, projectID, version, bytes.NewReader(zipBytes))
	if err != nil {
		t.Fatalf("expected successful dev deploy, got: %v", err)
	}
	if res.URL != "/games/101/dev/index.html" {
		t.Errorf("unexpected dev URL: %s", res.URL)
	}

	// Проверяем наличие файла dev/index.html
	devIndex := filepath.Join(tempBase, "101", "dev", "index.html")
	if _, err := os.Stat(devIndex); os.IsNotExist(err) {
		t.Fatalf("expected dev index.html symlinked, err: %v", err)
	}

	// 2. DeployProd
	prodRes, err := svc.DeployProd(ctx, projectID, version)
	if err != nil {
		t.Fatalf("expected successful prod deploy, got: %v", err)
	}
	if prodRes.URL != "/games/101/prod/index.html" {
		t.Errorf("unexpected prod URL: %s", prodRes.URL)
	}

	prodIndex := filepath.Join(tempBase, "101", "prod", "index.html")
	if _, err := os.Stat(prodIndex); os.IsNotExist(err) {
		t.Fatalf("expected prod index.html symlinked, err: %v", err)
	}

	// 3. UndeployProd
	if err := svc.UndeployProd(ctx, projectID); err != nil {
		t.Fatalf("expected successful undeploy prod, got: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(tempBase, "101", "prod")); !os.IsNotExist(err) {
		t.Errorf("expected prod symlink to be removed")
	}

	// 4. DeleteVersion
	if err := svc.DeleteVersion(ctx, projectID, version); err != nil {
		t.Fatalf("expected successful delete version, got: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tempBase, "101", "versions", version)); !os.IsNotExist(err) {
		t.Errorf("expected unpacked version dir to be removed")
	}

	// 5. DeleteProject
	if err := svc.DeleteProject(ctx, projectID); err != nil {
		t.Fatalf("expected successful delete project, got: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tempBase, "101")); !os.IsNotExist(err) {
		t.Errorf("expected project dir to be removed")
	}
}
