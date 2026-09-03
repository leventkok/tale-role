# datasets

Training and eval sets. Commit **synthetic, anonymized examples** only. Raw player logs stay out of git.

Eval: `npm test -w @tale-role/game-schema` (loads `synthetic/*.jsonl`).

`python llm/datasets/generate_storyteller.py` rebuilds `synthetic/storyteller.jsonl` as an **eval/contract** set (locale lock, dice digits, no PII). Do not train the live storyteller LoRA on it — that combinatorial salad is the clipped voice players already rejected. Live narration uses base instruct until a hand-written literary JSONL exists.
