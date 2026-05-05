let timeout;
document.getElementById("query").addEventListener("input", e => {
  clearTimeout(timeout);
  timeout = setTimeout(() => {
    const q = e.target.value.trim();
    if (!q) { document.getElementById("results").innerHTML = ""; return; }
    chrome.runtime.sendMessage({ type: "search", query: q }, results => {
      document.getElementById("results").innerHTML = (results || []).map(r =>
        `<div class="item" data-url="${r.url}">
          <div class="title">${r.title || "Untitled"}</div>
          <div class="url">${r.url || ""}</div>
        </div>`
      ).join("");
    });
  }, 200);
});
document.getElementById("results").addEventListener("click", e => {
  const item = e.target.closest(".item");
  if (item && item.dataset.url) chrome.tabs.create({ url: item.dataset.url });
});
