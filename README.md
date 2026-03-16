# logmx

> Aggregate and stream logs from multiple cloud platforms into a single terminal view.

Stop switching between Vercel, Railway, GCP, Render dashboards. `logmx` multiplexes logs from all your sources into one stream.

```sh
logmx tail --apps vercel,railway,gcp
```

```
[vercel] API server started
[railway] database connected
[vercel] ERROR timeout
[gcp]    worker processed job 1831
```

## Install

```sh
go install github.com/lucasnevespereira/logmx@latest
```

## Usage

```sh
# Tail logs from multiple sources
logmx tail --apps vercel,railway

# Filter by log level
logmx tail --level error

# Search recent logs
logmx search "timeout"

# List configured sources
logmx sources
```

## Documentation

- [Architecture](docs/ARCHITECTURE.md) — how it works internally
- [Roadmap](docs/ROADMAP.md) — what's planned

## Contributing

Contributions welcome. See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for an overview of the codebase before opening a PR.

## License

MIT
