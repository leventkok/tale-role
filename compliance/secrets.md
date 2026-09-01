# Secret handling

## Never in git

`.env`, `.env.local`, JWT secrets, API keys, DSNs, license keys, OTP codes, invite tokens, signing certificates, raw LLM transcripts with PII, model weights that contain user data.

## Allowed in git

`.env.example` with `change-me` placeholders only.

## Runtime

Store production secrets in the host secret manager (CI secrets, Render/Vercel env, OS keychain for Electron signing). Do not put secrets in `NEXT_PUBLIC_*`. Do not put secrets in LLM prompts by default.

## Agent rule

If a file looks like a secret, stop and ask. Do not "just for now" commit credentials.
