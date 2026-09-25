package service

import (
	"context"
	"testing"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestProjectService_GetProjectAnalytics(t *testing.T) {
	ctx := context.Background()

	projectRepo := newMockProjectRepo()
	draftRepo := newMockDraftRepo()
	buildRepo := newMockBuildRepo()
	releaseRepo := newMockReleaseRepo()
	deploymentRepo := newMockDeploymentRepo()
	memberRepo := newMockMemberRepo()
	invitationRepo := newMockInvitationRepo()
	blockRepo := newMockBlockRepo()

	svc := NewProjectService(
		projectRepo,
		draftRepo,
		buildRepo,
		releaseRepo,
		deploymentRepo,
		memberRepo,
		invitationRepo,
		blockRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		5,
	)

	mockAnalytics := &mockAnalyticsClient{
		res: &domain.GameAnalytics{
			ProjectID: 1,
			Summary: domain.AnalyticsSummary{
				UniquePlayers: 120,
				TotalRevenue:  25000.0,
			},
		},
	}
	svc.WithAnalyticsClient(mockAnalytics)

	// 1. Создаем проект для пользователя owner-1
	p, err := svc.CreateProject(ctx, "owner-1", "Тестовая игра", "Test Game", false)
	require.NoError(t, err)

	// 2. Владелец может просматривать аналитику
	filter := domain.AnalyticsFilter{DateFrom: "2026-06-01", DateTo: "2026-06-07"}
	res, err := svc.GetProjectAnalytics(ctx, p.ID, "owner-1", filter)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(120), res.Summary.UniquePlayers)
	require.Equal(t, p.ID, mockAnalytics.calledWithGameID)
	require.Equal(t, "2026-06-01", mockAnalytics.calledWithFilter.DateFrom)

	// 3. Посторонний пользователь не имеет доступа
	_, err = svc.GetProjectAnalytics(ctx, p.ID, "stranger-2", filter)
	require.ErrorIs(t, err, domain.ErrForbidden)

	// 4. Участник без PERM_VIEW_STATS не имеет доступа
	err = memberRepo.Add(ctx, &domain.Member{
		ProjectID:   p.ID,
		UserID:      "member-3",
		Permissions: []string{domain.PermEditInfo},
	})
	require.NoError(t, err)
	_, err = svc.GetProjectAnalytics(ctx, p.ID, "member-3", filter)
	require.ErrorIs(t, err, domain.ErrForbidden)

	// 5. Участник с PERM_VIEW_STATS имеет доступ
	err = memberRepo.Add(ctx, &domain.Member{
		ProjectID:   p.ID,
		UserID:      "analyst-4",
		Permissions: []string{domain.PermViewStats},
	})
	require.NoError(t, err)
	resAnalyst, err := svc.GetProjectAnalytics(ctx, p.ID, "analyst-4", filter)
	require.NoError(t, err)
	require.NotNil(t, resAnalyst)
}
