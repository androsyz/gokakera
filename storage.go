package gokakera

import (
	"context"
	"io"
)

// Storage is the interface that chunk and file storage backends must implement.
// Implementations are responsible for persisting individual chunks and assembling
// them into the final file when the upload is complete.
type Storage interface {
	// StoreChunk persists a single chunk identified by sessionID and index.
	StoreChunk(ctx context.Context, sessionID string, index int, data []byte) error
	// GetChunk retrieves a previously stored chunk.
	GetChunk(ctx context.Context, sessionID string, index int) ([]byte, error)
	// DeleteChunk removes a single chunk from storage.
	DeleteChunk(ctx context.Context, sessionID string, index int) error
	// DeleteChunks removes all chunks belonging to a session.
	DeleteChunks(ctx context.Context, sessionID string) error
	// AssembleFile writes the ordered chunk stream from src into the final file.
	AssembleFile(ctx context.Context, sessionID string, totalChunks int, src io.Reader) error
}
