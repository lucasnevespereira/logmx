# Architecture

## Overview

logmx is built around **providers** — one per platform — that each stream logs into a shared aggregator, which merges them and outputs to the terminal.

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
| fly.io            |
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

- `api.go` — client for the provider (HTTP API for Vercel, CLI wrapper for Railway)
- `logs.go` — log fetching and streaming via the provider's CLI

Supported: Vercel, Railway. Planned: Fly.io, Render.

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
    project: my-app
  - name: worker
    provider: railway
    project: 3d506c76-...
    service: 9f73aec2-...
    environment: f157c6e5-...
```

Auth tokens (Vercel) are stored in `~/.config/logmx/auth.json`. Railway authentication is managed by the Railway CLI itself (`~/.railway/config.json`).

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
    │   │   ├── api.go                  ← Railway CLI wrapper (auth, project discovery)
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
