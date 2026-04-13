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
	if _, err := rand.Read(b); err != nil {
		panic("gokakera: failed to generate random ID: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// CreateSession initialises a new upload session for a file with the given name and total size.
// It returns an error if totalSize is non-positive or exceeds the configured MaxFileSize.
func (k *Kakera) CreateSession(ctx context.Context, filename string, totalSize int64) (*Session, error) {
	if totalSize <= 0 {
		return nil, fmt.Errorf("invalid file size")
	}

	if totalSize > k.config.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size")
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
		Status:         StatusUploading,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(k.config.SessionTTL),
	}

	if err := k.config.SessionStore.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetSession retrieves an active session by ID.
// It returns an error if the session does not exist or has expired.
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
