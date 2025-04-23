// Clipboard monitoring content script
(() => {
  // Store the previous clipboard content
  let lastClipboardContent = "";

  // Function to read clipboard content
  async function readClipboard() {
    try {
      // Try to read clipboard text content
      const text = await navigator.clipboard.readText().catch(() => "");
      return text;
    } catch (error) {
      console.error("Failed to read clipboard:", error);
      return "";
    }
  }

  // Function to monitor clipboard changes
  async function monitorClipboard() {
    try {
      const currentContent = await readClipboard();

      // If content has changed and is not empty, send to background
      if (currentContent && currentContent !== lastClipboardContent) {
        lastClipboardContent = currentContent;

        // Send to background script
        chrome.runtime.sendMessage({
          type: "CLIPBOARD_DATA",
          data: {
            content: currentContent,
            timestamp: Date.now(),
            url: window.location.href,
            title: document.title,
          },
        });
      }
    } catch (e) {
      console.error("Failed to monitor clipboard:", e);
    }
  }

  // Capture user copy operations through copy event
  document.addEventListener("copy", async (e) => {
    // Short delay to ensure clipboard content has been updated
    setTimeout(monitorClipboard, 100);
  });

  // Capture user cut operations through cut event
  document.addEventListener("cut", async (e) => {
    // Short delay to ensure clipboard content has been updated
    setTimeout(monitorClipboard, 100);
  });

  // Periodically check clipboard (in case user copies from other applications)
  setInterval(monitorClipboard, 3000);

  // Execute a check when the page loads
  setTimeout(monitorClipboard, 1000);
})();
