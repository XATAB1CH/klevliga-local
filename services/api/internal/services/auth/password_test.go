package auth

import "testing"

func TestPasswordHashAndVerify(t *testing.T) {
	const password = "correct horse battery staple"

	hash, err := hashPassword(password)
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}

	ok, err := verifyPassword(password, hash)
	if err != nil {
		t.Fatalf("verifyPassword: %v", err)
	}
	if !ok {
		t.Fatal("expected password to verify successfully")
	}

	ok, err = verifyPassword("wrong password", hash)
	if err != nil {
		t.Fatalf("verifyPassword(wrong): %v", err)
	}
	if ok {
		t.Fatal("expected wrong password to fail verification")
	}
}

func TestPasswordHashIsSalted(t *testing.T) {
	a, err := hashPassword("same-password")
	if err != nil {
		t.Fatal(err)
	}

	b, err := hashPassword("same-password")
	if err != nil {
		t.Fatal(err)
	}

	if a == b {
		t.Fatal("expected different hashes for the same password (unique salt)")
	}
}
