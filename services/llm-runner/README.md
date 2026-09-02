# llm-runner

GPU process that **pulls our adapters from Hugging Face Hub** (`from_pretrained`) and serves `/v1/narrate` and `/v1/intent`.

This is **not** Hugging Face Inference API (paid, third-party). Hub is private object storage for weights. Inference runs on **our** host. `HF_TOKEN` stays in the host secret manager — never git, never the browser, never the game API.

## Production

1. Push LoRA/adapters to **private** Hub repos (Colab notebook last cell, or `huggingface-cli upload`).
2. Deploy this service on a GPU box (Render GPU, Fly, your VM).
3. Env on the **runner**:

```
HF_MODEL_ID=your-org/talerole-storyteller
HF_TOKEN=hf_...          # read-only, private repo
PORT=8091
```

Second process for mechanics with `HF_MODEL_ID=your-org/talerole-mechanics`.

4. Env on the **game API** (no HF token):

```
HF_STORYTELLER_MODEL=your-org/talerole-storyteller
HF_MECHANICS_MODEL=your-org/talerole-mechanics
LLM_STORYTELLER_URL=https://storyteller.your.host
LLM_MECHANICS_URL=https://mechanics.your.host
```

`/health/ready` is `"llm":"hub"` when Hub model ids and runner URLs are set. Runner down → stub, dice still from Go.

CI never downloads weights.
