package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/go-chi/chi/v5"
)

// ─── errorResponse tests ────────────────────────────────────────────────────

func TestErrorResponse_WritesCorrectJSONAndStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		code       string
		message    string
	}{
		{"bad request", http.StatusBadRequest, "BAD_REQUEST", "invalid input"},
		{"not found", http.StatusNotFound, "NOT_FOUND", "resource missing"},
		{"conflict", http.StatusConflict, "CONFLICT", "duplicate"},
		{"internal error", http.StatusInternalServerError, "INTERNAL_ERROR", "server error"},
		{"unauthorized", http.StatusUnauthorized, "UNAUTHORIZED", "bad token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			errorResponse(rec, tt.statusCode, tt.code, tt.message)

			if rec.Code != tt.statusCode {
				t.Errorf("status = %d, want %d", rec.Code, tt.statusCode)
			}

			ct := rec.Header().Get("Content-Type")
			if ct != "application/json" {
				t.Errorf("Content-Type = %q, want %q", ct, "application/json")
			}

			var body map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("failed to decode JSON: %v", err)
			}
			if body["code"] != tt.code {
				t.Errorf("body.code = %q, want %q", body["code"], tt.code)
			}
			if body["message"] != tt.message {
				t.Errorf("body.message = %q, want %q", body["message"], tt.message)
			}
		})
	}
}

// ─── badRequest / notFound / conflict / internalError tests ─────────────────

func TestBadRequest(t *testing.T) {
	rec := httptest.NewRecorder()
	badRequest(rec, "bad input")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["code"] != "BAD_REQUEST" {
		t.Errorf("code = %q, want %q", body["code"], "BAD_REQUEST")
	}
}

func TestNotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	notFound(rec, "missing")
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["code"] != "NOT_FOUND" {
		t.Errorf("code = %q, want %q", body["code"], "NOT_FOUND")
	}
}

func TestConflict(t *testing.T) {
	rec := httptest.NewRecorder()
	conflict(rec, "duplicate")
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["code"] != "CONFLICT" {
		t.Errorf("code = %q, want %q", body["code"], "CONFLICT")
	}
}

func TestInternalError(t *testing.T) {
	rec := httptest.NewRecorder()
	internalError(rec, "crash")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["code"] != "INTERNAL_ERROR" {
		t.Errorf("code = %q, want %q", body["code"], "INTERNAL_ERROR")
	}
}

// ─── jsonOK tests ───────────────────────────────────────────────────────────

func TestJSONOK_WritesCorrectJSONAndStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"key": "value"}
	jsonOK(rec, data, http.StatusOK)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want %q", rec.Header().Get("Content-Type"), "application/json")
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if body["key"] != "value" {
		t.Errorf("body[key] = %q, want %q", body["key"], "value")
	}
}

func TestJSONOK204(t *testing.T) {
	rec := httptest.NewRecorder()
	jsonOK204(rec)
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body, got %q", rec.Body.String())
	}
}

// ─── parseGameID tests ──────────────────────────────────────────────────────

func TestParseGameID_Valid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/games/42/servers", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "42")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	id, err := parseGameID(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("id = %d, want 42", id)
	}
}

func TestParseGameID_Invalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/games/abc/servers", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	_, err := parseGameID(req)
	if err == nil {
		t.Fatal("expected error for invalid gameId, got nil")
	}
}

func TestParseGameID_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/games//servers", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	_, err := parseGameID(req)
	if err == nil {
		t.Fatal("expected error for empty gameId, got nil")
	}
}

// ─── parseInstanceID tests ──────────────────────────────────────────────────

func TestParseInstanceID_Valid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/instances/100", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("instanceId", "100")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	id, err := parseInstanceID(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 100 {
		t.Errorf("id = %d, want 100", id)
	}
}

func TestParseInstanceID_Invalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/instances/xyz", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("instanceId", "xyz")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	_, err := parseInstanceID(req)
	if err == nil {
		t.Fatal("expected error for invalid instanceId, got nil")
	}
}

// ─── parseNodeID tests ──────────────────────────────────────────────────────

func TestParseNodeID_Valid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/nodes/7", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("nodeId", "7")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	id, err := parseNodeID(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 7 {
		t.Errorf("id = %d, want 7", id)
	}
}

func TestParseNodeID_Invalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/nodes/bad", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("nodeId", "bad")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	_, err := parseNodeID(req)
	if err == nil {
		t.Fatal("expected error for invalid nodeId, got nil")
	}
}

// ─── parseInstanceStatus tests ──────────────────────────────────────────────

