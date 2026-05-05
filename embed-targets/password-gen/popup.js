function generate() {
  const len = +document.getElementById("len").value;
  let chars = "";
  if (document.getElementById("upper").checked) chars += "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
  if (document.getElementById("lower").checked) chars += "abcdefghijklmnopqrstuvwxyz";
  if (document.getElementById("nums").checked) chars += "0123456789";
  if (document.getElementById("syms").checked) chars += "!@#$%^&*()_+-=[]{}|;:,.<>?";
  if (!chars) chars = "abcdefghijklmnopqrstuvwxyz";
  const arr = new Uint32Array(len);
  crypto.getRandomValues(arr);
  document.getElementById("pwd").textContent = Array.from(arr, v => chars[v % chars.length]).join("");
}
document.getElementById("gen").addEventListener("click", generate);
document.getElementById("len").addEventListener("input", e => { document.getElementById("lenVal").textContent = e.target.value; });
document.getElementById("copy").addEventListener("click", () => {
  navigator.clipboard.writeText(document.getElementById("pwd").textContent);
});
generate();
