document.addEventListener("wheel", (event) => {
  if (event.target.matches('input[type="number"]')) {
    event.preventDefault();
  }
}, { passive: false });

document.querySelectorAll("[data-copy-symbol]").forEach((button) => {
  button.addEventListener("click", async () => {
    const symbol = button.dataset.copySymbol;
    const status = button.parentElement.querySelector(".symbol-status");
    try {
      await navigator.clipboard.writeText(symbol);
      status.textContent = `${symbol} copied`;
    } catch {
      status.textContent = `Copy: ${symbol}`;
    }
  });
});
