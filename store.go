package gokakera

import "context"

// SessionStore is the interface that session persistence backends must implement.
type SessionStore interface {
	// Create persists a new session.
	Create(ctx context.Context, session *Session) error
	// Get retrieves a session by its ID.
	Get(ctx context.Context, sessionID string) (*Session, error)
	// Update overwrites an existing session with updated state.
	Update(ctx context.Context, session *Session) error
	// Delete removes a session from the store.
	Delete(ctx context.Context, sessionID string) error
	// ListExpired returns all sessions whose ExpiresAt is in the past.
	ListExpired(ctx context.Context) ([]*Session, error)
}
