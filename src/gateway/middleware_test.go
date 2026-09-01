package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddleware(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test OPTIONS preflight
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/projects", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for OPTIONS, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("unexpected allow origin: %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestJWTMetadataAnnotator(t *testing.T) {
	// Sample unsigned JWT payload: {"sub":"user-123","role":"admin"} -> eyJzdWIiOiJ1c2VyLTEyMyIsInJvbGUiOiJhZG1pbiJ9
	// Header: {"alg":"none"} -> eyJhbGciOiJub25lIn0
	token := "eyJhbGciOiJub25lIn0.eyJzdWIiOiJ1c2VyLTEyMyIsInJvbGUiOiJhZG1pbiJ9.sig"

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	md := jwtMetadataAnnotator(context.Background(), req)

	if auth := md.Get("authorization"); len(auth) == 0 || auth[0] != "Bearer "+token {
		t.Fatalf("expected authorization header in metadata, got %v", auth)
	}
	if userID := md.Get("x-user-id"); len(userID) == 0 || userID[0] != "user-123" {
		t.Fatalf("expected user-123, got %v", userID)
	}
	if role := md.Get("x-user-role"); len(role) == 0 || role[0] != "admin" {
		t.Fatalf("expected admin, got %v", role)
	}
}
