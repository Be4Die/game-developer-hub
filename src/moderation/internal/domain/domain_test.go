package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDomain_ModerationEntities(t *testing.T) {
	now := time.Now()
	reqID := int64(100)

	req := ModerationRequest{
		ID:          1,
		ProjectID:   42,
		OwnerID:     "user-1",
		ModeratorID: "mod-1",
		Status:      RequestStatusPending,
		Snapshot: ProjectSnapshot{
			ProjectID:          42,
			TitleRu:            "Игра",
			TitleEn:            "Game",
			ActiveBuildVersion: "1.0.0",
		},
		SubmittedAt: now,
	}

	assert.Equal(t, int64(1), req.ID)
	assert.Equal(t, RequestStatusPending, req.Status)
	assert.Equal(t, "1.0.0", req.Snapshot.ActiveBuildVersion)

	msg := ChatMessage{
		ID:          10,
		ProjectID:   42,
		RequestID:   &reqID,
		SenderID:    "mod-1",
		SenderRole:  SenderRoleModerator,
		MessageType: MessageTypeText,
		Content:     "Пожалуйста, добавьте иконку.",
		CreatedAt:   now,
	}

	assert.Equal(t, int64(10), msg.ID)
	assert.Equal(t, SenderRoleModerator, msg.SenderRole)
	assert.Equal(t, MessageTypeText, msg.MessageType)
	assert.Equal(t, "Пожалуйста, добавьте иконку.", msg.Content)
}
