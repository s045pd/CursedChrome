# cursed-go

Go backend for CursedChrome. Single binary replacing the original Node.js server.

## Requirements

- Go 1.25+
- PostgreSQL 14+
- Redis 6+

## Quick Start

```bash
cp .env.example .env
make build
./bin/cursed-server
```

Smoke check (no DB required):

```bash
make smoke
```

## Layout

```
cmd/cursed-server  entry point
internal/config    env loading
internal/db        GORM models + migrations
internal/auth      bcrypt, sessions, middleware
internal/api       REST routes (chi) + extension packaging
internal/ws        WebSocket server + RPC handlers
internal/proxy     HTTP forward proxy
internal/busx      Redis pub/sub bus (multi-instance)
internal/utils     shared helpers
test/              integration tests
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

## Tests

Unit tests under each package (sqlite in-memory, no Docker needed). Integration test under `test/integration/` exercises login, WebSocket handshake, ping/pong, and proxy forwarding.

```bash
make test          # ~3s
make test-race     # ~10s
```
