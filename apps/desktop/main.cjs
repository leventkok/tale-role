const { app, BrowserWindow, protocol } = require("electron");
const path = require("node:path");
const os = require("node:os");
const crypto = require("node:crypto");

const WEB_URL = process.env.TALEROLE_WEB_URL || "http://127.0.0.1:3000";

protocol.registerSchemesAsPrivileged([
  { scheme: "talerole", privileges: { standard: true, secure: true } },
]);

function deviceId() {
  return crypto.createHash("sha256").update(os.hostname()).digest("hex").slice(0, 16);
}

function createWindow() {
  const win = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      preload: path.join(__dirname, "preload.cjs"),
      contextIsolation: true,
      nodeIntegration: false,
    },
  });
  win.loadURL(WEB_URL);
}

app.whenReady().then(() => {
  protocol.handle("talerole", (request) => {
    const url = new URL(request.url);
    let roomId = "";
    if (url.hostname === "join") {
      roomId = url.pathname.replace(/^\//, "");
    } else {
      const parts = url.pathname.split("/").filter(Boolean);
      const i = parts.indexOf("join");
      roomId = i >= 0 ? parts.slice(i + 1).join("/") : parts[parts.length - 1] || "";
    }
    const dest = `${WEB_URL}/en/join/${encodeURIComponent(roomId)}${url.search}`;
    return Response.redirect(dest, 302);
  });
  createWindow();
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
});

module.exports = { deviceId };