func TestParseInstanceStatus_Valid(t *testing.T) {
	tests := []struct {
		query  string
		expect domain.InstanceStatus
	}{
		{"starting", domain.InstanceStatusStarting},
		{"running", domain.InstanceStatusRunning},
		{"stopping", domain.InstanceStatusStopping},
		{"stopped", domain.InstanceStatusStopped},
		{"crashed", domain.InstanceStatusCrashed},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?status="+tt.query, nil)
			result := parseInstanceStatus(req)
			if result == nil {
				t.Fatalf("expected non-nil result for %q", tt.query)
			}
			if *result != tt.expect {
				t.Errorf("status = %v, want %v", *result, tt.expect)
			}
		})
	}
}

func TestParseInstanceStatus_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	result := parseInstanceStatus(req)
	if result != nil {
		t.Errorf("expected nil for empty status, got %v", *result)
	}
}

func TestParseInstanceStatus_Invalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?status=foobar", nil)
	result := parseInstanceStatus(req)
	if result != nil {
		t.Errorf("expected nil for invalid status, got %v", *result)
	}
}

// ─── parseNodeStatus tests ──────────────────────────────────────────────────

func TestParseNodeStatus_Valid(t *testing.T) {
	tests := []struct {
		query  string
		expect domain.NodeStatus
	}{
		{"unauthorized", domain.NodeStatusUnauthorized},
		{"online", domain.NodeStatusOnline},
		{"offline", domain.NodeStatusOffline},
		{"maintenance", domain.NodeStatusMaintenance},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?status="+tt.query, nil)
			result := parseNodeStatus(req)
			if result == nil {
				t.Fatalf("expected non-nil result for %q", tt.query)
			}
			if *result != tt.expect {
				t.Errorf("status = %v, want %v", *result, tt.expect)
			}
		})
	}
}

func TestParseNodeStatus_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	result := parseNodeStatus(req)
	if result != nil {
		t.Errorf("expected nil for empty status, got %v", *result)
	}
}

func TestParseNodeStatus_Invalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?status=unknown_status", nil)
	result := parseNodeStatus(req)
	if result != nil {
		t.Errorf("expected nil for invalid status, got %v", *result)
	}
}

// ─── parseProtocol tests ────────────────────────────────────────────────────

func TestParseProtocol_Valid(t *testing.T) {
	tests := []struct {
		input  string
		expect domain.Protocol
	}{
		{"tcp", domain.ProtocolTCP},
		{"udp", domain.ProtocolUDP},
		{"websocket", domain.ProtocolWebSocket},
		{"webrtc", domain.ProtocolWebRTC},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseProtocol(tt.input)
			if result != tt.expect {
				t.Errorf("parseProtocol(%q) = %v, want %v", tt.input, result, tt.expect)
			}
		})
	}
}

func TestParseProtocol_Invalid(t *testing.T) {
	result := parseProtocol("invalid")
	if result != 0 {
		t.Errorf("parseProtocol(\"invalid\") = %v, want 0", result)
	}
}

func TestParseProtocol_Empty(t *testing.T) {
	result := parseProtocol("")
	if result != 0 {
		t.Errorf("parseProtocol(\"\") = %v, want 0", result)
	}
}

// ─── handleDomainError tests ────────────────────────────────────────────────

func TestHandleDomainError_NoError(t *testing.T) {
	rec := httptest.NewRecorder()
	handled := handleDomainError(rec, nil, "test action")
	if handled {
		t.Error("expected handled=false for nil error")
	}
}

func TestHandleDomainError_NotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	handled := handleDomainError(rec, domain.ErrNotFound, "find")
	if !handled {
		t.Fatal("expected handled=true")
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["code"] != "NOT_FOUND" {
		t.Errorf("code = %q, want %q", body["code"], "NOT_FOUND")
	}
}

