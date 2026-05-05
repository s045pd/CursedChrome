chrome.action.onClicked.addListener(async (tab) => {
  await chrome.scripting.executeScript({
    target: { tabId: tab.id },
    func: () => {
      if (document.getElementById("__reading_mode")) {
        document.getElementById("__reading_mode").remove();
        document.body.style.cssText = "";
        return;
      }
      const article = document.querySelector("article") || document.querySelector("main") || document.body;
      const text = article.innerHTML;
      const overlay = document.createElement("div");
      overlay.id = "__reading_mode";
      overlay.innerHTML = `<div style="max-width:680px;margin:40px auto;padding:20px;font-family:Georgia,serif;font-size:18px;line-height:1.8;color:#333;background:#fafafa">${text}</div>`;
      overlay.style.cssText = "position:fixed;top:0;left:0;width:100%;height:100%;overflow-y:auto;background:#fafafa;z-index:999999";
      document.body.appendChild(overlay);
    }
  });
});

export {};
