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

func TestUnit_DeploymentService_UpdateCSP(t *testing.T) {
	tempBase, err := os.MkdirTemp("", "agent-games-csp-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempBase)
	}()

	svc := NewDeploymentService(tempBase, "/games", 100, 1000)
	ctx := context.Background()
	projectID := int64(202)

	// Test online CSP with custom hosts
	hosts := []string{"wss://proxy.welwise.online:*", "https://proxy.welwise.online:*"}
	err = svc.UpdateCSP(ctx, projectID, "dev", true, hosts)
	if err != nil {
		t.Fatalf("unexpected error updating CSP: %v", err)
	}

	manifestPath := filepath.Join(tempBase, "202", "csp.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read generated csp.json: %v", err)
	}

	if !bytes.Contains(data, []byte("proxy.welwise.online")) {
		t.Errorf("expected csp.json to contain proxy host, got %s", string(data))
	}
	if !bytes.Contains(data, []byte("'self'")) {
		t.Errorf("expected csp.json to contain 'self', got %s", string(data))
	}
}
