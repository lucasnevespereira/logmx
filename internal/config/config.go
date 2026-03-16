package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func IsNotExist(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}

type Source struct {
	Name     string `yaml:"name"`
	Provider string `yaml:"provider"`
	Project  string `yaml:"project,omitempty"`
	Service  string `yaml:"service,omitempty"`
}

type Config struct {
	Sources []Source `yaml:"sources"`
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "logmx", "config.yaml")
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

func Init(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	example := `# logmx configuration
# Add your log sources here.
#
# sources:
#   - name: api
#     provider: vercel
#     project: prj_abc123
#   - name: frontend
#     provider: vercel
#     project: prj_def456
#   - name: worker
#     provider: railway
#     service: srv_xxx

sources: []
`
	return os.WriteFile(path, []byte(example), 0o644)
}
