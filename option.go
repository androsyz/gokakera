package gokakera

import "time"

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

type Option func(*Config)

func WithChunkSize(size int64) Option {
	return func(c *Config) {
		c.ChunkSize = size
	}
}

func WithMaxFileSize(size int64) Option {
	return func(c *Config) {
		c.MaxFileSize = size
	}
}

func WithSessionTTL(ttl time.Duration) Option {
	return func(c *Config) {
		c.SessionTTL = ttl
	}
}

func WithMaxConcurrency(n int) Option {
	return func(c *Config) {
		c.MaxConcurrency = n
	}
}

func WithChecksumType(t string) Option {
	return func(c *Config) {
		c.ChecksumType = t
	}
}

func WithStorage(s Storage) Option {
	return func(c *Config) {
		c.Storage = s
	}
}

func WithSessionStore(s SessionStore) Option {
	return func(c *Config) {
		c.SessionStore = s
	}
}

func WithOnProgress(fn ProgressFunc) Option {
	return func(c *Config) {
		c.OnProgress = fn
	}
}
