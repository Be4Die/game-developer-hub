package domain

import (
	"errors"
)

var (
	ErrChatNotFound        = errors.New("chat not found")
	ErrMessageNotFound     = errors.New("message not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrAlreadyParticipant  = errors.New("user is already a participant")
	ErrNotParticipant      = errors.New("user is not a participant")
	ErrInvalidChatType     = errors.New("invalid chat type")
)