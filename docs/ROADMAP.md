# Roadmap

## v0.0.1 — MVP

- [x] `logmx tail` CLI command
- [x] Demo connector with async streaming
- [x] Colored terminal output
- [x] `LogEntry` model with level, source, timestamp
- [x] Concurrent streaming from multiple sources
- [x] Filter by log level (`--level error`)
- [x] Config file (`~/.config/logmx/config.yaml`)
- [x] `logmx auth <provider>` — token-based auth (Vercel, Railway)
- [x] `logmx add` — interactive project picker
- [x] `logmx sources` / `logmx remove`
- [x] Vercel connector (polls deployment events, detects new deployments)
- [x] Railway connector (polls deployment logs via GraphQL)
- [x] Reconnection with exponential backoff
- [x] Log ordering by timestamp across sources

## v0.0.2 — More Providers & Polish

- [ ] GCP Cloud Logging connector
- [ ] Render connector
- [ ] Filter by keyword / regex (`logmx tail --grep "timeout"`)
- [ ] `logmx search "timeout"` — search recent logs across sources
- [ ] goreleaser + Homebrew tap

## Future Ideas

These are not committed but worth exploring:

- **Interactive TUI** — press `/` to search, `f` to filter
- **Local log caching** — `logmx search "database"` over recent history
- **Metrics view** — errors per service, warnings per minute
- **Docker / Kubernetes** — `logmx tail --apps docker,k8s`
