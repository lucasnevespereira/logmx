# Architecture

## Overview

logmx is built around the concept of **connectors** — one per cloud provider — that each stream logs into a shared aggregator, which merges them and outputs to the terminal.

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

Library: [`clap`](https://github.com/clap-rs/clap)

### Aggregator

Starts connectors concurrently, merges their async streams, and forwards log entries to output.

Libraries: [`tokio`](https://tokio.rs), [`futures`](https://docs.rs/futures)

### Connectors

Each connector authenticates to a provider, fetches logs, and streams them as `LogEntry` values.

Planned connectors: Vercel, Railway, GCP, Render, Heroku, Docker, Kubernetes.

## Log Model

All connectors normalize logs into a single structure:

```rust
struct LogEntry {
    timestamp: DateTime<Utc>,
    source: String,
    level: LogLevel,
    message: String,
}
```

Example JSON representation:

```json
{
  "timestamp": "2026-03-16T18:10:00Z",
  "source": "vercel",
  "level": "error",
  "message": "timeout while fetching API"
}
```

## Project Structure

```
logmx
├── Cargo.toml
└── src
    ├── main.rs
    ├── cli.rs
    ├── aggregator.rs
    ├── models.rs
    ├── connectors
    │   ├── mod.rs
    │   ├── vercel.rs
    │   ├── railway.rs
    │   └── gcp.rs
    └── utils.rs
```