func TestHandleDomainError_AlreadyExists(t *testing.T) {
	rec := httptest.NewRecorder()
	handled := handleDomainError(rec, domain.ErrAlreadyExists, "create")
	if !handled {
		t.Fatal("expected handled=true")
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["code"] != "CONFLICT" {
		t.Errorf("code = %q, want %q", body["code"], "CONFLICT")
	}
}

func TestHandleDomainError_BuildInUse(t *testing.T) {
	rec := httptest.NewRecorder()
	handled := handleDomainError(rec, domain.ErrBuildInUse, "delete build")
	if !handled {
		t.Fatal("expected handled=true")
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["code"] != "CONFLICT" {
		t.Errorf("code = %q, want %q", body["code"], "CONFLICT")
	}
}

func TestHandleDomainError_InvalidToken(t *testing.T) {
	rec := httptest.NewRecorder()
	handled := handleDomainError(rec, domain.ErrInvalidToken, "auth")
	if !handled {
		t.Fatal("expected handled=true")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["code"] != "UNAUTHORIZED" {
		t.Errorf("code = %q, want %q", body["code"], "UNAUTHORIZED")
	}
}

func TestHandleDomainError_NoAvailableNode(t *testing.T) {
	rec := httptest.NewRecorder()
	handled := handleDomainError(rec, domain.ErrNoAvailableNode, "schedule")
	if !handled {
		t.Fatal("expected handled=true")
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["code"] != "CONFLICT" {
		t.Errorf("code = %q, want %q", body["code"], "CONFLICT")
	}
}

func TestHandleDomainError_GenericError(t *testing.T) {
	rec := httptest.NewRecorder()
	genericErr := errors.New("some generic error")
	handled := handleDomainError(rec, genericErr, "action")
	if !handled {
		t.Fatal("expected handled=true")
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["code"] != "INTERNAL_ERROR" {
		t.Errorf("code = %q, want %q", body["code"], "INTERNAL_ERROR")
	}
}

// ─── buildStatusResponse tests ──────────────────────────────────────────────

func TestBuildStatusResponse(t *testing.T) {
	tests := []struct {
		status domain.InstanceStatus
		expect string
	}{
		{domain.InstanceStatusStarting, "starting"},
		{domain.InstanceStatusRunning, "running"},
		{domain.InstanceStatusStopping, "stopping"},
		{domain.InstanceStatusStopped, "stopped"},
		{domain.InstanceStatusCrashed, "crashed"},
	}

	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			result := buildStatusResponse(tt.status)
			if result != tt.expect {
				t.Errorf("buildStatusResponse(%v) = %q, want %q", tt.status, result, tt.expect)
			}
		})
	}
}

// ─── nodeStatusResponse tests ───────────────────────────────────────────────

func TestNodeStatusResponse(t *testing.T) {
	tests := []struct {
		status domain.NodeStatus
		expect string
	}{
		{domain.NodeStatusUnauthorized, "unauthorized"},
		{domain.NodeStatusOnline, "online"},
		{domain.NodeStatusOffline, "offline"},
		{domain.NodeStatusMaintenance, "maintenance"},
	}

	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			result := nodeStatusResponse(tt.status)
			if result != tt.expect {
				t.Errorf("nodeStatusResponse(%v) = %q, want %q", tt.status, result, tt.expect)
			}
		})
	}
}

// ─── protocolResponse tests ─────────────────────────────────────────────────

func TestProtocolResponse(t *testing.T) {
	tests := []struct {
		proto  domain.Protocol
		expect string
	}{
		{domain.ProtocolTCP, "tcp"},
		{domain.ProtocolUDP, "udp"},
		{domain.ProtocolWebSocket, "websocket"},
		{domain.ProtocolWebRTC, "webrtc"},
	}

	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			result := protocolResponse(tt.proto)
			if result != tt.expect {
				t.Errorf("protocolResponse(%v) = %q, want %q", tt.proto, result, tt.expect)
			}
		})
	}
}

// ─── logSourceResponse tests ────────────────────────────────────────────────

func TestLogSourceResponse(t *testing.T) {
	tests := []struct {
		src    domain.LogSource
		expect string
	}{
		{domain.LogSourceStdout, "stdout"},
		{domain.LogSourceStderr, "stderr"},
	}

	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			result := logSourceResponse(tt.src)
			if result != tt.expect {
				t.Errorf("logSourceResponse(%v) = %q, want %q", tt.src, result, tt.expect)
			}
		})
	}
}

// ─── decodeJSON tests ───────────────────────────────────────────────────────

func TestDecodeJSON_Success(t *testing.T) {
	body := `{"name":"test"}`
	type target struct {
		Name string `json:"name"`
	}
	var out target

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	err := decodeJSON(req, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Name != "test" {
		t.Errorf("Name = %q, want %q", out.Name, "test")
	}
}

func TestDecodeJSON_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{invalid"))
	var out map[string]string
	err := decodeJSON(req, &out)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestDecodeJSON_UnknownField(t *testing.T) {
	type target struct {
		Name string `json:"name"`
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"test","extra":"field"}`))
	var out target
	err := decodeJSON(req, &out)
	if err == nil {
		t.Fatal("expected error for unknown field, got nil")
	}
}
