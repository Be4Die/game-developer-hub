//go:build integration

package integration

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_ProjectService_FullFlow(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	ownerID := "user-service-test-1"

	// 1. Create Project
	proj, err := env.projectSvc.CreateProject(ctx, ownerID, "Игра Тест", "Game Test", false)
	require.NoError(t, err)
	assert.Greater(t, proj.ID, int64(0))

	// 2. Update Draft
	err = env.projectSvc.UpdateDraft(ctx, proj.ID, ownerID, domain.DraftMeta{
		TitleRu: "Обновленная игра",
		TitleEn: "Updated game",
		AboutRu: "Классное описание",
		AboutEn: "Cool description",
	})
	require.NoError(t, err)

	// 3. Upload Build
	zipContent := createTestGameZip(t)
	build, devURL, err := env.projectSvc.UploadBuildStream(ctx, proj.ID, ownerID, "1.0.0", bytes.NewReader(zipContent))
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", build.Version)
	assert.Contains(t, devURL, "index.html")

	// 4. Submit for Moderation
	reqID, err := env.projectSvc.SubmitForModeration(ctx, proj.ID, ownerID)
	require.NoError(t, err)
	assert.Greater(t, reqID, int64(0))

	// 5. Publish Release
	rel, err := env.projectSvc.PublishRelease(ctx, proj.ID, "1.0.0", "mod-super")
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", rel.Version)

	// 6. Get Published
	activeRel, err := env.projectSvc.GetPublished(ctx, proj.ID)
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", activeRel.Version)

	// 7. Unpublish
	err = env.projectSvc.Unpublish(ctx, proj.ID, ownerID)
	require.NoError(t, err)

	_, err = env.projectSvc.GetPublished(ctx, proj.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestIntegration_ValkeyLocker_Concurrency(t *testing.T) {
	env := setupIntegration(t)
	ctx := context.Background()

	key := "test-lock-key"

	// 1. First acquire succeeds
	unlock1, err := env.locker.Acquire(ctx, key, 5*time.Second)
	require.NoError(t, err)
	require.NotNil(t, unlock1)

	// 2. Concurrent acquire fails with ErrLockBusy
	unlock2, err := env.locker.Acquire(ctx, key, 5*time.Second)
	assert.ErrorIs(t, err, domain.ErrLockBusy)
	assert.Nil(t, unlock2)

	// 3. Release first lock
	unlock1()

	// 4. Subsequent acquire succeeds
	unlock3, err := env.locker.Acquire(ctx, key, 5*time.Second)
	require.NoError(t, err)
	require.NotNil(t, unlock3)
	unlock3()
}
