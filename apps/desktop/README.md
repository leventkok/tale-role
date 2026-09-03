# apps/desktop

Electron shell (Windows + macOS). Loads the live web app and registers `talerole://join/...`.

## Run locally

```powershell
cd apps\desktop
npm install
npm start
```

Live site (default `https://www.talerole.com`):

```powershell
npm start
```

Local web dev:

```powershell
$env:TALEROLE_WEB_URL="http://127.0.0.1:3000"
$env:TALEROLE_LOCALE="tr"
npm start
```

## Build installers

```powershell
cd apps\desktop
npm install
npm run pack
```

Icons are generated from `icons/icon.svg` (the Tale Role lantern mark) before each pack.

Outputs in `dist/`:

| OS | File |
| --- | --- |
| Windows | `Tale-Role-Setup.exe` |
| macOS | `Tale-Role.dmg` (build on a Mac) |

Unsigned builds only until you set signing env vars — see [docs/SIGNING.md](../../docs/SIGNING.md).

## Website download button

The home page links to the **latest GitHub Release** (`Tale-Role-Setup.exe`, `Tale-Role.dmg`). Pushing to `main` (when `apps/desktop` changes) runs `.github/workflows/desktop.yml` and publishes a new release automatically.

For local web testing, copy installers to `apps/web/public/downloads/` and set `NEXT_PUBLIC_DESKTOP_DOWNLOAD_*` in `.env.local`. See `apps/web/public/downloads/README.md`.

## Device license

`window.taleRoleDesktop` exposes `deviceId` and `platform`. Register the device from Account → Devices in the desktop app. JWT stays in HttpOnly cookies, not `localStorage`.
