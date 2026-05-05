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
