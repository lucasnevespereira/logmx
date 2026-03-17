# Contributing

Thanks for your interest in logmx!

## Getting Started

```sh
git clone https://github.com/lucasnevespereira/logmx.git
cd logmx
make build
```

## Development

```sh
make build    # build binary to bin/logmx
make run      # run from source
make lint     # run go vet + golangci-lint
```

Requires Go 1.25+ and [golangci-lint](https://golangci-lint.run/) v2.

## Adding a Provider

Each provider lives in `internal/provider/<name>/` with two files:

- `api.go` — HTTP client for the provider's API (list projects, validate tokens)
- `logs.go` — log fetching and streaming via the provider's CLI

See `internal/provider/vercel/` for reference. A new provider also needs:

1. An entry in `provider.ProviderDeps` for CLI dependency checking
2. A case in `connectorForSource()` in `cmd/logmx/commands/tail.go`
3. A case in `validateToken()` and `fetchProjectOptions()` in the commands package

## Pull Requests

- Keep PRs focused — one feature or fix per PR
- Make sure `make lint` passes before pushing
- Use conventional commit prefixes in PR titles: `feat:`, `fix:`, `chore:`, `docs:`

## Architecture

See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for an overview of the codebase.
