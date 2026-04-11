package gokakera

import "time"

type Chunk struct {
	SessionID string
	Index     int
	Data      []byte
	Checksum  string
}

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
