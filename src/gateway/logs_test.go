package main

import (
	"testing"
	"time"

	gwpb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestFormatLogEvent(t *testing.T) {
	assert.Equal(t, "{}", formatLogEvent(nil))

	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	entry := &gwpb.LogEntry{
		Timestamp: timestamppb.New(now),
		Source:    gwpb.LogSource_LOG_SOURCE_STDOUT,
		Message:   "Server started on port 8080",
	}

	formatted := formatLogEvent(entry)
	assert.Contains(t, formatted, `"message":"Server started on port 8080"`)
	assert.Contains(t, formatted, `"source":"LOG_SOURCE_STDOUT"`)
}
