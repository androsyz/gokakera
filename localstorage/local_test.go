package localstorage_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/androsyz/gokakera/localstorage"
)

func TestStoreAndGetChunk(t *testing.T) {
	ctx := context.Background()
	ls := localstorage.New(t.TempDir())
	data := []byte("helloworld")

	if err := ls.StoreChunk(ctx, "sess", 0, data); err != nil {
		t.Fatalf("StoreChunk: %v", err)
	}
	got, err := ls.GetChunk(ctx, "sess", 0)
	if err != nil {
		t.Fatalf("GetChunk: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("expected %q, got %q", data, got)
	}
}

func TestGetChunkMissingReturnsError(t *testing.T) {
	ctx := context.Background()
	ls := localstorage.New(t.TempDir())

	if _, err := ls.GetChunk(ctx, "sess", 99); err == nil {
		t.Error("expected error for missing chunk, got nil")
	}
}

func TestDeleteChunk(t *testing.T) {
	ctx := context.Background()
	ls := localstorage.New(t.TempDir())
	data := []byte("todelete")

	if err := ls.StoreChunk(ctx, "sess", 0, data); err != nil {
		t.Fatal(err)
	}
	if err := ls.DeleteChunk(ctx, "sess", 0); err != nil {
		t.Fatalf("DeleteChunk: %v", err)
	}
	if _, err := ls.GetChunk(ctx, "sess", 0); err == nil {
		t.Error("expected error after DeleteChunk, got nil")
	}
}

func TestDeleteChunksRemovesAll(t *testing.T) {
	ctx := context.Background()
	ls := localstorage.New(t.TempDir())
	const total = 3

	for i := range total {
		if err := ls.StoreChunk(ctx, "multi", i, []byte("data")); err != nil {
			t.Fatal(err)
		}
	}
	if err := ls.DeleteChunks(ctx, "multi"); err != nil {
		t.Fatalf("DeleteChunks: %v", err)
	}
	for i := range total {
		if _, err := ls.GetChunk(ctx, "multi", i); err == nil {
			t.Errorf("chunk %d should be deleted", i)
		}
	}
}

func TestAssembleFileWritesContent(t *testing.T) {
	ctx := context.Background()
	ls := localstorage.New(t.TempDir())

	if err := ls.AssembleFile(ctx, "sess", 1, strings.NewReader("assembled content")); err != nil {
		t.Fatalf("AssembleFile: %v", err)
	}
}

func TestAssembleFileFromPipe(t *testing.T) {
	ctx := context.Background()
	ls := localstorage.New(t.TempDir())
	content := "hello from assembler"

	pr, pw := io.Pipe()
	go func() {
		pw.Write([]byte(content))
		pw.Close()
	}()

	if err := ls.AssembleFile(ctx, "sess", 1, pr); err != nil {
		t.Fatalf("AssembleFile via pipe: %v", err)
	}
}
