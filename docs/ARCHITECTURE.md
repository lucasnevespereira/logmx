# Architecture

## Overview

logmx is built around **providers** — one per cloud platform — that each stream logs into a shared aggregator, which merges them and outputs to the terminal.

```
+-------------------+
| CLI               |
+---------+---------+
          |
          v
+-------------------+
| Aggregator        |
| merges log streams|
+---------+---------+
          |
          v
+-------------------+
| Providers         |
| vercel            |
| railway           |
| gcp               |
| render            |
+-------------------+
```

## Components

### CLI

Parses commands, starts the aggregator, renders logs to the terminal.

Library: [`cobra`](https://github.com/spf13/cobra)

### Aggregator

Starts providers concurrently via goroutines, merges their output into a single `chan LogEntry`, sorts by timestamp in 500ms windows, and forwards entries to the printer. Uses a `sync.WaitGroup` to close the channel cleanly when all providers finish.

### Providers

Each provider implements the `Connector` interface. A provider directory contains:

- `api.go` — HTTP client for the provider's API (list projects, validate tokens)
- `logs.go` — log fetching and streaming via the provider's CLI

Supported: Vercel, Railway. Planned: GCP, Render, Docker, Kubernetes.

## Log Model

All providers normalize logs into a single structure:

```go
type LogEntry struct {
    Timestamp time.Time
    Source    string
    Level    LogLevel
    Message  string
}
```

## Configuration

Sources are defined in `~/.config/logmx/config.yaml`:

```yaml
sources:
  - name: api
    provider: vercel
    project: prj_abc123
  - name: worker
    provider: railway
    service: srv_xxx
```

Auth tokens are stored in `~/.config/logmx/auth.json`.

## Project Structure

```
logmx/
├── cmd/logmx/
│   ├── main.go                         ← entry point
│   └── commands/
│       ├── root.go                     ← cobra root command
│       ├── tail.go                     ← tail command (fetch + follow)
│       ├── auth.go                     ← auth command
│       ├── setup.go                    ← interactive setup wizard
│       └── source.go                   ← source add/list/remove
└── internal/
    ├── provider/
    │   ├── provider.go                 ← Connector interface, deps registry
    │   ├── vercel/
    │   │   ├── api.go                  ← Vercel REST API client
    │   │   └── logs.go                 ← log fetching/streaming via CLI
    │   ├── railway/
    │   │   ├── api.go                  ← Railway GraphQL API client
    │   │   └── logs.go                 ← log fetching/streaming via CLI
    │   └── demo/
    │       └── demo.go                 ← demo connector (mock data)
    ├── aggregator/
    │   └── aggregator.go               ← merges provider channels
    ├── config/
    │   ├── config.go                   ← config.yaml management
    │   └── auth.go                     ← auth.json token store
    └── log/
        ├── entry.go                    ← LogEntry, LogLevel
        └── printer.go                  ← colored terminal output
```
