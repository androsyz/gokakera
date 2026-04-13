package gokakera

import (
	"context"
	"fmt"
	"io"
	"log"
)

// AssembleFile streams all stored chunks in order into the final file via the Storage backend.
// It transitions the session through "assembling" → "completed" (or "failed" on error).
// Chunk data is cleaned up from storage after a successful assembly.
func (k *Kakera) AssembleFile(ctx context.Context, sessionID string) error {
	session, err := k.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if len(session.ReceivedChunks) != session.TotalChunks {
		return fmt.Errorf("upload not complete: received %d of %d chunks", len(session.ReceivedChunks), session.TotalChunks)
	}

	session.Status = StatusAssembling
	if err := k.config.SessionStore.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		for i := 0; i < session.TotalChunks; i++ {
			chunk, err := k.config.Storage.GetChunk(ctx, sessionID, i)
			if err != nil {
				pw.CloseWithError(fmt.Errorf("failed to get chunk %d: %w", i, err))
				return
			}

			if _, err := pw.Write(chunk); err != nil {
				pw.CloseWithError(err)
				return
			}
		}
	}()

	if err := k.config.Storage.AssembleFile(ctx, sessionID, session.TotalChunks, pr); err != nil {
		session.Status = StatusFailed
		if updateErr := k.config.SessionStore.Update(ctx, session); updateErr != nil {
			log.Printf("gokakera: failed to mark session %s as failed: %v", sessionID, updateErr)
		}
		return fmt.Errorf("failed to assemble file: %w", err)
	}

	if err := k.config.Storage.DeleteChunks(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to cleanup chunks: %w", err)
	}

	session.Status = StatusCompleted
	if err := k.config.SessionStore.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	return nil
}
