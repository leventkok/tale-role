# Colab serve — one-time setup, then Run All

Serve **our** Hugging Face adapters from free Colab GPU. No manual `serve.py` upload. Optional Render auto-sync so you do not paste URLs by hand.

## One-time only (≈15 minutes)

### A. Hugging Face token (Colab Secrets)

1. [huggingface.co/settings/tokens](https://huggingface.co/settings/tokens) → **Read** token.
2. Colab notebook → **Secrets** (key icon) → add `HF_TOKEN` → enable notebook access.

### B. Render API (optional — skips dashboard URL paste)

1. [dashboard.render.com/u/account/api-keys](https://dashboard.render.com/u/account/api-keys) → create key.
2. Colab Secrets → `RENDER_API_KEY`.
3. Render → your **API** web service → Settings → copy **Service ID** (`srv-...`).
4. Colab Secrets → `RENDER_SERVICE_ID`.

### C. Render env (stable — set once, never change)

Render → API service → Environment:

| Key | Value |
| --- | --- |
| `HF_STORYTELLER_MODEL` | `levonov/talerole-storyteller` |
| `HF_MECHANICS_MODEL` | `levonov/talerole-mechanics` |

Do **not** put `HF_TOKEN` on Render. Colab loads weights; Render only calls the public runner URL.

Save → deploy once.

## Every Colab session (storyteller)

1. Open [serve_colab.ipynb](https://colab.research.google.com/github/leventkok/tale-role/blob/main/llm/notebooks/serve_colab.ipynb)
2. **Runtime → Change runtime type → T4 GPU**
3. **Runtime → Run all**
4. Wait until `PUBLIC URL:` and (if configured) `Render LLM_STORYTELLER_URL updated`.
5. Test: `{PUBLIC URL}/health/live` → `"weights": true`

Keep this tab open while demoing. Closing Colab stops inference (stub returns in game).

## Every Colab session (mechanics)

Same steps with [serve_colab_mechanics.ipynb](https://colab.research.google.com/github/leventkok/tale-role/blob/main/llm/notebooks/serve_colab_mechanics.ipynb) in a **second** tab.

## What persists vs what resets

| Item | Persists? |
| --- | --- |
| HF models on Hub | Yes |
| Colab Secrets (`HF_TOKEN`, Render keys) | Yes |
| Render `HF_*_MODEL` env | Yes |
| Colab GPU session | No — close tab = stop |
| `trycloudflare.com` URL | No — new each session |
| Render URL env | Auto-updated if `RENDER_API_KEY` set; else paste once per session |

## Troubleshooting

- **`weights: false`** — wrong `HF_TOKEN` or repo name; check Hub repo is private and token has read access.
- **Runner exited** — scroll up for OOM; use one model per Colab tab.
- **Render still stub** — wait for deploy after URL sync; check API logs for `inference hub`.
