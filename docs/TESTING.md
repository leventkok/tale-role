# How to test Tale Role

Work in `C:\Users\leven\Documents\development\project\talerole`.

There is **no real LLM** yet. Chronicle prose is a stub template. Dice, HP, and turn order come from Go.

Restarting the API wipes users and rooms (in-memory).

## One sitting — F1 + F2 + F3 stubs

Use three terminals. Keep them running.

### 0. Branch

Checkout the stack you want to click through (`feat/f3-llm-gateway` has auth + table + stub narrator).

### 1. API (terminal A)

```powershell
cd C:\Users\leven\Documents\development\project\talerole\apps\api
$env:TALEROLE_DEV_OTP="123456"
$env:TALEROLE_ADMIN_EMAIL="admin@tale.role"
$env:CORS_ALLOWED_ORIGINS="http://localhost:3000,http://localhost:3001"
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
| Host | `host@tale.role` | `longenough` | `123456` |
| Player | `player@tale.role` | `longenough` | `123456` |
| Spectator | `admin@tale.role` | `longenough` | `123456` |

**Host (normal window)**

1. Register `host@tale.role` → verify `123456` → land signed in.
2. Sign out, sign in again. Second login must **not** ask for OTP.
3. DevTools → Application → Cookies: `talerole_session` is HttpOnly. JWT is **not** in `localStorage`.

**Player (incognito / second browser)**

4. Register `player@tale.role` → verify `123456`.

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

14. Sign in as `admin@tale.role` (OTP `123456` if first time).
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

Email delivery, Postgres, Mongo, trained models, scene images, Electron signing, KVKK export.

## When our models land

No paid external LLM. Storyteller and mechanics are **our** fine-tunes, served from `services/llm-gateway` on our machines. Until adapters exist, the stub stays.

Order: synthetic datasets under `llm/datasets` → train adapters (weights never in git) → eval gate → gateway loads those adapters. F4 (universe wizard) is the first feature that needs a live narrator rather than a stub.
