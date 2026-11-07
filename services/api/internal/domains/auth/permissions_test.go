package auth

import (
	"reflect"
	"testing"
)

func TestPermissionBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name  string
		roles []string
		want  []Permission
	}{
		{"user", []string{RoleUser}, []Permission{}},
		{"moderator", []string{RoleUser, RoleModerator}, []Permission{ModerateContent}},
		{"support", []string{RoleUser, RoleSupport}, []Permission{HandleSupport}},
		{"combined", []string{RoleUser, RoleModerator, RoleSupport}, []Permission{ModerateContent, HandleSupport}},
		{"admin", []string{RoleUser, RoleAdmin}, []Permission{ManageUsers, ViewAudit, ModerateContent, HandleSupport}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			u := &User{Status: UserStatusActive, Roles: tc.roles}
			if got := u.Permissions(); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("permissions: got %v, want %v", got, tc.want)
			}
			if u.Can("private_chat.read") || u.Can("unknown") {
				t.Fatal("unrecognized permission was allowed")
			}
			u.Status = UserStatusBlocked
			if len(u.Permissions()) != 0 {
				t.Fatal("blocked account retained privileges")
			}
		})
	}
	var absent *User
	if absent.Can(ManageUsers) {
		t.Fatal("absent user was allowed")
	}
}

func TestNormalizeRoles(t *testing.T) {
	for _, roles := range [][]string{nil, {}, {RoleAdmin}, {RoleUser, RoleUser}, {RoleUser, "owner"}, {RoleUser, ""}} {
		if _, ok := NormalizeRoles(roles); ok {
			t.Fatalf("accepted invalid roles %v", roles)
		}
	}
	got, ok := NormalizeRoles([]string{RoleAdmin, RoleSupport, RoleUser, RoleModerator})
	if !ok || !reflect.DeepEqual(got, []string{RoleUser, RoleModerator, RoleSupport, RoleAdmin}) {
		t.Fatalf("normalization: %v, %v", got, ok)
	}
}
