package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

// ReportToken вычисляет токен, которым игровой сервер подписывает отчёты об онлайне.
// Токен привязан к инстансу и выводится из секрета ноды, поэтому его не нужно хранить:
// после перезапуска ноды уже запущенные серверы продолжают отчитываться.
func ReportToken(secret string, instanceID int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte("instance-report:" + strconv.FormatInt(instanceID, 10)))
	return hex.EncodeToString(mac.Sum(nil))
}

// ValidReportToken проверяет токен отчёта за постоянное время.
func ValidReportToken(secret string, instanceID int64, token string) bool {
	if secret == "" || token == "" {
		return false
	}
	return hmac.Equal([]byte(ReportToken(secret, instanceID)), []byte(token))
}
