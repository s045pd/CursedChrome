# Umbra-EDR Fork & Publish Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce a sanitized, rebranded fork of CursedChrome at `~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr/` and publish it to `https://github.com/s045pd/umbra-edr` as a public open-source EDR project.

**Architecture:** Five phases — copy with exclusions → apply Umbra rebrand mechanically → run sanitizer → run packager → final review then push. Each phase is locally-recoverable; only Phase 5 (publish) is irreversible. Agents are used for Phase 3 (`opensource-sanitizer`) and Phase 4 (`opensource-packager`); Phases 1–2 and 5 are manual because the rebrand details are project-specific.

**Tech Stack:** Go 1.25 (server), Vue 3 + Vite + Tailwind v4 (frontend), Chrome MV3 extensions, Docker Compose, GitHub via `gh` CLI, `rsvg-convert` for SVG → PNG.

**Source spec:** [docs/superpowers/specs/2026-05-07-umbra-rebrand-design.md](../specs/2026-05-07-umbra-rebrand-design.md)

---

## File Structure

The fork lives at `~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr/`. After Phases 1–4 it contains:

```
umbra-edr/
├── README.md                          (regenerated, OSS-ready, Phase 4)
├── LICENSE                            (MIT, copyright updated, Phase 4)
├── CONTRIBUTING.md                    (new, Phase 4)
├── SECURITY.md                        (new, authorized-use disclaimer, Phase 4)
├── CLAUDE.md                          (pruned, dev guidance only, Phase 4)
├── .env.example                       (new, no real values, Phase 2)
├── .gitignore                         (updated for umbra-server paths, Phase 2)
├── .github/
│   ├── ISSUE_TEMPLATE/                (new, Phase 4)
│   ├── PULL_REQUEST_TEMPLATE.md       (new, Phase 4)
│   └── workflows/
│       └── Build&Push.yml             (audited and rewritten, Phase 4)
├── setup.sh                           (new, Phase 4)
├── deploy.sh                          (Umbra-renamed, Phase 2)
├── Dockerfile                         (Umbra-renamed, Phase 2)
├── docker-compose.yaml                (Umbra-renamed, Phase 2)
├── images/
│   ├── umbra.svg                      (master icon, Phase 2)
│   ├── umbra-16.png                   (rendered, Phase 2)
│   ├── umbra-48.png                   (rendered, Phase 2)
│   ├── umbra-128.png                  (rendered, Phase 2)
│   └── umbra-512.png                  (rendered, Phase 2)
├── extension/
│   ├── manifest.json                  (defensive description, Phase 2)
│   ├── icons/icon{16,48,128}.png      (replaced, Phase 2)
│   └── src/                           (string sweep, Phase 2)
├── cookie-sync-extension/             (ShadowLink — Umbra icon reuse, Phase 2)
├── embed-targets/                     (carried unchanged from source)
├── gui-next/
│   ├── package.json                   (name updated, Phase 2)
│   ├── vite.config.ts                 (comment + LAN IP scrubbed, Phase 2)
│   ├── index.html                     (title updated, Phase 2)
│   └── src/
│       ├── assets/styles.css          (Umbra @theme block, Phase 2)
│       ├── layouts/AppShell.vue       (brand string, Phase 2)
│       └── pages/Login.vue            (brand string, Phase 2)
└── umbra-server/                      (was cursed-go/, Phase 2)
    ├── go.mod                         (module path updated)
    ├── Makefile                       (binary name updated)
    ├── Dockerfile                     (binary name updated)
    ├── cmd/umbra-server/              (was cmd/cursed-server/)
    └── internal/                      (54 import paths rewritten)
```

Excluded from the fork (must NOT exist under `umbra-edr/`):

- `.git/` — fresh history starts in Phase 5
- `.env`, `.env.*` — operator credentials
- `.chrome-data/` — operator browser profile
- `bypass-paywalls-chrome/` — third-party redistribution risk
- `ssl/` — operator's certs (regen instructions in README)
- `gui/dist/`, `cursed-go/bin/`, `cursed-go/deploy/cursed-server`, `cursed-go/deploy/gui-dist/`, `cursed-go/deploy/extensions/` — build artifacts
- `cursed-go/tools/node_modules/`, `gui-next/node_modules/` — dependencies
- `images/cursedchrome-diagram.png`, `images/cursed-chrome-web-panel.png`, `images/doll.svg`, `images/icon.png` — old branding
- `images/umbra-icon-sideview.svg`, `images/umbra-stitch/` — exploration artifacts
- `.DS_Store` — macOS metadata
- `bypass-paywalls-chrome.zip` — same redistribution issue

---

## Phase 1 — Fork & exclude

### Task 1: Verify source repo state and create exclusion file

**Files:**
- Create: `/tmp/umbra-rsync-exclude.txt`

- [ ] **Step 1: Confirm working directory and CursedChrome state**

Run:
```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/CursedChrome
pwd
git status --short | head -5
git log --oneline -3
```

Expected: pwd ends with `/CursedChrome`, branch is `cursed-go-rewrite`, last commit `f20e571` (spec revision).

- [ ] **Step 2: Verify target directory does not exist**

Run:
```bash
ls -d ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr 2>&1 | head -1
```

Expected: `ls: ...: No such file or directory`. If the directory exists, stop and confirm with the operator before proceeding.

- [ ] **Step 3: Write rsync exclusion file**

Create `/tmp/umbra-rsync-exclude.txt` with the following content:

```
.git/
.env
.env.*
.chrome-data/
.DS_Store
bypass-paywalls-chrome/
bypass-paywalls-chrome.zip
ssl/
gui/dist/
node_modules/
*.test
*.out
coverage.txt
cursed-go/bin/
cursed-go/cursed-server
cursed-go/deploy/cursed-server
cursed-go/deploy/gui-dist/
cursed-go/deploy/extensions/
cursed-go/tools/node_modules/
images/cursedchrome-diagram.png
images/cursed-chrome-web-panel.png
images/doll.svg
images/icon.png
images/umbra-icon-sideview.svg
images/umbra-stitch/
docs/superpowers/
.idea/
.vscode/
```

Note: `.vscode/` is excluded so that `launch.json` operator-paths don't leak. We re-add a generic `.vscode/launch.json` later if needed. `docs/superpowers/` is excluded because spec/plan files mention internal context.

