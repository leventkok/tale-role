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

Own-model training playbook: `llm/TRAINING.md`. Eval gate: `npm run eval` (synthetic JSONL, no PII, mechanics never emit dice).
