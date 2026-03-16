package connectors

import (
	"context"

	"github.com/lucasnevespereira/logmx/internal/models"
)

type Connector interface {
	Name() string
	Start(ctx context.Context, ch chan<- models.LogEntry) error
}
