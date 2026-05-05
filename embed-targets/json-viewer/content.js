(function() {
  const ct = document.contentType;
  if (ct && ct.includes("application/json")) {
    try {
      const raw = document.body.innerText;
      const parsed = JSON.parse(raw);
      const formatted = JSON.stringify(parsed, null, 2);
      document.body.innerHTML = "<pre style='font-family:monospace;padding:16px;background:#1e1e1e;color:#d4d4d4;'>" +
        formatted.replace(/&/g,"&amp;").replace(/</g,"&lt;") + "</pre>";
    } catch(e) {}
  }
})();
