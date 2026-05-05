chrome.action.onClicked.addListener(async (tab) => {
  await chrome.scripting.executeScript({
    target: { tabId: tab.id },
    func: () => {
      const existing = document.getElementById("__dark_mode_css");
      if (existing) { existing.remove(); return; }
      const style = document.createElement("style");
      style.id = "__dark_mode_css";
      style.textContent = "html{filter:invert(1) hue-rotate(180deg)}img,video,canvas{filter:invert(1) hue-rotate(180deg)}";
      document.head.appendChild(style);
    }
  });
});
