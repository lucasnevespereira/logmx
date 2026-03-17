package demo

import (
	"context"
	"time"

	"github.com/lucasnevespereira/logmx/internal/log"
	"github.com/lucasnevespereira/logmx/internal/provider"
)

var messages = []struct {
	level   log.LogLevel
	message string
}{
	{log.LevelInfo, "request completed in 42ms"},
	{log.LevelWarn, "memory usage at 80%"},
	{log.LevelError, "failed to connect to database"},
	{log.LevelDebug, "cache miss for key user:123"},
}

type DemoConnector struct {
	Source string
	Follow bool
}

func (d *DemoConnector) Name() string {
	return d.Source
}

func (d *DemoConnector) Start(ctx context.Context, ch chan<- log.LogEntry) error {
	for i := 0; d.Follow || i < len(messages); i++ {
		msg := messages[i%len(messages)]
		provider.Send(ctx, ch, log.LogEntry{
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
