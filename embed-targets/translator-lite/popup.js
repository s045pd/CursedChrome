document.getElementById("translate").addEventListener("click", () => {
  const text = document.getElementById("input").value;
  const lang = document.getElementById("lang").value;
  document.getElementById("result").textContent = `[${lang}] ${text}`;
});
