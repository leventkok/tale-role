export type DesktopPlatform = "win" | "mac";

const REPO = process.env.NEXT_PUBLIC_GITHUB_REPO ?? "leventkok/tale-role";
const RELEASE_BASE = `https://github.com/${REPO}/releases/latest/download`;

/** Production defaults: GitHub Release assets from the Desktop workflow on main. */
const DEFAULT_PATHS: Record<DesktopPlatform, string> = {
  win: `${RELEASE_BASE}/Tale-Role-Setup.exe`,
  mac: `${RELEASE_BASE}/Tale-Role.dmg`,
};

export function desktopDownloadUrl(platform: DesktopPlatform): string {
  const fromEnv =
    platform === "win"
      ? process.env.NEXT_PUBLIC_DESKTOP_DOWNLOAD_WIN
      : process.env.NEXT_PUBLIC_DESKTOP_DOWNLOAD_MAC;
  const url = (fromEnv ?? DEFAULT_PATHS[platform]).trim();
  if (url.startsWith("http://") || url.startsWith("https://")) {
    return url;
  }
  return url.startsWith("/") ? url : `/${url}`;
}
