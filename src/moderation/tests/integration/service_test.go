//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_ModerationService_RealPostgres(t *testing.T) {
	env := setupModIntegration(t)
	ctx := context.Background()

	projectID := int64(55)
	ownerID := "dev-user-55"
	modID := "mod-user-99"

	// 1. Submit Draft
	req, err := env.svc.SubmitDraft(ctx, projectID, ownerID, domain.ProjectSnapshot{
		ProjectID:          projectID,
		TitleRu:            "Супер Игра",
		TitleEn:            "Super Game",
		AboutRu:            "Обалденная игра",
		AboutEn:            "Awesome game",
		ActiveBuildVersion: "1.0.0",
		DevURL:             "/games/55/dev/index.html",
	})
	require.NoError(t, err)
	assert.Greater(t, req.ID, int64(0))
	assert.Equal(t, domain.RequestStatusPending, req.Status)

	// 2. Moderator Claims Request
	claimed, err := env.svc.ClaimRequest(ctx, req.ID, modID)
	require.NoError(t, err)
	assert.Equal(t, modID, claimed.ModeratorID)
	assert.Equal(t, domain.RequestStatusInReview, claimed.Status)

	// 3. Send Messages
	msg1, err := env.svc.SendMessage(ctx, projectID, ownerID, domain.SenderRoleDeveloper, "Когда релиз?")
	require.NoError(t, err)
	assert.Greater(t, msg1.ID, int64(0))

	msg2, err := env.svc.SendMessage(ctx, projectID, modID, domain.SenderRoleModerator, "Сейчас тестирую.")
	require.NoError(t, err)
	assert.Greater(t, msg2.ID, int64(0))

	// 4. Approve
	approved, prodURL, err := env.svc.Approve(ctx, projectID, modID, "Проверка успешно завершена!")
	require.NoError(t, err)
	assert.Equal(t, domain.RequestStatusApproved, approved.Status)
	assert.Equal(t, "/games/55/prod/index.html", prodURL)

	// 5. Verify Latest Request
	latest, err := env.svc.GetLatestRequestByProject(ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, domain.RequestStatusApproved, latest.Status)
}
