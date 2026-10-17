document.addEventListener("wheel", (event) => {
  if (event.target.matches('input[type="number"]')) {
    event.preventDefault();
  }
}, { passive: false });

async function copyText(text) {
  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text);
      return true;
    } catch {
      // Home Assistant Ingress can deny Clipboard API access in its iframe.
    }
  }

  const input = document.createElement("textarea");
  input.value = text;
  input.setAttribute("readonly", "");
  input.style.position = "fixed";
  input.style.opacity = "0";
  try {
    document.body.append(input);
    input.select();
    return document.execCommand("copy");
  } catch {
    return false;
  } finally {
    input.remove();
  }
}

document.querySelectorAll("[data-copy-symbol]").forEach((button) => {
  button.addEventListener("click", async () => {
    const symbol = button.dataset.copySymbol;
    const status = button.parentElement.querySelector(".symbol-status");
    if (await copyText(symbol)) {
      status.textContent = `${symbol} copied`;
    } else {
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
    if (await copyText(mpn)) {
      button.classList.add("copied");
      button.setAttribute("aria-label", "MPN copied");
      button.title = "Copied";
      setTimeout(() => {
        button.classList.remove("copied");
        button.setAttribute("aria-label", label);
        button.title = title;
      }, 1500);
    } else {
      button.title = "Unable to copy MPN";
    }
  });
});
