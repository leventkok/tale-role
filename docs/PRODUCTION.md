# Live topology

Tale Role is meant to run as hosted services, not on a laptop disk.

```
Browser → Next.js (Vercel)
        → apps/api (Render)
             ├─ MongoDB Atlas
             ├─ OTP mail (Resend HTTPS on Render free; SMTP only if the host allows 587)
             └─ HTTPS → llm-runner (GPU, later)
                          └─ Hugging Face Hub (private repos, from_pretrained)
```

`HF_TOKEN` only on the GPU runner. Game API only has model **ids** and runner **URLs** (`LLM_STORYTELLER_URL` / `LLM_MECHANICS_URL`, or comma-separated `LLM_*_URLS` replicas). No paid OpenAI/Anthropic/HF Inference API. Dice stay in Go.

The Cloudflare tunnel to this PC (`localhost:3000/8080/3001`) is a smoke path. Cut DNS over after Vercel + Render are healthy. Do not set `TALEROLE_DEV_OTP` on hosted API.

## Hosts

| Public name | Service | Origin |
| --- | --- | --- |
| `talerole.com`, `www.talerole.com` | `apps/web` | Vercel |
| `admin.talerole.com` | `apps/admin` | Vercel |
| `api.talerole.com` | `apps/api` | Render (`talerole-api`) |

Vercel `API_URL` is server-only and points at `https://api.talerole.com` (or the `*.onrender.com` URL before the custom domain). The browser talks to Next, not to Go.

## Secrets (dashboard only)

Never git. Never paste into chat.

On **Render** (`talerole-api`):

- `JWT_SECRET` — Blueprint generates one; rotate if it leaked
- `RESEND_API_KEY` — dashboard only. OTP over HTTPS (Render free cannot use SMTP 587)
- `RESEND_FROM=Tale Role <onboarding@resend.dev>` until `talerole.com` is verified in Resend
- `MONGO_URI` — Atlas connection string (`mongodb+srv://…`)
- `MONGO_DB=talerole`
- `SMTP_HOST=smtp.hostinger.com`
- `SMTP_PORT=587`
- `SMTP_FROM=Tale Role <noreply@talerole.com>`
- `SMTP_USER` / `SMTP_PASS` — mailbox, not the Cloudflare token
- `CORS_ALLOWED_ORIGINS=https://talerole.com,https://www.talerole.com,https://admin.talerole.com`

On **Vercel** (web + admin):

- `API_URL=https://api.talerole.com`

Leave `TALEROLE_DEV_OTP` unset. Atlas Network Access must allow Render egress (for an alpha, `0.0.0.0/0` is the blunt option).

## Cut over from the laptop tunnel

1. Merge this hosting PR. Wait for green CI.
2. Create Atlas (M0) + a Hostinger mailbox `noreply@talerole.com`.
3. Deploy `talerole-api` on Render (Blueprint `render.yaml`, Frankfurt). The workspace already has free `navgo-api`; a second free web service may be rejected — use Starter if so.
4. Confirm `https://<service>.onrender.com/health/ready` returns `"persistence":"mongo"` and `"mail":"resend"` (or `"mail":"smtp"` on a host that allows 587).
5. Create two Vercel projects from `leventkok/tale-role`: root `apps/web` and `apps/admin`. Set `API_URL`.
6. Cloudflare DNS: replace tunnel CNAMEs with Vercel (`cname.vercel-dns.com`) for apex/`www`/`admin`, and the Render hostname for `api`. Keep MX/TXT.
7. Zero Trust: remove the published application routes that pointed at `localhost`.
8. SSL/TLS stays **Full**. Do not put Cloudflare Access in front of the player app.

`PORT` is set by Render. The API then binds `0.0.0.0`. Local `go run` without `PORT` still binds `127.0.0.1:8080`.
