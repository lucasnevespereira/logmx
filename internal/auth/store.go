package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Store struct {
	path   string
	tokens map[string]string
}

func NewStore() *Store {
	home, _ := os.UserHomeDir()
	return &Store{
		path:   filepath.Join(home, ".config", "logmx", "auth.json"),
		tokens: make(map[string]string),
	}
}

func (s *Store) Load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &s.tokens)
}

func (s *Store) Save() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.tokens, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

func (s *Store) Set(provider, token string) {
	s.tokens[provider] = token
}

func (s *Store) Get(provider string) (string, error) {
	t, ok := s.tokens[provider]
	if !ok || t == "" {
		return "", fmt.Errorf("not authenticated with %s — run: logmx auth %s", provider, provider)
	}
	return t, nil
}

func (s *Store) Delete(provider string) {
	delete(s.tokens, provider)
}

func (s *Store) Providers() []string {
	var providers []string
	for p := range s.tokens {
		providers = append(providers, p)
	}
	return providers
}
