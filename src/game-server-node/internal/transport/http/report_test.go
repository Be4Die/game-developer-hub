package http

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/storage/memory"
)

const testNodeSecret = "node-secret-for-report-tests-000000"

func newReportFixture(t *testing.T) (*ReportHandler, *memory.Storage) {
	t.Helper()
	storage := memory.NewStorage()
	for _, id := range []int64{1, 2} {
		require.NoError(t, storage.RecordInstance(context.Background(), domain.Instance{ID: id, GameID: 10}))
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewReportHandler(log, storage, testNodeSecret), storage
}

func postReport(h http.Handler, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/report", strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestReportHandler_Auth(t *testing.T) {
	body := `{"instance_id":1,"player_count":7,"queue_size":2,"max_players":16}`

	tests := []struct {
		name     string
		token    string
		wantCode int
	}{
		{name: "valid token", token: domain.ReportToken(testNodeSecret, 1), wantCode: http.StatusOK},
		{name: "missing token", token: "", wantCode: http.StatusUnauthorized},
		{name: "token of another instance", token: domain.ReportToken(testNodeSecret, 2), wantCode: http.StatusUnauthorized},
		{name: "token signed by another node", token: domain.ReportToken("other-node-secret", 1), wantCode: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, storage := newReportFixture(t)

			rec := postReport(h, tt.token, body)
			assert.Equal(t, tt.wantCode, rec.Code)

			inst, err := storage.GetInstanceByID(context.Background(), 1)
			require.NoError(t, err)
			if tt.wantCode == http.StatusOK {
				require.NotNil(t, inst.PlayerCount)
				assert.Equal(t, uint32(7), *inst.PlayerCount)
				assert.Equal(t, uint32(16), inst.MaxPlayers)
			} else {
				assert.Nil(t, inst.PlayerCount, "rejected report must not change the instance")
			}
		})
	}
}

func TestReportHandler_ZeroPlayersIsRecorded(t *testing.T) {
	h, storage := newReportFixture(t)
	token := domain.ReportToken(testNodeSecret, 1)

	require.Equal(t, http.StatusOK, postReport(h, token, `{"instance_id":1,"player_count":5}`).Code)
	require.Equal(t, http.StatusOK, postReport(h, token, `{"instance_id":1,"player_count":0}`).Code)

	inst, err := storage.GetInstanceByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, inst.PlayerCount)
	assert.Equal(t, uint32(0), *inst.PlayerCount)
}

func TestReportHandler_RejectsOversizedBody(t *testing.T) {
	h, _ := newReportFixture(t)
	body := `{"instance_id":1,"pad":"` + strings.Repeat("x", maxReportBodyBytes) + `"}`

	rec := postReport(h, domain.ReportToken(testNodeSecret, 1), body)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestValidReportToken_EmptySecret(t *testing.T) {
	assert.False(t, domain.ValidReportToken("", 1, domain.ReportToken("", 1)),
		"node without a secret must not accept reports")
}
