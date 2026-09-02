# How to test Tale Role

Work in `C:\Users\leven\Documents\development\project\talerole`.

There is **no real LLM** yet. Chronicle prose is a stub template. Dice, HP, and turn order come from Go.

Without `MONGO_URI`, restarting the API wipes users and rooms. With Compose Mongo, they survive.

## One sitting — F1 + F2 + F3 stubs

Use three terminals. Keep them running.

### 0. Branch

Checkout the stack you want to click through (`feat/f9-smtp-otp` emails OTP via Mailhog; `feat/f8-graphql` adds `POST /graphql`; `feat/f7-mongo-hosting` has persistence when Mongo is up).

Optional Mongo + Mailhog:

```powershell
cd C:\Users\leven\Documents\development\project\talerole
.\infra\scripts\compose-up.ps1
```

Mailhog UI: http://127.0.0.1:8025

### 1. API (terminal A)

```powershell
cd C:\Users\leven\Documents\development\project\talerole\apps\api
$env:TALEROLE_DEV_OTP="123456"
$env:TALEROLE_ADMIN_EMAIL="admin@tale.role"
$env:CORS_ALLOWED_ORIGINS="http://localhost:3000,http://localhost:3001"
$env:MONGO_URI="mongodb://127.0.0.1:27017"
$env:MONGO_DB="talerole"
$env:SMTP_HOST="127.0.0.1"
$env:SMTP_PORT="1025"
go run ./cmd/server
```

Wait until it logs `listening` on `127.0.0.1:8080`.

### 2. Player web (terminal B)

```powershell
cd C:\Users\leven\Documents\development\project\talerole
$env:API_URL="http://127.0.0.1:8080"
npm run dev:web
```

Open http://127.0.0.1:3000

### 3. Spectator (terminal C)

```powershell
cd C:\Users\leven\Documents\development\project\talerole
$env:API_URL="http://127.0.0.1:8080"
npm run dev:admin
```

Leave http://127.0.0.1:3001 for later.

### 4. Accounts

| Role | Email | Password | OTP |
| --- | --- | --- | --- |
| Host | `host@tale.role` | `longenough` | Mailhog, or `123456` if `TALEROLE_DEV_OTP` is set |
| Player | `player@tale.role` | `longenough` | same |
| Spectator | `admin@tale.role` | `longenough` | same |

**Host (normal window)**

1. Register `host@tale.role` → verify the code from Mailhog (or `123456` with `TALEROLE_DEV_OTP`) → land signed in.
2. Sign out, sign in again. Second login must **not** ask for OTP.
3. DevTools → Application → Cookies: `talerole_session` is HttpOnly. JWT is **not** in `localStorage`.

**Player (incognito / second browser)**

4. Register `player@tale.role` → verify from Mailhog (or `123456`).

### 5. Table

**Host**

5. **Host** → name `Ashwood`, dice `d20`, access invite/id → **Create table**.
6. Copy the room id.

**Player**

7. **Play** → paste id → **Join**.

**Both**

8. Character name + six stats, each 1–6, **total 18** → save.
9. Host clicks **Roll initiative**. Turn order appears.
10. Player (or host) types an action, picks a skill, **Roll**. Chronicle shows dice **and** Storyteller prose (stub sentence, not a model).
11. Try **Pass** and **Wait**. Those skip dice.

### 6. PII + spectator

**Host or player**

12. On a roll, put an email in the action notes, e.g. `force the door, cc spy@tale.role`.
13. Prose must not repeat that email. Notes on the table may still show what you typed.

**Spectator (port 3001)**

14. Sign in as `admin@tale.role` (OTP from Mailhog, or `123456` if first time with `TALEROLE_DEV_OTP`).
15. Pack starts as `v1`. Click **Use v1-terse**.
16. Back on the table, play another turn. New prose should include `[v1-terse]`.
17. Traces list must show `[redacted]`, never `spy@tale.role`, and never raw `mechanic_intent` on the **player** chronicle.
18. Join the same room as this admin from the player app if you want; the roster must **not** list `system_admin`.

### Pass / fail

| Expect | Fail if |
| --- | --- |
| Cookie HttpOnly, no JWT in localStorage | Token appears in Application → Local Storage |
| Stats 18 required | Save works with total ≠ 18 |
| Dice from engine | Prose invents a different total than the dice line |
| Stub prose after a roll | Chronicle has only “action 14” with no sentence |
| Admin traces redact email | `spy@tale.role` in the :3001 trace list |
| Invisible spectator | `system_admin` in presence or turn order |

### Do not test yet

Postgres, trained production weights, signed Electron installers, production DPA.

## When our models land

No paid external LLM. Storyteller and mechanics are **our** fine-tunes, served from `services/llm-gateway` on our machines. Until adapters exist, the stub stays.

Order: synthetic datasets under `llm/datasets` → train adapters (weights never in git) → eval gate → gateway loads those adapters. F4 (universe wizard) is the first feature that needs a live narrator rather than a stub — the interview still compiles a prompt pack on the stub today.

## F4 — universe interview

Signed in: **Universe** → 3 steps (name/tone/taboos → theme + dice → NPC) → compile → Host with that universe selected. Room GET must show `theme_id` and must **not** include `compiled_prompt`. Dice still come from Go.

## F5 — scene panel

After a roll, the Storyteller **side panel** (not the chronicle) shows a stub SVG for the room theme. Turn JSON must not contain `image_svg`. No paid image API.

## F6 — privacy

Signed in: **Account** → download JSON (no password hash) → type `DELETE` to erase. After erase, `/api/me` is 401. **Privacy** is public. `/health/ready` reports `persistence: memory` (or `mongo` when `MONGO_URI` is set) and `llm: stub`.

## F7 — Mongo

`.\infra\scripts\compose-up.ps1` then API with `MONGO_URI`. Create a room, restart the API, join the same id. `/health/ready` → `"persistence":"mongo"`.

## F8 — GraphQL

Anonymous (no cookie):

```powershell
Invoke-RestMethod -Method Post -Uri http://127.0.0.1:3000/api/graphql -ContentType "application/json" -Body '{"query":"{ health { persistence llm } me { email } }"}'
```

`me` is null. Signed in, the same path forwards the session cookie. Room queries must not include `compiled_prompt`. Stolen `universe(id)` returns GraphQL errors, not the pack. Direct API: `POST http://127.0.0.1:8080/graphql`.

## F9 — SMTP OTP

Compose starts Mailhog. API with `SMTP_HOST=127.0.0.1` `SMTP_PORT=1025`. Register, open http://127.0.0.1:8025, copy the 6-digit code. `/health/ready` includes `"mail":"smtp"`. Register JSON must not contain the code. `TALEROLE_DEV_OTP` still bypasses randomness for local sitting; leave it unset to force Mailhog.

## F10 — local LLM

Without adapters, `/health/ready` stays `"llm":"stub"`. Train in Colab: `llm/notebooks/qlora_mechanics_7b.ipynb`. Put the export on a private disk, set `TALEROLE_ADAPTER_DIR`, start `python services/llm-runner/serve.py --role storyteller --port 8091`, set `LLM_STORYTELLER_URL`. Then ready reports `"llm":"local"`. Kill the runner: table still plays (stub fallback). Dice totals still come from Go. No paid API.
