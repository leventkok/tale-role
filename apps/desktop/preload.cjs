const { contextBridge } = require("electron");
const os = require("node:os");
const crypto = require("node:crypto");

contextBridge.exposeInMainWorld("taleRoleDesktop", {
  platform: process.platform,
  deviceId: crypto.createHash("sha256").update(os.hostname()).digest("hex").slice(0, 16),
});
