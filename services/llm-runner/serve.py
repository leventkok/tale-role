#!/usr/bin/env python3
"""Production runner: load OUR adapters from Hugging Face Hub and serve them.

Not a paid Inference API. from_pretrained pulls private repos with HF_TOKEN
onto this GPU host's cache, then we run the model here.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import urlparse

EMAIL = re.compile(r"[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}", re.I)

_pipe = None
_pipe_lock = threading.Lock()


def redact(text: str) -> str:
    return EMAIL.sub("[redacted]", text or "")


def load_pipeline(model_id: str, token: str | None):
    import torch
    from huggingface_hub import list_repo_files
    from transformers import AutoModelForCausalLM, AutoTokenizer, BitsAndBytesConfig, pipeline

    quant = None
    if os.environ.get("HF_LOAD_IN_4BIT") == "1":
        use_bf16 = torch.cuda.is_available() and torch.cuda.is_bf16_supported()
        quant = BitsAndBytesConfig(
            load_in_4bit=True,
            bnb_4bit_quant_type="nf4",
            bnb_4bit_compute_dtype=torch.bfloat16 if use_bf16 else torch.float16,
        )
    tok_id = model_id
    files = set(list_repo_files(model_id, token=token))
    if "adapter_config.json" in files:
        from peft import PeftConfig, PeftModel

        cfg = PeftConfig.from_pretrained(model_id, token=token)
        tok_id = cfg.base_model_name_or_path
        tok = AutoTokenizer.from_pretrained(tok_id, token=token, trust_remote_code=True)
        if tok.pad_token is None:
            tok.pad_token = tok.eos_token
        base = AutoModelForCausalLM.from_pretrained(
            tok_id,
            token=token,
            device_map="auto",
            trust_remote_code=True,
            quantization_config=quant,
        )
        model = PeftModel.from_pretrained(base, model_id, token=token)
        return pipeline("text-generation", model=model, tokenizer=tok, return_full_text=False)

    tok = AutoTokenizer.from_pretrained(tok_id, token=token, trust_remote_code=True)
    if tok.pad_token is None:
        tok.pad_token = tok.eos_token
    model = AutoModelForCausalLM.from_pretrained(
        model_id,
        token=token,
        device_map="auto",
        trust_remote_code=True,
        quantization_config=quant,
    )
    return pipeline("text-generation", model=model, tokenizer=tok, return_full_text=False)


class Handler(BaseHTTPRequestHandler):
    role = "storyteller"
    model_id = ""
    allow_unloaded = False

    def log_message(self, fmt: str, *args) -> None:  # noqa: A003
        return

    def _send(self, status: int, payload: dict) -> None:
        raw = json.dumps(payload).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)

    def ready(self) -> bool:
        return bool(self.model_id) and (_pipe is not None or self.allow_unloaded)

    def do_GET(self) -> None:  # noqa: N802
        if urlparse(self.path).path == "/health/live":
            self._send(
                200,
                {
                    "status": "alive",
                    "role": self.role,
                    "source": "huggingface-hub",
                    "model_id": self.model_id,
                    "weights": self.ready(),
                },
            )
            return
        self._send(404, {"error": "not found"})

    def do_POST(self) -> None:  # noqa: N802
        length = int(self.headers.get("Content-Length") or 0)
        body = json.loads(self.rfile.read(length) or b"{}")
        path = urlparse(self.path).path
        if not self.ready():
            self._send(503, {"error": "model not loaded"})
            return
        if path == "/v1/narrate":
            self._send(200, self.narrate(body))
            return
        if path == "/v1/intent":
            self._send(200, self.intent(body))
            return
        self._send(404, {"error": "not found"})

    def generate(self, prompt: str) -> str:
        global _pipe
        if _pipe is None:
            return ""
        out = _pipe(prompt, max_new_tokens=180, do_sample=False)
        text = out[0]["generated_text"] if out else ""
        return redact(text.strip())

    def narrate(self, req: dict) -> dict:
        notes = redact(str(req.get("notes") or ""))
        actor = req.get("actor_name") or "Someone"
        total = req.get("total") or 0
        dice = req.get("dice_system") or "d20"
        locale = "tr" if req.get("locale") == "tr" else "en"
        prompt = (
            "Narrate. Never invent dice or HP. Engine already rolled "
            f"{dice} total {total}. Actor {actor}. Notes: {notes}\n"
        )
        prose = self.generate(prompt)
        if not prose:
            if locale == "tr":
                prose = f"[hub] {actor} salonun loşluğunda adım atar. {notes} Fener sönmez."
            else:
                prose = f"[hub] {actor} steps into the hall. {notes} The lantern holds."
        return {"locale": locale, "prose": redact(prose), "npc_lines": []}

    def intent(self, req: dict) -> dict:
        kind = req.get("kind") or "action"
        skill = req.get("skill") or ("str" if kind == "action" else "")
        notes = redact(str(req.get("notes") or ""))
        prompt = f"Return mechanic JSON only. kind={kind} skill={skill} notes={notes}\n"
        raw = self.generate(prompt)
        if raw.startswith("{"):
            try:
                parsed = json.loads(raw[raw.find("{") : raw.rfind("}") + 1])
                parsed["notes"] = redact(str(parsed.get("notes") or notes))
                if "rolls" in parsed or "hp" in parsed:
                    raise ValueError("model invented engine state")
                return parsed
            except (json.JSONDecodeError, ValueError):
                pass
        out = {"kind": kind, "notes": notes}
        if skill:
            out["skill"] = skill
        if kind == "action":
            out["dc"] = 12
        return out


def main() -> None:
    p = argparse.ArgumentParser(description="Tale Role Hugging Face Hub runner")
    p.add_argument("--role", choices=("storyteller", "mechanics"), required=True)
    p.add_argument("--host", default=os.environ.get("RUNNER_HOST", "0.0.0.0"))
    p.add_argument("--port", type=int, default=int(os.environ.get("PORT", "8091")))
    p.add_argument("--hf-model", default=os.environ.get("HF_MODEL_ID", ""))
    p.add_argument("--allow-unloaded", action="store_true")
    args = p.parse_args()
    if not args.hf_model:
        raise SystemExit("HF_MODEL_ID / --hf-model is required (private Hub repo, not a paid Inference API)")
    Handler.role = args.role
    Handler.model_id = args.hf_model
    Handler.allow_unloaded = args.allow_unloaded or os.environ.get("TALEROLE_RUNNER_UNLOADED") == "1"
    token = os.environ.get("HF_TOKEN") or None
    global _pipe
    if not Handler.allow_unloaded:
        with _pipe_lock:
            _pipe = load_pipeline(args.hf_model, token)
    httpd = ThreadingHTTPServer((args.host, args.port), Handler)
    print(f"llm-runner {args.role} hub={args.hf_model} on {args.host}:{args.port}", flush=True)
    httpd.serve_forever()


if __name__ == "__main__":
    main()
