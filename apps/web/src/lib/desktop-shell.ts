"use client";

import { useEffect, useState } from "react";

type DesktopBridge = { platform?: string; deviceId?: string };

function desktopWindow(): Window & { taleRoleDesktop?: DesktopBridge } {
  return window as Window & { taleRoleDesktop?: DesktopBridge };
}

/** True only inside our Electron shell — never a generic Electron user-agent. */
export function isDesktopShell(): boolean {
  if (typeof window === "undefined") {
    return false;
  }
  if (desktopWindow().taleRoleDesktop) {
    return true;
  }
  if (document.documentElement.dataset.taleroleShell === "desktop") {
    return true;
  }
  if (document.body?.dataset.taleroleShell === "desktop") {
    return true;
  }
  return /\bTaleRoleDesktop\b/i.test(navigator.userAgent);
}

export function readDesktopBridge(): { platform: string; deviceId: string } | undefined {
  if (typeof window === "undefined") {
    return undefined;
  }
  const bridge = desktopWindow().taleRoleDesktop;
  const deviceId = (bridge?.deviceId || "").trim();
  if (!deviceId) {
    return undefined;
  }
  return { platform: (bridge?.platform || "").trim(), deviceId };
}

/** After mount — SSR HTML never has the Electron bridge, so the register button must wait. */
export function useDesktopBridge() {
  const [desktop, setDesktop] = useState<{ platform: string; deviceId: string } | undefined>(undefined);
  useEffect(() => {
    function tick() {
      const next = readDesktopBridge();
      if (next) {
        setDesktop(next);
        return true;
      }
      return false;
    }
    if (tick()) {
      return;
    }
    const id = window.setInterval(() => {
      if (tick()) {
        window.clearInterval(id);
      }
    }, 200);
    const stop = window.setTimeout(() => window.clearInterval(id), 5000);
    return () => {
      window.clearInterval(id);
      window.clearTimeout(stop);
    };
  }, []);
  return desktop;
}

export function desktopDeviceHeaders(): Record<string, string> {
  const bridge = readDesktopBridge();
  if (!bridge) {
    return {};
  }
  const headers: Record<string, string> = { "X-TaleRole-Device": bridge.deviceId };
  if (bridge.platform) {
    headers["X-TaleRole-Platform"] = bridge.platform;
  }
  return headers;
}

/** Next.js hydrates `<html>` without this flag and would otherwise wipe the preload mark. */
export function markDesktopShell(): void {
  if (typeof document === "undefined") {
    return;
  }
  document.documentElement.dataset.taleroleShell = "desktop";
  if (document.body) {
    document.body.dataset.taleroleShell = "desktop";
  }
}
