package filesystem

import (
	"bytes"
	"context"
	"testing"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Valid PNG header for mime detection
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

func TestBuildStorage_Operations(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewBuildStorage(tmpDir)
	ctx := context.Background()

	projectID := int64(42)
	version := "v1.0.0"
	dummyData := []byte("zip-binary-archive-content-12345")

	// 1. SaveArchive
	path, size, err := storage.SaveArchive(ctx, projectID, version, dummyData)
	require.NoError(t, err)
	assert.Equal(t, int64(len(dummyData)), size)
	assert.FileExists(t, path)

	// 2. GetArchivePath
	expectedPath := storage.GetArchivePath(projectID, version)
	assert.Equal(t, expectedPath, path)

	// 3. DeleteBuild
	err = storage.DeleteBuild(projectID, version)
	require.NoError(t, err)
	assert.NoFileExists(t, path)

	// 4. DeleteBuild non-existing (should not error)
	err = storage.DeleteBuild(projectID, "v9.9.9")
	require.NoError(t, err)

	// 5. DeleteProject
	_, _, err = storage.SaveArchive(ctx, projectID, "v2.0.0", dummyData)
	require.NoError(t, err)
	err = storage.DeleteProject(projectID)
	require.NoError(t, err)
	assert.NoDirExists(t, storage.projectDir(projectID))
}

func TestMediaStorage_Operations(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewMediaStorage(tmpDir)
	ctx := context.Background()

	projectID := int64(77)

	t.Run("valid png upload for icon and cover", func(t *testing.T) {
		iconPath, err := storage.SaveMedia(ctx, projectID, "icon", samplePng)
		require.NoError(t, err)
		assert.FileExists(t, iconPath)

		coverPath, err := storage.SaveMedia(ctx, projectID, "cover", samplePng)
		require.NoError(t, err)
		assert.FileExists(t, coverPath)
	})

	t.Run("invalid mime type for image", func(t *testing.T) {
		textData := []byte("this is just plain text, definitely not an image")
		_, err := storage.SaveMedia(ctx, projectID, "icon", textData)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("invalid media type", func(t *testing.T) {
		_, err := storage.SaveMedia(ctx, projectID, "unknown-type", samplePng)
		require.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("empty stream", func(t *testing.T) {
		_, err := storage.SaveMediaStream(ctx, projectID, "icon", bytes.NewReader([]byte{}))
		require.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("delete media", func(t *testing.T) {
		iconPath, err := storage.SaveMedia(ctx, projectID, "icon", samplePng)
		require.NoError(t, err)
		assert.FileExists(t, iconPath)

		err = storage.DeleteMedia(projectID, "icon")
		require.NoError(t, err)
		assert.NoFileExists(t, iconPath)

		// Delete non-existing media should not fail
		err = storage.DeleteMedia(projectID, "icon")
		require.NoError(t, err)
	})

	t.Run("snapshot media for release", func(t *testing.T) {
		iconPath, err := storage.SaveMedia(ctx, projectID, "icon", samplePng)
		require.NoError(t, err)

		snapshotPath, err := storage.SnapshotMediaForRelease(ctx, projectID, "v1.0.0", iconPath, "icon")
		require.NoError(t, err)
		assert.FileExists(t, snapshotPath)

		// Empty srcPath returns empty string
		emptySnapshot, err := storage.SnapshotMediaForRelease(ctx, projectID, "v1.0.0", "", "icon")
		require.NoError(t, err)
		assert.Empty(t, emptySnapshot)

		// Invalid media type returns error
		_, err = storage.SnapshotMediaForRelease(ctx, projectID, "v1.0.0", iconPath, "bad-type")
		require.ErrorIs(t, err, domain.ErrInvalidInput)
	})
}
