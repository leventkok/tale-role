# Tale Role

AI-hosted tabletop FRP. One model tells the story. Another proposes mechanics as JSON. A Go engine rolls dice, tracks HP, and owns the game state.

**North star:** the LLM narrates; the engine writes truth.

## Status

F1 platform shell: in-memory Go auth + OTP, Next.js BFF cookies, admin spectator origin, Electron `talerole://` stub.

F3 dual LLM: in-process Storyteller + mechanics stubs, PII redaction, live prompt-pack swap, admin traces.

F4 universe interview: owner compiles a versioned prompt pack; rooms bind `theme_id`; players never see `compiled_prompt`.

F5 scene panel: stub SVG beside the Storyteller, keyed by theme. Not in chat. Turn loop does not wait on a real image model.

F6 privacy: export + erase, eval gate in CI, unsigned Electron pack config. Inference is still stub.

F7 Mongo: `infra/scripts/compose-up.ps1` starts object storage. Set `MONGO_URI` so users, rooms, and universes survive API restart. Identity on Mongo for this capstone; masterfabric-go Postgres remains the later IAM home.

F8 GraphQL: `POST /graphql` (and Next.js `/api/graphql`) wraps table, universe, and `me`. REST stays. No GraphiQL. Room queries never include `compiled_prompt`.

F9 SMTP OTP: Compose Mailhog (`127.0.0.1:1025`, UI `:8025`) delivers the 6-digit code. Store still bcrypt-hashes it. `TALEROLE_DEV_OTP` is a local bypass. Codes never appear in JSON or logs.

F10 local LLM: Colab QLoRA notebook + `services/llm-runner`. Gateway uses `"local"` only when adapters are on disk and a runner URL is set; otherwise stub. No paid APIs. Dice still come from Go.

## Invariants

- Public GitHub is the only delivery path (branch → PR → green CI → merge)
- Default UI locale: **English** (`tr` is first-class)
- Tale Core dice: **d20 default**, `2d6` as a plugin
- No regional/national UI themes
- System admins spectate every room invisibly
- No paid third-party LLM APIs; we train and host our own Storyteller + mechanics models

## Monorepo

| Path | Role |
| --- | --- |
| `apps/web` | Player + GM (Next.js) |
| `apps/admin` | Operator console (separate origin) |
| `apps/desktop` | Electron (Win + Mac) |
| `apps/api` | [masterfabric-go](https://github.com/gurkanfikretgunak/masterfabric-go) layout + game context |
| `packages/i18n` | Locale registry (`en`, `tr`) |
| `packages/game-schema` | Turn, intent, universe, presence contracts |
| `services/llm-gateway` | Storyteller + mechanics models |
| `services/llm-runner` | Local HTTP process for adapters on disk |
| `compliance/` | Threat model and secret handling |

Identity lives in Postgres (masterfabric-go). World state lives in MongoDB.

## Develop

```bash
npm install
npm test
cd apps/api && go test ./...
```

Web: `npm run dev:web` (needs `API_URL`, see `apps/web/README.md`). Admin: `npm run dev:admin` on port 3001.

See [CONTRIBUTING.md](CONTRIBUTING.md), [AGENTS.md](AGENTS.md), and [SECURITY.md](SECURITY.md).

## License

GNU Affero General Public License v3.0 (AGPL-3.0), aligned with masterfabric-go.
