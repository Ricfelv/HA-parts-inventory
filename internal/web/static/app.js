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

document.querySelectorAll("[data-copy-mpn]").forEach((button) => {
  button.addEventListener("click", async () => {
    const mpn = button.dataset.copyMpn;
    const label = button.dataset.originalLabel ?? button.getAttribute("aria-label");
    const title = button.dataset.originalTitle ?? button.title;
    button.dataset.originalLabel = label;
    button.dataset.originalTitle = title;
    try {
      await navigator.clipboard.writeText(mpn);
      button.classList.add("copied");
      button.setAttribute("aria-label", "MPN copied");
      button.title = "Copied";
      setTimeout(() => {
        button.classList.remove("copied");
        button.setAttribute("aria-label", label);
        button.title = title;
      }, 1500);
    } catch {
      button.title = "Unable to copy MPN";
    }
  });
});
