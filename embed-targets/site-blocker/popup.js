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
