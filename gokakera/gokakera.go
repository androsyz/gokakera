package gokakera

import "time"

const (
	DefaultChunkSize      = 5 * 1024 * 1024         // 5MB
	DefaultMaxFileSize    = 10 * 1024 * 1024 * 1024 // 10GB
	DefaultSessionTTL     = 24 * time.Hour
	DefaultMaxConcurrency = 5
	DefaultChecksumType   = "crc32"
)

type Kakera struct {
	config *Config
}

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
