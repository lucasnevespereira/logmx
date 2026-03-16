# Architecture

## Overview

logmx is built around **connectors** — one per cloud provider — that each stream logs into a shared aggregator, which merges them and outputs to the terminal.

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
| Connectors        |
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

Starts connectors concurrently via goroutines, merges their output into a single `chan LogEntry`, and forwards entries to the printer.

### Connectors

Each connector implements the `Connector` interface, authenticates to a provider, and streams `LogEntry` values into the shared channel.

Planned connectors: Vercel, Railway, GCP, Render, Heroku, Docker, Kubernetes.

## Log Model

All connectors normalize logs into a single structure:

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

## Project Structure

```
logmx/
├── go.mod
├── cmd/
│   └── logmx/
│       ├── main.go              ← entry point
│       └── commands/
│           ├── root.go          ← cobra root command
│           ├── tail.go          ← tail command
│           ├── sources.go       ← sources command
│           └── init.go          ← init command
└── internal/
    ├── models/
    │   └── log.go               ← LogEntry, LogLevel
    ├── config/
    │   └── config.go            ← YAML config loading
    ├── aggregator/
    │   └── aggregator.go        ← merges connector channels
    └── connectors/
        ├── connector.go         ← Connector interface
        ├── demo/
        │   └── demo.go
        ├── vercel/
        │   └── vercel.go
        ├── railway/
        │   └── railway.go
        └── gcp/
            └── gcp.go
```
