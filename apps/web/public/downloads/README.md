# Desktop installers

Production downloads use **same-origin URLs** on the site:

- `/downloads/Tale-Role-Setup.exe`
- `/downloads/Tale-Role.dmg`

Vercel rewrites (see `apps/web/vercel.json`) serve the latest GitHub Release assets through `talerole.com` — the browser downloads the file; it does not open the GitHub website.

## Local testing

Build Windows installer and copy here (gitignored):

```powershell
cd apps\desktop
npm ci
npm run pack:win
copy dist\Tale-Role-Setup.exe ..\web\public\downloads\Tale-Role-Setup.exe
```

When the file exists under `public/downloads/`, Next.js serves it directly in dev.

Do not commit `.exe` / `.dmg` (too large; gitignored).
