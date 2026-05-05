async function render(filter = "") {
  const tabs = await chrome.tabs.query({});
  const container = document.getElementById("tabs");
  document.getElementById("count").textContent = tabs.length + " tabs open";
  const filtered = filter ? tabs.filter(t => (t.title + t.url).toLowerCase().includes(filter.toLowerCase())) : tabs;
  container.innerHTML = filtered.map(t => `
    <div class="tab" data-id="${t.id}">
      <img src="${t.favIconUrl || ''}" onerror="this.style.display='none'">
      <span title="${t.url}">${t.title || t.url}</span>
      <span class="close" data-close="${t.id}">&times;</span>
    </div>`).join("");
}
document.getElementById("search").addEventListener("input", e => render(e.target.value));
document.getElementById("tabs").addEventListener("click", async e => {
  const closeId = e.target.dataset.close;
  if (closeId) { await chrome.tabs.remove(+closeId); render(); return; }
  const tabEl = e.target.closest(".tab");
  if (tabEl) chrome.tabs.update(+tabEl.dataset.id, { active: true });
});
render();
