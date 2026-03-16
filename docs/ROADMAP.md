# Roadmap

## v0.1 — MVP

- [ ] `logmx tail` CLI command
- [ ] One working connector (Vercel or Railway)
- [ ] Async log streaming
- [ ] Colored terminal output
- [ ] Basic `LogEntry` model

## v0.2

- [ ] Multiple connectors (Vercel, Railway, GCP, Render)
- [ ] Concurrent streaming from multiple sources
- [ ] Filter by log level (`--level error`)
- [ ] Filter by keyword

## v0.3

- [ ] Config file with authentication tokens
- [ ] Reconnection handling
- [ ] Log ordering by timestamp across sources

## Future Ideas

These are not committed but worth exploring:

- **Interactive TUI** — press `/` to search, `f` to filter
- **Local log caching** — `logmx search "database"` over recent history
- **Metrics view** — errors per service, warnings per minute
- **Docker / Kubernetes** — `logmx tail --apps docker,k8s`
