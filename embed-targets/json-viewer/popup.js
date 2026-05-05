document.getElementById("fmt").addEventListener("click", () => {
  const input = document.getElementById("input");
  try {
    const parsed = JSON.parse(input.value);
    input.value = JSON.stringify(parsed, null, 2);
  } catch(e) {
    input.value = "Invalid JSON: " + e.message;
  }
});
