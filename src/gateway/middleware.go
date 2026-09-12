package main

import (
	"context"
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
	rawToken := extractRawToken(req)
	if rawToken != "" {
		md.Set("authorization", "Bearer "+rawToken)
	} else if auth := req.Header.Get("Authorization"); auth != "" {
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

// extractRawToken извлекает токен из заголовка Authorization, query параметров или cookies.
func extractRawToken(req *http.Request) string {
	if auth := req.Header.Get("Authorization"); auth != "" {
		trimmed := strings.TrimSpace(auth)
		if strings.HasPrefix(strings.ToLower(trimmed), "bearer ") {
			return strings.TrimSpace(trimmed[7:])
		}
		return trimmed
	}
	if tok := req.URL.Query().Get("token"); tok != "" {
		return strings.TrimSpace(tok)
	}
	if tok := req.URL.Query().Get("auth"); tok != "" {
		return strings.TrimSpace(tok)
	}
	if tok := req.URL.Query().Get("access_token"); tok != "" {
		return strings.TrimSpace(tok)
	}
	if cookie, err := req.Cookie("gdh_access_token"); err == nil && cookie.Value != "" {
		return strings.TrimSpace(cookie.Value)
	}
	if cookie, err := req.Cookie("gdh_session"); err == nil && cookie.Value != "" {
		return strings.TrimSpace(cookie.Value)
	}
	if cookie, err := req.Cookie("access_token"); err == nil && cookie.Value != "" {
		return strings.TrimSpace(cookie.Value)
	}
	return ""
}

// parseJWTClaims извлекает userID и userRole из JWT payload без валидации подписи (подпись проверяется в SSO).
func parseJWTClaims(req *http.Request) (userID, userRole string) {
	rawToken := extractRawToken(req)
	if rawToken == "" {
		return "", ""
	}

	parts := strings.Split(rawToken, ".")
	if len(parts) != 3 {
		return "", ""
	}

	payload, err := decodeJWTSegment(parts[1])
	if err != nil {
		return "", ""
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", ""
	}

	if sub, ok := claims["sub"].(string); ok {
		userID = sub
	} else if uid, ok := claims["user_id"].(string); ok {
		userID = uid
	} else if id, ok := claims["id"].(string); ok {
		userID = id
	}

	var rawRole any
	if rVal, ok := claims["role"]; ok {
		rawRole = rVal
	} else if rVal, ok := claims["user_role"]; ok {
		rawRole = rVal
	} else if rVal, ok := claims["roles"]; ok {
		rawRole = rVal
	}

	switch v := rawRole.(type) {
	case float64:
		userRole = fmt.Sprintf("%.0f", v)
	case int:
		userRole = fmt.Sprintf("%d", v)
	case int64:
		userRole = fmt.Sprintf("%d", v)
	case json.Number:
		userRole = v.String()
	case string:
		userRole = v
	case []any:
		if len(v) > 0 {
			userRole = fmt.Sprintf("%v", v[0])
		}
	}

	return userID, userRole
}
