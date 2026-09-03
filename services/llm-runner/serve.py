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
from typing import Any
from urllib.parse import urlparse

EMAIL = re.compile(r"[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}", re.I)
IM_END = "<|" + "im_end" + "|>"

STORYTELLER_SYSTEM = (
    "Narrate the engine result in the locale. "
    "On a successful action, include the total number in the prose. "
    "Never invent dice, HP, or turn order."
)
MECHANICS_SYSTEM = "Return only mechanic JSON. Never dice, HP, or turn order."

EN_OK = (
    "The latch yields; a gap opens.",
    "No shout follows. It worked.",
    "The room answers. The thing moves.",
)
EN_FAIL = (
    "The lock holds. A pin snaps.",
    "The tool slips. Time is gone.",
    "Boots scrape closer. The attempt dies.",
)
TR_OK = (
    "Mandal teslim olur; bir aralık açılır.",
    "Çığlık gelmez. İş bitti.",
    "Oda cevap verir. Şey hareket eder.",
)
TR_FAIL = (
    "Kilit durur. Bir pim kopar.",
    "Alet kayar. Zaman biter.",
    "Çizmeler yaklaşır. Deneme ölür.",
)

_pipe = None
_pipe_lock = threading.Lock()


def redact(text: str) -> str:
    return EMAIL.sub("[redacted]", text or "")


def extract_json_object(raw: str) -> dict[str, Any] | None:
    raw = (raw or "").strip()
    for stop in (IM_END, "<|" + "im_start" + "|>"):
        if stop in raw:
            raw = raw.split(stop, 1)[0].strip()
    start = raw.find("{")
    if start < 0:
        return None
    depth = 0
    for idx in range(start, len(raw)):
        ch = raw[idx]
        if ch == "{":
            depth += 1
        elif ch == "}":
            depth -= 1
            if depth == 0:
                try:
                    parsed = json.loads(raw[start : idx + 1])
                except json.JSONDecodeError:
                    return None
                return parsed if isinstance(parsed, dict) else None
    return None


def prose_looks_valid(prose: str) -> bool:
    text = (prose or "").strip()
    if len(text) < 12:
        return False
    if text.startswith("{") or text.startswith("["):
        return False
    lowered = text.casefold()
    if "never invent dice" in lowered or ("<|" + "im_start" + "|>") in text:
        return False
    if '"actor"' in text and '"room"' in text:
        return False
    if text.count("The bar splinters") > 1:
        return False
    return True


def pick_tag(locale: str, ok: bool | None, seed: str) -> str:
    pool = (TR_OK if ok else TR_FAIL) if locale == "tr" else (EN_OK if ok else EN_FAIL)
    idx = sum(ord(c) for c in seed) % len(pool)
    return pool[idx]


def normalize_kind(kind: str) -> str:
    kind = (kind or "action").strip().casefold()
    if kind in ("pass", "wait", "action"):
        return kind
    if kind == "say":
        return "say"
    return "action"


def storyteller_input(req: dict[str, Any]) -> tuple[str, dict[str, Any]]:
    locale = "tr" if req.get("locale") == "tr" else "en"
    kind = normalize_kind(str(req.get("kind") or "action"))
    actor = str(req.get("actor_name") or "Someone")
    room = str(req.get("room_name") or "Hall")
    notes = redact(str(req.get("notes") or ""))
    dice = str(req.get("dice_system") or "d20")
    presence = req.get("presence_names") or []
    if not isinstance(presence, list):
        presence = [actor]
    presence = [str(name) for name in presence if str(name).strip()]
    if actor not in presence:
        presence = [actor, *presence]

    rolls = req.get("rolls") or []
    if not isinstance(rolls, list):
        rolls = []
    rolls = [int(v) for v in rolls]

    total = req.get("total")
    if total is None:
        total = 0
    else:
        total = int(total)

    success = req.get("success")
    if kind in ("pass", "wait", "say"):
        rolls = []
        total = 0
        success = None
        model_kind = "wait" if kind == "say" else kind
    else:
        model_kind = kind

    payload = {
        "actor": actor,
        "room": room,
        "kind": model_kind,
        "dice": dice,
        "rolls": rolls,
        "total": total,
        "success": success,
        "notes": notes,
        "presence": presence,
    }
    return locale, payload


def fallback_storyteller(locale: str, payload: dict[str, Any], *, say: bool) -> dict[str, Any]:
    actor = payload["actor"]
    room = payload["room"]
    notes = payload["notes"]
    kind = payload["kind"]
    total = int(payload.get("total") or 0)
    success = payload.get("success")
    seed = f"{actor}:{room}:{total}:{notes}"

    if say:
        if locale == "tr":
            prose = f"{actor}, {room} içinde söz alır: {notes}" if notes else f"{actor}, {room} içinde sessizliği kırar."
        elif notes:
            prose = f'{actor} speaks in {room}. "{notes}"'
        else:
            prose = f"{actor} takes the floor in {room}."
        return {"locale": locale, "prose": redact(prose), "npc_lines": []}

    if kind == "pass":
        if locale == "tr":
            prose = f"{actor}, {room} içinde, {notes}. Bu kez zar yok. Masa durur."
        else:
            prose = f"{actor} in {room} chooses to {notes}. No dice this time. The table waits."
    elif kind == "wait":
        if locale == "tr":
            prose = f"{actor} {room} içinde bekler, {notes}, zar atılmaz. Masa durur."
        else:
            prose = f"{actor} waits in {room}, {notes}, and does not roll. The table waits."
    elif success is True:
        tag = pick_tag(locale, True, seed)
        if locale == "tr":
            prose = f"{actor}, {room} içinde, {notes}. Motorun {total}'i tutar. {tag}"
        else:
            prose = f"{actor} tries to {notes} in {room}. The engine's {total} lands true. {tag}"
    else:
        tag = pick_tag(locale, False, seed)
        if locale == "tr":
            prose = f"{actor}, {room} içinde, {notes}. Zar {total} der. {tag}"
        else:
            prose = f"{actor} tries to {notes} in {room}. The die reads {total}. {tag}"

    return {"locale": locale, "prose": redact(prose), "npc_lines": []}


