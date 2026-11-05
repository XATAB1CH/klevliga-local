package http

import (
	"net/mail"
	"strings"
	"unicode/utf8"
)

const (
	usernameMinLength = 3
	usernameMaxLength = 32
	passwordMinLength = 12
	passwordMaxLength = 128
	titleMinLength    = 5
	titleMaxLength    = 250
	bodyMaxLength     = 20000
	queryMaxLength    = 200
	reasonMinLength   = 10
	reasonMaxLength   = 3000
)

// Text fields are normalized before storage. Passwords deliberately bypass
// this helper so leading and trailing spaces remain part of the credential.
func normalizeText(value string) string {
	return strings.TrimSpace(value)
}

func validTextLength(value string, min, max int) bool {
	if !utf8.ValidString(value) {
		return false
	}
	length := utf8.RuneCountInString(value)
	return length >= min && length <= max
}

func validEmail(value string) bool {
	if value == "" || len(value) > 254 {
		return false
	}
	address, err := mail.ParseAddress(value)
	return err == nil && address.Name == "" && address.Address == value
}
