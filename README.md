# gokakera

Gokakera (欠片) — Go library for reliable chunked file uploads.

## Features

- Configurable chunk size
- Checksum validation per chunk (CRC32, SHA256)
- Parallel uploads with concurrency control
- Resumable uploads
- Idempotent chunk uploads (safe retries)
- Upload session management with expiry
- Auto-cleanup of expired sessions
- Pluggable storage backends (local, GCS, S3)
- Pluggable session stores (memory, Redis)
- Progress tracking via callbacks
- Middleware-friendly — works with net/http, Chi, Gin, or any router

## Installation

```bash
go get github.com/androsyz/gokakera
```

## Quick Start

```go
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/androsyz/gokakera"
	"github.com/androsyz/gokakera/localstorage"
	"github.com/androsyz/gokakera/memorystore"
)

func main() {
	kakera := gokakera.New(
		gokakera.WithChunkSize(5*1024*1024),
		gokakera.WithMaxFileSize(10*1024*1024*1024),
		gokakera.WithSessionTTL(24*time.Hour),
		gokakera.WithChecksumType("crc32"),
		gokakera.WithStorage(localstorage.New("./uploads")),
		gokakera.WithSessionStore(memorystore.New()),
	)

	kakera.StartCleanup(context.Background(), 5*time.Minute)

	mux := http.NewServeMux()
	kakera.RegisterRoutes(mux)

	log.Println("server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/upload/session` | Create upload session |
| PUT | `/upload/{sessionID}/chunk/{index}` | Upload a chunk |
| POST | `/upload/{sessionID}/complete` | Assemble final file |
| GET | `/upload/{sessionID}/status` | Check upload progress |

## Usage

### Create Upload Session

```bash
curl -X POST http://localhost:8080/upload/session \
  -H "Content-Type: application/json" \
  -d '{"filename": "video.mp4", "total_size": 15728640}'
```

Response:

```json
{
  "session_id": "a1b2c3d4...",
  "total_chunks": 3,
  "chunk_size": 5242880,
  "expires_at": "2026-04-14T10:00:00Z"
}
```

### Upload Chunk

```bash
curl -X PUT http://localhost:8080/upload/{sessionID}/chunk/0 \
  -H "X-Checksum: a1b2c3d4" \
  --data-binary @chunk0.bin
```

### Check Status

```bash
curl http://localhost:8080/upload/{sessionID}/status
```

Response:

```json
{
  "session_id": "a1b2c3d4...",
  "filename": "video.mp4",
  "status": "uploading",
  "total_chunks": 3,
  "received_chunks": 2,
  "percentage": 66.7,
  "expires_at": "2026-04-14T10:00:00Z"
}
```

### Complete Upload

```bash
curl -X POST http://localhost:8080/upload/{sessionID}/complete
```

## Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithChunkSize(n)` | 5 MB | Size of each chunk in bytes |
| `WithMaxFileSize(n)` | 10 GB | Maximum allowed file size |
| `WithSessionTTL(d)` | 24 hours | Session expiry duration |
| `WithMaxConcurrency(n)` | 5 | Max parallel chunk uploads |
| `WithChecksumType(t)` | crc32 | Checksum algorithm (crc32, sha256) |
| `WithStorage(s)` | - | Storage backend |
| `WithSessionStore(s)` | - | Session store |
| `WithOnProgress(fn)` | - | Progress callback |

## Storage Backends

### Local Storage

```go
import "github.com/androsyz/gokakera/localstorage"

storage := localstorage.New("./uploads")
```

### GCS and S3

Coming soon.

## Session Stores

### Memory Store

```go
import "github.com/androsyz/gokakera/memorystore"

store := memorystore.New()
```

### Redis Store

Coming soon.

## How It Works

1. Client creates an upload session with filename and total file size
2. Server returns session ID and chunk details
3. Client splits the file into chunks and sends each with a checksum
4. Server validates checksum, stores chunk, tracks progress
5. Client signals upload complete
6. Server assembles all chunks into the final file
7. Background worker cleans up expired sessions

## Running Tests

```bash
go test ./... -v
```

## License

MIT