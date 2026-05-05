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
