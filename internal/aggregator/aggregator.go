package aggregator

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/lucasnevespereira/logmx/internal/log"
	"github.com/lucasnevespereira/logmx/internal/provider"
)

const flushInterval = 500 * time.Millisecond

type Aggregator struct {
	connectors []provider.Connector
}

func New(conns []provider.Connector) *Aggregator {
	return &Aggregator{connectors: conns}
}

func (a *Aggregator) Run(ctx context.Context) <-chan log.LogEntry {
	raw := make(chan log.LogEntry, 100)
	out := make(chan log.LogEntry, 100)

	var wg sync.WaitGroup
	for _, c := range a.connectors {
		wg.Add(1)
		go func(c provider.Connector) {
			defer wg.Done()
			_ = c.Start(ctx, raw)
		}(c)
	}

	go func() {
		wg.Wait()
		close(raw)
	}()

	go func() {
		defer close(out)
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()

		var buf []log.LogEntry

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

func flush(buf []log.LogEntry, out chan<- log.LogEntry) []log.LogEntry {
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
