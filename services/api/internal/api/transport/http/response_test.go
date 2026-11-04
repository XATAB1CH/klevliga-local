package http

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTimeCursorRoundtrip(t *testing.T) {
	id := uuid.New()
	ts := time.Date(2026, 10, 29, 10, 0, 0, 123, time.UTC)

	s := encodeTimeCursor(ts, id)

	gotTs, gotID, err := decodeTimeCursor(s)
	if err != nil {
		t.Fatalf("decodeTimeCursor: %v", err)
	}
	if !gotTs.Equal(ts) {
		t.Fatalf("timestamp mismatch: got %v want %v", gotTs, ts)
	}
	if gotID != id {
		t.Fatalf("id mismatch: got %s want %s", gotID, id)
	}
}

func TestWaterCursorRoundtrip(t *testing.T) {
	id := uuid.New()
	const name = "Москва-река"

	s := encodeWaterCursor(name, id)

	gotName, gotID, err := decodeWaterCursor(s)
	if err != nil {
		t.Fatalf("decodeWaterCursor: %v", err)
	}
	if gotName != name {
		t.Fatalf("name mismatch: got %q want %q", gotName, name)
	}
	if gotID != id {
		t.Fatalf("id mismatch: got %s want %s", gotID, id)
	}
}

func TestDecodeCursorInvalid(t *testing.T) {
	if _, _, err := decodeTimeCursor("not-a-cursor"); err == nil {
		t.Fatal("expected error for invalid cursor")
	}
}
