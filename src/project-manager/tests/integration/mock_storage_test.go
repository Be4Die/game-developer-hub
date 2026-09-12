package integration

import (
	"context"
	"fmt"
	"io"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
)

type mockBuildStorage struct{}

func newMockBuildStorage() *mockBuildStorage {
	return &mockBuildStorage{}
}

func (s *mockBuildStorage) SaveArchiveStream(ctx context.Context, projectID int64, version string, src io.Reader) (string, int64, error) {
	n, err := io.Copy(io.Discard, src)
	if err != nil {
		return "", 0, err
	}
	return fmt.Sprintf("projects/%d/%s.zip", projectID, version), n, nil
}

func (s *mockBuildStorage) SaveArchive(ctx context.Context, projectID int64, version string, data []byte) (string, int64, error) {
	return fmt.Sprintf("projects/%d/%s.zip", projectID, version), int64(len(data)), nil
}

func (s *mockBuildStorage) GetArchivePath(projectID int64, version string) string {
	return fmt.Sprintf("projects/%d/%s.zip", projectID, version)
}

func (s *mockBuildStorage) DeleteBuild(projectID int64, version string) error { return nil }
func (s *mockBuildStorage) DeleteProject(projectID int64) error               { return nil }

type mockMediaStorage struct{}

func newMockMediaStorage() *mockMediaStorage {
	return &mockMediaStorage{}
}

func (s *mockMediaStorage) SaveMediaStream(ctx context.Context, projectID int64, mediaType string, src io.Reader) (string, error) {
	if _, err := io.Copy(io.Discard, src); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d/%s.png", projectID, mediaType), nil
}

func (s *mockMediaStorage) SaveMedia(ctx context.Context, projectID int64, mediaType string, data []byte) (string, error) {
	return fmt.Sprintf("%d/%s.png", projectID, mediaType), nil
}

func (s *mockMediaStorage) DeleteMedia(projectID int64, mediaType string) error { return nil }
func (s *mockMediaStorage) SnapshotMediaForRelease(ctx context.Context, projectID int64, version string, srcPath string, mediaType string) (string, error) {
	return fmt.Sprintf("%d/releases/%s/%s.png", projectID, version, mediaType), nil
}

type mockDeployer struct{}

func newMockDeployer() *mockDeployer {
	return &mockDeployer{}
}

func (d *mockDeployer) DeployDev(ctx context.Context, projectID int64, version string, archivePath string) (*domain.DeploymentResult, error) {
	return &domain.DeploymentResult{
		URL:          fmt.Sprintf("/games/%d/dev/index.html", projectID),
		UnpackedPath: fmt.Sprintf("s3://games/%d/versions/%s", projectID, version),
		Success:      true,
	}, nil
}

func (d *mockDeployer) DeployProd(ctx context.Context, projectID int64, version string, archivePath string) (*domain.DeploymentResult, error) {
	return &domain.DeploymentResult{
		URL:          fmt.Sprintf("/games/%d/prod/index.html", projectID),
		UnpackedPath: fmt.Sprintf("s3://games/%d/prod", projectID),
		Success:      true,
	}, nil
}

func (d *mockDeployer) UndeployProd(ctx context.Context, projectID int64) error { return nil }
func (d *mockDeployer) DeleteVersion(ctx context.Context, projectID int64, version string) error {
	return nil
}
func (d *mockDeployer) DeleteProject(ctx context.Context, projectID int64) error { return nil }
