package gokakera

import (
	"context"
	"fmt"
)

// UploadChunk stores a single chunk for an active session.
// The call is idempotent: uploading a chunk that was already received succeeds without re-storing it.
// It returns an error if the session is expired, the index is out of range, or the checksum does not match.
func (k *Kakera) UploadChunk(ctx context.Context, sessionID string, index int, data []byte, checksum string) error {
	mu := k.sessionMu(sessionID)
	mu.Lock()
	defer mu.Unlock()

	session, err := k.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.Status != StatusUploading {
		return fmt.Errorf("session is not accepting chunks")
	}

	if index < 0 || index >= session.TotalChunks {
		return fmt.Errorf("chunk index %d out of range (0-%d)", index, session.TotalChunks-1)
	}

	if session.ReceivedChunks[index] {
		return nil // already received, idempotent
	}

	computed, err := k.computeChecksum(data)
	if err != nil {
		return fmt.Errorf("failed to compute checksum %w", err)
	}

	if computed != checksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", checksum, computed)
	}

	if err := k.config.Storage.StoreChunk(ctx, sessionID, index, data); err != nil {
		return fmt.Errorf("failed to store chunk: %w", err)
	}

	session.ReceivedChunks[index] = true
	if err := k.config.SessionStore.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	if k.config.OnProgress != nil {
		k.config.OnProgress(Progress{
			SessionID:      sessionID,
			ChunkIndex:     index,
			ChunksReceived: len(session.ReceivedChunks),
			TotalChunks:    session.TotalChunks,
			BytesReceived:  min(int64(len(session.ReceivedChunks))*k.config.ChunkSize, session.TotalSize),
			TotalSize:      session.TotalSize,
			Percentage:     float64(len(session.ReceivedChunks)) / float64(session.TotalChunks) * 100,
		})
	}

	return nil
}

// IsUploadComplete reports whether all expected chunks have been received for the session.
func (k *Kakera) IsUploadComplete(ctx context.Context, sessionID string) (bool, error) {
	session, err := k.GetSession(ctx, sessionID)
	if err != nil {
		return false, err
	}

	return len(session.ReceivedChunks) == session.TotalChunks, nil
}
