package localstorage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	basePath string
}

func New(basePath string) *LocalStorage {
	return &LocalStorage{
		basePath: basePath,
	}
}

func (l *LocalStorage) StoreChunk(ctx context.Context, sessionID string, index int, data []byte) error {
	dir := filepath.Join(l.basePath, "chunks", sessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create chunk directory: %w", err)
	}

	path := filepath.Join(dir, fmt.Sprintf("%d", index))
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write chunk: %w", err)
	}

	return nil
}

func (l *LocalStorage) GetChunk(ctx context.Context, sessionID string, index int) ([]byte, error) {
	path := filepath.Join(l.basePath, "chunks", sessionID, fmt.Sprintf("%d", index))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read chunk: %w", err)
	}

	return data, nil
}

func (l *LocalStorage) DeleteChunk(ctx context.Context, sessionID string, index int) error {
	path := filepath.Join(l.basePath, "chunks", sessionID, fmt.Sprintf("%d", index))
	return os.Remove(path)
}

func (l *LocalStorage) DeleteChunks(ctx context.Context, sessionID string) error {
	path := filepath.Join(l.basePath, "chunks", sessionID)
	return os.RemoveAll(path)
}

func (l *LocalStorage) AssembleFile(ctx context.Context, sessionID string, totalChunks int, src io.Reader) error {
	dir := filepath.Join(l.basePath, "files")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create files directory: %w", err)
	}

	path := filepath.Join(dir, sessionID)
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, src); err != nil {
		return fmt.Errorf("failed to write assembled file: %w", err)
	}

	return nil
}
