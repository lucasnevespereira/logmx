package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type AuthStore struct {
	Tokens map[string]string `json:"tokens"`
}

func DefaultAuthPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "logmx", "auth.json")
}

func LoadAuth(path string) (*AuthStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &AuthStore{Tokens: make(map[string]string)}, nil
		}
		return nil, fmt.Errorf("reading auth: %w", err)
	}

	var s AuthStore
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing auth: %w", err)
	}
	if s.Tokens == nil {
		s.Tokens = make(map[string]string)
	}
	return &s, nil
}

func SaveAuth(path string, s *AuthStore) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
