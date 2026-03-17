# Roadmap

## v0.0.1 — MVP

- [x] `logmx tail` fetches recent runtime logs from all configured sources
- [x] Colored terminal output with log level parsing
- [x] Fetch logs from multiple sources concurrently
- [x] Filter by log level (`--level error`) and source (`--source api`)
- [x] Limit results per source (`--limit 50`)
- [x] Config file (`~/.config/logmx/config.yaml`)
- [x] Token-based auth stored locally (`~/.config/logmx/auth.json`)
- [x] `logmx auth` — paste API token, validated via provider API
- [x] `logmx setup` — TUI wizard (pick providers, paste tokens, install CLIs, pick sources)
- [x] `logmx source add/list/remove` — manage sources
- [x] Project listing via Vercel REST API and Railway GraphQL API
- [x] Log fetching via provider CLIs with `--token` passthrough
- [x] CLI dependency check + auto-install in setup wizard
- [x] Log ordering by timestamp across sources
- [x] Demo mode (no config needed)

## v0.0.2 — Streaming & CI

- [x] `logmx tail -f` for real-time log streaming
- [x] Architecture refactor — provider-centric package structure
- [x] CI pipeline (build + lint on push/PR)
- [x] Automated releases via GitHub Actions
- [x] Homebrew tap distribution

## v0.0.3 — More Providers

- [ ] Fly.io connector
- [ ] Render connector

## Future Ideas

- **Interactive TUI** — press `/` to search, `f` to filter, live dashboard
- **Local storage & search** — save logs to SQLite, query with `logmx search`
- **Metrics view** — errors per service, warnings per minute
- **Alerts** — `logmx watch --level error --notify slack`
- **Log export** — `logmx export --format json --since 24h`
