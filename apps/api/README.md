# apps/api

Tale Role identity API. Layout follows [masterfabric-go](https://github.com/gurkanfikretgunak/masterfabric-go): Chi, hexagonal folders, generic 5xx, CORS allow-list, JWT `jti`.

F1 uses an in-memory store so CI can run without Postgres. Postgres IAM lands when we vendor the full masterfabric-go runtime.

## Run

```bash
cd apps/api
go test ./...
go run ./cmd/server
```

## Auth

- `POST /api/v1/auth/register` — 6-digit OTP required (code is never in the JSON body)
- `POST /api/v1/auth/otp/verify`
- `POST /api/v1/auth/login`
- `GET /api/v1/me`
- `POST /api/v1/licenses/register` — Electron registered product
- `GET /api/v1/licenses/me`

OTP is emailed when `SMTP_HOST` is set (local: Mailhog). Tests inject a fixed issuer and skip SMTP. Codes are bcrypt-hashed at rest and never written to logs or JSON. `TALEROLE_DEV_OTP` is a local-only bypass.

## LLM

Turns call the in-process llm-gateway library after the engine writes dice: mechanics JSON is traced for admins; Storyteller prose is attached to the public turn. `system_admin` is omitted from that context. PII is redacted before prompts are stored.

