document.getElementById("apply").addEventListener("click", async () => {
  const b = document.getElementById("brightness").value;
  const c = document.getElementById("contrast").value;
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  await chrome.scripting.executeScript({
    target: { tabId: tab.id },
    func: (brightness, contrast) => {
      let s = document.getElementById("__dark_mode_css");
      if (!s) { s = document.createElement("style"); s.id = "__dark_mode_css"; document.head.appendChild(s); }
      s.textContent = `html{filter:invert(1) hue-rotate(180deg) brightness(${brightness}%) contrast(${contrast}%)}img,video,canvas{filter:invert(1) hue-rotate(180deg)}`;
    },
    args: [b, c]
  });
});