- [ ] **Step 4: Commit nothing — this task is local-only setup**

The exclusion file is in `/tmp/`, no git changes.

---

### Task 2: rsync source tree to fork directory

**Files:**
- Create: `~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr/` (entire tree)

- [ ] **Step 1: Run rsync with exclusion file**

Run:
```bash
rsync -av \
  --exclude-from=/tmp/umbra-rsync-exclude.txt \
  ~/workobj/BackupData/DocumentArchive/GitHub/CursedChrome/ \
  ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr/
```

Expected: `sent ... bytes` summary; no errors. Trailing slash on source path is critical (copies contents, not the parent directory).

- [ ] **Step 2: Verify excluded patterns are absent**

Run:
```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
find . \( \
  -name ".env" -o \
  -name ".env.*" -o \
  -name ".DS_Store" -o \
  -name "node_modules" -o \
  -path "*/bypass-paywalls-chrome*" -o \
  -path "*/ssl/*" -o \
  -path "*/.chrome-data/*" -o \
  -path "*/.git/*" -o \
  -path "*/gui/dist/*" \
  \) -print 2>/dev/null | head
```

Expected: empty output (no matches).

- [ ] **Step 3: Verify expected directories are present**

Run:
```bash
ls -la ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr/
```

Expected: visible — `cursed-go/`, `extension/`, `cookie-sync-extension/`, `embed-targets/`, `gui-next/`, `images/`, `LICENSE`, `Dockerfile`, `docker-compose.yaml`, `deploy.sh`, `README.md`, `CLAUDE.md`. NOT visible: `.git`, `.env`, `.chrome-data`, `bypass-paywalls-chrome`, `ssl`, `docs`.

- [ ] **Step 4: Commit nothing yet — fork has no `.git` until Phase 5**

---

### Task 3: Verify fork size and integrity

**Files:** none

- [ ] **Step 1: Compare fork size to source**

Run:
```bash
du -sh ~/workobj/BackupData/DocumentArchive/GitHub/CursedChrome
du -sh ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
```

Expected: fork is meaningfully smaller (no node_modules, no .git, no chrome-data, no bypass-paywalls). Typical: source 200MB+, fork <50MB.

- [ ] **Step 2: Sanity check — scan for obvious leaks**

Run:
```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
grep -rln "s045pd.x@gmail" . 2>/dev/null | head -5
grep -rln "192.168" . 2>/dev/null | head -10
grep -rln "/Users/s045pd" . 2>/dev/null | head -10
```

Expected: results may be non-empty — these are exactly what Phase 2/3 sanitization addresses. Note the file list mentally; revisit during Task 12 / 19.

---

## Phase 2 — Apply Umbra rebrand

All work from Task 4 onward happens inside `~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr/`. Use that as the implicit working directory unless a step says otherwise.

### Task 4: Rename `cursed-go/` to `umbra-server/`

**Files:**
- Modify: directory `cursed-go/` → `umbra-server/`
- Modify: directory `cursed-go/cmd/cursed-server/` → `umbra-server/cmd/umbra-server/`

- [ ] **Step 1: Move the top-level directory**

Run:
```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
mv cursed-go umbra-server
```

Expected: silent success; `umbra-server/` exists, `cursed-go/` does not.

- [ ] **Step 2: Move the cmd subdirectory**

Run:
```bash
mv umbra-server/cmd/cursed-server umbra-server/cmd/umbra-server
```

- [ ] **Step 3: Verify directory tree**

Run:
```bash
ls umbra-server/cmd/
ls umbra-server/internal/ | head -10
```

Expected: `cmd/umbra-server/` (single subdir), `internal/` contains `api/`, `auth/`, `busx/`, `config/`, `db/`, `proxy/`, `utils/`, `version/`, `ws/`.

---

### Task 5: Update Go module path

**Files:**
- Modify: `umbra-server/go.mod`
- Modify: all 54 `*.go` files containing `github.com/s045pd/cursed-go` imports

- [ ] **Step 1: Update `go.mod` module declaration**

Edit `umbra-server/go.mod` line 1:

Old:
```
module github.com/s045pd/cursed-go
```

New:
```
module github.com/s045pd/umbra
```

- [ ] **Step 2: Bulk-rewrite import paths across all Go files**

Run:
```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr/umbra-server
grep -rl "github.com/s045pd/cursed-go" --include="*.go" . | \
  xargs sed -i '' 's|github.com/s045pd/cursed-go|github.com/s045pd/umbra|g'
```

- [ ] **Step 3: Verify no `cursed-go` imports remain**

Run:
```bash
grep -rn "github.com/s045pd/cursed-go" --include="*.go" . | head
```

Expected: no output.

Run:
```bash
grep -rn "github.com/s045pd/umbra" --include="*.go" . | wc -l
```

Expected: 54 (matches original count).

- [ ] **Step 4: Tidy `go.mod` and verify build**

Run:
```bash
go mod tidy
go build ./...
```

Expected: `go build` succeeds with no output.

- [ ] **Step 5: Run tests**

Run:
```bash
go test ./...
```

Expected: PASS for all packages that had passing tests before. If tests need a DB, they may skip — that's acceptable.

---

### Task 6: Update binary name in build files

**Files:**
- Modify: `umbra-server/Makefile`
- Modify: `umbra-server/Dockerfile`
- Modify: `umbra-server/deploy/Dockerfile`
- Modify: `umbra-server/scripts/smoke.sh`
- Modify: `umbra-server/scripts/deploy.sh`

- [ ] **Step 1: Replace `cursed-server` with `umbra-server` in build files**

Run:
```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr/umbra-server
grep -rl "cursed-server" --include="Makefile" --include="Dockerfile" --include="*.sh" . | \
  xargs sed -i '' 's|cursed-server|umbra-server|g'
```

- [ ] **Step 2: Verify build output target**

Run:
```bash
grep -rn "cursed" Makefile Dockerfile deploy/Dockerfile scripts/ 2>/dev/null
```

Expected: empty (or only inside historical comments — review and remove any remaining).

- [ ] **Step 3: Build and confirm binary name**

Run:
```bash
make build
ls -la bin/
```

Expected: `bin/umbra-server` produced. No `bin/cursed-server`.

---

### Task 7: Rename cookie name prefix

**Files:**
- Modify: `umbra-server/internal/api/extension.go:264,271`

- [ ] **Step 1: Locate cookie names**

