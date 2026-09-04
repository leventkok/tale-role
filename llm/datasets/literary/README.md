# literary storyteller seed

Hand-written beats for a future Storyteller LoRA. This is **not** `synthetic/storyteller.jsonl`.

That combinatorial file is an eval/contract set. Training on it taught the clipped salad voice. Live narration stays on **base instruct** until this seed is large enough to QLoRA (`TALEROLE_STORYTELLER_ADAPTER=1`).

Rules for new rows:

- Invent the scene. Do not paste a live player table.
- No emails, OTP, `system_admin`, lobby titles as places, theme ids.
- Prose does not invent dice, HP, or turn order. The UI already shows the count.
- A miss fail-forwards without granting the deed.
- If the deed stays put, the actor does not walk into a new place.
- Match live ChatML: grow rows that `services/llm-runner/serve.py` `storyteller_user` can render.

Train later from `llm/notebooks/qlora_storyteller_7b.ipynb` using this folder, not the salad JSONL.

Batches:

- `storyteller.jsonl` — hand-written seed
- `generated-01.jsonl` — 40 literary action rows (en/tr, hit/miss, stay-put). Do not merge into `synthetic/storyteller.jsonl`.
