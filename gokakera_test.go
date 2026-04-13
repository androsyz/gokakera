package gokakera_test

import (
	"context"
	"testing"
	"time"

	"github.com/androsyz/gokakera"
	"github.com/androsyz/gokakera/localstorage"
	"github.com/androsyz/gokakera/memorystore"
)

const errExpectedNoError = "expected no error, got %v"

func setupKakera(t *testing.T) *gokakera.Kakera {
	t.Helper()

	dir := t.TempDir()

	return gokakera.New(
		gokakera.WithChunkSize(10),
		gokakera.WithMaxFileSize(1024),
		gokakera.WithSessionTTL(1*time.Hour),
		gokakera.WithChecksumType("crc32"),
		gokakera.WithStorage(localstorage.New(dir)),
		gokakera.WithSessionStore(memorystore.New()),
	)
}

func computeChecksum(t *testing.T, data []byte) string {
	t.Helper()

	k := gokakera.New(gokakera.WithChecksumType("crc32"))
	checksum, err := k.ExportComputeChecksum(data)
	if err != nil {
		t.Fatalf("failed to compute checksum: %v", err)
	}
	return checksum
}

func TestCreateSessionValidFile(t *testing.T) {
	ctx := context.Background()
	session, err := setupKakera(t).CreateSession(ctx, "test.pdf", 25)
	if err != nil {
		t.Fatalf(errExpectedNoError, err)
	}
	if session.Filename != "test.pdf" {
		t.Errorf("expected filename test.pdf, got %s", session.Filename)
	}
	if session.TotalChunks != 3 {
		t.Errorf("expected 3 chunks, got %d", session.TotalChunks)
	}
	if session.Status != gokakera.StatusUploading {
		t.Errorf("expected status %s, got %s", gokakera.StatusUploading, session.Status)
	}
}

func TestCreateSessionExactChunkSize(t *testing.T) {
	ctx := context.Background()
	session, err := setupKakera(t).CreateSession(ctx, "exact.pdf", 10)
	if err != nil {
		t.Fatalf(errExpectedNoError, err)
	}
	if session.TotalChunks != 1 {
		t.Errorf("expected 1 chunk, got %d", session.TotalChunks)
	}
	if session.Status != gokakera.StatusUploading {
		t.Errorf("expected status %s, got %s", gokakera.StatusUploading, session.Status)
	}
}

func TestCreateSessionExceedsMaxSize(t *testing.T) {
	ctx := context.Background()
	if _, err := setupKakera(t).CreateSession(ctx, "big.pdf", 2048); err == nil {
		t.Fatal("expected error for size exceeding max, got nil")
	}
}

func TestCreateSessionZeroSize(t *testing.T) {
	ctx := context.Background()
	if _, err := setupKakera(t).CreateSession(ctx, "empty.pdf", 0); err == nil {
		t.Fatal("expected error for zero size, got nil")
	}
}

func TestCreateSessionNegativeSize(t *testing.T) {
	ctx := context.Background()
	if _, err := setupKakera(t).CreateSession(ctx, "bad.pdf", -1); err == nil {
		t.Fatal("expected error for negative size, got nil")
	}
}

func TestUploadChunk(t *testing.T) {
	k := setupKakera(t)
	ctx := context.Background()

	session, _ := k.CreateSession(ctx, "test.txt", 25)
	validData := []byte("helloworld")
	validChecksum := computeChecksum(t, validData)

	tests := []struct {
		name     string
		index    int
		data     []byte
		checksum string
		wantErr  bool
	}{
		{
			name:     "valid chunk",
			index:    0,
			data:     validData,
			checksum: validChecksum,
			wantErr:  false,
		},
		{
			name:     "bad checksum",
			index:    1,
			data:     validData,
			checksum: "badchecksum",
			wantErr:  true,
		},
		{
			name:     "invalid index negative",
			index:    -1,
			data:     validData,
			checksum: validChecksum,
			wantErr:  true,
		},
		{
			name:     "invalid index too high",
			index:    99,
			data:     validData,
			checksum: validChecksum,
			wantErr:  true,
		},
		{
			name:     "idempotent upload",
			index:    0,
			data:     validData,
			checksum: validChecksum,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := k.UploadChunk(ctx, session.ID, tt.index, tt.data, tt.checksum)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf(errExpectedNoError, err)
			}
		})
	}
}

func TestSHA256Checksum(t *testing.T) {
	ctx := context.Background()
	k := gokakera.New(
		gokakera.WithChunkSize(10),
		gokakera.WithMaxFileSize(1024),
		gokakera.WithSessionTTL(1*time.Hour),
		gokakera.WithChecksumType("sha256"),
		gokakera.WithStorage(localstorage.New(t.TempDir())),
		gokakera.WithSessionStore(memorystore.New()),
	)

	session, err := k.CreateSession(ctx, "sha.txt", 10)
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("helloworld")
	sha256Checksum, err := k.ExportComputeChecksum(data)
	if err != nil {
		t.Fatalf("failed to compute sha256 checksum: %v", err)
	}

	// Correct sha256 checksum should succeed.
	if err := k.UploadChunk(ctx, session.ID, 0, data, sha256Checksum); err != nil {
		t.Errorf("expected no error with sha256 checksum, got %v", err)
	}

	// A crc32 checksum should fail when sha256 is configured.
	crc32k := gokakera.New(gokakera.WithChecksumType("crc32"))
	crc32Checksum, _ := crc32k.ExportComputeChecksum(data)

	session2, _ := k.CreateSession(ctx, "sha2.txt", 10)
	if err := k.UploadChunk(ctx, session2.ID, 0, data, crc32Checksum); err == nil {
		t.Error("expected checksum mismatch error when providing crc32 to sha256 instance")
	}
}

