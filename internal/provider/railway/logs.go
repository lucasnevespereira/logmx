package railway

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/lucasnevespereira/logmx/internal/log"
	"github.com/lucasnevespereira/logmx/internal/provider"
)

// Connector streams logs from a Railway service using the Railway CLI.
//
// Railway CLI requires a "linked" project to know which service to fetch logs from.
// Normally this is done via `railway link` which stores the link in ~/.railway/config.json
// keyed by the current working directory.
//
// Since logmx may tail multiple Railway sources in parallel, we can't use a single
// linked project. Instead, each Connector:
//  1. Creates a temporary directory
//  2. Writes a project link entry to ~/.railway/config.json keyed by that temp dir
//  3. Runs `railway logs --json` with cmd.Dir set to the temp dir
//  4. Cleans up the temp entry and directory on exit
//
// This lets each source resolve to its own project context without conflicts.
// A mutex (railwayCfgMu) prevents concurrent writes to the shared config file.
type Connector struct {
	Source        string
	ProjectID     string
	ServiceID     string
	EnvironmentID string
	Limit         int
	Follow        bool
}

func (c *Connector) Name() string {
	return c.Source
}

type logLine struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Severity  string `json:"severity"`
}

func (c *Connector) Start(ctx context.Context, ch chan<- log.LogEntry) error {
	tmpDir, err := os.MkdirTemp("", "logmx-railway-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Resolve symlinks so the path matches what Railway CLI stores
	// (e.g. macOS /var → /private/var).
	tmpDir, err = filepath.EvalSymlinks(tmpDir)
	if err != nil {
		return fmt.Errorf("resolving temp dir: %w", err)
	}

	if err := writeRailwayLink(tmpDir, c.ProjectID, c.ServiceID, c.EnvironmentID); err != nil {
		return fmt.Errorf("writing railway link: %w", err)
	}
	defer removeRailwayLink(tmpDir)

	args := []string{"logs", "--json"}
	if c.ServiceID != "" {
		args = append(args, "-s", c.ServiceID)
	}

	if !c.Follow {
		// Non-follow: fetch history and exit.
		args = append(args, "-n", fmt.Sprintf("%d", c.Limit))
		return c.runLogs(ctx, ch, tmpDir, args)
	}

	// Follow mode: Railway CLI may exit when a deployment is idle.
	// Retry with --since to only fetch new logs on reconnect.
	args = append(args, "--latest")
	first := true
	for {
		runArgs := args
		if !first {
			// On reconnect, only fetch logs from the last 30 seconds
			// to avoid duplicating the full history on each retry.
			runArgs = append(append([]string{}, args...), "--since", "30s")
		}
		first = false

		_ = c.runLogs(ctx, ch, tmpDir, runArgs)

		// If context is cancelled, stop retrying.
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(5 * time.Second):
			// Reconnect after a short delay.
		}
	}
}

// runLogs executes a single `railway logs` invocation and streams entries to ch.
func (c *Connector) runLogs(ctx context.Context, ch chan<- log.LogEntry, dir string, args []string) error {
	cmd := exec.CommandContext(ctx, "railway", args...)
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		provider.Send(ctx, ch, log.LogEntry{
			Timestamp: time.Now().UTC(),
			Source:    c.Source,
			Level:     log.LevelError,
			Message:   fmt.Sprintf("railway: %v", err),
		})
		return nil
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var l logLine
		if err := json.Unmarshal(line, &l); err != nil {
			continue
		}

		if l.Message == "" {
			continue
		}

		ts, _ := time.Parse(time.RFC3339Nano, l.Timestamp)
		if ts.IsZero() {
			ts = time.Now().UTC()
		}

		provider.Send(ctx, ch, log.LogEntry{
			Timestamp: ts,
			Source:    c.Source,
			Level:     parseLevel(l.Severity, l.Message),
			Message:   l.Message,
		})
	}

	_ = cmd.Wait()
	return nil
}

// railwayCfgMu guards concurrent reads/writes to ~/.railway/config.json
// when multiple Railway connectors start in parallel.
var railwayCfgMu sync.Mutex

// writeRailwayLink registers a temp directory as a linked project in the
// Railway CLI config so that `railway logs` run from that directory
// knows which project/environment/service to target.
func writeRailwayLink(dir, projectID, serviceID, environmentID string) error {
	railwayCfgMu.Lock()
	defer railwayCfgMu.Unlock()

	cfgPath := railwayConfigPath()

	data, _ := os.ReadFile(cfgPath)
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		cfg = map[string]any{}
	}

	projects, _ := cfg["projects"].(map[string]any)
	if projects == nil {
		projects = map[string]any{}
	}

	entry := map[string]any{
		"projectPath":     dir,
		"project":         projectID,
		"environment":     environmentID,
		"environmentName": "production",
	}
	if serviceID != "" {
		entry["service"] = serviceID
	}
	projects[dir] = entry
	cfg["projects"] = projects

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfgPath, out, 0o644)
}

// removeRailwayLink removes the temp directory entry from the Railway CLI config.
func removeRailwayLink(dir string) {
	railwayCfgMu.Lock()
	defer railwayCfgMu.Unlock()

	cfgPath := railwayConfigPath()

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		return
	}

	projects, _ := cfg["projects"].(map[string]any)
	if projects == nil {
		return
	}
	delete(projects, dir)
	cfg["projects"] = projects

	out, _ := json.MarshalIndent(cfg, "", "  ")
	_ = os.WriteFile(cfgPath, out, 0o644)
}

func railwayConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".railway", "config.json")
}

func parseLevel(severity, text string) log.LogLevel {
	switch strings.ToLower(severity) {
	case "error", "err", "critical", "fatal":
		return log.LevelError
	case "warn", "warning":
		return log.LevelWarn
	case "debug":
		return log.LevelDebug
	case "info":
		return log.LevelInfo
	}

	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "error") || strings.Contains(lower, "fatal"):
		return log.LevelError
	case strings.Contains(lower, "warn"):
		return log.LevelWarn
	default:
		return log.LevelInfo
	}
}
