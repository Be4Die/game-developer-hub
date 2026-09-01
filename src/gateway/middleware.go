package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

// corsMiddleware обрабатывает CORS-заголовки и preflight OPTIONS-запросы.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Api-Key,x-api-key")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// jwtMetadataAnnotator извлекает Authorization и claims пользователя из JWT токена в gRPC metadata.
func jwtMetadataAnnotator(_ context.Context, req *http.Request) metadata.MD {
	md := metadata.MD{}
	if auth := req.Header.Get("Authorization"); auth != "" {
		md.Set("authorization", auth)
	}

	userID, userRole := parseJWTClaims(req)
	if userID != "" {
		md.Set("x-user-id", userID)
	}
	if userRole != "" {
		md.Set("x-user-role", userRole)
	}

	return md
}

// parseJWTClaims извлекает userID и userRole из JWT payload без валидации подписи (подпись проверяется в SSO).
func parseJWTClaims(req *http.Request) (userID, userRole string) {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return "", ""
	}

	parts := strings.Split(strings.TrimPrefix(authHeader, "Bearer "), ".")
	if len(parts) != 3 {
		return "", ""
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", ""
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", ""
	}

	if sub, ok := claims["sub"].(string); ok {
		userID = sub
	}

	if role, ok := claims["role"].(float64); ok {
		userRole = fmt.Sprintf("%.0f", role)
	} else if role, ok := claims["role"].(string); ok {
		userRole = role
	}

	return userID, userRole
}
