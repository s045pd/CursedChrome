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
