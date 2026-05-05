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
