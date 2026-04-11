package gokakera

import (
	"context"
	"log"
	"time"
)

func (k *Kakera) StartCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				k.cleanup(ctx)
			}
		}
	}()
}

func (k *Kakera) cleanup(ctx context.Context) {
	sessions, err := k.config.SessionStore.ListExpired(ctx)
	if err != nil {
		log.Printf("gokakera: failed to list expired sessions: %v", err)
		return
	}

	for _, session := range sessions {
		if err := k.config.Storage.DeleteChunks(ctx, session.ID); err != nil {
			log.Printf("gokakera: failed to delete chunks for session %s: %v", session.ID, err)
			continue
		}

		if err := k.config.SessionStore.Delete(ctx, session.ID); err != nil {
			log.Printf("gokakera: failed to delete session %s: %v", session.ID, err)
		}
	}
}
