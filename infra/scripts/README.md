# infra/scripts

Host bootstrap, seed, and adapter rollout scripts. Agents generate files that match `packages/game-schema` — they do not invent credentials.

## Local object store (MongoDB)

```powershell
.\infra\scripts\compose-up.ps1
```

```bash
./infra/scripts/compose-up.sh
```

Compose file: `infra/compose/docker-compose.yml` (no auth on localhost). Production credentials stay in the host secret manager.

Then start the API with `MONGO_URI=mongodb://127.0.0.1:27017` and `MONGO_DB=talerole`. `/health/ready` reports `"persistence":"mongo"`. Restart keeps users, rooms, and universes.

Mailhog is on SMTP `127.0.0.1:1025` and UI http://127.0.0.1:8025. Set `SMTP_HOST=127.0.0.1` and `SMTP_PORT=1025`. OTP codes appear in Mailhog, never in API JSON or logs. `TALEROLE_DEV_OTP` remains a local bypass.

Own-model training playbook: `llm/TRAINING.md`. Eval gate: `npm run eval` (synthetic JSONL, no PII, mechanics never emit dice).
