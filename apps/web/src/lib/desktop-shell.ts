export function isDesktopShell(): boolean {
  if (typeof window === "undefined") {
    return false;
  }
  return Boolean((window as Window & { taleRoleDesktop?: unknown }).taleRoleDesktop);
}
