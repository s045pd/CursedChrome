# cursed-gui

Modern rewrite of the CursedChrome admin panel.

## Stack

- **Vite 5** — fast dev server, ESM-native build
- **Vue 3.5** — Composition API + `<script setup>` + TypeScript
- **Pinia** — state management
- **Tailwind CSS v4** — design tokens + utility-first styling
- **VueUse / wavesurfer.js / ECharts** — supporting libs

## Layout

```
src/
├── api/             # fetch wrapper + endpoint declarations
├── stores/          # auth, bots
├── pages/           # Login / Dashboard / BotDetail / Settings
├── layouts/         # AppShell
├── components/
│   ├── ui/          # Btn, Field, Drawer, StatusDot
│   ├── bot/         # BotRow
│   └── data/        # Tabs, History, Cookies, Bookmarks, Downloads,
│                    # Screenshots, Keyboard, Audio, Activity, Remote, Config
├── composables/     # useTime, useClipboard
├── types/           # api.ts (single source of truth for response shapes)
├── router/
└── assets/          # design tokens + global styles
```

Total: 2.7k LOC across 32 files (replaces 4k LOC monolith in `gui/`).

## Design

Single-accent dark theme (Linear/Datadog inspired). High data density,
tabular-numeric monospace for IDs and timestamps, subtle borders,
single accent color for emphasis. Light theme available via `html.light`.

All colors live in `src/assets/styles.css` as `@theme` tokens — no hex
literals scattered through templates.

## Commands

```bash
npm install
npm run dev          # dev server on :5173, proxies /api → 192.168.13.202:8118
npm run typecheck    # vue-tsc
npm run build        # writes to ../gui/dist (consumed by cursed-go)
```

The build output lives under `../gui/dist` so the existing
`cursed-go/scripts/deploy.sh` picks it up automatically — no separate
deploy step.

## Routes

| Route | Component |
|---|---|
| `/login` | Login |
| `/` | Dashboard (bot list + filters + bulk actions) |
| `/bots/:id` | BotDetail with 11 tabs |
| `/settings` | Account password rotation + CA download |

## Bot detail tabs

1. **Tabs** — open browser tabs with favicon + active marker
2. **History** — chronological with visit counts + filter
3. **Cookies** — searchable, copy as JSON or Netscape format
4. **Bookmarks** — link-out
5. **Downloads** — filename + state + size
6. **Screenshots** — grouped by day, lightbox preview
7. **Keyboard** — search with highlight, time range
8. **Audio** — start/stop recording, wavesurfer playback per session
9. **Activity** — heatmap-style 30-day activity grid
10. **Remote control** — push the bot to a URL
11. **Config** — edit name, proxy creds, telemetry switches

## Compared to old `gui/`

| | Old `gui/` (Vue 2) | New `gui-next/` (Vue 3) |
|---|---|---|
| LOC | 4086 | 2749 |
| Files | 19 | 32 (smaller, focused) |
| Largest file | Main.vue 923 | BotDetail.vue 230 |
| Build tool | Vue CLI 4 (slow) | Vite 5 (fast) |
| Types | none (JS) | TypeScript strict |
| State | mixed `data()` + props | Pinia store |
| Modal renders | Bootstrap-Vue (broken backdrop, raw JSON overflow) | dedicated route + page |
| Theme | Bootstrap defaults | design tokens, dark/light toggle |
| Bundle gz | ~250 KB | 104 KB main + lazy chunks |
