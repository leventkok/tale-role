# Training our models

Tale Role does **not** call paid third-party LLM APIs. Storyteller and mechanics are our fine-tunes, loaded by `services/llm-gateway`.

Until adapters exist, the gateway uses the **stub**. That is expected.

## What lives in git

- Synthetic JSONL under `llm/datasets/synthetic/` (no player logs, no PII)
- Model cards (`llm/storyteller/card.json`, `llm/mechanics/card.json`)
- This playbook

## What never lives in git

- `.safetensors`, `.gguf`, `.bin`, raw transcripts, emails, API keys

Put trained adapters in private object storage or a disk path **outside** the repo, then point the process at that directory:

```powershell
$env:TALEROLE_ADAPTER_DIR="D:\talerole-adapters"
```

Gateway reports `weights_ready` only if that directory contains a recognizable adapter file. Inference is `"local"` only when that is true **and** a runner URL is set (`LLM_STORYTELLER_URL` / `LLM_MECHANICS_URL` / `LLM_RUNNER_URL`). Without either, runtime stays on `stub`.

## Colab (7B QLoRA)

Open `llm/notebooks/qlora_mechanics_7b.ipynb` on a GPU runtime. It trains the mechanics adapter on synthetic JSONL. Export the adapter folder to private disk — never commit `.safetensors`. Storyteller can use the same 7B recipe as a stand-in; the 32B card waits for a larger host.

## Recipe (your GPU host)

1. Grow `llm/datasets/synthetic/` with more invented scenes. Keep `locale` `en` or `tr`.
2. Run `npm test -w @tale-role/game-schema` — eval rejects PII, `system_admin`, and mechanics rows that invent dice.
3. Fine-tune QLoRA (Colab notebook or this playbook) on the open base named in each card. Constrained decoding for mechanics JSON.
4. Export adapters to `$TALEROLE_ADAPTER_DIR/storyteller` and `.../mechanics`.
5. Start `services/llm-runner/serve.py` per role (optional split; one process is fine).
6. Point the API at those URLs. Eval again on a held-out slice before swapping the live pack.
7. Serve through `services/llm-gateway` only — never from the browser.

Do not paste player tables into the train set. Synthesize.
