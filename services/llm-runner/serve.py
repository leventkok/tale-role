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

MECHANICS_SYSTEM = "Return only mechanic JSON. Never dice, HP, or turn order."

TR_LETTERS = set("çğıöşüÇĞİÖŞÜ")
EN_STOP = re.compile(r"\b(the|and|of|to|in|is|with|that|this|for|are|was|not|you|your)\b", re.I)
TR_STOP = re.compile(r"\b(ve|bir|için|ile|bu|da|de|ki|ama|gibi|olan|sonra|içinde|üzerinde)\b", re.I)
STOCK = (
    "the watch is unblinded",
    "hold the line",
    "keep the hand on the hilt",
    "the bar splinters",
    "the engine's die reads",
    "never invent dice",
    "what will you do",
)

EN_OK = (
    "A gap opens where the wood used to hold.",
    "No shout follows. The work is done.",
    "The room answers and the thing moves.",
)
EN_FAIL = (
    "The lock holds. A pin snaps in the dark.",
    "The tool slips. Time is gone.",
    "Boots scrape closer. The attempt dies.",
)
TR_OK = (
    "Ahşabın tuttuğu yerde bir aralık açılır.",
    "Çığlık gelmez. İş biter.",
    "Oda cevap verir; şey yerinden oynar.",
)
TR_FAIL = (
    "Kilit durur. Karanlıkta bir pim kopar.",
    "Alet kayar. Zaman biter.",
    "Çizmeler yaklaşır. Deneme ölür.",
)
EN_OPEN = (
    "Night holds {room}. {cast} stand at the threshold. The tale begins before anyone moves.",
    "A hush settles over {room}. Lanternlight finds {cast}. Something in the dark is already listening.",
    "{room} waits with a held breath. {cast} have arrived. The storyteller takes the floor.",
)
TR_OPEN = (
    "{room} sessizliğe gömülür. {cast} eşiğe durur. Zar atılmadan hikâye başlar.",
    "{room} üzerinde bir durgunluk var. Fener {cast} yüzünü bulur. Karanlıkta biri dinliyor.",
    "{room} bekler. {cast} gelmiştir. Anlatıcı sözü alır.",
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


def prior_list(req: dict[str, Any]) -> list[str]:
    raw = req.get("prior") or []
    if not isinstance(raw, list):
        return []
    out: list[str] = []
    for item in raw:
        text = redact(str(item or "").strip())
        if text:
            out.append(text)
    return out[:3]


def too_similar(prose: str, prior: list[str]) -> bool:
    p = prose.casefold()
    for old in prior:
        o = (old or "").casefold().strip()
        if len(o) < 16:
            continue
        if o[:40] in p:
            return True
        for sent in re.split(r"[.!?]", o):
            sent = sent.strip()
            if len(sent) >= 24 and sent in p:
                return True
    return False


def locale_matches(prose: str, locale: str) -> bool:
    text = prose.strip()
    if not text:
        return False
    tr_marks = sum(1 for ch in text if ch in TR_LETTERS)
    en_n = len(EN_STOP.findall(text))
    tr_n = len(TR_STOP.findall(text))
    if locale == "tr":
        if tr_marks > 0:
            return en_n <= tr_n + 2
        return tr_n >= en_n and en_n < 6
    if tr_marks > 0:
        return False
    return en_n >= tr_n


def prose_looks_valid(prose: str, locale: str = "en", prior: list[str] | None = None, *, opening: bool = False) -> bool:
    text = (prose or "").strip()
    if len(text) < 12:
        return False
    if text.startswith("{") or text.startswith("["):
        return False
    lowered = text.casefold()
    if ("<|" + "im_start" + "|>") in text:
        return False
    if '"actor"' in text and '"room"' in text:
        return False
    if any(stock in lowered for stock in STOCK):
        return False
    if text.count(".") + text.count("!") + text.count("?") > 12:
        return False
    if too_similar(text, prior or []):
        return False
    if opening:
        return True
    if not locale_matches(text, locale):
        return False
    return True


def pick_line(pool: tuple[str, ...], seed: str) -> str:
    idx = sum(ord(c) for c in seed) % len(pool)
    return pool[idx]


def pick_tag(locale: str, ok: bool | None, seed: str) -> str:
    pool = (TR_OK if ok else TR_FAIL) if locale == "tr" else (EN_OK if ok else EN_FAIL)
    return pick_line(pool, seed)


def normalize_kind(kind: str) -> str:
    kind = (kind or "action").strip().casefold()
    if kind in ("pass", "wait", "action", "say", "story"):
        return kind
    return "action"


def storyteller_system(locale: str, *, opening: bool, prior: list[str]) -> str:
    lang = "Turkish" if locale == "tr" else "English"
    lock = (
        f"Write the prose field entirely in {lang}. "
        "Do not mix English and Turkish. Do not switch language mid-sentence. "
        "Never invent dice, HP, or turn order. Never echo the user JSON."
    )
    if opening:
        body = (
            "You are the table storyteller. Open the tale in 3 to 5 literary sentences. "
            "Set place, weather, and tension. Name who is present. "
            "Do not mention dice, rolls, waiting, or passing. "
            + lock
            + ' Return only JSON {"prose":"...","npc_lines":[]}.'
        )
    else:
        body = (
            f"Narrate the engine result in {lang}. "
            "On a successful action, include the total number in the prose. "
            "Vary metaphor. Do not reuse stock training lines. "
            + lock
        )
    if prior:
        clipped = " | ".join(p[:120] for p in prior if p)
        if clipped:
            body += f" Do not repeat these prior lines: {clipped}"
    return body


def storyteller_input(req: dict[str, Any]) -> tuple[str, dict[str, Any]]:
    locale = "tr" if req.get("locale") == "tr" else "en"
    kind = normalize_kind(str(req.get("kind") or "action"))
    actor = str(req.get("actor_name") or "Someone")
    room = str(req.get("room_name") or "Hall")
    notes = redact(str(req.get("notes") or req.get("opening") or ""))
    dice = str(req.get("dice_system") or "d20")
    theme = redact(str(req.get("theme_id") or ""))
    presence = req.get("presence_names") or []
    if not isinstance(presence, list):
        presence = [actor]
    presence = [str(name) for name in presence if str(name).strip() and str(name) != "system_admin"]
    if kind == "story":
        actor = "Anlatıcı" if locale == "tr" else "Storyteller"
    elif actor not in presence:
        presence = [actor, *presence]

    rolls = req.get("rolls") or []
    if not isinstance(rolls, list):
        rolls = []
    rolls = [int(v) for v in rolls]

    total = req.get("total")
    total = 0 if total is None else int(total)
    success = req.get("success")
    if kind in ("pass", "wait", "say", "story"):
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
        "theme": theme,
    }
    return locale, payload


