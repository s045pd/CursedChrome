document.getElementById("capture").addEventListener("click", async () => {
  chrome.runtime.sendMessage({ type: "capture" }, (res) => {
    if (res && res.url) {
      const img = document.getElementById("preview");
      img.src = res.url;
      img.style.display = "block";
      document.getElementById("download").href = res.url;
      document.getElementById("actions").style.display = "block";
    }
  });
});