def parse_storyteller_response(raw: str) -> dict[str, Any] | None:
    parsed = extract_json_object(raw)
    if not parsed:
        return None
    prose = redact(str(parsed.get("prose") or "").strip())
    if not prose_looks_valid(prose):
        return None
    npc_lines: list[dict[str, str]] = []
    raw_lines = parsed.get("npc_lines")
    if isinstance(raw_lines, list):
        for line in raw_lines:
            if not isinstance(line, dict):
                continue
            npc_id = str(line.get("npc_id") or "").strip()
            text = redact(str(line.get("text") or "").strip())
            if not npc_id or not text or npc_id == "system_admin":
                continue
            npc_lines.append({"npc_id": npc_id, "text": text})
    return {"prose": prose, "npc_lines": npc_lines}


def mechanics_input(req: dict[str, Any]) -> dict[str, Any]:
    kind = normalize_kind(str(req.get("kind") or "action"))
    if kind == "say":
        kind = "wait"
    skill = str(req.get("skill") or ("str" if kind == "action" else ""))
    notes = redact(str(req.get("notes") or ""))
    payload: dict[str, Any] = {"notes": notes, "kind": kind}
    if skill:
        payload["skill"] = skill
    return payload


def parse_mechanics_response(raw: str, notes: str) -> dict[str, Any] | None:
    parsed = extract_json_object(raw)
    if not parsed:
        return None
    if "rolls" in parsed or "hp" in parsed:
        return None
    kind = str(parsed.get("kind") or "").strip()
    if not kind:
        return None
    parsed["kind"] = kind
    parsed["notes"] = redact(str(parsed.get("notes") or notes))
    if "skill" in parsed and parsed["skill"] is not None:
        parsed["skill"] = str(parsed["skill"])
    return parsed


def chat_prompt(tokenizer: Any, system: str, user: str) -> str:
    messages = [
        {"role": "system", "content": system},
        {"role": "user", "content": user},
    ]
    if hasattr(tokenizer, "apply_chat_template"):
        return tokenizer.apply_chat_template(
            messages,
            tokenize=False,
            add_generation_prompt=True,
        )
    return (
        "<|" + "im_start|>system\n" + system + IM_END + "\n"
        "<|" + "im_start|>user\n" + user + IM_END + "\n"
        "<|" + "im_start|>assistant\n"
    )


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

    def generate(self, prompt: str, *, max_new_tokens: int = 240) -> str:
        global _pipe
        if _pipe is None:
            return ""
        out = _pipe(
            prompt,
            max_new_tokens=max_new_tokens,
            do_sample=False,
            repetition_penalty=1.08,
            return_full_text=False,
        )
        text = out[0]["generated_text"] if out else ""
        return redact(str(text).strip())

    def narrate(self, req: dict) -> dict:
        locale, payload = storyteller_input(req)
        say = normalize_kind(str(req.get("kind") or "action")) == "say"
        if say:
            return fallback_storyteller(locale, payload, say=True)

        global _pipe
        tokenizer = getattr(_pipe, "tokenizer", None) if _pipe is not None else None
        user = json.dumps(payload, ensure_ascii=False, separators=(",", ":"))
        if tokenizer is None:
            parsed = None
        else:
            prompt = chat_prompt(tokenizer, STORYTELLER_SYSTEM, user)
            raw = self.generate(prompt)
            parsed = parse_storyteller_response(raw)

        if parsed:
            return {"locale": locale, "prose": parsed["prose"], "npc_lines": parsed["npc_lines"]}
        return fallback_storyteller(locale, payload, say=False)

    def intent(self, req: dict) -> dict:
        payload = mechanics_input(req)
        notes = payload["notes"]
        kind = payload["kind"]
        skill = str(payload.get("skill") or "")

        global _pipe
        tokenizer = getattr(_pipe, "tokenizer", None) if _pipe is not None else None
        if tokenizer is not None:
            user = json.dumps(payload, ensure_ascii=False, separators=(",", ":"))
            prompt = chat_prompt(tokenizer, MECHANICS_SYSTEM, user)
            raw = self.generate(prompt, max_new_tokens=120)
            parsed = parse_mechanics_response(raw, notes)
            if parsed:
                return parsed

        out: dict[str, Any] = {"kind": kind, "notes": notes}
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
