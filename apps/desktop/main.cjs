const { app, BrowserWindow, Menu, nativeImage, protocol, shell } = require("electron");
const path = require("node:path");
const os = require("node:os");
const crypto = require("node:crypto");

const WEB_URL = (process.env.TALEROLE_WEB_URL || "https://www.talerole.com").replace(/\/$/, "");
const LOCALE = process.env.TALEROLE_LOCALE === "tr" ? "tr" : "en";
const APP_ID = "role.tale.desktop";
const APP_ICON = nativeImage.createFromPath(path.join(__dirname, "icons", "icon.png"));

protocol.registerSchemesAsPrivileged([
  { scheme: "talerole", privileges: { standard: true, secure: true } },
]);

let mainWindow = null;

function deviceId() {
  return crypto.createHash("sha256").update(os.hostname()).digest("hex").slice(0, 16);
}

function joinDest(requestUrl) {
  const url = new URL(requestUrl);
  let roomId = "";
  if (url.hostname === "join") {
    roomId = url.pathname.replace(/^\//, "");
  } else {
    const parts = url.pathname.split("/").filter(Boolean);
    const i = parts.indexOf("join");
    roomId = i >= 0 ? parts.slice(i + 1).join("/") : parts[parts.length - 1] || "";
  }
  return `${WEB_URL}/${LOCALE}/join/${encodeURIComponent(roomId)}${url.search}`;
}

function loadJoin(requestUrl) {
  const dest = joinDest(requestUrl);
  if (mainWindow) {
    mainWindow.loadURL(dest);
    mainWindow.show();
    mainWindow.focus();
  }
}

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1280,
    height: 840,
    minWidth: 900,
    minHeight: 640,
    title: "Tale Role",
    icon: APP_ICON,
    autoHideMenuBar: true,
    backgroundColor: "#120e0a",
    webPreferences: {
      preload: path.join(__dirname, "preload.cjs"),
      contextIsolation: true,
      nodeIntegration: false,
    },
  });
  mainWindow.loadURL(`${WEB_URL}/${LOCALE}`);
  mainWindow.setMenuBarVisibility(false);
  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    if (url.startsWith("http:") || url.startsWith("https:")) {
      void shell.openExternal(url);
    }
    return { action: "deny" };
  });
  mainWindow.on("closed", () => {
    mainWindow = null;
  });
}

const locked = app.requestSingleInstanceLock();
if (!locked) {
  app.quit();
} else {
  if (process.platform === "win32") {
    app.setAppUserModelId(APP_ID);
  }

  if (process.defaultApp) {
    if (process.argv.length >= 2) {
      app.setAsDefaultProtocolClient("talerole", process.execPath, [path.resolve(process.argv[1])]);
    }
  } else {
    app.setAsDefaultProtocolClient("talerole");
  }

  app.on("second-instance", (_event, argv) => {
    const deep = argv.find((a) => typeof a === "string" && a.startsWith("talerole:"));
    if (deep) {
      loadJoin(deep);
    } else if (mainWindow) {
      mainWindow.show();
      mainWindow.focus();
    }
  });

  app.on("open-url", (event, url) => {
    event.preventDefault();
    loadJoin(url);
  });

  app.whenReady().then(() => {
    Menu.setApplicationMenu(null);
    if (process.platform === "darwin" && app.dock) {
      app.dock.setIcon(APP_ICON);
    }
    protocol.handle("talerole", (request) => Response.redirect(joinDest(request.url), 302));
    createWindow();
    const fromArg = process.argv.find((a) => typeof a === "string" && a.startsWith("talerole:"));
    if (fromArg) {
      loadJoin(fromArg);
    }
  });

  app.on("activate", () => {
    if (!mainWindow) {
      createWindow();
    }
  });

  app.on("window-all-closed", () => {
    if (process.platform !== "darwin") {
      app.quit();
    }
  });
}

module.exports = { deviceId };
