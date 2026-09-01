# Tale Role

AI-hosted tabletop FRP. One model tells the story. Another proposes mechanics as JSON. A Go engine rolls dice, tracks HP, and owns the game state.

**North star:** the LLM narrates; the engine writes truth.

## Status

Foundation (F0). Auth, lobbies, and models are not shipped yet.

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
```

Copy `.env.example` to `.env.local`. Never commit real secrets.

See [CONTRIBUTING.md](CONTRIBUTING.md), [AGENTS.md](AGENTS.md), and [SECURITY.md](SECURITY.md).

## License

GNU Affero General Public License v3.0 (AGPL-3.0), aligned with masterfabric-go.
