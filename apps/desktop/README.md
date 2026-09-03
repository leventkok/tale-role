# apps/desktop

Electron shell (Windows + macOS). Loads the live web app and registers `talerole://join/...`.

```bash
# live site (default)
npx electron apps/desktop

# local web
$env:TALEROLE_WEB_URL="http://127.0.0.1:3000"
npx electron apps/desktop
```

Turkish join links: `$env:TALEROLE_LOCALE="tr"`.

Unsigned local pack (no certificates):

```bash
cd apps/desktop
npx electron-builder --publish never
```

Signing is a human step. See [docs/SIGNING.md](../../docs/SIGNING.md). Do not commit certificates.

Device id is a hostname hash on `window.taleRoleDesktop`. The JWT stays in the HttpOnly web cookie; never `localStorage`.
