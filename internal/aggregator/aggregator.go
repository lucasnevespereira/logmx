package aggregator

import (
	"context"
	"sort"
	"time"

	"github.com/lucasnevespereira/logmx/internal/connectors"
	"github.com/lucasnevespereira/logmx/internal/models"
)

const flushInterval = 500 * time.Millisecond

type Aggregator struct {
	connectors []connectors.Connector
}

func New(conns []connectors.Connector) *Aggregator {
	return &Aggregator{connectors: conns}
}

func (a *Aggregator) Run(ctx context.Context) <-chan models.LogEntry {
	raw := make(chan models.LogEntry, 100)
	out := make(chan models.LogEntry, 100)

	for _, c := range a.connectors {
		c.Start(ctx, raw)
	}

	// Collect entries in short windows, sort by timestamp, then flush.
	go func() {
		defer close(out)
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()

		var buf []models.LogEntry

		for {
			select {
			case entry, ok := <-raw:
				if !ok {
					flush(buf, out)
					return
				}
				buf = append(buf, entry)

			case <-ticker.C:
				buf = flush(buf, out)

			case <-ctx.Done():
				flush(buf, out)
				return
			}
		}
	}()

	return out
}

func flush(buf []models.LogEntry, out chan<- models.LogEntry) []models.LogEntry {
	if len(buf) == 0 {
		return buf
	}

	sort.Slice(buf, func(i, j int) bool {
		return buf[i].Timestamp.Before(buf[j].Timestamp)
	})

	for _, e := range buf {
		out <- e
	}

	return buf[:0]
}
