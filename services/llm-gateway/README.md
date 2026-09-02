# llm-gateway

Storyteller (LLM1) and mechanics (LLM2) inference. Scale independently.

F3 ships **stub adapters** plus versioned prompt packs (`v1`, `v1-terse`). Production: private Hugging Face Hub repos + `services/llm-runner` on our GPU host. Set `HF_STORYTELLER_MODEL` / `HF_MECHANICS_MODEL` and runner URLs. Inference is `"hub"` only then. Runner down → stub. Dice still come from Go.

Fine-tune weights stay in Hub (private) — never git. Swap is live: the next turn uses the new pack.

PII redaction (email, long digit runs, phone-like strings) happens before a prompt is recorded or returned.

## Run

```bash
cd services/llm-gateway
go test ./...
LLM_GATEWAY_ADMIN_TOKEN=change-me go run ./cmd/gateway
```

The game API embeds the same library in-process. Point `LLM_STORYTELLER_URL` at the deployed runner for live inference.

Admin routes on this process require `Authorization: Bearer $LLM_GATEWAY_ADMIN_TOKEN`.
