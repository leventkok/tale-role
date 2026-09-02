# datasets

Training and eval sets. Commit **synthetic, anonymized examples** only. Raw player logs stay out of git.

Eval: `npm test -w @tale-role/game-schema` (loads `synthetic/*.jsonl`).

`python llm/datasets/generate_storyteller.py` rebuilds `synthetic/storyteller.jsonl` with the same quotas and uniqueness checks. Do not hand-type thousands of rows.
