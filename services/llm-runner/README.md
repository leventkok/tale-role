# llm-runner

Local HTTP process for **our** adapters. No paid third-party APIs. Weights stay on disk (`TALEROLE_ADAPTER_DIR`), never git.

Until PEFT weights exist, `--allow-unloaded` serves bound templates so the Go gateway path can be wired. Real QLoRA export comes from `llm/notebooks/qlora_mechanics_7b.ipynb`.

Storyteller and mechanics can run as **two processes** if you want them on separate GPUs or hosts. One process is enough until you measure load.

```powershell
cd C:\Users\leven\Documents\development\project\talerole
$env:TALEROLE_ADAPTER_DIR="D:\talerole-adapters"
python services/llm-runner/serve.py --role storyteller --port 8091 --adapter-dir $env:TALEROLE_ADAPTER_DIR\storyteller
python services/llm-runner/serve.py --role mechanics --port 8092 --adapter-dir $env:TALEROLE_ADAPTER_DIR\mechanics
```

API / gateway:

```powershell
$env:LLM_STORYTELLER_URL="http://127.0.0.1:8091"
$env:LLM_MECHANICS_URL="http://127.0.0.1:8092"
$env:TALEROLE_ADAPTER_DIR="D:\talerole-adapters"
```

`/health/ready` reports `"llm":"local"` only when the adapter dir looks ready **and** a runner URL is set. If the runner is down, the gateway falls back to the stub so the table still rolls in Go.
