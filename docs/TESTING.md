# How to test Tale Role

Work in `C:\Users\leven\Documents\development\project\talerole`. Public PRs stay the delivery path.

## Now — F1 (do this first)

PR: https://github.com/leventkok/tale-role/pull/2 (CI is green). Merge it on GitHub when you are happy.

1. Terminal A:

```powershell
cd C:\Users\leven\Documents\development\project\talerole\apps\api
$env:TALEROLE_DEV_OTP="123456"
$env:TALEROLE_ADMIN_EMAIL="admin@tale.role"
go run ./cmd/server
```

2. Terminal B (repo root):

```powershell
cd C:\Users\leven\Documents\development\project\talerole
$env:API_URL="http://127.0.0.1:8080"
npm run dev:web
```

3. Open http://127.0.0.1:3000
4. Create account → verify with `123456` → you should land signed in
5. Sign out, sign in again (no OTP if already verified)
6. Confirm DevTools → Application → Cookies: `talerole_session` is **HttpOnly** (not in localStorage)
7. Optional: `npm run dev:admin` → http://127.0.0.1:3001 spectator note

You are **not** testing email, Postgres, LLMs, or Electron polish yet.

## Next — F2 table (after the F2 PR is up)

Same two terminals. Then:

1. Sign in as a host → **Host** → create a d20 table → copy room id
2. Second browser / incognito: register another user → **Play** → join with that id
3. Both save a character (six stats, total **18**)
4. Host clicks **Roll initiative**
5. Try **Roll**, **Pass**, **Wait**
6. Admin account (`admin@tale.role` if env is set) can join; the player presence list must **not** show `system_admin`

In-memory store: restarting the API wipes users and rooms.

## Later — F3 dual LLM (after the F3 PR)

Same API + web terminals. Optional: `npm run dev:admin`.

1. Play a turn — the chronicle should show Storyteller prose. Dice still come from the Go engine.
2. Notes that contain an email must appear as `[redacted]` in admin traces (http://127.0.0.1:3001), never as raw PII.
3. Spectator signs in with `TALEROLE_ADMIN_EMAIL` on :3001, swaps `v1` → `v1-terse`, plays another turn; voice changes. Players do not see mechanic JSON.
4. `system_admin` still never appears in presence or Storyteller context.

Fine-tune weights are not in this phase.

## Later (do not test yet)

- F4: universe wizard + themes
- F5: scene images beside Storyteller
- F6: KVKK export/delete, signed desktop builds, real model eval