func TestExpiredSession(t *testing.T) {
	ctx := context.Background()
	k := gokakera.New(
		gokakera.WithChunkSize(10),
		gokakera.WithMaxFileSize(1024),
		gokakera.WithSessionTTL(-1*time.Hour), // already expired
		gokakera.WithStorage(localstorage.New(t.TempDir())),
		gokakera.WithSessionStore(memorystore.New()),
	)

	session, err := k.CreateSession(ctx, "expired.txt", 10)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if _, err := k.GetSession(ctx, session.ID); err == nil {
		t.Error("expected error for expired session, got nil")
	}

	data := []byte("helloworld")
	checksum := computeChecksum(t, data)
	if err := k.UploadChunk(ctx, session.ID, 0, data, checksum); err == nil {
		t.Error("expected error uploading to expired session, got nil")
	}
}

func TestProgressCallback(t *testing.T) {
	ctx := context.Background()

	var received []gokakera.Progress
	k := gokakera.New(
		gokakera.WithChunkSize(10),
		gokakera.WithMaxFileSize(1024),
		gokakera.WithSessionTTL(1*time.Hour),
		gokakera.WithStorage(localstorage.New(t.TempDir())),
		gokakera.WithSessionStore(memorystore.New()),
		gokakera.WithOnProgress(func(p gokakera.Progress) {
			received = append(received, p)
		}),
	)

	session, err := k.CreateSession(ctx, "progress.txt", 25)
	if err != nil {
		t.Fatal(err)
	}

	chunks := [][]byte{
		[]byte("helloworld"),
		[]byte("helloworld"),
		[]byte("hello"),
	}

	for i, chunk := range chunks {
		checksum := computeChecksum(t, chunk)
		if err := k.UploadChunk(ctx, session.ID, i, chunk, checksum); err != nil {
			t.Fatalf("chunk %d: %v", i, err)
		}
	}

	if len(received) != len(chunks) {
		t.Fatalf("expected %d progress events, got %d", len(chunks), len(received))
	}

	last := received[len(received)-1]
	if last.ChunksReceived != last.TotalChunks {
		t.Errorf("last progress: expected ChunksReceived == TotalChunks, got %d/%d", last.ChunksReceived, last.TotalChunks)
	}
	if last.Percentage != 100 {
		t.Errorf("expected 100%% on last chunk, got %.1f%%", last.Percentage)
	}
	if last.BytesReceived > last.TotalSize {
		t.Errorf("BytesReceived %d exceeds TotalSize %d", last.BytesReceived, last.TotalSize)
	}
}

func TestStartCleanup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := memorystore.New()
	k := gokakera.New(
		gokakera.WithChunkSize(10),
		gokakera.WithMaxFileSize(1024),
		gokakera.WithSessionTTL(-1*time.Hour), // already expired
		gokakera.WithStorage(localstorage.New(t.TempDir())),
		gokakera.WithSessionStore(store),
	)

	session, err := k.CreateSession(ctx, "cleanup.txt", 10)
	if err != nil {
		t.Fatal(err)
	}

	// Session should be present in the store before cleanup.
	if _, err := store.Get(ctx, session.ID); err != nil {
		t.Fatal("session should exist in store before cleanup")
	}

	k.StartCleanup(ctx, 50*time.Millisecond)
	time.Sleep(200 * time.Millisecond)

	// After cleanup the session should have been deleted from the store.
	if _, err := store.Get(ctx, session.ID); err == nil {
		t.Error("expected session to be deleted from store after cleanup ran")
	}
}

func TestFullUploadFlow(t *testing.T) {
	k := setupKakera(t)
	ctx := context.Background()

	session, err := k.CreateSession(ctx, "test.txt", 25)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	chunks := [][]byte{
		[]byte("helloworld"),
		[]byte("helloworld"),
		[]byte("hello"),
	}

	for i, chunk := range chunks {
		checksum := computeChecksum(t, chunk)
		err := k.UploadChunk(ctx, session.ID, i, chunk, checksum)
		if err != nil {
			t.Fatalf("chunk %d failed: %v", i, err)
		}
	}

	complete, err := k.IsUploadComplete(ctx, session.ID)
	if err != nil {
		t.Fatalf("failed to check completion: %v", err)
	}

	if !complete {
		t.Fatal("expected upload to be complete")
	}

	err = k.AssembleFile(ctx, session.ID)
	if err != nil {
		t.Fatalf("assembly failed: %v", err)
	}

	session, _ = k.GetSession(ctx, session.ID)
	if session.Status != "completed" {
		t.Errorf("expected status completed, got %s", session.Status)
	}
}
