# Tale Role — agent instructions

Public GitHub is the only delivery path. Local-only work does not count as shipped.

## Rule 0

Every meaningful change: `branch` → conventional commit → public PR → green CI → merge. Do not push to `main`. Do not leave days of unpushed commits.

Commit style (cursor-security):

```
feat: add locale registry with english default

Players without a preference land on en. Turkish remains a first-class catalog.

```

Prefixes: `feat:`, `fix:`, `ci:`, `chore:`, `docs:`, `security:`. Title is imperative; lowercase after the prefix. PR body has summary, test checklist, and a security note.

## Secrets

Never commit `.env`, JWT secrets, API keys, DSNs, license keys, OTP codes, invite tokens, signing `.pem`, or PII transcripts. `.env.example` uses `change-me` only. `NEXT_PUBLIC_*` is not a secret vault. Session tokens belong in HttpOnly cookies, not `localStorage`.

## Product invariants

- North star: the LLM narrates; the Go rules engine writes truth (dice, HP, inventory, turn order).
- Tale Core dice: infrastructure supports `d20` and `2d6`; **default is `d20`**.
- UI locales: `en` (default) and `tr`. Add a language with a catalog + LLM locale, not hardcoded strings.
- No local/national UI themes (no Anatolian, Ottoman, or regional packs).
- System admins join every room as an invisible `system_admin` spectator. They never appear in roster, turn order, or Storyteller context.
- Prompt is not policy. Authorization lives in Go / server actions.
- **No paid third-party LLM APIs** (no OpenAI, Anthropic, etc.). Storyteller and mechanics are our own fine-tunes, served by `services/llm-gateway` on our hardware. Weights stay in private object storage — never git.

## Least agency

Agents write files that match schemas in `packages/game-schema` and scripts under `infra/scripts`. Do not invent host credentials. High-impact actions (prod deploy, secret rotation, license issuance) need a human in the loop.

## Layout

- `apps/web`, `apps/admin`, `apps/desktop`, `apps/api` (masterfabric-go layout)
- `packages/i18n`, `packages/game-schema`, `packages/ui`, `packages/rules-plugins`
- `services/llm-gateway`, `services/image-worker`
- `compliance/` threat model and secret handling
