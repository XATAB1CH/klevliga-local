package auth

import "slices"

const (
	RoleUser      = "user"
	RoleModerator = "moderator"
	RoleSupport   = "support"
	RoleAdmin     = "admin"
)

type Permission string

const (
	ManageUsers     Permission = "users.manage"
	ViewAudit       Permission = "audit.view"
	ModerateContent Permission = "content.moderate"
	HandleSupport   Permission = "support.handle"
)

func (u *User) HasRole(role string) bool {
	return u != nil && slices.Contains(u.Roles, role)
}

// Can is transport-independent: both HTTP and future event handlers use the
// same policy. Unknown permissions fail closed, including private chat access.
func (u *User) Can(permission Permission) bool {
	if u == nil || u.Status != UserStatusActive {
		return false
	}
	switch permission {
	case ManageUsers, ViewAudit:
		return u.HasRole(RoleAdmin)
	case ModerateContent:
		return u.HasRole(RoleAdmin) || u.HasRole(RoleModerator)
	case HandleSupport:
		return u.HasRole(RoleAdmin) || u.HasRole(RoleSupport)
	default:
		return false
	}
}

func (u *User) Permissions() []Permission {
	permissions := make([]Permission, 0, 4)
	for _, permission := range []Permission{ManageUsers, ViewAudit, ModerateContent, HandleSupport} {
		if u.Can(permission) {
			permissions = append(permissions, permission)
		}
	}
	return permissions
}

// NormalizeRoles validates a full replacement and returns a stable order.
// Baseline user rights cannot be removed; duplicates and unknown roles fail.
func NormalizeRoles(roles []string) ([]string, bool) {
	if len(roles) < 1 || len(roles) > 4 || !slices.Contains(roles, RoleUser) {
		return nil, false
	}
	seen := make(map[string]bool, len(roles))
	for _, role := range roles {
		if seen[role] || !slices.Contains([]string{RoleUser, RoleModerator, RoleSupport, RoleAdmin}, role) {
			return nil, false
		}
		seen[role] = true
	}
	normalized := make([]string, 0, len(roles))
	for _, role := range []string{RoleUser, RoleModerator, RoleSupport, RoleAdmin} {
		if seen[role] {
			normalized = append(normalized, role)
		}
	}
	return normalized, true
}
