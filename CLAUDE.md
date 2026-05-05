# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CursedChrome is an enterprise-level browser monitoring EDR (Endpoint Detection and Response) system that consists of:

- **Backend Server** (`cursed-go/`): Go single-binary server with WebSocket, HTTP proxy, and REST API
- **Frontend GUI** (`gui-next/`): Vue 3 + Vite + Tailwind CSS web interface (builds to `gui/dist/`)
- **Chrome Extensions**: Manifest V3 extensions for monitoring capabilities
  - `extension/` — Main monitoring extension (background service worker, content scripts, offscreen)
  - `cookie-sync-extension/` — Standalone cookie synchronization extension
  - `embed-targets/` — Lightweight MV3 extensions used as embed hosts
  - `bypass-paywalls-chrome/` — Paywall bypass (gitignored, optional)

## Architecture

### Go Backend (`cursed-go/`)

| Package | Purpose |
|---------|---------|
| `cmd/cursed-server` | Entry point |
| `internal/config` | Env-based configuration |
| `internal/db` | GORM models + PostgreSQL migrations |
| `internal/auth` | bcrypt, gorilla/sessions, middleware |
| `internal/api` | chi router, REST endpoints, extension packaging |
| `internal/ws` | WebSocket server (port 4343), RPC handlers |
| `internal/proxy` | HTTP forward proxy (port 8080) |
| `internal/busx` | Redis pub/sub bus (for multi-instance) |
| `internal/utils` | Shared helpers, logging |

**Ports**: 8118 (API + GUI), 4343 (WebSocket), 8080 (HTTP Proxy)

### Frontend (`gui-next/`)

- **Framework**: Vue 3 + Vite
- **Styling**: Tailwind CSS v4
- **Build output**: `../gui/dist/` (served by Go backend at `/`)
- **Dev proxy**: Vite proxies `/api/*` to Go backend during development

### Chrome Extensions

All extensions use **Manifest V3** with:
- Service Workers for background functionality
- Content Security Policy
- Chrome Storage API for persistent data
- Declarative Net Request for request modification

## Development

### Backend

```bash
cd cursed-go
make build          # Compile to ./bin/cursed-server
make test           # Run all tests
make test-race      # Tests with race detector
make smoke          # Smoke test (no DB required)
```

### Frontend

```bash
cd gui-next
npm install
npm run dev         # Dev server at :5173 with API proxy
npm run build       # Production build to ../gui/dist/
```

### Full Stack (Docker Compose)

```bash
docker compose up --build
```

First start prints admin credentials in logs.

### Extension Development

Load extensions in Chrome via `chrome://extensions/` in developer mode:
- `extension/` — Main monitoring extension
- `cookie-sync-extension/` — Cookie synchronization
- `embed-targets/*` — Embed host extensions

## Deployment

```bash
PORTAINER_PASS=xxx ./deploy.sh
```

The deploy script builds frontend, cross-compiles Go binary, packages Docker context with extensions, and deploys via Portainer API.

### Environment Variables

```
DATABASE_HOST, DATABASE_PORT, DATABASE_NAME, DATABASE_USER, DATABASE_PASSWORD
REDIS_HOST, REDIS_PORT
BCRYPT_ROUNDS (default 10)
API_PORT (8118), WS_PORT (4343), PROXY_PORT (8080)
GUI_DIST_PATH (default /work/gui/dist)
EXTENSION_SRC_PATH (path to extensions directory)
```

## Important Notes

- Go backend is a complete replacement for the original Node.js server
- All RPC action names, message formats, and API URLs are backward-compatible
- Database schema is compatible (GORM AutoMigrate is idempotent against old Sequelize schema)
- Extension packaging (injection, obfuscation, embed) is handled by `internal/api/extension.go`
