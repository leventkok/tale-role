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

OTP delivery (email) is not wired yet; tests inject a fixed issuer. Production must send mail and keep codes out of logs.
