// Package apperr содержит общие доменные ошибки, единые для всех слоёв приложения.
package apperr

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrValidation         = errors.New("validation error")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrNotImplemented     = errors.New("not implemented")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
