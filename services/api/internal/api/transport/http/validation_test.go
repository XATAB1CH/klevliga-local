package http

import (
	"strings"
	"testing"
)

func TestTextValidationCountsUnicodeCharacters(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		min   int
		max   int
		valid bool
	}{
		{"short Cyrillic title", "Щука", titleMinLength, titleMaxLength, false},
		{"minimum Cyrillic title", "Ловля", titleMinLength, titleMaxLength, true},
		{"short Cyrillic username", "Ян", usernameMinLength, usernameMaxLength, false},
		{"minimum Cyrillic username", "Илья", usernameMinLength, usernameMaxLength, true},
		{"emoji is one character", "🐟🐟", usernameMinLength, usernameMaxLength, false},
		{"maximum Cyrillic title", strings.Repeat("я", titleMaxLength), titleMinLength, titleMaxLength, true},
		{"overlong title", strings.Repeat("я", titleMaxLength+1), titleMinLength, titleMaxLength, false},
		{"maximum body", strings.Repeat("я", bodyMaxLength), 1, bodyMaxLength, true},
		{"overlong body", strings.Repeat("я", bodyMaxLength+1), 1, bodyMaxLength, false},
		{"empty body", "", 1, bodyMaxLength, false},
		{"invalid UTF8", string([]byte{0xff, 0xfe}), 1, bodyMaxLength, false},
		{"password minimum", strings.Repeat("я", passwordMinLength), passwordMinLength, passwordMaxLength, true},
		{"password maximum", strings.Repeat("я", passwordMaxLength), passwordMinLength, passwordMaxLength, true},
		{"password overlong", strings.Repeat("я", passwordMaxLength+1), passwordMinLength, passwordMaxLength, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := validTextLength(tc.value, tc.min, tc.max); got != tc.valid {
				t.Errorf("got valid=%v, want %v", got, tc.valid)
			}
		})
	}
	if validTextLength(normalizeText(" \t\n\u00a0"), 1, bodyMaxLength) {
		t.Fatal("whitespace-only normalized body was accepted")
	}
	if got := normalizeText(" \tЛовля щуки\n"); got != "Ловля щуки" {
		t.Fatalf("normalization: %q", got)
	}
}

func TestEmailValidationAcceptsOnlyMailboxAddresses(t *testing.T) {
	for _, tc := range []struct {
		value string
		valid bool
	}{
		{"angler@example.test", true},
		{"angler+reports@example.test", true},
		{"missing-at-sign", false},
		{"@example.test", false},
		{"Angler <angler@example.test>", false},
		{"angler@example.test,other@example.test", false},
		{"angler@example.test\r\nX-Injected: yes", false},
		{strings.Repeat("a", 255) + "@example.test", false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			if got := validEmail(tc.value); got != tc.valid {
				t.Errorf("got valid=%v, want %v", got, tc.valid)
			}
		})
	}
}
