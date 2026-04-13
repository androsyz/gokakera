package gokakera_test

import (
	"context"
	"testing"
	"time"

	"github.com/androsyz/gokakera"
	"github.com/androsyz/gokakera/localstorage"
	"github.com/androsyz/gokakera/memorystore"
)

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

func TestCreateSession(t *testing.T) {
	k := setupKakera(t)
	ctx := context.Background()

	tests := []struct {
		name       string
		filename   string
		totalSize  int64
		wantErr    bool
		wantChunks int
		wantStatus string
	}{
		{
			name:       "valid file",
			filename:   "test.pdf",
			totalSize:  25,
			wantErr:    false,
			wantChunks: 3,
			wantStatus: "uploading",
		},
		{
			name:       "exact chunk size",
			filename:   "exact.pdf",
			totalSize:  10,
			wantErr:    false,
			wantChunks: 1,
			wantStatus: "uploading",
		},
		{
			name:      "exceeds max size",
			filename:  "big.pdf",
			totalSize: 2048,
			wantErr:   true,
		},
		{
			name:      "zero size",
			filename:  "empty.pdf",
			totalSize: 0,
			wantErr:   true,
		},
		{
			name:      "negative size",
			filename:  "bad.pdf",
			totalSize: -1,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := k.CreateSession(ctx, tt.filename, tt.totalSize)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if session.Filename != tt.filename {
				t.Errorf("expected filename %s, got %s", tt.filename, session.Filename)
			}

			if session.TotalChunks != tt.wantChunks {
				t.Errorf("expected %d chunks, got %d", tt.wantChunks, session.TotalChunks)
			}

			if session.Status != tt.wantStatus {
				t.Errorf("expected status %s, got %s", tt.wantStatus, session.Status)
			}
		})
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
				t.Fatalf("expected no error, got %v", err)
			}
		})
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
