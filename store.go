package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type tokenStore struct {
	path string
	mu   sync.RWMutex
	data storedToken
}

type storedToken struct {
	Token     string    `json:"token"`
	UserName  string    `json:"user_name"`
	UpdatedAt time.Time `json:"updated_at"`
}

func newTokenStore(dir string) (*tokenStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	store := &tokenStore{
		path: filepath.Join(dir, "token.json"),
	}

	raw, err := os.ReadFile(store.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return store, nil
		}
		return nil, fmt.Errorf("read token file: %w", err)
	}

	if len(raw) == 0 {
		return store, nil
	}

	if err := json.Unmarshal(raw, &store.data); err != nil {
		return nil, fmt.Errorf("parse token file: %w", err)
	}

	return store, nil
}

func (s *tokenStore) Get() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Token
}

func (s *tokenStore) Set(token, userName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = storedToken{
		Token:     token,
		UserName:  userName,
		UpdatedAt: time.Now().UTC(),
	}

	payload, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal token: %w", err)
	}

	if err := os.WriteFile(s.path, payload, 0o600); err != nil {
		return fmt.Errorf("write token file: %w", err)
	}

	return nil
}

func (s *tokenStore) Status() storedToken {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}
