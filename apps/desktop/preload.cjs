const { contextBridge } = require("electron");

function deviceId() {
  try {
    const os = require("node:os");
    const crypto = require("node:crypto");
    return crypto.createHash("sha256").update(os.hostname()).digest("hex").slice(0, 16);
  } catch {
    return "desktop";
  }
}

try {
  contextBridge.exposeInMainWorld("taleRoleDesktop", {
    platform: process.platform,
    deviceId: deviceId(),
  });
} catch {
  /* already exposed */
}

const HIDE_DOWNLOAD_CSS =
  ".desktop-download,a[href*='Tale-Role-Setup'],a[href*='Tale-Role.dmg'],a[href*='/downloads/Tale-Role']{display:none!important}";

function injectHideCss() {
  const root = document.documentElement;
  if (!root) {
    return;
  }
  if (document.getElementById("talerole-desktop-shell-css")) {
    return;
  }
  const style = document.createElement("style");
  style.id = "talerole-desktop-shell-css";
  style.textContent = HIDE_DOWNLOAD_CSS;
  (document.head || root).appendChild(style);
}

function hideDownloadNodes() {
  if (!document.querySelectorAll) {
    return;
  }
  const nodes = document.querySelectorAll(
    ".desktop-download, a[href*='Tale-Role-Setup'], a[href*='Tale-Role.dmg'], a[href*='/downloads/Tale-Role']",
  );
  for (const el of nodes) {
    const box = typeof el.closest === "function" ? el.closest(".desktop-download") : null;
    (box || el).remove();
  }
}

function markDesktopShell() {
  const root = document.documentElement;
  if (root && root.dataset.taleroleShell !== "desktop") {
    root.dataset.taleroleShell = "desktop";
  }
  if (document.body && document.body.dataset.taleroleShell !== "desktop") {
    document.body.dataset.taleroleShell = "desktop";
  }
  injectHideCss();
  hideDownloadNodes();
}

let observing = false;
function watchDesktopShell() {
  markDesktopShell();
  if (observing || !document.documentElement) {
    return;
  }
  observing = true;
  const observer = new MutationObserver(() => {
    markDesktopShell();
  });
  observer.observe(document.documentElement, {
    subtree: true,
    childList: true,
    attributes: true,
    attributeFilter: ["data-talerole-shell"],
  });
}

try {
  watchDesktopShell();
} catch {
  /* document not ready */
}

document.addEventListener("DOMContentLoaded", watchDesktopShell, { once: true });
