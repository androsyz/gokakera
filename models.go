package gokakera

import "time"

// Chunk represents a single piece of a chunked upload.
// Reserved for future use by storage implementations that need to pass chunk metadata.
type Chunk struct {
	SessionID string
	Index     int
	Data      []byte
	Checksum  string
}

// Session tracks the state of a single chunked upload.
// ReceivedChunks maps chunk index to true once that chunk has been stored.
type Session struct {
	ID             string
	Filename       string
	TotalSize      int64
	TotalChunks    int
	ReceivedChunks map[int]bool
	Status         string
	CreatedAt      time.Time
	ExpiresAt      time.Time
}
