//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Moderation_RequestRepo(t *testing.T) {
	env := setupModIntegration(t)
	ctx := context.Background()

	req := &domain.ModerationRequest{
		ProjectID: 101,
		OwnerID:   "dev-101",
		Status:    domain.RequestStatusPending,
		Snapshot: domain.ProjectSnapshot{
			ProjectID:          101,
			TitleRu:            "Тестовая игра",
			TitleEn:            "Test Game",
			AboutRu:            "Описание",
			AboutEn:            "Description",
			ActiveBuildVersion: "1.0.0",
			DevURL:             "/games/101/dev/index.html",
		},
	}

	// 1. Create
	id, err := env.requestRepo.Create(ctx, req)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// 2. Get
	fetched, err := env.requestRepo.Get(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, id, fetched.ID)
	assert.Equal(t, int64(101), fetched.ProjectID)
	assert.Equal(t, "dev-101", fetched.OwnerID)
	assert.Equal(t, "Тестовая игра", fetched.Snapshot.TitleRu)
	assert.Equal(t, domain.RequestStatusPending, fetched.Status)

	// 3. GetLatestByProject
	latest, err := env.requestRepo.GetLatestByProject(ctx, 101)
	require.NoError(t, err)
	assert.Equal(t, id, latest.ID)

	// 4. List (Filter by Status)
	statusPending := domain.RequestStatusPending
	list, total, err := env.requestRepo.List(ctx, domain.RequestFilter{
		Status: &statusPending,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, list, 1)

	// 5. Claim
	err = env.requestRepo.Claim(ctx, id, "mod-202")
	require.NoError(t, err)

	claimed, err := env.requestRepo.Get(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "mod-202", claimed.ModeratorID)
	assert.Equal(t, domain.RequestStatusInReview, claimed.Status)
	assert.NotNil(t, claimed.StartedReviewAt)

	// Duplicate claim by another moderator -> ErrAlreadyClaimed
	err = env.requestRepo.Claim(ctx, id, "mod-303")
	assert.ErrorIs(t, err, domain.ErrAlreadyClaimed)

	// 6. Resolve (Approve)
	err = env.requestRepo.Resolve(ctx, id, domain.RequestStatusApproved, "", "mod-202")
	require.NoError(t, err)

	resolved, err := env.requestRepo.Get(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, domain.RequestStatusApproved, resolved.Status)
	assert.NotNil(t, resolved.ResolvedAt)
}

func TestIntegration_Moderation_MessageRepo(t *testing.T) {
	env := setupModIntegration(t)
	ctx := context.Background()

	msg1 := &domain.ChatMessage{
		ProjectID:   101,
		SenderID:    "dev-101",
		SenderRole:  domain.SenderRoleDeveloper,
		MessageType: domain.MessageTypeText,
		Content:     "Hello, reviewer!",
	}

	msg2 := &domain.ChatMessage{
		ProjectID:   101,
		SenderID:    "mod-202",
		SenderRole:  domain.SenderRoleModerator,
		MessageType: domain.MessageTypeText,
		Content:     "Everything looks great!",
	}

	// 1. Create Messages
	id1, err := env.messageRepo.Create(ctx, msg1)
	require.NoError(t, err)
	assert.Greater(t, id1, int64(0))

	id2, err := env.messageRepo.Create(ctx, msg2)
	require.NoError(t, err)
	assert.Greater(t, id2, int64(0))

	// 2. ListByProject
	messages, total, err := env.messageRepo.ListByProject(ctx, 101, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, messages, 2)
	assert.Equal(t, "Hello, reviewer!", messages[0].Content)
	assert.Equal(t, "Everything looks great!", messages[1].Content)

	// 3. ListActiveChats
	chats, chatTotal, err := env.messageRepo.ListActiveChats(ctx, 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, chatTotal, 1)
	assert.NotEmpty(t, chats)
}
