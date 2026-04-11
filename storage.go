package gokakera

import (
	"context"
	"io"
)

type Storage interface {
	StoreChunk(ctx context.Context, sessionID string, index int, data []byte) error
	GetChunk(ctx context.Context, sessionID string, index int) ([]byte, error)
	DeleteChunk(ctx context.Context, sessionID string, index int) error
	DeleteChunks(ctx context.Context, sessionID string) error
	AssembleFile(ctx context.Context, sessionID string, totalChunks int, src io.Reader) error
}
