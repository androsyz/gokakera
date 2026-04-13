package gokakera

import "time"

// Config holds all configuration for a Kakera instance.
// Use the With* option functions to set individual fields.
type Config struct {
	ChunkSize      int64
	MaxFileSize    int64
	SessionTTL     time.Duration
	MaxConcurrency int
	ChecksumType   string
	Storage        Storage
	SessionStore   SessionStore
	OnProgress     ProgressFunc
}

// Option is a functional option that configures a Kakera instance.
type Option func(*Config)

// WithChunkSize sets the maximum size of each chunk in bytes.
func WithChunkSize(size int64) Option {
	return func(c *Config) {
		c.ChunkSize = size
	}
}

// WithMaxFileSize sets the maximum total file size accepted by CreateSession.
func WithMaxFileSize(size int64) Option {
	return func(c *Config) {
		c.MaxFileSize = size
	}
}

// WithSessionTTL sets how long a session remains valid after creation.
func WithSessionTTL(ttl time.Duration) Option {
	return func(c *Config) {
		c.SessionTTL = ttl
	}
}

// WithMaxConcurrency sets the maximum number of concurrent chunk uploads allowed.
func WithMaxConcurrency(n int) Option {
	return func(c *Config) {
		c.MaxConcurrency = n
	}
}

// WithChecksumType sets the checksum algorithm used to validate chunks.
// Supported values: "crc32" (default), "sha256".
func WithChecksumType(t string) Option {
	return func(c *Config) {
		c.ChecksumType = t
	}
}

// WithStorage sets the Storage backend used to persist and retrieve chunks.
func WithStorage(s Storage) Option {
	return func(c *Config) {
		c.Storage = s
	}
}

// WithSessionStore sets the SessionStore backend used to persist upload sessions.
func WithSessionStore(s SessionStore) Option {
	return func(c *Config) {
		c.SessionStore = s
	}
}

// WithOnProgress registers a callback that is invoked after each successful chunk upload.
func WithOnProgress(fn ProgressFunc) Option {
	return func(c *Config) {
		c.OnProgress = fn
	}
}
