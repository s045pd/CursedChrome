#!/usr/bin/env bash
# End-to-end smoke for cursed-go. Boots the binary in SKIP_DB stub mode
# (no Postgres / Redis required) and verifies the public surface.
set -euo pipefail

cd "$(dirname "$0")/.."

BIN=./bin/cursed-server
API_PORT=18118

echo "[smoke] build"
make build >/dev/null

echo "[smoke] launch (SKIP_DB=1)"
DATABASE_HOST=localhost DATABASE_NAME=x DATABASE_USER=x DATABASE_PASSWORD=x \
    REDIS_HOST=localhost \
    API_PORT=${API_PORT} WS_PORT=14343 PROXY_PORT=18080 SKIP_DB=1 \
    "$BIN" &
PID=$!

cleanup() {
    if kill -0 "$PID" 2>/dev/null; then
        kill "$PID" 2>/dev/null || true
        wait "$PID" 2>/dev/null || true
    fi
}
trap cleanup EXIT

# Wait for the API to be reachable.
for _ in $(seq 1 25); do
    if curl -fsS "http://127.0.0.1:${API_PORT}/health" >/dev/null 2>&1; then
        break
    fi
    sleep 0.2
done

echo "[smoke] /health"
curl -fsS "http://127.0.0.1:${API_PORT}/health" | grep -q '"success":true'

echo "[smoke] /version"
curl -fsS "http://127.0.0.1:${API_PORT}/version" | grep -q '"name":"cursed-go"'

echo "[smoke] login without DB returns 500/401 (not 200)"
status=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H 'Content-Type: application/json' \
    -d '{"username":"admin","password":"x"}' \
    "http://127.0.0.1:${API_PORT}/api/v1/login" || true)
if [[ "$status" == "200" ]]; then
    echo "[smoke] FAIL: login succeeded without DB"; exit 1
fi

echo "[smoke] /api/v1/me without session returns 401"
status=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:${API_PORT}/api/v1/me")
if [[ "$status" != "401" ]]; then
    echo "[smoke] FAIL: /api/v1/me status=$status, want 401"; exit 1
fi

echo "[smoke] CSP header present"
hdr=$(curl -s -D - -o /dev/null "http://127.0.0.1:${API_PORT}/health")
echo "$hdr" | grep -qi "content-security-policy"

echo "[smoke] OK"
