package railway

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lucasnevespereira/logmx/internal/models"
	provider "github.com/lucasnevespereira/logmx/internal/provider/railway"
	"github.com/lucasnevespereira/logmx/internal/retry"
)

type Connector struct {
	Source    string
	ServiceID string
	Token    string
}

func (c *Connector) Name() string {
	return c.Source
}

func (c *Connector) Start(ctx context.Context, ch chan<- models.LogEntry) error {
	client := provider.NewClient(c.Token)

	go func() {
		var deploymentID string
		seen := make(map[string]bool)

		retry.Backoff(ctx, 30*time.Second, func() error {
			// Resolve deployment if needed
			if deploymentID == "" {
				deployment, err := client.GetActiveDeployment(c.ServiceID)
				if err != nil {
					send(ctx, ch, errEntry(c.Source, "failed to get deployment: "+err.Error()))
					return err
				}
				deploymentID = deployment.ID
			}

			logs, err := client.QueryLogs(deploymentID, 50)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				return err
			}

			for _, l := range logs {
				key := l.Timestamp + "|" + l.Message
				if seen[key] {
					continue
				}
				seen[key] = true

				ts, _ := time.Parse(time.RFC3339Nano, l.Timestamp)
				if ts.IsZero() {
					ts = time.Now().UTC()
				}

				send(ctx, ch, models.LogEntry{
					Timestamp: ts,
					Source:    c.Source,
					Level:     parseLevel(l.Severity, l.Message),
					Message:   l.Message,
				})
			}

			// Cap the seen set
			if len(seen) > 5000 {
				seen = make(map[string]bool)
			}

			time.Sleep(3 * time.Second)
			return nil // success — resets backoff
		})
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
	case strings.Contains(lower, "debug"):
		return models.LevelDebug
	default:
		return models.LevelInfo
	}
}

func errEntry(source, msg string) models.LogEntry {
	return models.LogEntry{
		Timestamp: time.Now().UTC(),
		Source:    source,
		Level:     models.LevelError,
		Message:   fmt.Sprintf("[logmx] %s", msg),
	}
}

func send(ctx context.Context, ch chan<- models.LogEntry, e models.LogEntry) {
	select {
	case ch <- e:
	case <-ctx.Done():
	}
}
