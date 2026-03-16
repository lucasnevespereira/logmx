package demo

import (
	"context"
	"time"

	"github.com/lucasnevespereira/logmx/internal/models"
)

type DemoConnector struct {
	Source string
}

func (d *DemoConnector) Name() string {
	return d.Source
}

func (d *DemoConnector) Start(ctx context.Context, ch chan<- models.LogEntry) error {
	messages := []struct {
		level   models.LogLevel
		message string
	}{
		{models.LevelInfo, "request completed in 42ms"},
		{models.LevelWarn, "memory usage at 80%"},
		{models.LevelError, "failed to connect to database"},
		{models.LevelDebug, "cache miss for key user:123"},
	}

	go func() {
		i := 0
		for {
			msg := messages[i%len(messages)]
			entry := models.LogEntry{
				Timestamp: time.Now().UTC(),
				Source:    d.Source,
				Level:     msg.level,
				Message:   msg.message,
			}

			select {
			case ch <- entry:
			case <-ctx.Done():
				return
			}

			i++
			time.Sleep(800 * time.Millisecond)
		}
	}()

	return nil
}
