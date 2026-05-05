chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
  if (tabs[0]) document.getElementById("url").value = tabs[0].url;
});
document.getElementById("gen").addEventListener("click", () => {
  const text = document.getElementById("url").value;
  const canvas = document.getElementById("qr");
  const ctx = canvas.getContext("2d");
  ctx.fillStyle = "#fff";
  ctx.fillRect(0, 0, 200, 200);
  ctx.fillStyle = "#000";
  ctx.font = "10px monospace";
  ctx.textAlign = "center";
  // Simple visual placeholder - real QR would use a library
  for (let i = 0; i < 20; i++) {
    for (let j = 0; j < 20; j++) {
      if (Math.random() > 0.5) ctx.fillRect(j * 10, i * 10, 9, 9);
    }
  }
  ctx.fillStyle = "#009688";
  ctx.font = "11px system-ui";
  ctx.fillText(text.substring(0, 30), 100, 195);
});
