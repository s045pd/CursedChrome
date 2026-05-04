# cursed-go

Go rewrite of the CursedChrome Node.js backend (server.js + api-server.js + database.js).

## Status

Active rewrite. See [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) for stage progress.

## Requirements

- Go 1.25+
- PostgreSQL 14+
- Redis 6+

## Quick start

```bash
cp .env.example .env
make build
./bin/cursed-server
```

Smoke check:

```bash
make smoke
```

## Layout

```
cmd/cursed-server  entry point
internal/config    env loading
internal/db        models + migrations
internal/auth      bcrypt, sessions, middleware
internal/api       REST routes (chi)
internal/ws        WebSocket server + RPC handlers
internal/proxy     HTTP forward proxy
internal/busx      Redis pub/sub bus
internal/utils     shared helpers
test/              integration & smoke
```

## Commands

| Make target | What it does |
|-------------|--------------|
| `make build` | Compile static binary into `./bin/` |
| `make test` | Run all unit tests |
| `make test-race` | Same with race detector |
| `make vet` | `go vet` |
| `make lint` | `golangci-lint` (if installed) |
| `make smoke` | Start binary, hit /health, /version |
| `make tidy` | `go mod tidy` |
