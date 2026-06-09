package domain

import (
	"errors"
	"time"
)

var (
	ErrModerationNotFound = errors.New("moderation not found")
	ErrAlreadyUnderReview = errors.New("game already under review")
	ErrAlreadyModerated   = errors.New("game already moderated")
)

type ModerationStatus string

const (
	ModerationStatusPending  ModerationStatus = "pending"
	ModerationStatusApproved ModerationStatus = "approved"
	ModerationStatusRejected ModerationStatus = "rejected"
)

type GameModeration struct {
	ID              int64
	GameID          int64
	DeveloperID     string
	GameName        string
	GameDescription string
	ModeratorID     string
	Status          ModerationStatus
	RejectionReason string
	SubmittedAt     time.Time
	ReviewedAt      *time.Time
}

func (g *GameModeration) Approve(moderatorID string) error {
	if g.Status != ModerationStatusPending {
		return ErrAlreadyModerated
	}
	g.Status = ModerationStatusApproved
	g.ModeratorID = moderatorID
	now := time.Now()
	g.ReviewedAt = &now
	return nil
}

func (g *GameModeration) Reject(moderatorID, reason string) error {
	if g.Status != ModerationStatusPending {
		return ErrAlreadyModerated
	}
	g.Status = ModerationStatusRejected
	g.ModeratorID = moderatorID
	g.RejectionReason = reason
	now := time.Now()
	g.ReviewedAt = &now
	return nil
}
