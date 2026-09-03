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
