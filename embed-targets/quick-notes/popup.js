const note = document.getElementById("note");
const status = document.getElementById("status");
chrome.storage.sync.get("noteText", data => { note.value = data.noteText || ""; });
document.getElementById("save").addEventListener("click", () => {
  chrome.storage.sync.set({ noteText: note.value }, () => {
    status.textContent = "Saved at " + new Date().toLocaleTimeString();
  });
});
document.getElementById("clear").addEventListener("click", () => {
  note.value = "";
  chrome.storage.sync.set({ noteText: "" });
  status.textContent = "Cleared";
});
