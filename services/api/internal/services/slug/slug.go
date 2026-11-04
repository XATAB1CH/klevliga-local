// Package slug — генерация уникальных URL-совместимых slug для тем.
package slug

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// Unique строит slug из заголовка, дополняя его коротким случайным суффиксом,
// что гарантирует уникальность без повторных попыток записи.
func Unique(title string) string {
	base := Slugify(title)
	if base == "" {
		base = "topic"
	}

	return base + "-" + randomHex(6)
}

// Slugify нормализует строку в ASCII-slug (кириллица отбрасывается).
func Slugify(title string) string {
	var b strings.Builder
	lastDash := false

	for _, r := range strings.ToLower(title) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastDash = false
		default:
			if b.Len() > 0 && !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}

	return strings.Trim(b.String(), "-")
}

func randomHex(n int) string {
	buf := make([]byte, n)

	_, _ = rand.Read(buf)

	return hex.EncodeToString(buf)
}
