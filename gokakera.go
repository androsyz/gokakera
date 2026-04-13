// Package gokakera provides resumable chunked file uploads with session management,
// checksum validation, progress tracking, and pluggable storage backends.
package gokakera

import "time"

const (
	DefaultChunkSize      = 5 * 1024 * 1024         // 5MB
	DefaultMaxFileSize    = 10 * 1024 * 1024 * 1024 // 10GB
	DefaultSessionTTL     = 24 * time.Hour
	DefaultMaxConcurrency = 5
	DefaultChecksumType   = "crc32"
)

// Session status values.
const (
	StatusUploading  = "uploading"
	StatusAssembling = "assembling"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
)

// Kakera is the main entry point for the chunked upload library.
// It manages upload sessions, validates chunks, and coordinates storage.
type Kakera struct {
	config *Config
}

// New creates a Kakera instance with the given options.
// It applies all options over a set of sensible defaults:
// 5 MB chunk size, 10 GB max file, 24 h session TTL, crc32 checksums.
func New(opts ...Option) *Kakera {
	cfg := &Config{
		ChunkSize:      DefaultChunkSize,
		MaxFileSize:    DefaultMaxFileSize,
		SessionTTL:     DefaultSessionTTL,
		MaxConcurrency: DefaultMaxConcurrency,
		ChecksumType:   DefaultChecksumType,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Kakera{
		config: cfg,
	}
}
