package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
)

type ModerationService struct {
	log  *slog.Logger
	repo domain.ModerationRepository
}

func NewModerationService(log *slog.Logger, repo domain.ModerationRepository) *ModerationService {
	return &ModerationService{
		log:  log,
		repo: repo,
	}
}

func (s *ModerationService) SubmitForReview(ctx context.Context, gameID int64, developerID, gameName, gameDescription string) (*domain.GameModeration, error) {
	existing, err := s.repo.GetByGameID(ctx, gameID)
	if err != nil && !errors.Is(err, domain.ErrModerationNotFound) {
		return nil, err
	}
	if existing != nil && existing.Status == domain.ModerationStatusPending {
		return nil, domain.ErrAlreadyUnderReview
	}

	moderation := &domain.GameModeration{
		GameID:          gameID,
		DeveloperID:     developerID,
		GameName:        gameName,
		GameDescription: gameDescription,
		Status:          domain.ModerationStatusPending,
		SubmittedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, moderation); err != nil {
		return nil, err
	}

	s.log.Info("game submitted for moderation",
		slog.Int64("game_id", gameID),
		slog.String("developer_id", developerID))

	return moderation, nil
}

func (s *ModerationService) GetModerationStatus(ctx context.Context, gameID int64) (*domain.GameModeration, error) {
	moderation, err := s.repo.GetByGameID(ctx, gameID)
	if err != nil {
		return nil, err
	}
	return moderation, nil
}

func (s *ModerationService) ApproveGame(ctx context.Context, gameID int64, moderatorID string) (*domain.GameModeration, error) {
	moderation, err := s.repo.GetByGameID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if err := moderation.Approve(moderatorID); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, moderation); err != nil {
		return nil, err
	}

	s.log.Info("game approved",
		slog.Int64("game_id", gameID),
		slog.String("moderator_id", moderatorID))

	return moderation, nil
}

func (s *ModerationService) RejectGame(ctx context.Context, gameID int64, moderatorID, reason string) (*domain.GameModeration, error) {
	moderation, err := s.repo.GetByGameID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if err := moderation.Reject(moderatorID, reason); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, moderation); err != nil {
		return nil, err
	}

	s.log.Info("game rejected",
		slog.Int64("game_id", gameID),
		slog.String("moderator_id", moderatorID),
		slog.String("reason", reason))

	return moderation, nil
}

func (s *ModerationService) ListPendingGames(ctx context.Context, limit, offset int) ([]domain.GameModeration, int, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.GetPending(ctx, limit, offset)
}
