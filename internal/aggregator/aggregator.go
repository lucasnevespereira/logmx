package aggregator

import (
	"context"

	"github.com/lucasnevespereira/logmx/internal/connectors"
	"github.com/lucasnevespereira/logmx/internal/models"
)

type Aggregator struct {
	connectors []connectors.Connector
}

func New(conns []connectors.Connector) *Aggregator {
	return &Aggregator{connectors: conns}
}

func (a *Aggregator) Run(ctx context.Context) <-chan models.LogEntry {
	ch := make(chan models.LogEntry, 100)

	for _, c := range a.connectors {
		c.Start(ctx, ch)
	}

	// Close the channel when context is cancelled
	go func() {
		<-ctx.Done()
		close(ch)
	}()

	return ch
}
