package auth

// Коды ошибок домена auth для единого формата ApiError.
const (
	ErrCodeEmailTaken     = "EMAIL_TAKEN"
	ErrCodeUsernameTaken  = "USERNAME_TAKEN"
	ErrCodeInvalidLogin   = "INVALID_LOGIN"
	ErrCodeSessionExpired = "SESSION_EXPIRED"
	ErrCodeAccountBlocked = "ACCOUNT_BLOCKED"
)
