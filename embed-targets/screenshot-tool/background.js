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
