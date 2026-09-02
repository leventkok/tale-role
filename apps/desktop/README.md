# apps/desktop

Electron shell (Windows + macOS). Loads the web app and registers `talerole://join/...`.

```bash
# from repo root, with apps/web running
npx electron apps/desktop
```

Device id is a hostname hash exposed via `window.taleRoleDesktop` for `POST /api/v1/licenses/register`. The JWT stays in the HttpOnly web cookie; never put it in `localStorage`.

Signing (Win/Mac) is a human step. See [docs/SIGNING.md](../../docs/SIGNING.md). Do not commit certificates.
