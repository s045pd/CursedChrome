#!/usr/bin/env bash
set -euo pipefail

# ── Config ──────────────────────────────────────────
PORTAINER_URL="${PORTAINER_URL:-http://192.168.13.202:9000}"
PORTAINER_USER="${PORTAINER_USER:-admin}"
PORTAINER_PASS="${PORTAINER_PASS:-}"
ENDPOINT_ID="${ENDPOINT_ID:-2}"
STACK_ID="${STACK_ID:-51}"
IMAGE_TAG="${IMAGE_TAG:-s045pd/cursed_chrome:latest}"

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"

# ── Helpers ─────────────────────────────────────────
red()   { printf '\033[31m%s\033[0m\n' "$*"; }
green() { printf '\033[32m%s\033[0m\n' "$*"; }
info()  { printf '\033[36m→\033[0m %s\n' "$*"; }

die() { red "ERROR: $*" >&2; exit 1; }

# ── Pre-checks ──────────────────────────────────────
[[ -z "$PORTAINER_PASS" ]] && die "Set PORTAINER_PASS env var"
command -v go   >/dev/null || die "go not found"
command -v node >/dev/null || die "node not found"
command -v curl >/dev/null || die "curl not found"

# ── Step 1: Build Vue 3 frontend ────────────────────
info "Building Vue 3 frontend..."
cd "$ROOT_DIR/gui-next"
npm run build --silent
green "Frontend built → gui/dist/"

# ── Step 2: Cross-compile Go binary ─────────────────
info "Cross-compiling Go backend (linux/amd64)..."
cd "$ROOT_DIR/cursed-go"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -trimpath -ldflags="-s -w" \
  -o /tmp/cursed-server ./cmd/cursed-server
green "Go binary built → /tmp/cursed-server ($(du -h /tmp/cursed-server | cut -f1))"

# ── Step 3: Package Docker context ──────────────────
info "Packaging Docker context..."
TMPDIR=$(mktemp -d)
cp /tmp/cursed-server "$TMPDIR/cursed-server"
cp -r "$ROOT_DIR/gui/dist" "$TMPDIR/gui-dist"
cp "$ROOT_DIR/Dockerfile" "$TMPDIR/Dockerfile"
tar cf /tmp/cursed-deploy.tar -C "$TMPDIR" .
rm -rf "$TMPDIR"
green "Context packaged → /tmp/cursed-deploy.tar ($(du -h /tmp/cursed-deploy.tar | cut -f1))"

# ── Step 4: Authenticate with Portainer ─────────────
info "Authenticating with Portainer..."
JWT=$(curl -sf -X POST "$PORTAINER_URL/api/auth" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"$PORTAINER_USER\",\"password\":\"$PORTAINER_PASS\"}" \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['jwt'])")
[[ -z "$JWT" ]] && die "Failed to get Portainer JWT"
green "Authenticated"

# ── Step 5: Build image on remote Docker ────────────
info "Building Docker image on remote server..."
BUILD_OUT=$(curl -sf -X POST \
  "$PORTAINER_URL/api/endpoints/$ENDPOINT_ID/docker/build?t=$IMAGE_TAG&t=cursed-go:latest&nocache=1" \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/x-tar" \
  --data-binary @/tmp/cursed-deploy.tar \
  --max-time 300)

if echo "$BUILD_OUT" | grep -q '"errorDetail"'; then
  red "Build failed:"
  echo "$BUILD_OUT" | grep '"error"' | tail -1
  exit 1
fi
green "Image built and tagged: $IMAGE_TAG"

# ── Step 6: Stop old container ──────────────────────
info "Stopping old cursedchrome container..."
OLD_ID=$(curl -sf "$PORTAINER_URL/api/endpoints/$ENDPOINT_ID/docker/containers/json?filters=%7B%22name%22%3A%5B%22cursed-cursedchrome%22%5D%7D" \
  -H "Authorization: Bearer $JWT" \
  | python3 -c "import sys,json; cs=json.load(sys.stdin); print(cs[0]['Id'] if cs else '')" 2>/dev/null || true)

if [[ -n "$OLD_ID" ]]; then
  curl -sf -X DELETE \
    "$PORTAINER_URL/api/endpoints/$ENDPOINT_ID/docker/containers/$OLD_ID?force=true" \
    -H "Authorization: Bearer $JWT" >/dev/null 2>&1 || true
  green "Old container removed"
else
  info "No old container found, skipping"
fi

# ── Step 7: Redeploy stack ──────────────────────────
info "Redeploying Portainer stack..."
COMPOSE=$(curl -sf "$PORTAINER_URL/api/stacks/$STACK_ID/file" \
  -H "Authorization: Bearer $JWT" \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['StackFileContent'])")

BODY=$(python3 -c "
import json
compose = '''$COMPOSE'''
print(json.dumps({'StackFileContent': compose, 'Env': [], 'Prune': False, 'PullImage': False}))
")

RESULT=$(curl -sf -X PUT \
  "$PORTAINER_URL/api/stacks/$STACK_ID?endpointId=$ENDPOINT_ID" \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d "$BODY")

STATUS=$(echo "$RESULT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('Status',0))" 2>/dev/null || echo "0")
[[ "$STATUS" == "1" ]] || die "Stack deploy returned status=$STATUS"
green "Stack redeployed"

# ── Step 8: Verify health ──────────────────────────
info "Waiting for health check..."
for i in $(seq 1 30); do
  if curl -sf http://192.168.13.202:8118/health | grep -q '"success":true'; then
    green "Health check passed!"
    echo ""
    green "Deploy complete. New image running at http://192.168.13.202:8118"
    exit 0
  fi
  sleep 2
done

die "Health check timed out after 60s"
