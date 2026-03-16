package connectors

import (
	"context"

	"github.com/lucasnevespereira/logmx/internal/models"
)

type Connector interface {
	Name() string
	Start(ctx context.Context, ch chan<- models.LogEntry) error
}

// Send writes an entry to the channel or returns if the context is cancelled.
func Send(ctx context.Context, ch chan<- models.LogEntry, e models.LogEntry) {
	select {
	case ch <- e:
	case <-ctx.Done():
	}
}
