export type DesktopPlatform = "win" | "mac";

export const DESKTOP_FILENAMES: Record<DesktopPlatform, string> = {
  win: "Tale-Role-Setup.exe",
  mac: "Tale-Role.dmg",
};

/** Same-origin paths; Vercel rewrites proxy to Release assets in production. */
export function desktopDownloadPath(platform: DesktopPlatform): string {
  return `/downloads/${DESKTOP_FILENAMES[platform]}`;
}
