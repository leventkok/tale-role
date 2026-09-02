# Live topology

Tale Role is meant to run as hosted services, not on a laptop disk.

```
Browser → Next.js (Vercel or similar)
        → apps/api (Go)
             ├─ Mongo / later Postgres (hosted)
             ├─ SMTP (real provider; Mailhog is local-only)
             └─ HTTPS → llm-runner (GPU)
                          └─ Hugging Face Hub (private repos, from_pretrained)
```

`HF_TOKEN` only on the GPU runner. Game API only has model **ids** and runner **URLs**. No paid OpenAI/Anthropic/HF Inference API. Dice stay in Go.
