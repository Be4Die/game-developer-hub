package cryptoprov_test

import (
	"testing"

	cryptoprov "github.com/Be4Die/game-developer-hub/sso/internal/infrastructure/crypto"
)

func TestHashToken(t *testing.T) {
	t.Parallel()

	t.Run("hash token", func(t *testing.T) {
		t.Parallel()

		token := "test-refresh-token"
		hash := cryptoprov.HashToken(token)

		if hash == "" {
			t.Fatal("expected non-empty hash")
		}

		// SHA-256 produces 64 hex characters
		if len(hash) != 64 {
			t.Errorf("expected hash length 64, got %d", len(hash))
		}
	})

	t.Run("same input produces same hash", func(t *testing.T) {
		t.Parallel()

		token := "test-token"
		hash1 := cryptoprov.HashToken(token)
		hash2 := cryptoprov.HashToken(token)

		if hash1 != hash2 {
			t.Error("expected same hash for same input")
		}
	})

	t.Run("different inputs produce different hashes", func(t *testing.T) {
		t.Parallel()

		hash1 := cryptoprov.HashToken("token-1")
		hash2 := cryptoprov.HashToken("token-2")

		if hash1 == hash2 {
			t.Error("expected different hashes for different inputs")
		}
	})

	t.Run("empty token", func(t *testing.T) {
		t.Parallel()

		hash := cryptoprov.HashToken("")

		if hash == "" {
			t.Fatal("expected non-empty hash for empty input")
		}

		if len(hash) != 64 {
			t.Errorf("expected hash length 64, got %d", len(hash))
		}
	})

	t.Run("hash is hex encoded", func(t *testing.T) {
		t.Parallel()

		hash := cryptoprov.HashToken("test")

		// Check all characters are valid hex
		for _, c := range hash {
			isDigit := c >= '0' && c <= '9'
			isLowerHex := c >= 'a' && c <= 'f'
			if !isDigit && !isLowerHex {
				t.Errorf("invalid hex character: %c", c)
			}
		}
	})
}
