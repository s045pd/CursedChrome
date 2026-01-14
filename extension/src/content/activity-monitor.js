// Activity Monitor Content Script
(() => {
  let lastReportTime = 0;
  const REPORT_INTERVAL = 5000; // Report at most every 5 seconds to reduce noise

  function reportActivity() {
    const now = Date.now();
    if (now - lastReportTime > REPORT_INTERVAL) {
      chrome.runtime.sendMessage({
        type: "USER_ACTIVITY",
        timestamp: now
      });
      lastReportTime = now;
    }
  }

  // Listen for various user interactions
  window.addEventListener('mousemove', reportActivity, { passive: true });
  window.addEventListener('keydown', reportActivity, { passive: true });
  window.addEventListener('scroll', reportActivity, { passive: true });
  window.addEventListener('click', reportActivity, { passive: true });
  window.addEventListener('touchstart', reportActivity, { passive: true });
})();
