package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"google.golang.org/grpc/metadata"
)

// corsMiddleware обрабатывает CORS-заголовки и preflight OPTIONS-запросы.
//
// allowedOrigins — список разрешённых источников (CORS_ALLOWED_ORIGINS). Пустой список означает режим
// разработки: отражается любой Origin. В продуктивной среде список обязателен (см. loadAllowedOrigins).
func corsMiddleware(allowedOrigins []string, next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[strings.TrimRight(strings.TrimSpace(o), "/")] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			_, ok := allowed[origin]
			if len(allowed) == 0 || ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Api-Key,x-api-key")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loadAllowedOrigins читает CORS_ALLOWED_ORIGINS (список через запятую).
// В продуктивной среде пустой список — ошибка конфигурации.
func loadAllowedOrigins() ([]string, error) {
	var origins []string
	for _, o := range strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}
	if len(origins) == 0 && isProd() {
		return nil, errors.New("prod: CORS_ALLOWED_ORIGINS is required")
	}
	return origins, nil
}

// jwtMetadataAnnotator пробрасывает токен пользователя в gRPC-метаданные.
// Личность и роль шлюз НЕ передаёт: сервисы извлекают их из токена после проверки подписи.
func jwtMetadataAnnotator(_ context.Context, req *http.Request) metadata.MD {
	md := metadata.MD{}
	if rawToken := extractRawToken(req); rawToken != "" {
		md.Set("authorization", "Bearer "+rawToken)
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
