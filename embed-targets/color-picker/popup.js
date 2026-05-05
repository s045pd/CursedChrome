const picker = document.getElementById("picker");
const preview = document.getElementById("preview");
const info = document.getElementById("info");
picker.addEventListener("input", e => {
  const c = e.target.value;
  preview.style.background = c;
  const r = parseInt(c.slice(1,3),16), g = parseInt(c.slice(3,5),16), b = parseInt(c.slice(5,7),16);
  info.textContent = `HEX: ${c}\nRGB: rgb(${r}, ${g}, ${b})`;
});
chrome.storage.local.get("history", ({ history }) => {
  const container = document.getElementById("history");
  (history || []).forEach(c => {
    const el = document.createElement("div");
    el.className = "swatch";
    el.style.background = c;
    el.title = c;
    el.addEventListener("click", () => { picker.value = c; picker.dispatchEvent(new Event("input")); });
    container.appendChild(el);
  });
});
