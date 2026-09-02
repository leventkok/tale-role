# llm-gateway

Storyteller (LLM1) and mechanics (LLM2) inference. Scale independently.

F3 ships **stub adapters** plus versioned prompt packs (`v1`, `v1-terse`). Point `TALEROLE_ADAPTER_DIR` at our private adapters and `LLM_STORYTELLER_URL` / `LLM_MECHANICS_URL` at `services/llm-runner`. Inference is `"local"` only then; otherwise the stub stays. Runner down → stub fallback. Dice still come from Go.

Fine-tune weights stay in object storage — never git. Swap is live: the next turn uses the new pack.

PII redaction (email, long digit runs, phone-like strings) happens before a prompt is recorded or returned.

## Run

```bash
cd services/llm-gateway
go test ./...
LLM_GATEWAY_ADMIN_TOKEN=change-me go run ./cmd/gateway
```

The game API embeds the same library in-process so a table still narrates without a second process. Optional: two runner processes (storyteller vs mechanics) if you want them on separate machines.

Admin routes on this process require `Authorization: Bearer $LLM_GATEWAY_ADMIN_TOKEN`.
