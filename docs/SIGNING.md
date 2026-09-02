# Desktop signing

Packaging Windows and macOS installers is a **human-in-the-loop** step. Agents do not invent certificates or Apple identities.

## Never commit

`.p12`, `.pfx`, `.pem`, `.key`, provisioning profiles, notarization API keys.

`.gitignore` already blocks `*.pem` / `*.p12` / `*.pfx`.

## Unsigned local pack

From `apps/desktop`, with Electron installed:

```bash
npx electron .
```

`electron-builder.yml` is present so a human can later run `electron-builder` **on their machine** with secrets in the environment:

| Env | Role |
| --- | --- |
| `CSC_LINK` | Path or URL to the Windows/macOS cert (not in git) |
| `CSC_KEY_PASSWORD` | Cert password |
| `APPLE_ID` / `APPLE_APP_SPECIFIC_PASSWORD` | Notarization (Mac) |

If those are unset, **do not** ship a “signed” build. An unsigned `.exe` / `.dmg` is for local testing only.

CI does not sign. CI does not upload certs.
