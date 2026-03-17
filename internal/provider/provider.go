package provider

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/lucasnevespereira/logmx/internal/log"
)

// Connector fetches or streams log entries from a single source.
// Start blocks until the source is exhausted or the context is cancelled.
type Connector interface {
	Name() string
	Start(ctx context.Context, ch chan<- log.LogEntry) error
}

// Send writes an entry to the channel or returns if the context is cancelled.
func Send(ctx context.Context, ch chan<- log.LogEntry, e log.LogEntry) {
	select {
	case ch <- e:
	case <-ctx.Done():
	}
}

// Dependency describes an external CLI required by a provider.
type Dependency struct {
	Name       string
	Binary     string
	InstallCmd string
}

var ProviderDeps = map[string]Dependency{
	"vercel": {
		Name:       "Vercel CLI",
		Binary:     "vercel",
		InstallCmd: "npm i -g vercel",
	},
	"railway": {
		Name:       "Railway CLI",
		Binary:     "railway",
		InstallCmd: "bash -c 'curl -fsSL cli.new | bash'",
	},
}

func CheckDep(prov string) error {
	dep, ok := ProviderDeps[prov]
	if !ok {
		return nil
	}

	_, err := exec.LookPath(dep.Binary)
	if err != nil {
		return fmt.Errorf("%s not found — install it with: %s", dep.Name, dep.InstallCmd)
	}
	return nil
}

func MissingDeps(providers []string) []Dependency {
	var missing []Dependency
	seen := make(map[string]bool)
	for _, p := range providers {
		if seen[p] {
			continue
		}
		seen[p] = true

		dep, ok := ProviderDeps[p]
		if !ok {
			continue
		}
		if _, err := exec.LookPath(dep.Binary); err != nil {
			missing = append(missing, dep)
		}
	}
	return missing
}
