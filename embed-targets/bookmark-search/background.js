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
