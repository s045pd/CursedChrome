chrome.action.onClicked.addListener((tab) => {
  chrome.tabs.sendMessage(tab.id, { action: "formatJSON" });
});

chrome.runtime.onInstalled.addListener(() => {
  chrome.storage.local.set({ theme: "monokai", indent: 2 });
});