Run:
```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr/umbra-server
grep -n "cursed_" internal/api/extension.go
```

Expected: 2 matches around lines 264 and 271 (cookie name strings like `cursed_session`, `cursed_proxy`, etc).

- [ ] **Step 2: Replace prefix**

Run:
```bash
sed -i '' 's|cursed_|umbra_|g' internal/api/extension.go
```

- [ ] **Step 3: Verify**

Run:
```bash
grep -n "cursed_" internal/api/extension.go
grep -n "umbra_" internal/api/extension.go
```

Expected: no `cursed_` matches; 2 `umbra_` matches.

- [ ] **Step 4: Build to confirm no syntax breakage**

Run:
```bash
go build ./...
```

Expected: success.

---

### Task 8: Render Umbra master icon and PNGs

**Files:**
- Create: `images/umbra.svg`
- Create: `images/umbra-16.png`
- Create: `images/umbra-48.png`
- Create: `images/umbra-128.png`
- Create: `images/umbra-512.png`
- Delete: `images/umbra-icon-topdown.svg` (after promotion)

- [ ] **Step 1: Promote topdown SVG to master**

Run:
```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
mv images/umbra-icon-topdown.svg images/umbra.svg
```

- [ ] **Step 2: Render PNGs at 4 sizes**

Run:
```bash
for size in 16 48 128 512; do
  rsvg-convert -w $size -h $size images/umbra.svg -o images/umbra-${size}.png
done
```

- [ ] **Step 3: Verify**

Run:
```bash
ls -la images/umbra*.png images/umbra.svg
file images/umbra-128.png
```

Expected: 4 PNG files exist; `file` reports each as `PNG image data, NxN, 8-bit/color RGBA`.

- [ ] **Step 4: Visual spot-check (manual)**

Open `images/umbra-128.png` in Preview. Confirm:
- Dark canopy disc with 8 spokes
- One amber wedge in upper-right
- Center ferrule (dark dot with amber inside)
- No pixelation

If the SVG produces a poor-quality 16px render (artifacts in spokes or wedge), consider adjusting stroke widths or wedge angle in `umbra.svg` and re-running step 2.

---

### Task 9: Replace extension icons with Umbra renders

**Files:**
- Replace: `extension/icons/icon16.png`
- Replace: `extension/icons/icon48.png`
- Replace: `extension/icons/icon128.png`
- Replace: `cookie-sync-extension/icons/icon16.png`
- Replace: `cookie-sync-extension/icons/icon48.png`
- Replace: `cookie-sync-extension/icons/icon128.png`

- [ ] **Step 1: Copy renders into both extensions**

Run:
```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
for size in 16 48 128; do
  cp images/umbra-${size}.png extension/icons/icon${size}.png
  cp images/umbra-${size}.png cookie-sync-extension/icons/icon${size}.png
done
```

- [ ] **Step 2: Verify file replacement**

Run:
```bash
ls -la extension/icons/ cookie-sync-extension/icons/
```

Expected: each `icon{16,48,128}.png` matches the size of its `images/umbra-N.png` source.

---

### Task 10: GUI design tokens — Umbra theme

**Files:**
- Modify: `gui-next/src/assets/styles.css` (the `@theme` block, lines 10-46)

- [ ] **Step 1: Read the current `@theme` block**

Read `gui-next/src/assets/styles.css` lines 1-90. Confirm the current tokens are Linear/Datadog cyan-on-navy.

- [ ] **Step 2: Replace the `@theme` block with Umbra tokens**

Replace lines 3-46 (everything from the `Design tokens` comment through the closing `}` of `@theme`) with:

```css
/* ----------------------------------------------------------------
 * Umbra design tokens.
 *
 * Geometric Minimalism, dark-mode first. Single amber accent over
 * navy/midnight surfaces. Sharp edges (radius 0). Tonal layering
 * replaces traditional shadows.
 * ---------------------------------------------------------------- */
@theme {
  --color-bg-base: oklch(8% 0.012 256);        /* #070a12 deep midnight */
  --color-bg-raised: oklch(20% 0.024 256);     /* #1c2536 navy surface */
  --color-bg-overlay: oklch(25% 0.028 257);    /* #252f44 modal */
  --color-bg-hover: oklch(28% 0.030 257);

  --color-border-subtle: oklch(100% 0 0 / 0.08);
  --color-border-strong: oklch(100% 0 0 / 0.16);

  --color-fg-base: oklch(96% 0.005 90);        /* near-white headings */
  --color-fg-muted: oklch(70% 0.012 240);
  --color-fg-faint: oklch(50% 0.012 240);

  --color-accent: oklch(83% 0.165 84);         /* #fbbf24 amber */
  --color-accent-strong: oklch(89% 0.155 88);
  --color-accent-soft: oklch(63% 0.158 51);    /* #d97706 amber-dark */

  --color-success: oklch(72% 0.18 145);
  --color-success-soft: oklch(30% 0.08 145);
  --color-warn: oklch(83% 0.165 84);           /* warn folds into accent */
  --color-warn-soft: oklch(45% 0.10 70);
  --color-danger: oklch(70% 0.20 25);
  --color-danger-soft: oklch(32% 0.10 25);

  --font-sans:
    "Space Grotesk", "Inter", ui-sans-serif, system-ui, -apple-system,
    BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  --font-mono:
    "JetBrains Mono", ui-monospace, "SF Mono", Menlo, Consolas, monospace;

  --radius-sm: 0px;
  --radius-md: 0px;
  --radius-lg: 0px;

  --shadow-soft: 0 1px 2px 0 oklch(0% 0 0 / 0.6);
  --shadow-pop: 0 12px 32px -8px oklch(0% 0 0 / 0.8);
}
```

- [ ] **Step 3: Update the light-theme override block (lines 49-66)**

Light mode is no longer a primary aesthetic for Umbra (geometric-minimalism dark-first). Replace the `:root.light` block with a minimal opt-in light palette:

