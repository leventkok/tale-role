# Training our models

Tale Role does **not** call paid third-party LLM APIs (no OpenAI, Anthropic, Hugging Face Inference API). Storyteller and mechanics are **our** fine-tunes.

Weights live in **private Hugging Face Hub repos**. The GPU runner does `from_pretrained(HF_MODEL_ID)` on our host. The game API never holds tensors or `HF_TOKEN`.

Until Hub repos and a runner URL exist, the gateway uses the **stub**. That is expected.

## What lives in git

- Synthetic JSONL under `llm/datasets/synthetic/` (no player logs, no PII)
- Model cards (`llm/storyteller/card.json`, `llm/mechanics/card.json`)
- This playbook and the Colab notebook

## What never lives in git

- `.safetensors`, `.gguf`, `.bin`, raw transcripts, emails, `HF_TOKEN`

## Colab (7B QLoRA)

Open `llm/notebooks/qlora_mechanics_7b.ipynb` or `llm/notebooks/qlora_storyteller_7b.ipynb` on a **separate** GPU runtime. Last cell pushes that adapter to a **private** Hub repo. Do not train both roles in one Colab session. The 32B storyteller card waits for a larger GPU host.

## Recipe (live)

1. Grow `llm/datasets/synthetic/`. Keep `locale` `en` or `tr`.
2. `npm test -w @tale-role/game-schema` — eval rejects PII, `system_admin`, and mechanics rows that invent dice.
3. Fine-tune QLoRA. Constrained decoding for mechanics JSON.
4. Upload adapters to private Hub repos (`your-org/talerole-storyteller`, `your-org/talerole-mechanics`).
5. Deploy `services/llm-runner` on a GPU host with `HF_MODEL_ID` + `HF_TOKEN`.
6. Game API: `HF_STORYTELLER_MODEL`, `HF_MECHANICS_MODEL`, `LLM_STORYTELLER_URL`, `LLM_MECHANICS_URL`. Inference becomes `"hub"`.
7. Never serve from the browser. Dice still come from Go.

Do not paste player tables into the train set. Synthesize.
