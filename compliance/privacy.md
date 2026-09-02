# Privacy (KVKK / GDPR)

This instance stores account and table data **in memory**. A process restart is a full wipe. That is not a production retention policy.

## Records

- Email, bcrypt password hash, verification flag
- Optional Electron device license (`device_id`, platform)
- Rooms, characters, and universe packs the user owns
- Action notes players type (may contain secrets — treat as personal data)

Never stored in git: `.env`, OTP codes, JWT secrets, signing certs, model weights, raw player transcripts.

## Rights on this API

- `GET /api/v1/me/export` — machine-readable copy (no password hash, no OTP)
- `DELETE /api/v1/me` — erase user, licenses, owned universes, hosted rooms; anonymize turns in rooms they joined

Authorization stays in Go. Prompt is not policy. Admins remain invisible `system_admin` spectators.

## Models

No paid third-party LLM or image APIs. Stubs today; our adapters on our hardware later. Do not paste player tables into Colab or train sets.
