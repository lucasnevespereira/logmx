package connectors

import (
	"context"

	"github.com/lucasnevespereira/logmx/internal/models"
)

// Connector fetches or streams log entries from a single source.
// Start blocks until the source is exhausted or the context is cancelled.
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
