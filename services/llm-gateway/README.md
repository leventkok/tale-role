# llm-gateway

Storyteller (LLM1) and mechanics (LLM2) inference. Scale independently.

F3 ships **stub adapters** plus versioned prompt packs (`v1`, `v1-terse`). Our fine-tunes land later via `TALEROLE_ADAPTER_DIR` (files on our disk — no paid API). Inference stays stub until the local runner is wired.

Fine-tune weights stay in object storage — never git. Swap is live: the next turn uses the new pack.

PII redaction (email, long digit runs, phone-like strings) happens before a prompt is recorded or returned.

## Run

```bash
cd services/llm-gateway
go test ./...
LLM_GATEWAY_ADMIN_TOKEN=change-me go run ./cmd/gateway
```

The game API embeds the same library in-process so a table still narrates without a second process. Point `LLM_GATEWAY_URL` at this service later to split scale.

Admin routes on this process require `Authorization: Bearer $LLM_GATEWAY_ADMIN_TOKEN`.
