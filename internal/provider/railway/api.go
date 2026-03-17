// Package railway wraps the Railway CLI to provide authentication,
// project discovery, and log streaming.
//
// Unlike Vercel (which uses API tokens), Railway authentication is
// delegated entirely to the Railway CLI. Users run `railway login`
// once and the CLI stores its own session in ~/.railway/config.json.
//
// Auth flow:
//   - logmx setup  → railway.Login()           (opens browser)
//   - logmx auth   → railway.LoginBrowserless() (terminal-only, for re-auth)
//
// Project discovery:
//   - railway list --json → returns projects, services, and environments
//
// Log streaming:
//   - See logs.go for the temp-dir linking strategy used to run
//     `railway logs` against specific projects without interactive `railway link`.
package railway

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type Project struct {
	ID           string
	Name         string
	Services     []Service
	Environments []Environment
}

type Service struct {
	ID   string
	Name string
}

type Environment struct {
	ID   string
	Name string
}

// CheckLogin verifies the user is logged in to Railway CLI via `railway whoami`.
func CheckLogin() (string, error) {
	out, err := exec.Command("railway", "whoami").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("not logged in — run 'railway login' first")
	}
	return string(out), nil
}

// Login runs `railway login` which opens the browser for authentication.
// Used during `logmx setup` for a smooth first-time experience.
func Login() error {
	cmd := exec.Command("railway", "login")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// LoginBrowserless runs `railway login --browserless` for terminal-only auth.
// Used by `logmx auth railway` for re-authentication or headless environments.
func LoginBrowserless() error {
	cmd := exec.Command("railway", "login", "--browserless")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ListProjects uses `railway list --json` to discover projects, services, and environments.
func ListProjects() ([]Project, error) {
	out, err := exec.Command("railway", "list", "--json").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("railway list: %s", string(out))
	}

	var raw []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Services struct {
			Edges []struct {
				Node struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"services"`
		Environments struct {
			Edges []struct {
				Node struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"environments"`
	}

	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parsing railway list: %w", err)
	}

	var projects []Project
	for _, r := range raw {
		p := Project{ID: r.ID, Name: r.Name}
		for _, se := range r.Services.Edges {
			p.Services = append(p.Services, Service{
				ID:   se.Node.ID,
				Name: se.Node.Name,
			})
		}
		for _, ee := range r.Environments.Edges {
			p.Environments = append(p.Environments, Environment{
				ID:   ee.Node.ID,
				Name: ee.Node.Name,
			})
		}
		projects = append(projects, p)
	}
	return projects, nil
}
