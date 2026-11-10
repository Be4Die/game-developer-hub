package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// tokenVerifier проверяет подпись и срок действия JWT, выпущенных SSO.
//
// Шлюз использует результат проверки только для собственных HTTP-обработчиков (вложения, скачивания).
// В gRPC-сервисы пробрасывается исходный токен: каждый сервис проверяет его самостоятельно
// и не доверяет заголовкам x-user-*.
type tokenVerifier struct {
	secret []byte
	issuer string
}

const minJWTSecretLength = 32

// newTokenVerifierFromEnv создаёт верификатор из JWT_SECRET и JWT_ISSUER.
// Секрет обязателен: без него шлюз не может отличить настоящий токен от поддельного.
func newTokenVerifierFromEnv() (*tokenVerifier, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}
	if isProd() {
		if len(secret) < minJWTSecretLength || strings.Contains(secret, "dev-") || strings.Contains(secret, "local-testing") {
			return nil, fmt.Errorf("prod: JWT_SECRET must be a unique random value of at least %d characters", minJWTSecretLength)
		}
	}
	return &tokenVerifier{secret: []byte(secret), issuer: envOr("JWT_ISSUER", "sso")}, nil
}

// verify возвращает идентификатор и числовую роль ("1" разработчик, "2" модератор, "3" администратор)
// из токена с корректной подписью. Для любого другого токена ok == false.
func (v *tokenVerifier) verify(rawToken string) (userID, role string, ok bool) {
	if v == nil || rawToken == "" {
		return "", "", false
	}

	token, err := jwt.Parse(rawToken, func(t *jwt.Token) (any, error) {
		if _, isHMAC := t.Method.(*jwt.SigningMethodHMAC); !isHMAC {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.secret, nil
	}, jwt.WithIssuer(v.issuer), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return "", "", false
	}

	claims, isMap := token.Claims.(jwt.MapClaims)
	if !isMap {
		return "", "", false
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", "", false
	}

	role = "1"
	if r, isNum := claims["role"].(float64); isNum && (r == 2 || r == 3) {
		role = strconv.Itoa(int(r))
	}
	return sub, role, true
}

// authenticate извлекает токен из запроса и проверяет его.
// Возвращает пустой userID, если токена нет или он не прошёл проверку.
func (v *tokenVerifier) authenticate(r *http.Request) (userID, role, rawToken string) {
	rawToken = extractRawToken(r)
	userID, role, ok := v.verify(rawToken)
	if !ok {
		return "", "", ""
	}
	return userID, role, rawToken
}

// blockedMetadataHeader сообщает, запрещено ли клиенту передавать заголовок в gRPC-метаданные.
// Через префикс Grpc-Metadata- нельзя подставить служебные ключи, по которым сервисы
// определяют пользователя и межсервисные вызовы.
func blockedMetadataHeader(key string) bool {
	lower := strings.ToLower(key)
	name, hasPrefix := strings.CutPrefix(lower, "grpc-metadata-")
	if !hasPrefix {
		name = lower
	}
	return strings.HasPrefix(name, "x-user-") ||
		strings.HasPrefix(name, "x-internal-") ||
		name == "x-backup-ticket" ||
		(hasPrefix && name == "authorization")
}
