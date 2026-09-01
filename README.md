# Tale Role

AI-hosted tabletop FRP. One model tells the story. Another proposes mechanics as JSON. A Go engine rolls dice, tracks HP, and owns the game state.

**North star:** the LLM narrates; the engine writes truth.

## Status

F1 platform shell: in-memory Go auth + OTP, Next.js BFF cookies, admin spectator origin, Electron `talerole://` stub.

## Invariants

- Public GitHub is the only delivery path (branch → PR → green CI → merge)
- Default UI locale: **English** (`tr` is first-class)
- Tale Core dice: **d20 default**, `2d6` as a plugin
- No regional/national UI themes
- System admins spectate every room invisibly

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
