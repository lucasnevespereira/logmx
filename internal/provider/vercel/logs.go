package vercel

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/lucasnevespereira/logmx/internal/log"
	"github.com/lucasnevespereira/logmx/internal/provider"
)

type Connector struct {
	Source    string
	ProjectID string
	Token     string
	Limit     int
	Follow    bool
}

func (c *Connector) Name() string {
	return c.Source
}

type logLine struct {
	Message        string `json:"message"`
	Timestamp      int64  `json:"timestamp"`
	Level          string `json:"level"`
	RequestPath    string `json:"requestPath"`
	Path           string `json:"path"`
	StatusCode     int    `json:"statusCode"`
	ResponseStatus int    `json:"responseStatusCode"`
}

func (c *Connector) Start(ctx context.Context, ch chan<- log.LogEntry) error {
	limit := c.Limit
	if limit == 0 {
		limit = 100
	}

	args := []string{"logs", "--json", "--project", c.ProjectID, "--limit", fmt.Sprintf("%d", limit)}
	if c.Follow {
		args = append(args, "--follow")
	}
	if c.Token != "" {
		args = append(args, "--token", c.Token)
	}

	cmd := exec.CommandContext(ctx, "vercel", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		provider.Send(ctx, ch, log.LogEntry{
			Timestamp: time.Now().UTC(),
			Source:    c.Source,
			Level:     log.LevelError,
			Message:   fmt.Sprintf("vercel: %v", err),
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

		entry := parseEntry(c.Source, l)
		if entry.Message == "" {
			continue
		}

		provider.Send(ctx, ch, entry)
	}

	_ = cmd.Wait()
	return nil
}

func parseEntry(source string, l logLine) log.LogEntry {
	msg := l.Message
	path := l.RequestPath
	if path == "" {
		path = l.Path
	}
	statusCode := l.ResponseStatus
	if statusCode == 0 {
		statusCode = l.StatusCode
	}
	if msg == "" && path != "" {
		msg = fmt.Sprintf("%s %d", path, statusCode)
	}

	ts := time.UnixMilli(l.Timestamp)
	if l.Timestamp == 0 {
		ts = time.Now().UTC()
	}

	return log.LogEntry{
		Timestamp: ts,
		Source:    source,
		Level:     parseLevel(l.Level, statusCode),
		Message:   msg,
	}
}

func parseLevel(level string, statusCode int) log.LogLevel {
	switch strings.ToLower(level) {
	case "error", "fatal":
		return log.LevelError
	case "warning", "warn":
		return log.LevelWarn
	case "debug":
		return log.LevelDebug
	}

	if statusCode >= 500 {
		return log.LevelError
	}
	if statusCode >= 400 {
		return log.LevelWarn
	}

	return log.LevelInfo
}
