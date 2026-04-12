package memorystore

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/androsyz/gokakera"
)

type MemoryStore struct {
	sessions map[string]*gokakera.Session
	mu       sync.RWMutex
}

func New() *MemoryStore {
	return &MemoryStore{
		sessions: make(map[string]*gokakera.Session),
	}
}

func (m *MemoryStore) Create(ctx context.Context, session *gokakera.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[session.ID] = session
	return nil
}

func (m *MemoryStore) Get(ctx context.Context, sessionID string) (*gokakera.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

func (m *MemoryStore) Update(ctx context.Context, session *gokakera.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sessions[session.ID]; !ok {
		return fmt.Errorf("session not found: %s", session.ID)
	}

	m.sessions[session.ID] = session
	return nil
}

func (m *MemoryStore) Delete(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, sessionID)
	return nil
}

func (m *MemoryStore) ListExpired(ctx context.Context) ([]*gokakera.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var expired []*gokakera.Session
	now := time.Now()

	for _, session := range m.sessions {
		if now.After(session.ExpiresAt) {
			expired = append(expired, session)
		}
	}

	return expired, nil
}
