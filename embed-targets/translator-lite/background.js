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
