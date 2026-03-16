package cli

import (
	"fmt"
	"os/exec"
)

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
		InstallCmd: "npm i -g @railway/cli",
	},
}

func CheckDep(provider string) error {
	dep, ok := ProviderDeps[provider]
	if !ok {
		return nil // no CLI dep for this provider
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
