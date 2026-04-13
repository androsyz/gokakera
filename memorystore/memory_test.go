package memorystore_test

import (
	"context"
	"testing"
	"time"

	"github.com/androsyz/gokakera"
	"github.com/androsyz/gokakera/memorystore"
)

func newSession(id string, ttl time.Duration) *gokakera.Session {
	return &gokakera.Session{
		ID:             id,
		Filename:       "test.txt",
		TotalSize:      100,
		TotalChunks:    10,
		ReceivedChunks: make(map[int]bool),
		Status:         "uploading",
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(ttl),
	}
}

func TestMemoryStoreCreateAndGet(t *testing.T) {
	ctx := context.Background()
	store := memorystore.New()
	session := newSession("abc123", 1*time.Hour)

	if err := store.Create(ctx, session); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := store.Get(ctx, session.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != session.ID {
		t.Errorf("expected ID %s, got %s", session.ID, got.ID)
	}
}

func TestMemoryStoreGetMissingReturnsError(t *testing.T) {
	ctx := context.Background()
	store := memorystore.New()

	if _, err := store.Get(ctx, "nonexistent"); err == nil {
		t.Error("expected error for missing session, got nil")
	}
}

func TestMemoryStoreUpdate(t *testing.T) {
	ctx := context.Background()
	store := memorystore.New()
	session := newSession("upd1", 1*time.Hour)

	if err := store.Create(ctx, session); err != nil {
		t.Fatal(err)
	}
	session.Status = "assembling"
	if err := store.Update(ctx, session); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err := store.Get(ctx, session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "assembling" {
		t.Errorf("expected status assembling, got %s", got.Status)
	}
}

func TestMemoryStoreUpdateMissingReturnsError(t *testing.T) {
	ctx := context.Background()
	store := memorystore.New()

	if err := store.Update(ctx, newSession("ghost", 1*time.Hour)); err == nil {
		t.Error("expected error updating non-existent session")
	}
}

func TestMemoryStoreListExpired(t *testing.T) {
	ctx := context.Background()
	store := memorystore.New()

	active := newSession("active1", 1*time.Hour)
	expired := newSession("expired1", -1*time.Hour)

	if err := store.Create(ctx, active); err != nil {
		t.Fatal(err)
	}
	if err := store.Create(ctx, expired); err != nil {
		t.Fatal(err)
	}

	list, err := store.ListExpired(ctx)
	if err != nil {
		t.Fatal(err)
	}

	foundExpired := false
	for _, s := range list {
		if s.ID == expired.ID {
			foundExpired = true
		}
		if s.ID == active.ID {
			t.Error("active session should not appear in ListExpired")
		}
	}
	if !foundExpired {
		t.Error("expected expired session in ListExpired result")
	}
}

func TestMemoryStoreDelete(t *testing.T) {
	ctx := context.Background()
	store := memorystore.New()
	session := newSession("del1", 1*time.Hour)

	if err := store.Create(ctx, session); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete(ctx, session.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get(ctx, session.ID); err == nil {
		t.Error("expected error after Delete, got nil")
	}
}
