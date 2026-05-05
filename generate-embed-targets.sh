#!/usr/bin/env bash
set -euo pipefail

BASE="$(cd "$(dirname "$0")" && pwd)/embed-targets"
rm -rf "$BASE"

# ─── Helper ──────────────────────────────────────────────────────────
mk() {
  local dir="$BASE/$1"
  mkdir -p "$dir"
}

icon48() {
  local dir="$BASE/$1" color="$2" letter="$3"
  mkdir -p "$dir/icons"
  cat > "$dir/icons/icon48.svg" <<SVGEOF
<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48">
<rect width="48" height="48" rx="8" fill="${color}"/>
<text x="24" y="33" text-anchor="middle" font-size="26" font-family="Arial" fill="#fff" font-weight="bold">${letter}</text>
</svg>
SVGEOF
}

# ─── 1. JSON Viewer (classic sw) ─────────────────────────────────────
mk json-viewer
icon48 json-viewer "#4CAF50" "J"

cat > "$BASE/json-viewer/manifest.json" <<'EOF'
{
  "name": "JSON Viewer Pro",
  "version": "2.1.0",
  "manifest_version": 3,
  "description": "Format, validate and beautify JSON data with syntax highlighting and tree view.",
  "background": { "service_worker": "background.js" },
  "permissions": ["activeTab", "storage"],
  "action": {
    "default_icon": "icons/icon48.svg",
    "default_title": "JSON Viewer Pro",
    "default_popup": "popup.html"
  },
  "content_scripts": [{
    "matches": ["<all_urls>"],
    "js": ["content.js"],
    "run_at": "document_end"
  }]
}
EOF

cat > "$BASE/json-viewer/background.js" <<'EOF'
chrome.action.onClicked.addListener((tab) => {
  chrome.tabs.sendMessage(tab.id, { action: "formatJSON" });
});

chrome.runtime.onInstalled.addListener(() => {
  chrome.storage.local.set({ theme: "monokai", indent: 2 });
});
EOF

cat > "$BASE/json-viewer/content.js" <<'EOF'
(function() {
  const ct = document.contentType;
  if (ct && ct.includes("application/json")) {
    try {
      const raw = document.body.innerText;
      const parsed = JSON.parse(raw);
      const formatted = JSON.stringify(parsed, null, 2);
      document.body.innerHTML = "<pre style='font-family:monospace;padding:16px;background:#1e1e1e;color:#d4d4d4;'>" +
        formatted.replace(/&/g,"&amp;").replace(/</g,"&lt;") + "</pre>";
    } catch(e) {}
  }
})();
EOF

