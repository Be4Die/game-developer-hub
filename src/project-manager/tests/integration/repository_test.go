//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_ProjectRepo_CRUD(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	// 1. Create project
	projID, err := env.projectRepo.Create(ctx, &domain.Project{
		OwnerID: "user-repo-1",
		Status:  domain.ProjectStatusDraft,
	})
	require.NoError(t, err)
	assert.Greater(t, projID, int64(0))

	// 2. Get project
	fetched, err := env.projectRepo.Get(ctx, projID)
	require.NoError(t, err)
	assert.Equal(t, projID, fetched.ID)
	assert.Equal(t, "user-repo-1", fetched.OwnerID)
	assert.Equal(t, domain.ProjectStatusDraft, fetched.Status)

	// 3. UpdateStatus
	err = env.projectRepo.UpdateStatus(ctx, projID, domain.ProjectStatusPublished)
	require.NoError(t, err)
	fetchedAfterUpdate, err := env.projectRepo.Get(ctx, projID)
	require.NoError(t, err)
	assert.Equal(t, domain.ProjectStatusPublished, fetchedAfterUpdate.Status)

	// 4. ListByOwner
	list, err := env.projectRepo.ListByOwner(ctx, "user-repo-1", 10, 0)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// 5. Delete
	err = env.projectRepo.Delete(ctx, projID)
	require.NoError(t, err)

	_, err = env.projectRepo.Get(ctx, projID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestIntegration_DraftRepo_CRUD(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	projID, err := env.projectRepo.Create(ctx, &domain.Project{
		OwnerID: "user-draft-1",
		Status:  domain.ProjectStatusDraft,
	})
	require.NoError(t, err)

	draft := &domain.Draft{
		ProjectID:          projID,
		TitleRu:            "Черновик игры",
		TitleEn:            "Draft game",
		AboutRu:            "Описание",
		AboutEn:            "Description",
		ActiveBuildVersion: "1.0.0",
		DevURL:             "/games/1/dev/index.html",
	}

	// 1. Create Draft
	err = env.draftRepo.Create(ctx, draft)
	require.NoError(t, err)

	// 2. Get Draft
	fetched, err := env.draftRepo.Get(ctx, projID)
	require.NoError(t, err)
	assert.Equal(t, "Черновик игры", fetched.TitleRu)
	assert.Equal(t, "1.0.0", fetched.ActiveBuildVersion)

	// 3. Update Draft
	draft.TitleRu = "Обновленный заголовок"
	err = env.draftRepo.Update(ctx, draft)
	require.NoError(t, err)

	// 4. UpdateMedia
	err = env.draftRepo.UpdateMedia(ctx, projID, "icon", "/path/to/icon.png")
	require.NoError(t, err)

	fetchedMedia, err := env.draftRepo.Get(ctx, projID)
	require.NoError(t, err)
	assert.Equal(t, "/path/to/icon.png", fetchedMedia.IconPath)
	assert.Equal(t, "Обновленный заголовок", fetchedMedia.TitleRu)
}

func TestIntegration_BuildRepo_CRUD(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	projID, err := env.projectRepo.Create(ctx, &domain.Project{
		OwnerID: "user-build-1",
		Status:  domain.ProjectStatusDraft,
	})
	require.NoError(t, err)

	build := &domain.Build{
		ProjectID:    projID,
		Version:      "1.0.0",
		FilePath:     "/archives/1/1.0.0.zip",
		FileSize:     1024,
		IsUnpacked:   false,
		UnpackedPath: "",
	}

	// 1. Create Build
	err = env.buildRepo.Create(ctx, build)
	require.NoError(t, err)
	assert.Greater(t, build.ID, int64(0))

	// 2. Get
	fetched, err := env.buildRepo.Get(ctx, projID, "1.0.0")
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", fetched.Version)
	assert.Equal(t, int64(1024), fetched.FileSize)
	assert.False(t, fetched.IsUnpacked)

	// 3. MarkUnpacked
	err = env.buildRepo.MarkUnpacked(ctx, projID, "1.0.0", "/unpacked/1/1.0.0")
	require.NoError(t, err)
	fetchedAfterUnpack, err := env.buildRepo.Get(ctx, projID, "1.0.0")
	require.NoError(t, err)
	assert.True(t, fetchedAfterUnpack.IsUnpacked)
	assert.Equal(t, "/unpacked/1/1.0.0", fetchedAfterUnpack.UnpackedPath)

	// 4. ListByProject
	build2 := &domain.Build{
		ProjectID:  projID,
		Version:    "1.0.1",
		FilePath:   "/archives/1/1.0.1.zip",
		FileSize:   2048,
		IsUnpacked: true,
	}
	err = env.buildRepo.Create(ctx, build2)
	require.NoError(t, err)

	list, err := env.buildRepo.ListByProject(ctx, projID, 10)
	require.NoError(t, err)
	assert.Len(t, list, 2)

	// 5. Delete Build
	err = env.buildRepo.Delete(ctx, build.ID)
	require.NoError(t, err)

	_, err = env.buildRepo.Get(ctx, projID, "1.0.0")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestIntegration_ReleaseRepo_CRUD(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	projID, err := env.projectRepo.Create(ctx, &domain.Project{
		OwnerID: "user-rel-1",
		Status:  domain.ProjectStatusDraft,
	})
	require.NoError(t, err)

	rel := &domain.Release{
		ProjectID:   projID,
		Version:     "1.0.0",
		TitleRu:     "Релиз 1",
		TitleEn:     "Release 1",
		ProdURL:     "https://play.game.local/1",
		PublishedBy: "mod-1",
	}

	// 1. Create Release
	relID, err := env.releaseRepo.Create(ctx, rel)
	require.NoError(t, err)
	assert.Greater(t, relID, int64(0))

	// 2. GetActive
	active, err := env.releaseRepo.GetActive(ctx, projID)
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", active.Version)
	assert.True(t, active.IsActive)

	// 3. Deactivate
	err = env.releaseRepo.Deactivate(ctx, projID)
	require.NoError(t, err)

	_, err = env.releaseRepo.GetActive(ctx, projID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestIntegration_DeploymentRepo_CRUD(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	projID, err := env.projectRepo.Create(ctx, &domain.Project{
		OwnerID: "user-dep-1",
		Status:  domain.ProjectStatusDraft,
	})
	require.NoError(t, err)

	record := &domain.DeploymentRecord{
		ProjectID:   projID,
		Environment: domain.DeploymentEnvDev,
		Version:     "1.0.0",
		Status:      domain.DeploymentStatusSuccess,
	}

	// 1. Create
	err = env.deploymentRepo.Create(ctx, record)
	require.NoError(t, err)
	assert.Greater(t, record.ID, int64(0))

	// 2. ListByProject
	records, err := env.deploymentRepo.ListByProject(ctx, projID, 10)
	require.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "1.0.0", records[0].Version)
	assert.Equal(t, domain.DeploymentStatusSuccess, records[0].Status)
}
