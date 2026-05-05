# Deploying cursed-go

Drop-in replacement for the Node.js backend. The Chrome extensions and Vue GUI need no changes.

## Local

```bash
make build         # binary into ./bin/cursed-server
make test          # go test ./...
make test-race     # go test -race ./...
make smoke         # boot in stub mode and curl /health, /version
```

## Docker Compose

From the repo root:

```bash
docker compose up --build
```

First start writes `default admin user created` to logs with the auto-generated admin password.

## Portainer Deploy

From the repo root:

```bash
PORTAINER_PASS=xxx ./deploy.sh
```

The script builds the Vue 3 frontend, cross-compiles the Go binary (linux/amd64), packages everything into a Docker context, builds the image on the remote host via Portainer API, redeploys the stack, and verifies health.

## Ports

| Port | Purpose |
|------|---------|
| 8118 | REST API + Vue GUI |
| 4343 | WebSocket bot endpoint |
| 8080 | HTTP forward proxy (mapped to 8119 externally) |

## Environment Variables

All variables match the original Node.js server so existing Portainer stacks keep working:

```
DATABASE_HOST, DATABASE_PORT, DATABASE_NAME, DATABASE_USER, DATABASE_PASSWORD
REDIS_HOST, REDIS_PORT
BCRYPT_ROUNDS (default 10)
API_PORT (8118), WS_PORT (4343), PROXY_PORT (8080)
GUI_DIST_PATH (default /work/gui/dist)
EXTENSION_SRC_PATH (path to extensions directory)
SKIP_DB=1     # smoke-only: boot with no DB / RPC
```

## Health Check

- `GET /health` returns `{"success":true}`
- `GET /version` returns the build version

## Rollback

The old `s045pd/cursed_chrome:latest` Node.js image is unmodified. Reverting is an image tag change in Portainer; the database volume is shared.
