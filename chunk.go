package gokakera

import (
	"context"
	"fmt"
)

func (k *Kakera) UploadChunk(ctx context.Context, sessionID string, index int, data []byte, checksum string) error {
	session, err := k.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.Status != "uploading" {
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
			BytesReceived:  int64(len(session.ReceivedChunks)) * k.config.ChunkSize,
			TotalSize:      session.TotalSize,
			Percentage:     float64(len(session.ReceivedChunks)) / float64(session.TotalChunks) * 100,
		})
	}

	return nil
}

func (k *Kakera) IsUploadComplete(ctx context.Context, sessionID string) (bool, error) {
	session, err := k.GetSession(ctx, sessionID)
	if err != nil {
		return false, err
	}

	return len(session.ReceivedChunks) == session.TotalChunks, nil
}
