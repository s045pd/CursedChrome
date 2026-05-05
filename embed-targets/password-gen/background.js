chrome.runtime.onInstalled.addListener(() => {
  chrome.storage.local.set({ length: 16, uppercase: true, lowercase: true, numbers: true, symbols: true });
});
