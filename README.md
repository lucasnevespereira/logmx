# logmx

[![CI](https://img.shields.io/github/actions/workflow/status/lucasnevespereira/logmx/ci.yml?style=flat-square)](https://github.com/lucasnevespereira/logmx/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/lucasnevespereira/logmx?style=flat-square)](https://github.com/lucasnevespereira/logmx/releases/latest)

> Aggregate and stream logs from multiple platforms into a single terminal view.

Stop switching between Vercel, Railway dashboards. `logmx` streams all your logs into one place.

```
21:04:12 vercel   INFO  request completed in 42ms
21:04:12 railway  INFO  database connected
21:04:13 vercel   ERROR connection timeout after 30s
21:04:13 railway  WARN  memory usage at 80%
```

## Install

```sh
brew install lucasnevespereira/tools/logmx
```

Or with Go:

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
# Setup
logmx setup                        # Interactive setup wizard

# View logs
logmx tail                         # Show recent logs and exit
logmx tail -f                      # Stream logs in real time
logmx tail -n 50                   # Show last 50 logs per source
logmx tail --source my-api         # Filter by source
logmx tail --level error           # Filter by log level
logmx tail -f --source my-api      # Stream a specific source

# Manage providers & sources
logmx auth vercel                  # Add or refresh a provider token
logmx source add                   # Add sources (pick provider)
logmx source add --from vercel     # Add sources from a specific provider
logmx source list                  # List configured sources
logmx source remove my-docs        # Remove a source
```

`tail` works like unix `tail` — shows recent logs and exits. Add `-f` to follow in real time, just like `tail -f`.

## Supported Providers

| Provider | Auth      | Streaming via        |
| -------- | --------- | -------------------- |
| Vercel   | API token | `vercel` CLI (npm)   |
| Railway  | API token | `@railway/cli` (npm) |

## Documentation

- [Architecture](docs/ARCHITECTURE.md) — how it works internally
- [Roadmap](docs/ROADMAP.md) — what's planned

## Contributing

Contributions welcome. See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for an overview of the codebase before opening a PR.

## License

[MIT](LICENSE)