def fallback_opening(locale: str, payload: dict[str, Any]) -> str:
    room = str(payload.get("room") or "").strip()
    notes = str(payload.get("notes") or "").strip()
    if notes:
        return notes
    seed = f"{room}:{locale}"
    if locale == "tr":
        if room:
            return pick_line(TR_OPEN, seed).format(room=room, cast="kahramanlar")
        return "Fener yanar. Eşikte bir duraklama var. Anlatıcı sözü alır."
    if room:
        return pick_line(EN_OPEN, seed).format(room=room, cast="the company")
    return "A hush. Lanternlight. The storyteller takes the floor."


def fallback_storyteller(locale: str, payload: dict[str, Any], *, say: bool) -> dict[str, Any]:
    actor = payload["actor"]
    room = payload["room"]
    notes = payload["notes"]
    kind = payload["kind"]
    total = int(payload.get("total") or 0)
    success = payload.get("success")
    seed = f"{actor}:{room}:{total}:{notes}"

    if kind == "story":
        return {"locale": locale, "prose": redact(fallback_opening(locale, payload)), "npc_lines": []}

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
            prose = f"{actor} {room}'da bu eli bırakır. {notes} Zar yok; masa dinler."
        else:
            prose = f"{actor} yields the beat in {room}. {notes} No roll. The hall listens."
    elif kind == "wait":
        if locale == "tr":
            prose = f"{actor} {room}'da bekler. {notes} Nefes tutulur; henüz zar yok."
        else:
            prose = f"{actor} holds still in {room}. {notes} Breath only. No roll yet."
    elif success is True:
        tag = pick_tag(locale, True, seed)
        if locale == "tr":
            prose = f"{actor} {notes}. {room} yol verir. Sayı {total} — motor zaten yazdı. {tag}"
        else:
            prose = f"{actor} {notes}. {room} gives way. The count is {total}, already written. {tag}"
    else:
        tag = pick_tag(locale, False, seed)
        if locale == "tr":
            prose = f"{actor} {notes}. {room} direnir. Sayı {total}. {tag}"
        else:
            prose = f"{actor} {notes}. {room} holds. The count is {total}. {tag}"

    return {"locale": locale, "prose": redact(prose), "npc_lines": []}