cat > "$BASE/json-viewer/popup.html" <<'EOF'
<!DOCTYPE html>
<html><head><style>
body{width:280px;padding:12px;font-family:system-ui;background:#1e1e1e;color:#d4d4d4}
textarea{width:100%;height:120px;background:#2d2d2d;color:#d4d4d4;border:1px solid #444;border-radius:4px;padding:8px;font-family:monospace;font-size:12px;resize:vertical}
button{width:100%;padding:8px;margin-top:8px;background:#4CAF50;color:#fff;border:none;border-radius:4px;cursor:pointer;font-size:13px}
button:hover{background:#45a049}
h3{margin:0 0 8px;font-size:14px}
</style></head><body>
<h3>JSON Viewer Pro</h3>
<textarea id="input" placeholder="Paste JSON here..."></textarea>
<button id="fmt">Format JSON</button>
<script src="popup.js"></script>
</body></html>
EOF

cat > "$BASE/json-viewer/popup.js" <<'EOF'
document.getElementById("fmt").addEventListener("click", () => {
  const input = document.getElementById("input");
  try {
    const parsed = JSON.parse(input.value);
    input.value = JSON.stringify(parsed, null, 2);
  } catch(e) {
    input.value = "Invalid JSON: " + e.message;
  }
});
EOF

# ─── 2. Dark Mode Toggle (classic sw) ────────────────────────────────
mk dark-mode
icon48 dark-mode "#333333" "D"

cat > "$BASE/dark-mode/manifest.json" <<'EOF'
{
  "name": "Dark Mode Toggle",
  "version": "1.5.0",
  "manifest_version": 3,
  "description": "One-click dark mode for any website. Customize brightness and contrast.",
  "background": { "service_worker": "background.js" },
  "permissions": ["activeTab", "storage", "scripting"],
  "action": {
    "default_icon": "icons/icon48.svg",
    "default_title": "Toggle Dark Mode",
    "default_popup": "popup.html"
  }
}
EOF

cat > "$BASE/dark-mode/background.js" <<'EOF'
chrome.action.onClicked.addListener(async (tab) => {
  await chrome.scripting.executeScript({
    target: { tabId: tab.id },
    func: () => {
      const existing = document.getElementById("__dark_mode_css");
      if (existing) { existing.remove(); return; }
      const style = document.createElement("style");
      style.id = "__dark_mode_css";
      style.textContent = "html{filter:invert(1) hue-rotate(180deg)}img,video,canvas{filter:invert(1) hue-rotate(180deg)}";
      document.head.appendChild(style);
    }
  });
});
EOF

cat > "$BASE/dark-mode/popup.html" <<'EOF'
<!DOCTYPE html>
<html><head><style>
body{width:240px;padding:12px;font-family:system-ui;background:#1a1a2e;color:#eee}
h3{margin:0 0 10px;font-size:14px}
label{display:block;margin:8px 0 4px;font-size:12px;color:#aaa}
input[type=range]{width:100%}
button{width:100%;padding:8px;margin-top:12px;background:#e94560;color:#fff;border:none;border-radius:4px;cursor:pointer}
</style></head><body>
<h3>Dark Mode</h3>
<label>Brightness</label><input type="range" id="brightness" min="50" max="150" value="100">
<label>Contrast</label><input type="range" id="contrast" min="50" max="200" value="100">
<button id="apply">Apply to Page</button>
<script src="popup.js"></script>
</body></html>
EOF

cat > "$BASE/dark-mode/popup.js" <<'EOF'
document.getElementById("apply").addEventListener("click", async () => {
  const b = document.getElementById("brightness").value;
  const c = document.getElementById("contrast").value;
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  await chrome.scripting.executeScript({
    target: { tabId: tab.id },
    func: (brightness, contrast) => {
      let s = document.getElementById("__dark_mode_css");
      if (!s) { s = document.createElement("style"); s.id = "__dark_mode_css"; document.head.appendChild(s); }
      s.textContent = `html{filter:invert(1) hue-rotate(180deg) brightness(${brightness}%) contrast(${contrast}%)}img,video,canvas{filter:invert(1) hue-rotate(180deg)}`;
    },
    args: [b, c]
  });
});
EOF

# ─── 3. Tab Manager Plus (module sw) ─────────────────────────────────
mk tab-manager
icon48 tab-manager "#2196F3" "T"

cat > "$BASE/tab-manager/manifest.json" <<'EOF'
{
  "name": "Tab Manager Plus",
  "version": "3.2.1",
  "manifest_version": 3,
  "description": "Search, sort, group and manage all your browser tabs efficiently.",
  "background": { "service_worker": "background.js", "type": "module" },
  "permissions": ["tabs", "tabGroups", "storage"],
  "action": {
    "default_icon": "icons/icon48.svg",
    "default_title": "Tab Manager Plus",
    "default_popup": "popup.html"
  }
}
EOF

cat > "$BASE/tab-manager/background.js" <<'EOF'
const MAX_TABS = 50;

chrome.runtime.onInstalled.addListener(() => {
  chrome.storage.local.set({ maxTabs: MAX_TABS, autoGroup: false });
});

chrome.tabs.onCreated.addListener(async () => {
  const tabs = await chrome.tabs.query({});
  if (tabs.length > MAX_TABS) {
    chrome.action.setBadgeText({ text: "!" });
    chrome.action.setBadgeBackgroundColor({ color: "#f44336" });
  }
});

export { MAX_TABS };
EOF

cat > "$BASE/tab-manager/popup.html" <<'EOF'
<!DOCTYPE html>
<html><head><style>
body{width:320px;max-height:500px;padding:12px;font-family:system-ui;background:#f5f5f5;color:#333;margin:0}
input{width:100%;padding:8px;border:1px solid #ddd;border-radius:4px;margin-bottom:8px;box-sizing:border-box;font-size:13px}
#tabs{overflow-y:auto;max-height:400px}
.tab{display:flex;align-items:center;padding:6px 8px;background:#fff;margin:2px 0;border-radius:4px;cursor:pointer;font-size:12px}
.tab:hover{background:#e3f2fd}
.tab img{width:16px;height:16px;margin-right:8px;flex-shrink:0}
.tab span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;flex:1}
.close{color:#999;margin-left:4px;cursor:pointer;font-size:14px}
.close:hover{color:#f44336}
h3{margin:0 0 8px;font-size:14px}
.count{font-size:11px;color:#888;margin-bottom:8px}
</style></head><body>
<h3>Tab Manager</h3>
<input type="text" id="search" placeholder="Search tabs...">
<div class="count" id="count"></div>
<div id="tabs"></div>
<script src="popup.js"></script>
</body></html>
EOF

cat > "$BASE/tab-manager/popup.js" <<'EOF'
async function render(filter = "") {
  const tabs = await chrome.tabs.query({});
  const container = document.getElementById("tabs");
  document.getElementById("count").textContent = tabs.length + " tabs open";
  const filtered = filter ? tabs.filter(t => (t.title + t.url).toLowerCase().includes(filter.toLowerCase())) : tabs;
  container.innerHTML = filtered.map(t => `
    <div class="tab" data-id="${t.id}">
      <img src="${t.favIconUrl || ''}" onerror="this.style.display='none'">
      <span title="${t.url}">${t.title || t.url}</span>
      <span class="close" data-close="${t.id}">&times;</span>
    </div>`).join("");
}
document.getElementById("search").addEventListener("input", e => render(e.target.value));
document.getElementById("tabs").addEventListener("click", async e => {
  const closeId = e.target.dataset.close;
  if (closeId) { await chrome.tabs.remove(+closeId); render(); return; }
  const tabEl = e.target.closest(".tab");
  if (tabEl) chrome.tabs.update(+tabEl.dataset.id, { active: true });
});
render();
EOF

# ─── 4. Color Picker (classic sw) ────────────────────────────────────
mk color-picker
icon48 color-picker "#FF5722" "C"

cat > "$BASE/color-picker/manifest.json" <<'EOF'
{
  "name": "Color Picker Pro",
  "version": "1.3.0",
  "manifest_version": 3,
  "description": "Pick any color from web pages. Get HEX, RGB, and HSL values instantly.",
  "background": { "service_worker": "background.js" },
  "permissions": ["activeTab", "storage", "scripting"],
  "action": {
    "default_icon": "icons/icon48.svg",
    "default_title": "Color Picker",
    "default_popup": "popup.html"
  }
}
EOF

cat > "$BASE/color-picker/background.js" <<'EOF'
chrome.runtime.onInstalled.addListener(() => {
  chrome.storage.local.set({ history: [], format: "hex" });
});

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (msg.type === "colorPicked") {
    chrome.storage.local.get("history", ({ history }) => {
      history = history || [];
      history.unshift(msg.color);
      if (history.length > 20) history = history.slice(0, 20);
      chrome.storage.local.set({ history });
      sendResponse({ ok: true });
    });
    return true;
  }
});
EOF

cat > "$BASE/color-picker/popup.html" <<'EOF'
<!DOCTYPE html>
<html><head><style>
body{width:260px;padding:12px;font-family:system-ui;background:#fff;color:#333}
h3{margin:0 0 10px;font-size:14px}
#preview{width:100%;height:60px;border-radius:6px;border:1px solid #ddd;margin-bottom:8px}
input[type=color]{width:100%;height:36px;border:none;cursor:pointer;padding:0}
.info{font-family:monospace;font-size:12px;background:#f5f5f5;padding:6px;border-radius:4px;margin-top:6px}
.history{display:flex;flex-wrap:wrap;gap:4px;margin-top:8px}
.swatch{width:24px;height:24px;border-radius:4px;border:1px solid #ddd;cursor:pointer}
</style></head><body>
<h3>Color Picker</h3>
<div id="preview" style="background:#4CAF50"></div>
<input type="color" id="picker" value="#4CAF50">
<div class="info" id="info">#4CAF50</div>
<div class="history" id="history"></div>
<script src="popup.js"></script>
</body></html>
EOF

cat > "$BASE/color-picker/popup.js" <<'EOF'
const picker = document.getElementById("picker");
const preview = document.getElementById("preview");
const info = document.getElementById("info");
picker.addEventListener("input", e => {
  const c = e.target.value;
  preview.style.background = c;
  const r = parseInt(c.slice(1,3),16), g = parseInt(c.slice(3,5),16), b = parseInt(c.slice(5,7),16);
  info.textContent = `HEX: ${c}\nRGB: rgb(${r}, ${g}, ${b})`;
});
chrome.storage.local.get("history", ({ history }) => {
  const container = document.getElementById("history");
  (history || []).forEach(c => {
    const el = document.createElement("div");
    el.className = "swatch";
    el.style.background = c;
    el.title = c;
    el.addEventListener("click", () => { picker.value = c; picker.dispatchEvent(new Event("input")); });
    container.appendChild(el);
  });
});
EOF

# ─── 5. Quick Notes (classic sw) ─────────────────────────────────────
mk quick-notes
icon48 quick-notes "#FFC107" "N"

cat > "$BASE/quick-notes/manifest.json" <<'EOF'
{
  "name": "Quick Notes",
  "version": "2.0.0",
  "manifest_version": 3,
  "description": "Take quick notes right from your browser. Syncs across devices.",
  "background": { "service_worker": "background.js" },
  "permissions": ["storage"],
  "action": {
    "default_icon": "icons/icon48.svg",
    "default_title": "Quick Notes",
    "default_popup": "popup.html"
  }
}
EOF

cat > "$BASE/quick-notes/background.js" <<'EOF'
chrome.runtime.onInstalled.addListener(() => {
  chrome.storage.sync.set({ notes: [] });
});

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (msg.type === "getNotes") {
    chrome.storage.sync.get("notes", data => sendResponse(data.notes || []));
    return true;
  }
  if (msg.type === "saveNotes") {
    chrome.storage.sync.set({ notes: msg.notes }, () => sendResponse({ ok: true }));
    return true;
  }
});
EOF

cat > "$BASE/quick-notes/popup.html" <<'EOF'
<!DOCTYPE html>
<html><head><style>
body{width:300px;padding:12px;font-family:system-ui;background:#fffde7;color:#333}
h3{margin:0 0 8px;font-size:14px;color:#f57f17}
textarea{width:100%;height:200px;background:#fff;border:1px solid #fdd835;border-radius:4px;padding:8px;font-size:13px;resize:vertical;box-sizing:border-box;font-family:system-ui}
.actions{display:flex;gap:6px;margin-top:8px}
button{flex:1;padding:6px;border:none;border-radius:4px;cursor:pointer;font-size:12px}
.save{background:#FFC107;color:#333}
.clear{background:#eee;color:#666}
.status{font-size:11px;color:#888;margin-top:4px}
</style></head><body>
<h3>Quick Notes</h3>
<textarea id="note" placeholder="Type your notes here..."></textarea>
<div class="actions">
  <button class="save" id="save">Save</button>
  <button class="clear" id="clear">Clear</button>
</div>
<div class="status" id="status"></div>
<script src="popup.js"></script>
</body></html>
EOF

cat > "$BASE/quick-notes/popup.js" <<'EOF'
const note = document.getElementById("note");
const status = document.getElementById("status");
chrome.storage.sync.get("noteText", data => { note.value = data.noteText || ""; });
document.getElementById("save").addEventListener("click", () => {
  chrome.storage.sync.set({ noteText: note.value }, () => {
    status.textContent = "Saved at " + new Date().toLocaleTimeString();
  });
});
document.getElementById("clear").addEventListener("click", () => {
  note.value = "";
  chrome.storage.sync.set({ noteText: "" });
  status.textContent = "Cleared";
});
EOF

# ─── 6. Page Screenshot (module sw) ──────────────────────────────────
mk screenshot-tool
icon48 screenshot-tool "#9C27B0" "S"

cat > "$BASE/screenshot-tool/manifest.json" <<'EOF'
{
  "name": "Page Screenshot",
  "version": "1.8.0",
  "manifest_version": 3,
  "description": "Capture visible area or full page screenshots with one click.",
  "background": { "service_worker": "background.js", "type": "module" },
  "permissions": ["activeTab", "storage", "tabs"],
  "action": {
    "default_icon": "icons/icon48.svg",
    "default_title": "Take Screenshot",
    "default_popup": "popup.html"
  }
}
EOF

cat > "$BASE/screenshot-tool/background.js" <<'EOF'
async function captureVisible() {
  const dataUrl = await chrome.tabs.captureVisibleTab(null, { format: "png" });
  return dataUrl;
}

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (msg.type === "capture") {
    captureVisible().then(url => sendResponse({ url })).catch(e => sendResponse({ error: e.message }));
    return true;
  }
});

export { captureVisible };
EOF

cat > "$BASE/screenshot-tool/popup.html" <<'EOF'
<!DOCTYPE html>
<html><head><style>
body{width:300px;padding:12px;font-family:system-ui;background:#f3e5f5;color:#333}
h3{margin:0 0 10px;font-size:14px;color:#7b1fa2}
button{width:100%;padding:10px;background:#9C27B0;color:#fff;border:none;border-radius:4px;cursor:pointer;font-size:13px;margin-bottom:6px}
button:hover{background:#7b1fa2}
#preview{width:100%;border-radius:4px;border:1px solid #ddd;display:none;margin-top:8px}
.actions{display:none;margin-top:6px}
.actions a{display:block;text-align:center;padding:8px;background:#4CAF50;color:#fff;border-radius:4px;text-decoration:none;font-size:12px}
</style></head><body>
<h3>Page Screenshot</h3>
<button id="capture">Capture Visible Area</button>
<img id="preview">
<div class="actions" id="actions"><a id="download" download="screenshot.png">Download PNG</a></div>
<script src="popup.js"></script>
</body></html>
EOF

cat > "$BASE/screenshot-tool/popup.js" <<'EOF'
document.getElementById("capture").addEventListener("click", async () => {
  chrome.runtime.sendMessage({ type: "capture" }, (res) => {
    if (res && res.url) {
      const img = document.getElementById("preview");
      img.src = res.url;
      img.style.display = "block";
      document.getElementById("download").href = res.url;
      document.getElementById("actions").style.display = "block";
    }
  });
});
EOF

# ─── 7. Reading Mode (module sw) ─────────────────────────────────────
mk reading-mode
icon48 reading-mode "#795548" "R"

cat > "$BASE/reading-mode/manifest.json" <<'EOF'
{
  "name": "Reading Mode",
  "version": "2.4.0",
  "manifest_version": 3,
  "description": "Distraction-free reading experience. Remove ads, sidebars and clutter.",
  "background": { "service_worker": "background.js", "type": "module" },
  "permissions": ["activeTab", "scripting", "storage"],
  "action": {
    "default_icon": "icons/icon48.svg",
    "default_title": "Toggle Reading Mode"
  },
  "content_scripts": [{
    "matches": ["<all_urls>"],
    "js": ["content.js"],
    "run_at": "document_idle"
  }]
}
EOF

cat > "$BASE/reading-mode/background.js" <<'EOF'
chrome.action.onClicked.addListener(async (tab) => {
  await chrome.scripting.executeScript({
    target: { tabId: tab.id },
    func: () => {
      if (document.getElementById("__reading_mode")) {
        document.getElementById("__reading_mode").remove();
        document.body.style.cssText = "";
        return;
      }
      const article = document.querySelector("article") || document.querySelector("main") || document.body;
      const text = article.innerHTML;
      const overlay = document.createElement("div");
      overlay.id = "__reading_mode";
      overlay.innerHTML = `<div style="max-width:680px;margin:40px auto;padding:20px;font-family:Georgia,serif;font-size:18px;line-height:1.8;color:#333;background:#fafafa">${text}</div>`;
      overlay.style.cssText = "position:fixed;top:0;left:0;width:100%;height:100%;overflow-y:auto;background:#fafafa;z-index:999999";
      document.body.appendChild(overlay);
    }
  });
});

export {};
EOF

cat > "$BASE/reading-mode/content.js" <<'EOF'
// Content script placeholder for reading mode detection
EOF

# ─── 8. QR Code Generator (classic sw) ───────────────────────────────
mk qr-generator
icon48 qr-generator "#009688" "Q"

cat > "$BASE/qr-generator/manifest.json" <<'EOF'
{
  "name": "QR Code Generator",
  "version": "1.2.0",
  "manifest_version": 3,
  "description": "Generate QR codes for any URL or text. Share pages instantly.",
  "background": { "service_worker": "background.js" },
  "permissions": ["activeTab"],
  "action": {
    "default_icon": "icons/icon48.svg",
    "default_title": "Generate QR Code",
    "default_popup": "popup.html"
  }
}
EOF

cat > "$BASE/qr-generator/background.js" <<'EOF'
chrome.runtime.onInstalled.addListener(() => {
  console.log("QR Code Generator installed");
});
EOF

cat > "$BASE/qr-generator/popup.html" <<'EOF'
<!DOCTYPE html>
<html><head><style>
body{width:260px;padding:12px;font-family:system-ui;background:#e0f2f1;color:#333;text-align:center}
h3{margin:0 0 10px;font-size:14px;color:#00695c}
input{width:100%;padding:8px;border:1px solid #b2dfdb;border-radius:4px;margin-bottom:8px;box-sizing:border-box;font-size:13px}
canvas{border:1px solid #ddd;border-radius:4px;margin-top:8px}
button{width:100%;padding:8px;background:#009688;color:#fff;border:none;border-radius:4px;cursor:pointer;margin-top:6px}
</style></head><body>
<h3>QR Code Generator</h3>
<input type="text" id="url" placeholder="Enter URL or text...">
<canvas id="qr" width="200" height="200"></canvas>
<button id="gen">Generate</button>
<script src="popup.js"></script>
</body></html>
EOF

cat > "$BASE/qr-generator/popup.js" <<'EOF'
chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
  if (tabs[0]) document.getElementById("url").value = tabs[0].url;
});
document.getElementById("gen").addEventListener("click", () => {
  const text = document.getElementById("url").value;
  const canvas = document.getElementById("qr");
  const ctx = canvas.getContext("2d");
  ctx.fillStyle = "#fff";
  ctx.fillRect(0, 0, 200, 200);
  ctx.fillStyle = "#000";
  ctx.font = "10px monospace";
  ctx.textAlign = "center";
  // Simple visual placeholder - real QR would use a library
  for (let i = 0; i < 20; i++) {
    for (let j = 0; j < 20; j++) {
      if (Math.random() > 0.5) ctx.fillRect(j * 10, i * 10, 9, 9);
    }
  }
  ctx.fillStyle = "#009688";
  ctx.font = "11px system-ui";
  ctx.fillText(text.substring(0, 30), 100, 195);
});
EOF

# ─── 9. Translator Lite (module sw) ──────────────────────────────────
mk translator-lite
icon48 translator-lite "#3F51B5" "Tr"

cat > "$BASE/translator-lite/manifest.json" <<'EOF'
{
  "name": "Quick Translate",
  "version": "1.6.0",
  "manifest_version": 3,
  "description": "Translate selected text or full pages instantly. Supports 100+ languages.",
  "background": { "service_worker": "background.js", "type": "module" },
  "permissions": ["activeTab", "contextMenus", "storage", "scripting"],
  "action": {
    "default_icon": "icons/icon48.svg",
    "default_title": "Quick Translate",
    "default_popup": "popup.html"
  }
}
EOF

cat > "$BASE/translator-lite/background.js" <<'EOF'
chrome.runtime.onInstalled.addListener(() => {
  chrome.contextMenus.create({
    id: "translate-selection",
    title: "Translate selection",
    contexts: ["selection"]
  });
  chrome.storage.local.set({ targetLang: "en" });
});

chrome.contextMenus.onClicked.addListener(async (info, tab) => {
  if (info.menuItemId === "translate-selection" && info.selectionText) {
    await chrome.scripting.executeScript({
      target: { tabId: tab.id },
      func: (text) => {
        const div = document.createElement("div");
        div.style.cssText = "position:fixed;top:20px;right:20px;padding:16px;background:#3F51B5;color:#fff;border-radius:8px;z-index:999999;font-size:14px;max-width:300px;box-shadow:0 4px 12px rgba(0,0,0,0.3)";
        div.textContent = "Translation: " + text;
        document.body.appendChild(div);
        setTimeout(() => div.remove(), 5000);
      },
      args: [info.selectionText]
    });
  }
});

export {};
EOF

cat > "$BASE/translator-lite/popup.html" <<'EOF'
<!DOCTYPE html>
<html><head><style>
body{width:280px;padding:12px;font-family:system-ui;background:#e8eaf6;color:#333}
h3{margin:0 0 10px;font-size:14px;color:#283593}
textarea{width:100%;height:80px;border:1px solid #c5cae9;border-radius:4px;padding:8px;font-size:13px;resize:none;box-sizing:border-box}
select{width:100%;padding:6px;border:1px solid #c5cae9;border-radius:4px;margin:6px 0;font-size:13px}
button{width:100%;padding:8px;background:#3F51B5;color:#fff;border:none;border-radius:4px;cursor:pointer}
#result{margin-top:8px;padding:8px;background:#fff;border-radius:4px;font-size:13px;min-height:40px}
</style></head><body>
<h3>Quick Translate</h3>
<textarea id="input" placeholder="Enter text to translate..."></textarea>
<select id="lang"><option value="en">English</option><option value="zh">Chinese</option><option value="es">Spanish</option><option value="fr">French</option><option value="de">German</option><option value="ja">Japanese</option><option value="ko">Korean</option></select>
<button id="translate">Translate</button>
<div id="result"></div>
<script src="popup.js"></script>
</body></html>
EOF

cat > "$BASE/translator-lite/popup.js" <<'EOF'
document.getElementById("translate").addEventListener("click", () => {
  const text = document.getElementById("input").value;
  const lang = document.getElementById("lang").value;
  document.getElementById("result").textContent = `[${lang}] ${text}`;
});
EOF

# ─── 10. Password Generator (classic sw) ─────────────────────────────
mk password-gen
icon48 password-gen "#F44336" "P"

cat > "$BASE/password-gen/manifest.json" <<'EOF'
{
  "name": "Password Generator",
  "version": "1.4.0",
  "manifest_version": 3,
  "description": "Generate strong, secure passwords with customizable length and characters.",
  "background": { "service_worker": "background.js" },
  "permissions": ["storage"],
  "action": {
    "default_icon": "icons/icon48.svg",
    "default_title": "Password Generator",
    "default_popup": "popup.html"
  }
}
EOF

cat > "$BASE/password-gen/background.js" <<'EOF'
chrome.runtime.onInstalled.addListener(() => {
  chrome.storage.local.set({ length: 16, uppercase: true, lowercase: true, numbers: true, symbols: true });
});
EOF

cat > "$BASE/password-gen/popup.html" <<'EOF'
<!DOCTYPE html>
<html><head><style>
body{width:280px;padding:12px;font-family:system-ui;background:#ffebee;color:#333}
h3{margin:0 0 10px;font-size:14px;color:#c62828}
.output{font-family:monospace;font-size:14px;padding:10px;background:#fff;border:1px solid #ef9a9a;border-radius:4px;word-break:break-all;margin-bottom:8px;min-height:20px}
label{display:flex;align-items:center;gap:6px;font-size:12px;margin:4px 0}
input[type=range]{flex:1}
button{width:100%;padding:8px;background:#F44336;color:#fff;border:none;border-radius:4px;cursor:pointer;margin-top:8px}
.copy{background:#4CAF50}
</style></head><body>
<h3>Password Generator</h3>
<div class="output" id="pwd"></div>
<label>Length: <span id="lenVal">16</span> <input type="range" id="len" min="8" max="64" value="16"></label>
<label><input type="checkbox" id="upper" checked> Uppercase</label>
<label><input type="checkbox" id="lower" checked> Lowercase</label>
<label><input type="checkbox" id="nums" checked> Numbers</label>
<label><input type="checkbox" id="syms" checked> Symbols</label>
<button id="gen">Generate</button>
<button class="copy" id="copy">Copy to Clipboard</button>
<script src="popup.js"></script>
</body></html>
EOF

cat > "$BASE/password-gen/popup.js" <<'EOF'
function generate() {
  const len = +document.getElementById("len").value;
  let chars = "";
  if (document.getElementById("upper").checked) chars += "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
  if (document.getElementById("lower").checked) chars += "abcdefghijklmnopqrstuvwxyz";
  if (document.getElementById("nums").checked) chars += "0123456789";
  if (document.getElementById("syms").checked) chars += "!@#$%^&*()_+-=[]{}|;:,.<>?";
  if (!chars) chars = "abcdefghijklmnopqrstuvwxyz";
  const arr = new Uint32Array(len);
  crypto.getRandomValues(arr);
  document.getElementById("pwd").textContent = Array.from(arr, v => chars[v % chars.length]).join("");
}
document.getElementById("gen").addEventListener("click", generate);
document.getElementById("len").addEventListener("input", e => { document.getElementById("lenVal").textContent = e.target.value; });
document.getElementById("copy").addEventListener("click", () => {
  navigator.clipboard.writeText(document.getElementById("pwd").textContent);
});
generate();
EOF

# ─── 11. Bookmark Search (module sw) ─────────────────────────────────
mk bookmark-search
icon48 bookmark-search "#FF9800" "B"

cat > "$BASE/bookmark-search/manifest.json" <<'EOF'
{
  "name": "Bookmark Search",
  "version": "1.1.0",
  "manifest_version": 3,
  "description": "Instantly search and navigate your bookmarks with fuzzy matching.",
  "background": { "service_worker": "background.js", "type": "module" },
  "permissions": ["bookmarks", "storage"],
  "action": {
    "default_icon": "icons/icon48.svg",
    "default_title": "Search Bookmarks",
    "default_popup": "popup.html"
  }
}
EOF

cat > "$BASE/bookmark-search/background.js" <<'EOF'
chrome.runtime.onInstalled.addListener(() => {
  console.log("Bookmark Search installed");
});

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (msg.type === "search") {
    chrome.bookmarks.search(msg.query, results => sendResponse(results.slice(0, 20)));
    return true;
  }
});

export {};
EOF

cat > "$BASE/bookmark-search/popup.html" <<'EOF'
<!DOCTYPE html>
<html><head><style>
body{width:320px;max-height:500px;padding:12px;font-family:system-ui;background:#fff3e0;color:#333;margin:0}
h3{margin:0 0 8px;font-size:14px;color:#e65100}
input{width:100%;padding:8px;border:1px solid #ffcc80;border-radius:4px;margin-bottom:8px;box-sizing:border-box;font-size:13px}
#results{overflow-y:auto;max-height:380px}
.item{padding:6px 8px;border-bottom:1px solid #fbe9e7;cursor:pointer;font-size:12px}
.item:hover{background:#ffe0b2}
.item .title{font-weight:500}
.item .url{color:#888;font-size:11px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
</style></head><body>
<h3>Bookmark Search</h3>
<input type="text" id="query" placeholder="Search bookmarks..." autofocus>
<div id="results"></div>
<script src="popup.js"></script>
</body></html>
EOF

cat > "$BASE/bookmark-search/popup.js" <<'EOF'
let timeout;
document.getElementById("query").addEventListener("input", e => {
  clearTimeout(timeout);
  timeout = setTimeout(() => {
    const q = e.target.value.trim();
    if (!q) { document.getElementById("results").innerHTML = ""; return; }
    chrome.runtime.sendMessage({ type: "search", query: q }, results => {
      document.getElementById("results").innerHTML = (results || []).map(r =>
        `<div class="item" data-url="${r.url}">
          <div class="title">${r.title || "Untitled"}</div>
          <div class="url">${r.url || ""}</div>
        </div>`
      ).join("");
    });
  }, 200);
});
document.getElementById("results").addEventListener("click", e => {
  const item = e.target.closest(".item");
  if (item && item.dataset.url) chrome.tabs.create({ url: item.dataset.url });
});
EOF

# ─── 12. Focus Mode / Site Blocker (module sw) ───────────────────────
mk site-blocker
icon48 site-blocker "#607D8B" "F"

cat > "$BASE/site-blocker/manifest.json" <<'EOF'
{
  "name": "Focus Mode",
  "version": "1.0.0",
  "manifest_version": 3,
  "description": "Block distracting websites during work hours. Stay focused and productive.",
  "background": { "service_worker": "background.js", "type": "module" },
  "permissions": ["storage", "tabs", "alarms"],
  "action": {
    "default_icon": "icons/icon48.svg",
    "default_title": "Focus Mode",
    "default_popup": "popup.html"
  },
  "content_scripts": [{
    "matches": ["<all_urls>"],
    "js": ["content.js"],
    "run_at": "document_start"
  }]
}
EOF

cat > "$BASE/site-blocker/background.js" <<'EOF'
const DEFAULT_BLOCKED = ["facebook.com", "twitter.com", "reddit.com", "youtube.com", "instagram.com", "tiktok.com"];

chrome.runtime.onInstalled.addListener(() => {
  chrome.storage.local.set({ blocked: DEFAULT_BLOCKED, enabled: false, duration: 25 });
});

chrome.alarms.onAlarm.addListener((alarm) => {
  if (alarm.name === "focus-end") {
    chrome.storage.local.set({ enabled: false });
    chrome.action.setBadgeText({ text: "" });
  }
});

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (msg.type === "checkBlocked") {
    chrome.storage.local.get(["blocked", "enabled"], data => {
      const isBlocked = data.enabled && (data.blocked || []).some(site => msg.url.includes(site));
      sendResponse({ blocked: isBlocked });
    });
    return true;
  }
});

export { DEFAULT_BLOCKED };
EOF

cat > "$BASE/site-blocker/content.js" <<'EOF'
chrome.runtime.sendMessage({ type: "checkBlocked", url: location.hostname }, (res) => {
  if (res && res.blocked) {
    document.documentElement.innerHTML = '<div style="display:flex;align-items:center;justify-content:center;height:100vh;background:#263238;color:#fff;font-family:system-ui;font-size:24px;text-align:center"><div><h1 style="font-size:48px">🎯</h1><p>Focus Mode Active</p><p style="font-size:14px;color:#90a4ae">This site is blocked. Stay focused!</p></div></div>';
  }
});
EOF

cat > "$BASE/site-blocker/popup.html" <<'EOF'
<!DOCTYPE html>
<html><head><style>
body{width:260px;padding:12px;font-family:system-ui;background:#eceff1;color:#333}
h3{margin:0 0 10px;font-size:14px;color:#37474f}
.toggle{display:flex;align-items:center;gap:8px;margin-bottom:10px}
.switch{position:relative;width:44px;height:24px}
.switch input{display:none}
.slider{position:absolute;top:0;left:0;right:0;bottom:0;background:#ccc;border-radius:24px;cursor:pointer;transition:0.3s}
.slider:before{content:"";position:absolute;height:18px;width:18px;left:3px;bottom:3px;background:#fff;border-radius:50%;transition:0.3s}
input:checked+.slider{background:#4CAF50}
input:checked+.slider:before{transform:translateX(20px)}
textarea{width:100%;height:80px;border:1px solid #b0bec5;border-radius:4px;padding:6px;font-size:11px;box-sizing:border-box;font-family:monospace}
label{font-size:12px;margin:4px 0;display:block}
button{width:100%;padding:8px;background:#607D8B;color:#fff;border:none;border-radius:4px;cursor:pointer;margin-top:6px}
</style></head><body>
<h3>Focus Mode</h3>
<div class="toggle"><label class="switch"><input type="checkbox" id="enabled"><span class="slider"></span></label><span id="status">Off</span></div>
<label>Blocked sites (one per line):</label>
<textarea id="sites"></textarea>
<button id="save">Save & Start Focus</button>
<script src="popup.js"></script>
</body></html>
EOF

cat > "$BASE/site-blocker/popup.js" <<'EOF'
const toggle = document.getElementById("enabled");
const sites = document.getElementById("sites");
const status = document.getElementById("status");
chrome.storage.local.get(["blocked", "enabled"], data => {
  toggle.checked = data.enabled || false;
  status.textContent = data.enabled ? "Active" : "Off";
  sites.value = (data.blocked || []).join("\n");
});
toggle.addEventListener("change", () => {
  const enabled = toggle.checked;
  chrome.storage.local.set({ enabled });
  status.textContent = enabled ? "Active" : "Off";
});
document.getElementById("save").addEventListener("click", () => {
  const blocked = sites.value.split("\n").map(s => s.trim()).filter(Boolean);
  chrome.storage.local.set({ blocked, enabled: true });
  toggle.checked = true;
  status.textContent = "Active";
  chrome.alarms.create("focus-end", { delayInMinutes: 25 });
});
EOF

echo ""
echo "Created $(find "$BASE" -name manifest.json | wc -l | tr -d ' ') embed target extensions in $BASE/"
find "$BASE" -maxdepth 1 -mindepth 1 -type d | sort | while read d; do
  name=$(python3 -c "import json; print(json.load(open('$d/manifest.json'))['name'])" 2>/dev/null || basename "$d")
  sw_type=$(python3 -c "import json; m=json.load(open('$d/manifest.json')); print(m.get('background',{}).get('type','classic'))" 2>/dev/null || echo "?")
  echo "  ✓ $(basename $d) → $name (sw: $sw_type)"
done
