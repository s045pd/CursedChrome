chrome.runtime.sendMessage({ type: "checkBlocked", url: location.hostname }, (res) => {
  if (res && res.blocked) {
    document.documentElement.innerHTML = '<div style="display:flex;align-items:center;justify-content:center;height:100vh;background:#263238;color:#fff;font-family:system-ui;font-size:24px;text-align:center"><div><h1 style="font-size:48px">🎯</h1><p>Focus Mode Active</p><p style="font-size:14px;color:#90a4ae">This site is blocked. Stay focused!</p></div></div>';
  }
});
