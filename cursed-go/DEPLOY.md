# Deploying cursed-go

This Go rewrite is a drop-in replacement for the Node.js backend. The Chrome extension and the Vue GUI need **no changes**.

## Layout

```
cursed-go/
├── cmd/cursed-server/   single binary
├── internal/...         business code
├── Dockerfile           multi-stage, ~25 MB final image
├── docker-compose.yaml  redis + postgres + cursed-go
├── Makefile             build / test / lint / smoke
└── scripts/smoke.sh     end-to-end smoke
```

## Local

```bash
make build         # binary into ./bin/cursed-server
make test          # go test ./...
make test-race     # go test -race ./...
make smoke         # boot in stub mode and curl /health, /version, ...
```

## Compose

```bash
cd cursed-go
docker compose build
docker compose up -d
docker compose logs -f cursed-go
```

The first start writes a one-time `default admin user created` log line containing the auto-generated admin password. Save it, then immediately log in via the Vue GUI at http://localhost:8118/ and rotate it.

## Replacing the Portainer stack

The old `cursed` Portainer stack pinned to image `s045pd/cursed_chrome:latest`. To swap it for `cursed-go`:

1. Build & push the new image (or build on the host):
   ```bash
   docker build -t s045pd/cursed-go:latest cursed-go/
   ```
2. Edit the `cursed` stack in Portainer:
   * change the `cursedchrome` service image to `s045pd/cursed-go:latest`
   * keep all `DATABASE_*`, `REDIS_HOST` env vars (already correct)
   * the existing `cursedchrome-db` volume is reused — no data loss; GORM AutoMigrate is idempotent against the Sequelize schema
3. Update the stack. WebSockets and the Web panel will reattach automatically.

## Ports

| Port | Purpose |
|------|---------|
| 8118 | REST API + Vue GUI |
| 4343 | WebSocket bot endpoint |
| 8080 | HTTP forward proxy (mapped to 8119 externally) |

## Environment variables

All environment variables match the Node.js server.js so existing Portainer stacks keep working:

```
DATABASE_HOST, DATABASE_PORT, DATABASE_NAME, DATABASE_USER, DATABASE_PASSWORD
REDIS_HOST, REDIS_PORT
BCRYPT_ROUNDS (default 10)
API_PORT (8118), WS_PORT (4343), PROXY_PORT (8080)
GUI_DIST_PATH (default /work/gui/dist)
BAK_SERVER (optional)
SKIP_DB=1     # smoke-only: boot with no DB / RPC
```

## Health & smoke

* `GET /health` returns `{"success":true}` always
* `GET /version` returns the build version
* `make smoke` boots the binary in stub mode (no Postgres/Redis required) and exercises:
  - `/health`, `/version`
  - `/api/v1/login` returns 5xx without DB
  - `/api/v1/me` returns 401 without session cookie
  - CSP/security headers are emitted

## Tests

* unit tests under each package (sqlite in-memory, no Docker needed)
* end-to-end integration test under `test/integration/` exercises:
  1. login with seeded admin
  2. WebSocket bot AUTH handshake + persistence
  3. PING → PONG round-trip
  4. Proxy → SEND_REQUEST_VIA_BROWSER → fake bot → response forwarded

Run them with:

```bash
make test          # ~3s
make test-race     # ~10s (with -race)
```

## Rollback

The old `s045pd/cursed_chrome:latest` image is unmodified. Reverting is just an image tag change in Portainer; the database volume is shared between the two implementations.
