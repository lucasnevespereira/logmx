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

	"github.com/lucasnevespereira/logmx/internal/connectors"
	"github.com/lucasnevespereira/logmx/internal/models"
)

type Connector struct {
	Source    string
	ProjectID string
	ServiceID string
	Token     string
	Limit     int
}

func (c *Connector) Name() string {
	return c.Source
}

type railwayLogLine struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Severity  string `json:"severity"`
}

func (c *Connector) Start(ctx context.Context, ch chan<- models.LogEntry) error {
	go func() {
		args := []string{"logs", "--json"}
		if c.ServiceID != "" {
			args = append(args, "-s", c.ServiceID)
		}

		cmd := exec.CommandContext(ctx, "railway", args...)
		if c.Token != "" {
			cmd.Env = append(os.Environ(), "RAILWAY_TOKEN="+c.Token)
		}

		out, err := cmd.Output()
		if err != nil {
			connectors.Send(ctx, ch, models.LogEntry{
				Timestamp: time.Now().UTC(),
				Source:    c.Source,
				Level:     models.LevelError,
				Message:   fmt.Sprintf("railway logs: %v", err),
			})
			return
		}

		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var log railwayLogLine
			if err := json.Unmarshal(line, &log); err != nil {
				continue
			}

			if log.Message == "" {
				continue
			}

			ts, _ := time.Parse(time.RFC3339Nano, log.Timestamp)
			if ts.IsZero() {
				ts = time.Now().UTC()
			}

			connectors.Send(ctx, ch, models.LogEntry{
				Timestamp: ts,
				Source:    c.Source,
				Level:     parseLevel(log.Severity, log.Message),
				Message:   log.Message,
			})
		}
	}()
	return nil
}

func parseLevel(severity, text string) models.LogLevel {
	switch strings.ToLower(severity) {
	case "error", "err", "critical", "fatal":
		return models.LevelError
	case "warn", "warning":
		return models.LevelWarn
	case "debug":
		return models.LevelDebug
	case "info":
		return models.LevelInfo
	}

	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "error") || strings.Contains(lower, "fatal"):
		return models.LevelError
	case strings.Contains(lower, "warn"):
		return models.LevelWarn
	default:
		return models.LevelInfo
	}
}
