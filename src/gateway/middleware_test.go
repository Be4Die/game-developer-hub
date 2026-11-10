package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestCORSMiddleware(t *testing.T) {
	handler := corsMiddleware(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestCORSMiddleware_AllowList(t *testing.T) {
	handler := corsMiddleware([]string{"https://hub.example.com"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for origin, want := range map[string]string{
		"https://hub.example.com":  "https://hub.example.com",
		"https://evil.example.com": "",
	} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		req.Header.Set("Origin", origin)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != want {
			t.Errorf("origin %s: expected allow-origin %q, got %q", origin, want, got)
		}
	}
}

func TestJWTMetadataAnnotator_ForwardsTokenOnly(t *testing.T) {
	// Неподписанный токен с ролью администратора: шлюз не должен превращать его содержимое в x-user-*.
	token := "eyJhbGciOiJub25lIn0.eyJzdWIiOiJ1c2VyLTEyMyIsInJvbGUiOiJhZG1pbiJ9.sig" //nolint:gosec // test token

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	md := jwtMetadataAnnotator(context.Background(), req)

	if auth := md.Get("authorization"); len(auth) == 0 || auth[0] != "Bearer "+token {
		t.Fatalf("expected authorization header in metadata, got %v", auth)
	}
	if v := md.Get("x-user-id"); len(v) != 0 {
		t.Fatalf("x-user-id must not be derived from an unverified token, got %v", v)
	}
	if v := md.Get("x-user-role"); len(v) != 0 {
		t.Fatalf("x-user-role must not be derived from an unverified token, got %v", v)
	}
}

func TestBlockedMetadataHeader(t *testing.T) {
	for key, blocked := range map[string]bool{
		"Grpc-Metadata-X-User-Id":        true,
		"Grpc-Metadata-x-user-role":      true,
		"grpc-metadata-x-internal-token": true,
		"Grpc-Metadata-Authorization":    true,
		"X-User-Role":                    true,
		"X-Backup-Ticket":                true,
		"Authorization":                  false,
		"Grpc-Metadata-X-Request-Id":     false,
		"Content-Type":                   false,
	} {
		if got := blockedMetadataHeader(key); got != blocked {
			t.Errorf("%s: expected blocked=%v, got %v", key, blocked, got)
		}
	}
}

func TestTokenVerifier(t *testing.T) {
	const secret, issuer = "0123456789abcdef0123456789abcdef", "sso"
	v := &tokenVerifier{secret: []byte(secret), issuer: issuer}

	sign := func(key string, claims jwt.MapClaims) string {
		s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(key))
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	exp := time.Now().Add(time.Hour).Unix()

	userID, role, ok := v.verify(sign(secret, jwt.MapClaims{"sub": "mod-1", "role": 2, "iss": issuer, "exp": exp}))
	if !ok || userID != "mod-1" || role != "2" {
		t.Fatalf("valid token: got %q %q %v", userID, role, ok)
	}

	rejected := map[string]string{
		"forged signature": sign("another-secret-another-secret-00", jwt.MapClaims{"sub": "a", "role": 3, "iss": issuer, "exp": exp}),
		"wrong issuer":     sign(secret, jwt.MapClaims{"sub": "a", "role": 3, "iss": "evil", "exp": exp}),
		"expired":          sign(secret, jwt.MapClaims{"sub": "a", "role": 3, "iss": issuer, "exp": time.Now().Add(-time.Minute).Unix()}),
		"no expiry":        sign(secret, jwt.MapClaims{"sub": "a", "role": 3, "iss": issuer}),
		"unsigned":         "eyJhbGciOiJub25lIn0.eyJzdWIiOiJhIiwicm9sZSI6M30.",
		"empty":            "",
	}
	for name, token := range rejected {
		if _, _, ok := v.verify(token); ok {
			t.Errorf("%s token must be rejected", name)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/x?token="+rejected["forged signature"], nil)
	if userID, _, _ := v.authenticate(req); userID != "" {
		t.Errorf("forged token must not authenticate, got %q", userID)
	}
}
