package vercel

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/lucasnevespereira/logmx/internal/connectors"
	"github.com/lucasnevespereira/logmx/internal/models"
)

type Connector struct {
	Source    string
	ProjectID string
	Token     string
	Limit     int
}

func (c *Connector) Name() string {
	return c.Source
}

type vercelLogLine struct {
	Message        string `json:"message"`
	Timestamp      int64  `json:"timestamp"`
	Level          string `json:"level"`
	RequestPath    string `json:"requestPath"`
	Path           string `json:"path"`
	StatusCode     int    `json:"statusCode"`
	ResponseStatus int    `json:"responseStatusCode"`
}

func (c *Connector) Start(ctx context.Context, ch chan<- models.LogEntry) error {
	go func() {
		limit := c.Limit
		if limit == 0 {
			limit = 100
		}

		args := []string{"logs", "--json", "--project", c.ProjectID, "--limit", fmt.Sprintf("%d", limit)}
		if c.Token != "" {
			args = append(args, "--token", c.Token)
		}

		cmd := exec.CommandContext(ctx, "vercel", args...)
		out, err := cmd.Output()
		if err != nil {
			connectors.Send(ctx, ch, models.LogEntry{
				Timestamp: time.Now().UTC(),
				Source:    c.Source,
				Level:     models.LevelError,
				Message:   fmt.Sprintf("vercel logs: %v", err),
			})
			return
		}

		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var log vercelLogLine
			if err := json.Unmarshal(line, &log); err != nil {
				continue
			}

			msg := log.Message
			path := log.RequestPath
			if path == "" {
				path = log.Path
			}
			statusCode := log.ResponseStatus
			if statusCode == 0 {
				statusCode = log.StatusCode
			}
			if msg == "" && path != "" {
				msg = fmt.Sprintf("%s %d", path, statusCode)
			}
			if msg == "" {
				continue
			}

			ts := time.UnixMilli(log.Timestamp)
			if log.Timestamp == 0 {
				ts = time.Now().UTC()
			}

			connectors.Send(ctx, ch, models.LogEntry{
				Timestamp: ts,
				Source:    c.Source,
				Level:     parseLevel(log.Level, statusCode),
				Message:   msg,
			})
		}
	}()
	return nil
}

func parseLevel(level string, statusCode int) models.LogLevel {
	switch strings.ToLower(level) {
	case "error", "fatal":
		return models.LevelError
	case "warning", "warn":
		return models.LevelWarn
	case "debug":
		return models.LevelDebug
	}

	if statusCode >= 500 {
		return models.LevelError
	}
	if statusCode >= 400 {
		return models.LevelWarn
	}

	return models.LevelInfo
}
