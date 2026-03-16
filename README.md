# logmx

> Aggregate and stream logs from multiple cloud platforms into a single terminal view.

Stop switching between Vercel, Railway, GCP, and Render dashboards. `logmx` streams all your logs into one place.

```
21:04:12 vercel   INFO  request completed in 42ms
21:04:12 railway  INFO  database connected
21:04:13 vercel   ERROR connection timeout after 30s
21:04:13 railway  WARN  memory usage at 80%
```

## Install

```sh
go install github.com/lucasnevespereira/logmx/cmd/logmx@latest
```

## Quick Start

```sh
logmx setup
```

The setup wizard walks you through everything:

1. Pick your providers (Vercel, Railway)
2. Paste an API token — validated instantly
3. Installs streaming CLIs if missing
4. Lists your projects via API — pick which ones to stream

Then start streaming:

```sh
logmx tail
```

That's it. All your logs, one stream, sorted by timestamp.

## Usage

```sh
logmx setup                        # Interactive setup wizard
logmx tail                         # Stream all sources
logmx tail --source my-api,worker  # Stream specific sources
logmx tail --level error           # Filter by log level
logmx auth vercel                  # Add or refresh a provider token
logmx source add                   # Add sources (pick provider)
logmx source add --from vercel     # Add sources from a specific provider
logmx source list                  # List configured sources
logmx source remove my-docs        # Remove a source
```

## Supported Providers

| Provider | Auth      | Streaming via          |
| -------- | --------- | ---------------------- |
| Vercel   | API token | `vercel` CLI (npm)     |
| Railway  | API token | `@railway/cli` (npm)   |
| GCP      | Planned   | —                      |
| Render   | Planned   | —                      |

## Documentation

- [Architecture](docs/ARCHITECTURE.md) — how it works internally
- [Roadmap](docs/ROADMAP.md) — what's planned

## Contributing

Contributions welcome. See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for an overview of the codebase before opening a PR.

## License

[MIT](LICENSE)
