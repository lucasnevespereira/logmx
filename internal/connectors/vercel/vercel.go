package vercel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lucasnevespereira/logmx/internal/models"
	provider "github.com/lucasnevespereira/logmx/internal/provider/vercel"
	"github.com/lucasnevespereira/logmx/internal/retry"
)

type Connector struct {
	Source    string
	ProjectID string
	Token    string
}

func (c *Connector) Name() string {
	return c.Source
}

func (c *Connector) Start(ctx context.Context, ch chan<- models.LogEntry) error {
	client := provider.NewClient(c.Token)

	go func() {
		var deploymentID string
		var since int64

		retry.Backoff(ctx, 30*time.Second, func() error {
			// Resolve deployment if needed
			if deploymentID == "" {
				deployments, err := client.ListDeployments(c.ProjectID, 1)
				if err != nil {
					send(ctx, ch, errEntry(c.Source, "failed to list deployments: "+err.Error()))
					return err
				}
				if len(deployments) == 0 {
					send(ctx, ch, errEntry(c.Source, "no deployments found"))
					return fmt.Errorf("no deployments")
				}
				deploymentID = deployments[0].UID
				since = 0
			}

			events, err := client.GetDeploymentEvents(ctx, deploymentID, since, 100)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				return err
			}

			for _, ev := range events {
				if ev.Text == "" {
					continue
				}
				entry := models.LogEntry{
					Timestamp: time.UnixMilli(ev.Date),
					Source:    c.Source,
					Level:     parseLevel(ev.Type, ev.Text),
					Message:   ev.Text,
				}
				send(ctx, ch, entry)

				if ev.Date > since {
					since = ev.Date + 1
				}
			}

			// Check for new deployments when idle
			if len(events) == 0 {
				newDeps, err := client.ListDeployments(c.ProjectID, 1)
				if err == nil && len(newDeps) > 0 && newDeps[0].UID != deploymentID {
					deploymentID = newDeps[0].UID
					since = 0
					send(ctx, ch, models.LogEntry{
						Timestamp: time.Now().UTC(),
						Source:    c.Source,
						Level:     models.LevelInfo,
						Message:   fmt.Sprintf("new deployment detected: %s", deploymentID[:12]),
					})
				}
			}

			time.Sleep(2 * time.Second)
			return nil // success — resets backoff
		})
	}()

	return nil
}

func parseLevel(evType, text string) models.LogLevel {
	if evType == "stderr" {
		return models.LevelError
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
		Message:   msg,
	}
}

func send(ctx context.Context, ch chan<- models.LogEntry, e models.LogEntry) {
	select {
	case ch <- e:
	case <-ctx.Done():
	}
}
