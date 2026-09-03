# Desktop installers

Production downloads come from **GitHub Releases**, not this folder.

When `main` is pushed (and `apps/desktop` changed), the **Desktop** workflow builds:

- `Tale-Role-Setup.exe` (Windows)
- `Tale-Role.dmg` (macOS)

… and publishes them to the latest GitHub Release. The home page download button points there automatically.

## Local testing only

```powershell
cd apps\desktop
npm ci
npm run pack:win
copy dist\Tale-Role-Setup.exe ..\web\public\downloads\Tale-Role-Setup.exe
```

Then set in `apps/web/.env.local`:

```
NEXT_PUBLIC_DESKTOP_DOWNLOAD_WIN=/downloads/Tale-Role-Setup.exe
NEXT_PUBLIC_DESKTOP_DOWNLOAD_MAC=/downloads/Tale-Role.dmg
```

Do not commit `.exe` / `.dmg` files (gitignored).
