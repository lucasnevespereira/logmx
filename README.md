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

### 1. Authenticate with your providers

```sh
logmx auth vercel
# → Paste your token (get one at https://vercel.com/account/tokens)
# → Verifying... authenticated as johndoe

logmx auth railway
# → Paste your token (get one at https://railway.app/account/tokens)
# → Verifying... authenticated as johndoe@example.com
```

Tokens are stored locally in `~/.config/logmx/auth.json` (file permissions: `0600`).

### 2. Pick which projects to tail

```sh
logmx add
# → Authenticated providers:
#     [1] vercel
#     [2] railway
# → Provider [1-2]: 1
# → Fetching projects from vercel...
#     [1] my-api
#     [2] my-frontend
#     [3] my-docs
# → Select sources (comma-separated, e.g. 1,3): 1,2
# → Added 2 source(s). Run 'logmx tail' to start streaming.
```

Repeat `logmx add` for each provider. Your sources are saved in `~/.config/logmx/config.yaml`.

### 3. Stream your logs

```sh
logmx tail
```

That's it. Logs from all your sources, one stream, sorted by timestamp.

## Usage

```sh
# Stream all sources
logmx tail

# Stream specific sources only
logmx tail --source my-api,worker

# Filter by log level
logmx tail --level error

# List configured sources
logmx sources

# Remove a source
logmx remove my-docs

# Try it without any config (demo mode)
logmx tail
```

## Supported Providers

| Provider | Status    | Auth                  |
| -------- | --------- | --------------------- |
| Vercel   | Supported | Personal access token |
| Railway  | Supported | API token             |
| GCP      | Planned   | —                     |
| Render   | Planned   | —                     |

## Documentation

- [Architecture](docs/ARCHITECTURE.md) — how it works internally
- [Roadmap](docs/ROADMAP.md) — what's planned

## Contributing

Contributions welcome. See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for an overview of the codebase before opening a PR.

## License

[MIT](LICENSE)
