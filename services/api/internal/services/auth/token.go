package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// newSessionToken генерирует случайный токен сессии и его SHA-256 хэш.
// В БД сохраняется только хэш; клиенту отдаётся сырой токен.
func newSessionToken() (raw string, hash []byte, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", nil, err
	}

	raw = base64.RawURLEncoding.EncodeToString(buf)
	hash = hashToken(raw)

	return raw, hash, nil
}

func hashToken(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))

	return sum[:]
}
