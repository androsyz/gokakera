package gokakera

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (k *Kakera) CreateSession(ctx context.Context, filename string, totalSize int64) (*Session, error) {
	if totalSize <= 0 {
		return nil, fmt.Errorf("invalid file size")
	}

	if totalSize > k.config.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed suze")
	}

	totalChunks := int(totalSize / k.config.ChunkSize)
	if totalSize%k.config.ChunkSize != 0 {
		totalChunks++
	}

	session := &Session{
		ID:             generateID(),
		Filename:       filename,
		TotalSize:      totalSize,
		TotalChunks:    totalChunks,
		ReceivedChunks: make(map[int]bool),
		Status:         "uploading",
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(k.config.SessionTTL),
	}

	if err := k.config.SessionStore.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

func (k *Kakera) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	session, err := k.config.SessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session has expired")
	}

	return session, nil
}