```css
/* Light theme overrides — opt-in only, kept for accessibility */
:root.light {
  --color-bg-base: oklch(98% 0.003 90);
  --color-bg-raised: oklch(95% 0.005 90);
  --color-bg-overlay: oklch(98% 0.003 90);
  --color-bg-hover: oklch(92% 0.008 90);

  --color-border-subtle: oklch(0% 0 0 / 0.10);
  --color-border-strong: oklch(0% 0 0 / 0.20);

  --color-fg-base: oklch(15% 0.005 240);
  --color-fg-muted: oklch(40% 0.012 240);
  --color-fg-faint: oklch(55% 0.012 240);

  --color-accent-soft: oklch(92% 0.10 84);
  --color-success-soft: oklch(92% 0.06 145);
  --color-warn-soft: oklch(94% 0.10 84);
  --color-danger-soft: oklch(94% 0.08 25);
}
```

- [ ] **Step 4: Add Space Grotesk via Google Fonts in `gui-next/index.html`**

Edit `gui-next/index.html`. After the existing `<link>` lines in `<head>`, add (or replace any existing font link):

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
```

- [ ] **Step 5: Build and verify GUI compiles**

Run:
```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr/gui-next
npm install
npm run build
```

Expected: `npm install` finishes. `npm run build` produces `../gui/dist/`. No TypeScript errors. (If `npm run build` fails because of missing `gui/dist/` write permission, run `mkdir -p ../gui/dist` first.)

- [ ] **Step 6: Visual spot-check (manual)**

Run `npm run dev` and open in browser. Confirm:
- Background is deep midnight, not Linear-cyan
- Accents are amber, not blue
- Font reads as Space Grotesk
- Sharp corners on cards/buttons (no rounding)

If any token leaks (e.g., a hardcoded hex in a Vue component overrides the theme), grep for it: `grep -rn "#137" gui-next/src/` and replace with token references.

---

### Task 11: Update `gui-next/package.json` and `vite.config.ts`

**Files:**
- Modify: `gui-next/package.json`
- Modify: `gui-next/vite.config.ts`

- [ ] **Step 1: Rename package**

Edit `gui-next/package.json` line 2:

Old:
```json
"name": "cursed-gui",
```

New:
```json
"name": "umbra-gui",
```

- [ ] **Step 2: Rewrite `vite.config.ts` comment header and remove LAN IP**

Edit `gui-next/vite.config.ts` lines 6-19. Replace:

```ts
// CursedChrome GUI build config.
//
// Output goes into ../gui/dist so the Go server can keep serving from
// the same path (GUI_DIST_PATH = /work/gui/dist) without needing any
// backend change. The /api/v1/* and /favicon.ico paths are proxied to
// the live API during dev so cookie-based session works against the
// running Go server on :8118.
//
// Proxy target defaults to local backend (the F5 launch.json config or a
// local `make run`). To point at a remote server during dev, set
// VITE_API_TARGET, e.g.:
//
//   VITE_API_TARGET=http://192.168.13.202:8118 npm run dev
const apiTarget = process.env.VITE_API_TARGET || 'http://127.0.0.1:8118'
```

With:

```ts
// Umbra GUI build config.
//
// Output goes into ../gui/dist so the Go server keeps serving the SPA
// from a stable path (GUI_DIST_PATH = /work/gui/dist).
//
// /api/v1/* and /favicon.ico are proxied to the running Go backend
// during dev so cookie-based session auth works end-to-end. To point
// at a non-local backend, set VITE_API_TARGET, e.g.:
//
//   VITE_API_TARGET=http://your-server:8118 npm run dev
const apiTarget = process.env.VITE_API_TARGET || 'http://127.0.0.1:8118'
```

- [ ] **Step 3: Verify**

Run:
```bash
grep -n "CursedChrome\|192.168\|cursed-gui" gui-next/vite.config.ts gui-next/package.json
```

Expected: no matches.

---

### Task 12: GUI brand strings — `index.html`, `AppShell.vue`, `Login.vue`

**Files:**
- Modify: `gui-next/index.html`
- Modify: `gui-next/src/layouts/AppShell.vue`
- Modify: `gui-next/src/pages/Login.vue`

- [ ] **Step 1: Update `<title>` in `gui-next/index.html`**

Find the line with `<title>CursedChrome — EDR Console</title>` and replace with:

```html
<title>Umbra · Browser-layer EDR</title>
```

- [ ] **Step 2: Update brand wordmark in `AppShell.vue`**

In `gui-next/src/layouts/AppShell.vue` find:

```html
<span class="font-semibold tracking-tight text-[13px]">CursedChrome</span>
```

Replace with:

```html
<span class="font-semibold tracking-tight text-[13px]">Umbra</span>
```

- [ ] **Step 3: Update `Login.vue`**

In `gui-next/src/pages/Login.vue` find:

```html
<h1 class="text-[15px] font-semibold tracking-tight">CursedChrome</h1>
```

Replace with:

```html
<h1 class="text-[15px] font-semibold tracking-tight">Umbra</h1>
```

- [ ] **Step 4: Verify**

Run:
```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
grep -rn "CursedChrome" gui-next/
```

Expected: no matches.

- [ ] **Step 5: Rebuild GUI**

Run:
```bash
cd gui-next && npm run build && cd ..
```

Expected: success.

---

### Task 13: Extension manifests + popup brand strings + defensive language

**Files:**
- Modify: `extension/manifest.json`
- Modify: `cookie-sync-extension/manifest.json` (if it exists)
- Modify: `cookie-sync-extension/src/browser_action/browser_action.html`
- Modify: `cookie-sync-extension/src/browser_action/main.js`

- [ ] **Step 1: Rewrite `extension/manifest.json`**

The current `name` is `"CursedChrome Implant"` and `description` mentions "implant code... inject/disguise". Replace lines 2-5:

Old:
```json
{
  "name": "CursedChrome Implant",
  "version": "0.0.1",
  "manifest_version": 3,
  "description": "Example Chrome extension with implant code. You should probably inject/disguise the implant instead of installing this extension directly.",
```

New:
```json
{
  "name": "Umbra Sensor",
  "version": "0.1.0",
  "manifest_version": 3,
  "description": "Browser-layer EDR sensor for authorized enterprise monitoring. Reports activity, screen captures, and credential events to a central Umbra server. Authorized use only.",
```

Also update `homepage_url` if present:

Old:
```json
"homepage_url": "https://thehackerblog.com",
```

New:
```json
"homepage_url": "https://github.com/s045pd/umbra-edr",
```

- [ ] **Step 2: Update `cookie-sync-extension` manifest if name field references CursedChrome**

Run:
```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
cat cookie-sync-extension/manifest.json | head -10
```

If the `name` is anything other than "ShadowLink" or "Umbra", update it to:

```json
"name": "ShadowLink",
"description": "Cookie + credential synchronization sidecar for the Umbra EDR platform. Authorized use only.",
```

- [ ] **Step 3: Sweep popup HTML and JS**

Run:
```bash
grep -n "CursedChrome\|cursed-chrome\|cursed-go" cookie-sync-extension/src/browser_action/browser_action.html cookie-sync-extension/src/browser_action/main.js
```

For each match, replace `CursedChrome` → `Umbra`, `cursed-chrome` → `umbra-edr`, `cursed-go` → `umbra-server`. Use `sed -i '' 's|...|...|g' <file>` or manual edits.

- [ ] **Step 4: Verify**

Run:
```bash
grep -rn "CursedChrome\|implant\|inject/disguise\|thehackerblog" extension/ cookie-sync-extension/ 2>/dev/null
```

Expected: no matches.

- [ ] **Step 5: Visual smoke-test the extension manifests**

Run:
```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
python3 -c "import json; print(json.load(open('extension/manifest.json'))['name'])"
python3 -c "import json; print(json.load(open('cookie-sync-extension/manifest.json'))['name'])"
```

Expected: `Umbra Sensor` and `ShadowLink` respectively.

---

### Task 14: MITM TLS cert subject line

**Files:**
- Modify: `umbra-server/internal/proxy/mitm.go`

- [ ] **Step 1: Locate and replace cert subject strings**

Run:
```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr/umbra-server
grep -n "CursedChrome" internal/proxy/mitm.go
```

Expected: 2 matches (CN and Org fields). Replace both with `Umbra`:

```bash
sed -i '' 's|CursedChrome|Umbra|g' internal/proxy/mitm.go
```

- [ ] **Step 2: Verify**

Run:
```bash
grep -n "CursedChrome\|Umbra" internal/proxy/mitm.go
```

Expected: 0 `CursedChrome`, ≥2 `Umbra`.

- [ ] **Step 3: Build to confirm**

Run:
```bash
go build ./...
```

Expected: success.

---

### Task 15: Top-level deployment files

**Files:**
- Modify: `Dockerfile`
- Modify: `docker-compose.yaml`
- Modify: `deploy.sh`
- Modify: `umbra-server/internal/config/config.go` (default values)
- Create: `.env.example`

- [ ] **Step 1: Sweep top-level Dockerfile**

Run:
```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
grep -n "cursed\|CursedChrome" Dockerfile
```

For each match, replace `cursed-go` → `umbra-server`, `cursed-server` → `umbra-server`, `CursedChrome` → `Umbra`, `cursedchrome` → `umbra`. Use sed or manual edits.

- [ ] **Step 2: Sweep `docker-compose.yaml`**

```bash
grep -n "cursed\|CursedChrome" docker-compose.yaml
```

Apply the same replacements. Pay attention to:
- `services:` keys (e.g., `cursedchrome:` → `umbra:`)
- `container_name:` values
- `image:` tags
- `volumes:` named volumes (`cursedchrome_pgdata` → `umbra_pgdata`)
- `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD` env values — change defaults to `umbra` and remove the hardcoded password (require it from env).

- [ ] **Step 3: Sweep `deploy.sh`**

```bash
grep -n "cursed\|CursedChrome" deploy.sh
```

Apply replacements. Stack name `cursedchrome` → `umbra-edr`, image tag `cursed-go` → `umbra-server`.

- [ ] **Step 4: Update Go config defaults**

In `umbra-server/internal/config/config.go`, find the lines that set defaults for `DATABASE_NAME`, `DATABASE_USER`, `DATABASE_PASSWORD`. Change:

- `DATABASE_NAME` default: `cursedchrome` → `umbra`
- `DATABASE_USER` default: `cursedchrome` → `umbra`
- `DATABASE_PASSWORD` default: `cursedchrome` → remove default; if env var is empty, return an error like `errors.New("DATABASE_PASSWORD env var is required")` and refuse to start.

After the edit, verify:

```bash
go build ./...
go test ./internal/config/...
```

Expected: build passes, config tests still pass (existing tests already supply env vars).

- [ ] **Step 5: Create `.env.example`**

Create `.env.example` at the fork root with:

```dotenv
# Umbra EDR — environment configuration template.
# Copy to `.env` and fill in real values. Never commit `.env`.

# Database (PostgreSQL)
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=umbra
DATABASE_USER=umbra
DATABASE_PASSWORD=__set_a_strong_password__

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Crypto
BCRYPT_ROUNDS=10

# Network ports
API_PORT=8118
WS_PORT=4343
PROXY_PORT=8080

# Filesystem paths (inside container)
GUI_DIST_PATH=/work/gui/dist
EXTENSION_SRC_PATH=/work/extensions
```

- [ ] **Step 6: Verify final state**

Run:
```bash
grep -rn "cursed\|CursedChrome\|cursedchrome" Dockerfile docker-compose.yaml deploy.sh .env.example umbra-server/internal/config/config.go 2>/dev/null
```

Expected: no matches.

---

### Task 16: Update `.gitignore` paths

**Files:**
- Modify: `.gitignore`

- [ ] **Step 1: Update Go-related paths**

The current `.gitignore` references `cursed-go/`. Update it to `umbra-server/`. Replace the section:

Old:
```
# Go build artifacts
cursed-go/bin/
cursed-go/cursed-server
cursed-go/deploy/cursed-server
cursed-go/deploy/gui-dist/
cursed-go/deploy/extensions/
```

New:
```
# Go build artifacts
umbra-server/bin/
umbra-server/umbra-server
umbra-server/deploy/umbra-server
umbra-server/deploy/gui-dist/
umbra-server/deploy/extensions/
```

- [ ] **Step 2: Verify**

```bash
grep "cursed" .gitignore
```

Expected: no matches.

---

### Task 17: README sweep — top-level + umbra-server + DEPLOY

**Files:**
- Modify: `README.md` (top-level)
- Modify: `umbra-server/README.md`
- Modify: `umbra-server/DEPLOY.md`
- Modify: `CLAUDE.md`

Note: this task only does mechanical string replacement and image link fixes. Phase 4 (Task 22) regenerates the top-level `README.md` as a proper OSS landing page.

- [ ] **Step 1: Bulk replace in markdown files**

Run:
```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
for f in README.md umbra-server/README.md umbra-server/DEPLOY.md CLAUDE.md; do
  [ -f "$f" ] || continue
  sed -i '' \
    -e 's|CursedChrome|Umbra|g' \
    -e 's|cursed-chrome|umbra-edr|g' \
    -e 's|cursed-go|umbra-server|g' \
    -e 's|cursed-server|umbra-server|g' \
    -e 's|cursedchrome|umbra|g' \
    "$f"
done
```

- [ ] **Step 2: Fix image references**

The README references `images/doll.svg`, `images/cursedchrome-diagram.png`, `images/cursed-chrome-web-panel.png`. These files no longer exist in the fork. Replace references:

In `README.md` find:
```markdown
<img src="./images/doll.svg" height="100" width="100" />
```
Replace with:
```markdown
<img src="./images/umbra.svg" height="100" width="100" />
```

For `cursedchrome-diagram.png` and `cursed-chrome-web-panel.png`, **delete the entire `<p align="center">...</p>` block that wraps them** (the diagrams will be regenerated in a follow-up; their absence is documented in the spec non-goals).

- [ ] **Step 3: Verify**

```bash
grep -rn "cursed\|CursedChrome\|doll.svg\|cursedchrome-diagram\|cursed-chrome-web-panel" README.md umbra-server/README.md umbra-server/DEPLOY.md CLAUDE.md 2>/dev/null
```

Expected: no matches.

---

### Task 18: End-to-end Phase 2 verification

**Files:** none (verification only)

- [ ] **Step 1: Final grep for `cursed` traces**

Run:
```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
grep -rln "cursed" \
  --include="*.json" --include="*.md" --include="*.go" --include="*.vue" \
  --include="*.ts" --include="*.js" --include="*.html" --include="*.css" \
  --include="*.yml" --include="*.yaml" --include="Dockerfile" --include="Makefile" \
  --include="*.sh" \
  . 2>/dev/null | grep -v node_modules
```

Expected: empty. If any file still has `cursed`, address inline (it's a missed mechanical edit) and re-run.

- [ ] **Step 2: Build umbra-server**

```bash
cd umbra-server && make build && make test && cd ..
```

Expected: `bin/umbra-server` exists, all tests pass.

- [ ] **Step 3: Build gui-next**

```bash
cd gui-next && npm run build && cd ..
```

Expected: `gui/dist/index.html` exists, no errors.

- [ ] **Step 4: Start the stack and smoke-test**

Run:
```bash
docker compose up -d --build
sleep 10
curl -sf http://localhost:8118/api/v1/health | head
docker compose logs umbra | tail -20
docker compose down -v
```

Expected: health endpoint returns 200, logs show no fatal errors. Container/volume names visible as `umbra*`, never `cursed*`. Bring stack down at the end.

- [ ] **Step 5: Visual extension check**

In Chrome:
1. Open `chrome://extensions/` with "Developer mode" on.
2. "Load unpacked" → select `~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr/extension/`.
3. Confirm extension shows as "Umbra Sensor" with the new icon.
4. Repeat for `cookie-sync-extension/` — should show as "ShadowLink".
5. Remove both unpacked extensions before continuing.

---

## Phase 3 — Sanitize

### Task 19: Run `opensource-sanitizer` agent

**Files:** none directly modified (the agent reports; we fix in Task 20)

- [ ] **Step 1: Dispatch the sanitizer agent**

Use the `Agent` tool with `subagent_type: "opensource-sanitizer"` and prompt:

```
Audit the open-source fork at ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr/ for
release-blocking issues. Report any:

- Hardcoded secrets, API keys, tokens, passwords (must be 0)
- Operator's local filesystem paths (e.g., /Users/s045pd/...)
- Operator's email (s045pd.x@gmail.com) outside LICENSE copyright
- Real production hostnames or LAN IP addresses (e.g., 192.168.x.x)
- Personal references in TODO/FIXME comments
- Remaining "cursed" / "CursedChrome" string traces
- Hidden files like .DS_Store
- Any file that looks operator-private rather than project-shared

Produce a PASS / FAIL / PASS-WITH-WARNINGS report. Categorize each finding by
severity (CRITICAL = blocks release, HIGH = should fix, LOW = optional). For each
finding, include the file path, line number, the offending content, and a
suggested fix. Do not modify any files — only report.
```

Wait for the agent's report.

- [ ] **Step 2: Save the report**

Save the agent's output to `/tmp/umbra-sanitizer-report.md` for reference during Task 20.

---

### Task 20: Address sanitizer findings

**Files:** depends on the sanitizer report

- [ ] **Step 1: Triage findings**

Read `/tmp/umbra-sanitizer-report.md`. For each finding:

- **CRITICAL** — must fix before publish (e.g., a leaked secret).
- **HIGH** — fix unless there's a clear reason not to.
- **LOW** — fix if it's a 1-line change; otherwise note for later.

- [ ] **Step 2: Apply fixes**

For each fix:
1. Open the offending file.
2. Apply the suggested fix or an equivalent.
3. After all fixes, verify with the same grep patterns the sanitizer used.

Common fixes:
- Replace `/Users/s045pd/path` with `~/path` or relative path.
- Replace `s045pd.x@gmail.com` (outside LICENSE) with placeholder or remove the line.
- Replace `192.168.x.x` with `localhost` or `your-server.example.com`.
- Delete found `.DS_Store` files: `find . -name ".DS_Store" -delete`.

- [ ] **Step 3: Re-run grep verification**

```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
grep -rln "/Users/s045pd\|s045pd\.x@\|192\.168\." . 2>/dev/null | grep -v LICENSE
```

Expected: empty.

- [ ] **Step 4: Re-dispatch sanitizer for confirmation**

Repeat Task 19 with the prompt addition: "This is a re-run after fixes. Confirm whether all prior CRITICAL and HIGH findings are now resolved."

Block on PASS or PASS-WITH-WARNINGS where all warnings have been individually accepted by the operator.

---

## Phase 4 — Package OSS

### Task 21: Run `opensource-packager` agent

**Files:**
- Create: `README.md` (regenerated)
- Update: `LICENSE` (copyright year + holder)
- Create: `CONTRIBUTING.md`
- Create: `SECURITY.md`
- Create: `setup.sh`
- Create: `.github/ISSUE_TEMPLATE/bug_report.md`
- Create: `.github/ISSUE_TEMPLATE/feature_request.md`
- Create: `.github/PULL_REQUEST_TEMPLATE.md`
- Update: `.github/workflows/Build&Push.yml` (audited)

- [ ] **Step 1: Dispatch the packager agent**

Use the `Agent` tool with `subagent_type: "opensource-packager"` and prompt:

```
Generate the open-source packaging files for the Umbra EDR project at
~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr/.

Project facts:
- Name: Umbra
- Tagline: Browser-layer EDR · Light through shadow
- License: MIT (copyright year 2026, holder s045pd)
- Stack: Go 1.25 server, Vue 3 + Vite + Tailwind v4 frontend, Chrome MV3 extensions
- Architecture: server (umbra-server/) + frontend (gui-next/, builds to gui/dist/) +
  extensions (extension/, cookie-sync-extension/, embed-targets/) + Docker Compose
- Ports: 8118 (API + GUI), 4343 (WebSocket), 8080 (HTTP forward proxy)
- Default DB: umbra / umbra / __set_a_strong_password__
- Public repo URL: https://github.com/s045pd/umbra-edr (about to be created)
- This is a dual-use security tool. The README and SECURITY.md MUST include a clear
  "Authorized monitoring environments only" disclaimer.

Generate:
1. README.md — top-level OSS landing page. Sections: tagline, what is it, screenshots
   placeholder, quick start (docker compose up), local dev, architecture, env vars
   table, deployment, contributing pointer, license, AUTHORIZED USE ONLY notice.
2. CONTRIBUTING.md — branch model, PR template pointer, test expectations, code style.
3. SECURITY.md — vulnerability reporting policy + AUTHORIZED USE ONLY disclaimer.
4. LICENSE — MIT, copyright "Copyright (c) 2026 s045pd". Update existing file in place.
5. setup.sh — one-shot dev environment bootstrap. Validates Go ≥1.25, Node ≥20, docker,
   rsvg-convert; copies .env.example to .env; runs go mod download; runs npm install
   in gui-next/; prints next steps.
6. .github/ISSUE_TEMPLATE/bug_report.md, feature_request.md.
7. .github/PULL_REQUEST_TEMPLATE.md.
8. Audit .github/workflows/Build&Push.yml — replace any 'cursed' references and any
   private registry tokens with placeholders.

Do not include the existing CLAUDE.md regeneration in scope (Task 22 handles it).
```

Wait for the agent to complete its work and report what it generated.

- [ ] **Step 2: Spot-check the regenerated README**

Read the new `README.md`. Confirm:
- The hero/tagline is present.
- Quick start works (mentally trace `docker compose up`).
- The "Authorized Use Only" disclaimer is prominent (top of README and in SECURITY.md).
- All asset references use `images/umbra*.png` or `images/umbra.svg`, not the deleted images.
- No `cursed` strings remain.

- [ ] **Step 3: Verify file presence**

```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
ls -la README.md LICENSE CONTRIBUTING.md SECURITY.md setup.sh
ls -la .github/ISSUE_TEMPLATE/ .github/PULL_REQUEST_TEMPLATE.md .github/workflows/
```

Expected: every file listed in this task's "Files" header exists.

- [ ] **Step 4: Test setup.sh on a clean shell**

```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
bash setup.sh
```

Expected: prerequisites validated, `.env` created from `.env.example`, dependencies fetched, completion message printed. (If the script edits files in unexpected ways, review and adjust.)

---

### Task 22: Prune `CLAUDE.md`

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Open and review current `CLAUDE.md`**

The file is project guidance for Claude Code. Currently it describes the codebase from an internal perspective. Review for:
- Operator-private context ("when I deploy I use Portainer at..." → genericize)
- References to `cursed` strings (should already be fixed by Task 17, double-check)
- Operator-specific paths
- Anything that reads as personal-project rather than open-source-project

- [ ] **Step 2: Rewrite as public dev guide**

Replace contents with a clean OSS-friendly version. The file should:
- Start with a 1-line product description: "Umbra is a browser-layer EDR..."
- Describe architecture (server, GUI, extensions) concisely.
- List dev commands for each component.
- List testing commands.
- Note any non-obvious development patterns (e.g., "the GUI builds into ../gui/dist so the Go server can serve from a stable path").
- NOT include private deployment details, operator email, or LAN IPs.

A good target length is 80–150 lines.

- [ ] **Step 3: Verify**

```bash
grep -n "s045pd\.x@\|192\.168\|/Users/\|cursed" CLAUDE.md
```

Expected: empty (the only `s045pd` reference acceptable is the GitHub repo URL `s045pd/umbra-edr`).

---

### Task 23: Final `.gitignore` audit

**Files:**
- Modify: `.gitignore` (if needed)

- [ ] **Step 1: Confirm `.gitignore` covers all build outputs and secrets**

The `.gitignore` should include:

```
# Secrets and operator-specific config
.env
.env.*
*.key
*.crt
.chrome-data/
ssl/

# OS / IDE
.DS_Store
.idea/
.vscode/

# Build outputs
gui/dist/
dist/
node_modules/
*.test
*.out
coverage.txt
umbra-server/bin/
umbra-server/umbra-server
umbra-server/deploy/umbra-server
umbra-server/deploy/gui-dist/
umbra-server/deploy/extensions/

# Optional third-party drop-in (operator brings their own)
bypass-paywalls-chrome/
bypass-paywalls-chrome.zip
external_extension/
```

Compare against the current `.gitignore` and add any missing lines.

- [ ] **Step 2: Verify nothing sensitive would be tracked**

Simulate a fresh `git init` to see what would be added:

```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
git init -q
git add -A --dry-run 2>&1 | grep -E "\.env$|\.DS_Store$|node_modules|chrome-data|/ssl/|bypass-paywalls" | head
rm -rf .git
```

Expected: dry-run output is empty for all the listed patterns (none would be added). Cleanup `.git/` afterwards because Phase 5 redoes the init properly.

---

## Phase 5 — Publish

### Task 24: Operator review gate

**Files:** none

- [ ] **Step 1: Stop and request operator review**

This task is a manual halt. The implementation agent must NOT proceed past this point without explicit operator confirmation.

Print to the operator (in the conversation):

```
Phase 1–4 complete. The local fork at ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr/
is ready for publish. Before I run `gh repo create` and push, please:

1. Open the directory in a file browser or `tree umbra-edr -L 2 | head -50`.
2. Spot-check README.md, SECURITY.md, manifest.json files.
3. Run `cd umbra-edr && bash setup.sh && cd umbra-server && make build` for one end-to-end sanity check.
4. Confirm the public repo target is `s045pd/umbra-edr` and the visibility is PUBLIC.

Reply "ship it" to proceed. Reply with any corrections you want made first.
```

Wait for operator response. Do not proceed on silence.

---

### Task 25: Initial git commit on the fork

**Files:**
- Create: `umbra-edr/.git/` (git repo)

- [ ] **Step 1: Initialize fresh git history**

```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
git init -b main
git status --short | head -20
```

Expected: `git init` succeeds, branch is `main`, status shows all files as untracked.

- [ ] **Step 2: Stage everything**

```bash
git add -A
git status --short | wc -l
```

Note the file count. Skim the staged file list with `git diff --cached --stat | tail -20` for sanity.

- [ ] **Step 3: First commit**

```bash
git commit -m "$(cat <<'EOF'
feat: initial Umbra release

Browser-layer EDR for authorized enterprise monitoring. Forked from a
private predecessor and rebranded; this is the first public release.

Architecture: Go server, Vue 3 + Vite frontend, Chrome MV3 extensions,
Docker Compose deployment.

Authorized monitoring environments only. See SECURITY.md.
EOF
)"
```

Expected: single commit on `main` with the initial tree.

---

### Task 26: Create public GitHub repo and push

**Files:** none locally; remote `s045pd/umbra-edr` created

- [ ] **Step 1: Confirm `gh` is authenticated to the right account**

```bash
gh auth status
```

Expected: logged in as `s045pd` on `github.com`. If not logged in, run `gh auth login` and complete the flow.

- [ ] **Step 2: Verify the target repo name is available**

```bash
gh repo view s045pd/umbra-edr 2>&1 | head -3
```

Expected: `GraphQL: Could not resolve to a Repository...` (i.e., does not exist yet).

- [ ] **Step 3: Create the repo and push**

```bash
cd ~/workobj/BackupData/DocumentArchive/GitHub/umbra-edr
gh repo create s045pd/umbra-edr \
  --public \
  --description "Umbra · Browser-layer EDR · Light through shadow. Authorized monitoring only." \
  --source=. \
  --remote=origin \
  --push
```

Expected: repo created, initial commit pushed to `origin/main`. URL printed.

- [ ] **Step 4: Add topics**

```bash
gh repo edit s045pd/umbra-edr \
  --add-topic edr \
  --add-topic browser-monitoring \
  --add-topic chrome-extension \
  --add-topic go \
  --add-topic vue
```

- [ ] **Step 5: Verify the repo is live**

```bash
gh repo view s045pd/umbra-edr
```

Expected: shows the repo with description, topics, and latest commit.

Open https://github.com/s045pd/umbra-edr in a browser and confirm the README renders, the Umbra icon appears in image references, and the AUTHORIZED USE notice is prominent.

---

### Task 27: Smoke-test the published repo

**Files:** temporary clone in `/tmp`

- [ ] **Step 1: Clone fresh and run setup**

```bash
mkdir -p /tmp/umbra-smoke && cd /tmp/umbra-smoke
git clone https://github.com/s045pd/umbra-edr.git
cd umbra-edr
bash setup.sh
```

Expected: clone succeeds, setup completes. No "command not found" or missing-prereq errors that a reasonable open-source consumer wouldn't hit.

- [ ] **Step 2: Build the server**

```bash
cd umbra-server && make build
ls bin/
```

Expected: `bin/umbra-server` produced.

- [ ] **Step 3: Build the GUI**

```bash
cd ../gui-next && npm install && npm run build
ls ../gui/dist/
```

Expected: build succeeds, `dist/index.html` produced.

- [ ] **Step 4: Cleanup**

```bash
cd /tmp && rm -rf umbra-smoke
```

- [ ] **Step 5: Final report to operator**

Reply in the conversation:

```
Umbra-EDR is live at https://github.com/s045pd/umbra-edr.

- Public repo, MIT licensed, Authorized Use disclaimer in README + SECURITY.md.
- Smoke-tested clone-and-build from a clean /tmp directory.
- Original s045pd/CursedChrome repo untouched and still private.

Next steps (out of scope of this plan):
- Re-create the architecture diagram in the new visual style.
- Optional: dedicated ShadowLink mark distinct from Umbra.
- Optional: write a CHANGELOG and tag v0.1.0.
```

---

## Self-review

The plan was self-reviewed against the spec at write time. All seven sections of the spec map to tasks:

- §3.1 Naming → Tasks 8–10, 12–17 (display strings + manifests)
- §3.2 Visual identity → Tasks 8–9 (SVG promotion + PNG renders + extension icon swap)
- §3.3 Design system → Task 10 (Tailwind v4 `@theme` block) + Task 11 (font load) + Task 12 (brand strings)
- §4.1 Layer 1 display strings → Tasks 12, 13, 14, 17
- §4.2 Layer 2 source identifiers → Tasks 4, 5, 6, 7
- §4.3 Layer 3 deployment defaults → Task 15
- §4.4 Layer 4 assets → Tasks 1 (exclusion), 8 (master + renders), 9 (extension icon swap)
- §5.1 Phase 1 fork → Tasks 1, 2, 3
- §5.2 Phase 2 rebrand → Tasks 4–18
- §5.3 Phase 3 sanitize → Tasks 19, 20
- §5.4 Phase 4 package → Tasks 21, 22, 23
- §5.5 Phase 5 publish → Tasks 24, 25, 26, 27
- §5.6 Sequencing & rollback → embedded in task ordering and the Task 24 review gate
- §5.7 Risk register → addressed by exclusion list (Task 1), sanitizer agent (Task 19), and review gate (Task 24)
- §6 Deliverables → cross-referenced in each task's Files header
- §7 Items deferred to plan: PNG render parameters fixed in Task 8 (rsvg-convert), SVG→PNG tool decision likewise fixed, default branch name fixed in Task 25 (`main`)

No placeholders remain. Method/file references are consistent across tasks (e.g., the binary is `umbra-server` everywhere, the master icon is `images/umbra.svg` everywhere, the GitHub URL is `s045pd/umbra-edr` everywhere).
