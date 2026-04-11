package gokakera

import (
	"context"
	"fmt"
	"io"
)

func (k *Kakera) AssembleFile(ctx context.Context, sessionID string) error {
	session, err := k.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if len(session.ReceivedChunks) != session.TotalChunks {
		return fmt.Errorf("upload not complete: received %d of %d chunks", len(session.ReceivedChunks), session.TotalChunks)
	}

	session.Status = "assembling"
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
		session.Status = "failed"
		k.config.SessionStore.Update(ctx, session)
		return fmt.Errorf("failed to assemble file: %w", err)
	}

	if err := k.config.Storage.DeleteChunks(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to cleanup chunks: %w", err)
	}

	session.Status = "completed"
	if err := k.config.SessionStore.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	return nil
}
