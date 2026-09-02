# apps/web

Player + GM Next.js app. Default locale `en`. Session is an **HttpOnly** cookie set by the Next.js BFF (`API_URL` → Go). The JWT is never written to `localStorage`. `POST /api/graphql` proxies to Go `/graphql` and forwards the cookie when present (anonymous `health` / `me` still work).

```bash
# terminal 1
cd apps/api && go run ./cmd/server

# terminal 2 (repo root)
cp .env.example apps/web/.env.local   # then keep API_URL, drop secrets
npm run dev:web
```

Open http://127.0.0.1:3000