def parse_storyteller_response(
    raw: str, locale: str = "en", prior: list[str] | None = None, *, opening: bool = False
) -> dict[str, Any] | None:
    parsed = extract_json_object(raw)
    prose = ""
    npc_raw: Any = []
    if parsed:
        prose = redact(str(parsed.get("prose") or "").strip())
        npc_raw = parsed.get("npc_lines")
    elif raw and not raw.lstrip().startswith("{"):
        prose = redact(raw.strip().split("\n\n")[0][:800])
    if not prose_looks_valid(prose, locale, prior, opening=opening):
        return None
    npc_lines: list[dict[str, str]] = []
    if isinstance(npc_raw, list):
        for line in npc_raw:
            if not isinstance(line, dict):
                continue
            npc_id = str(line.get("npc_id") or "").strip()
            text = redact(str(line.get("text") or "").strip())
            if not npc_id or not text or npc_id == "system_admin":
                continue
            if not opening and not locale_matches(text, locale):
                continue
            npc_lines.append({"npc_id": npc_id, "text": text})
    return {"prose": prose, "npc_lines": npc_lines}


def mechanics_input(req: dict[str, Any]) -> dict[str, Any]:
    kind = normalize_kind(str(req.get("kind") or "action"))
    if kind in ("say", "story"):
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


class ReusableHTTPServer(ThreadingHTTPServer):
    allow_reuse_address = True


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

    def generate(self, prompt: str, *, max_new_tokens: int = 240, sample: bool = True) -> str:
        global _pipe
        if _pipe is None:
            return ""
        kwargs: dict[str, Any] = {
            "max_new_tokens": max_new_tokens,
            "repetition_penalty": 1.18,
            "return_full_text": False,
        }
        if sample:
            kwargs.update(do_sample=True, temperature=0.82, top_p=0.92)
        else:
            kwargs["do_sample"] = False
        out = _pipe(prompt, **kwargs)
        text = out[0]["generated_text"] if out else ""
        return redact(str(text).strip())

    def narrate(self, req: dict) -> dict:
        locale, payload = storyteller_input(req)
        kind = payload["kind"]
        say = kind == "say" or normalize_kind(str(req.get("kind") or "")) == "say"
        if say:
            return fallback_storyteller(locale, payload, say=True)
        prior = prior_list(req)
        opening = kind == "story"

        global _pipe
        tokenizer = getattr(_pipe, "tokenizer", None) if _pipe is not None else None
        parsed = None
        if tokenizer is not None:
            system = storyteller_system(locale, opening=opening, prior=prior)
            if opening:
                user = (
                    f"locale={locale}\nroom={payload['room']}\n"
                    f"theme={payload.get('theme') or ''}\n"
                    f"present={', '.join(payload.get('presence') or [])}\n"
                    f"opening={payload['notes']}\nBegin."
                )
            else:
                user = json.dumps(
                    {k: payload[k] for k in ("actor", "room", "kind", "dice", "rolls", "total", "success", "notes", "presence")},
                    ensure_ascii=False,
                    separators=(",", ":"),
                )
            prompt = chat_prompt(tokenizer, system, user)
            raw = self.generate(prompt, max_new_tokens=280 if opening else 240)
            parsed = parse_storyteller_response(raw, locale, prior, opening=opening)

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
    httpd = ReusableHTTPServer((args.host, args.port), Handler)
    print(f"llm-runner {args.role} hub={args.hf_model} on {args.host}:{args.port}", flush=True)
    httpd.serve_forever()


if __name__ == "__main__":
    main()
