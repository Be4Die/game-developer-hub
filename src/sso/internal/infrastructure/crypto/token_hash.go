// Package cryptoprov предоставляет криптографические утилиты.
package cryptoprov

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken хеширует токен через SHA-256 для безопасного хранения.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
