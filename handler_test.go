package gokakera_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/androsyz/gokakera"
	"github.com/androsyz/gokakera/localstorage"
	"github.com/androsyz/gokakera/memorystore"
)

const (
	testFilename      = "test.txt"
	uploadSessionPath = "/upload/session"
	headerXChecksum   = "X-Checksum"
	errExpected400    = "expected 400, got %d"
	errExpected200    = "expected 200, got %d: %s"
)

func setupServer(t *testing.T) (*gokakera.Kakera, *http.ServeMux) {
	t.Helper()
	k := gokakera.New(
		gokakera.WithChunkSize(10),
		gokakera.WithMaxFileSize(1024),
		gokakera.WithSessionTTL(1*time.Hour),
		gokakera.WithChecksumType("crc32"),
		gokakera.WithStorage(localstorage.New(t.TempDir())),
		gokakera.WithSessionStore(memorystore.New()),
	)
	mux := http.NewServeMux()
	k.RegisterRoutes(mux)
	return k, mux
}

func TestHandleCreateSession(t *testing.T) {
	_, mux := setupServer(t)

	t.Run("valid request returns 201 with session fields", func(t *testing.T) {
		body, _ := json.Marshal(map[string]any{"filename": testFilename, "total_size": 25})
		req := httptest.NewRequest(http.MethodPost, uploadSessionPath, bytes.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var resp map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if _, ok := resp["session_id"]; !ok {
			t.Error("response missing session_id")
		}
		if _, ok := resp["total_chunks"]; !ok {
			t.Error("response missing total_chunks")
		}
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, uploadSessionPath, bytes.NewReader([]byte("not-json")))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf(errExpected400, w.Code)
		}
	})

	t.Run("zero total_size returns 400", func(t *testing.T) {
		body, _ := json.Marshal(map[string]any{"filename": testFilename, "total_size": 0})
		req := httptest.NewRequest(http.MethodPost, uploadSessionPath, bytes.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf(errExpected400, w.Code)
		}
	})

	t.Run("size exceeds max returns 400", func(t *testing.T) {
		body, _ := json.Marshal(map[string]any{"filename": testFilename, "total_size": 9999})
		req := httptest.NewRequest(http.MethodPost, uploadSessionPath, bytes.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf(errExpected400, w.Code)
		}
	})
}

func TestHandleUploadChunk(t *testing.T) {
	k, mux := setupServer(t)
	ctx := context.Background()

	session, err := k.CreateSession(ctx, testFilename, 25)
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("helloworld")
	checksum := computeChecksum(t, data)

	t.Run("valid chunk returns 200", func(t *testing.T) {
		url := fmt.Sprintf("/upload/%s/chunk/0", session.ID)
		req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(data))
		req.Header.Set(headerXChecksum, checksum)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf(errExpected200, w.Code, w.Body.String())
		}
	})

	t.Run("missing X-Checksum header returns 400", func(t *testing.T) {
		url := fmt.Sprintf("/upload/%s/chunk/1", session.ID)
		req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(data))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf(errExpected400, w.Code)
		}
	})

	t.Run("non-numeric index returns 400", func(t *testing.T) {
		url := fmt.Sprintf("/upload/%s/chunk/abc", session.ID)
		req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(data))
		req.Header.Set(headerXChecksum, checksum)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf(errExpected400, w.Code)
		}
	})

	t.Run("bad checksum returns 400", func(t *testing.T) {
		url := fmt.Sprintf("/upload/%s/chunk/1", session.ID)
		req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(data))
		req.Header.Set(headerXChecksum, "badchecksum")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf(errExpected400, w.Code)
		}
	})
}

func TestHandleStatus(t *testing.T) {
	k, mux := setupServer(t)
	ctx := context.Background()

	session, err := k.CreateSession(ctx, testFilename, 10)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("existing session returns 200 with status", func(t *testing.T) {
		url := fmt.Sprintf("/upload/%s/status", session.ID)
		req := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf(errExpected200, w.Code, w.Body.String())
		}
		var resp map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp["status"] != gokakera.StatusUploading {
			t.Errorf("expected status %q, got %v", gokakera.StatusUploading, resp["status"])
		}
	})

	t.Run("unknown session returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/upload/doesnotexist/status", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestHandleComplete(t *testing.T) {
	k, mux := setupServer(t)
	ctx := context.Background()

	session, err := k.CreateSession(ctx, testFilename, 10)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("incomplete upload returns 400", func(t *testing.T) {
		url := fmt.Sprintf("/upload/%s/complete", session.ID)
		req := httptest.NewRequest(http.MethodPost, url, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf(errExpected400, w.Code)
		}
	})

	// Upload the single chunk, then complete.
	data := []byte("helloworld")
	checksum := computeChecksum(t, data)
	if err := k.UploadChunk(ctx, session.ID, 0, data, checksum); err != nil {
		t.Fatal(err)
	}

	t.Run("fully uploaded session returns 200", func(t *testing.T) {
		url := fmt.Sprintf("/upload/%s/complete", session.ID)
		req := httptest.NewRequest(http.MethodPost, url, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf(errExpected200, w.Code, w.Body.String())
		}
	})
}
