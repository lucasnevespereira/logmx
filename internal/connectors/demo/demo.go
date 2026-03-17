package demo

import (
	"context"
	"time"

	"github.com/lucasnevespereira/logmx/internal/connectors"
	"github.com/lucasnevespereira/logmx/internal/models"
)

var messages = []struct {
	level   models.LogLevel
	message string
}{
	{models.LevelInfo, "request completed in 42ms"},
	{models.LevelWarn, "memory usage at 80%"},
	{models.LevelError, "failed to connect to database"},
	{models.LevelDebug, "cache miss for key user:123"},
}

type DemoConnector struct {
	Source string
	Follow bool
}

func (d *DemoConnector) Name() string {
	return d.Source
}

func (d *DemoConnector) Start(ctx context.Context, ch chan<- models.LogEntry) error {
	for i := 0; d.Follow || i < len(messages); i++ {
		msg := messages[i%len(messages)]
		connectors.Send(ctx, ch, models.LogEntry{
			Timestamp: time.Now().UTC(),
			Source:    d.Source,
			Level:     msg.level,
			Message:   msg.message,
		})

		if d.Follow {
			select {
			case <-time.After(800 * time.Millisecond):
			case <-ctx.Done():
				return nil
			}
		}
	}

	return nil
}
