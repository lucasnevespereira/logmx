package railway

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/lucasnevespereira/logmx/internal/log"
	"github.com/lucasnevespereira/logmx/internal/provider"
)

type Connector struct {
	Source    string
	ProjectID string
	ServiceID string
	Token     string
	Limit     int
	Follow    bool
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
	args := []string{"logs", "--json"}
	if c.ServiceID != "" {
		args = append(args, "-s", c.ServiceID)
	}
	if c.Follow {
		args = append(args, "--follow")
	}

	cmd := exec.CommandContext(ctx, "railway", args...)
	if c.Token != "" {
		cmd.Env = append(os.Environ(), "RAILWAY_TOKEN="+c.Token)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

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
