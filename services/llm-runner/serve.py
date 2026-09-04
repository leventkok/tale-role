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
PII_ASK = re.compile(
    r"(e-?mail|e-?posta|telefon numar|phone number|cep telefon|tc kimlik|kredi kart|credit card|social security)",
    re.I,
)
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
    "nöbet dönmez",
    "nöbet dönüyor",
    "rün karanlık",
    "kilit durur",
    "pim kopar",
    "kahkaha bitince",
    "kahkaha kopar",
    "çandan önce",
    "gelene dek bekle",
    "zar ",
    "yol açılır",
    "taş cevap verir",
    "taş susar",
    "motorun sayısı",
    "hamleyi tamamlar:",
    "hamleyi kaçırır:",
    "uzaktan bir ses sahneyi",
    "the way opens",
    "alet kayar",
    "zaman biter",
    "e-posta",
    "email address",
    "phone number",
    "telefon numar",
    "tc kimlik",
    "kredi kart",
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


PROSE_KEY = re.compile(r'"prose"\s*:\s*"', re.I)


def recover_prose(raw: str) -> str:
    parsed = extract_json_object(raw)
    if parsed:
        text = redact(str(parsed.get("prose") or "").strip())
        if text:
            return text
    blob = (raw or "").strip()
    if not blob:
        return ""
    match = PROSE_KEY.search(blob)
    if match:
        chunk = blob[match.end() :]
        out: list[str] = []
        idx = 0
        while idx < len(chunk):
            ch = chunk[idx]
            if ch == "\\":
                if idx + 1 >= len(chunk):
                    break
                nxt = chunk[idx + 1]
                if nxt == "n":
                    out.append(" ")
                elif nxt == "t":
                    out.append(" ")
                elif nxt in '"\\/':
                    out.append(nxt)
                elif nxt == "u" and idx + 5 < len(chunk):
                    try:
                        out.append(chr(int(chunk[idx + 2 : idx + 6], 16)))
                        idx += 6
                        continue
                    except ValueError:
                        out.append(nxt)
                else:
                    out.append(nxt)
                idx += 2
                continue
            if ch == '"':
                break
            out.append(ch)
            idx += 1
        return redact("".join(out).strip())
    if blob.startswith("{") or blob.startswith("["):
        return ""
    return redact(blob.split("\n\n")[0][:800].strip())


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
    p = prose.casefold().strip()
    if len(p) < 40:
        return False
    for old in prior:
        o = (old or "").casefold().strip()
        if len(o) < 40:
            continue
        if o[:80] in p or p[:80] in o:
            return True
        if p in o:
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


def prose_looks_valid(
    prose: str,
    locale: str = "en",
    prior: list[str] | None = None,
    *,
    opening: bool = False,
    table_title: str = "",
    host: str = "",
    success: bool | None = None,
    notes: str = "",
) -> bool:
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
    if is_salad(text):
        return False
    if re.search(r"\bsayı\s+\d", lowered) or re.search(r"\bthe count is\s+\d", lowered):
        return False
    if not opening and echoes_deed(text, notes):
        return False
    if not opening and first_person_notes(text):
        return False
    if PII_ASK.search(text):
        return False
    if table_title_leak(text, table_title, host):
        return False
    if not opening and clause_salad(text):
        return False
    if len(text) < 90:
        return False
    if text.count(".") + text.count("!") + text.count("?") > 12:
        return False
    if too_similar(text, prior or []):
        return False
    if opening:
        return True
    if not engine_outcome_ok(text, success):
        return False
    if actor_moved_against_deed(text, notes):
        return False
    if not locale_matches(text, locale):
        return False
    return True


def engine_outcome_ok(prose: str, success: bool | None) -> bool:
    if success is not False:
        return True
    low = (prose or "").casefold()
    for tell in (
        "without hesitation",
        "recognizing the",
        "the way opens",
        "follows through",
        "tereddüt etmeden",
        "çekinmeden",
    ):
        if tell in low:
            return False
    return True


STAY_PUT_TELLS = (
    "without going",
    "without taking a step",
    "stay where",
    "staying where",
    "without walking",
    "without stepping",
    "without moving",
    "don't go",
    "do not go",
    "remain here",
    "yerinde kal",
    "yerimde kal",
    "yerde kal",
    "olduğum yerde",
    "oldugum yerde",
    "adım atmadan",
    "adım atmıyor",
    "koridora inmeden",
    "koridora inmiyor",
    "inmeden",
)


def stay_put_deed(notes: str) -> bool:
    low = (notes or "").casefold()
    return any(p in low for p in STAY_PUT_TELLS)


def actor_moved_against_deed(prose: str, notes: str) -> bool:
    if not stay_put_deed(notes):
        return False
    low = (prose or "").casefold()
    return any(
        tell in low
        for tell in (
            "steps into",
            "steps toward",
            "he walks",
            "she walks",
            "they walk",
            "walks into",
            "walked down",
            "walks down",
            "enters the corridor",
            "leading him deeper",
            "leading her deeper",
            "adım atar",
            "koridora iner",
        )
    )


def table_title_leak(prose: str, title: str, host: str) -> bool:
    name = (title or "").strip()
    if not name:
        return False
    if len(name) < 6 and not any(ch in name for ch in " \t-"):
        return False
    low = name.casefold()
    if low in (host or "").casefold():
        return False
    return low in (prose or "").casefold()


def clause_salad(prose: str) -> bool:
    parts = [p.strip() for p in re.split(r"[.!?]", prose or "") if p.strip()]
    if len(parts) < 6:
        return False
    lengths = [len(p.split()) for p in parts]
    if max(lengths) >= 10:
        return False
    return sum(lengths) / len(parts) < 4


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


def apply_storyteller_adapter() -> bool:
    return os.environ.get("TALEROLE_STORYTELLER_ADAPTER", "").strip() == "1"


ADAPTER_WEIGHT_NAMES = (
    "adapter_model.safetensors",
    "adapter_model.bin",
    "adapter_model.pt",
    "adapter_model.pth",
)


def hub_has_adapter_weights(files: set[str]) -> bool:
    names = {path.rsplit("/", 1)[-1] for path in files}
    return "adapter_config.json" in names and any(name in names for name in ADAPTER_WEIGHT_NAMES)


def hub_has_full_weights(files: set[str]) -> bool:
    for path in files:
        name = path.rsplit("/", 1)[-1]
        if name in {"model.safetensors", "pytorch_model.bin"}:
            return True
        if name.startswith("model-") and name.endswith(".safetensors"):
            return True
        if name.startswith("pytorch_model-") and name.endswith(".bin"):
            return True
    return False


def patch_torchao_for_peft() -> None:
    """Colab peft 0.15 imports a torchao symbol that some wheels omit."""
    try:
        import torchao.quantization as quant
    except Exception:
        return
    if getattr(quant, "LinearActivationQuantizedTensor", None) is None:
        class LinearActivationQuantizedTensor:  # noqa: N801
            pass

        quant.LinearActivationQuantizedTensor = LinearActivationQuantizedTensor


def tale_locale(ui: Any, opening: str, notes: str) -> str:
    for text in (opening, notes):
        t = (text or "").strip()
        if len(t) < 12:
            continue
        if any(ch in TR_LETTERS for ch in t):
            return "tr"
        if locale_matches(t, "en"):
            return "en"
    return "tr" if ui == "tr" else "en"


def format_cast(raw: Any) -> str:
    if not isinstance(raw, list):
        return ""
    lines: list[str] = []
    for row in raw:
        if not isinstance(row, dict):
            continue
        name = redact(str(row.get("name") or "")).strip()
        if not name or name == "system_admin":
            continue
        if len(lines) >= 8:
            break
        bits = [name]
        for key in ("species", "path"):
            val = redact(str(row.get(key) or "")).strip()
            if val:
                bits.append(val)
        line = ", ".join(bits)
        back = redact(str(row.get("backstory") or "")).strip()
        if back:
            line += f": {back[:300]}"
        lines.append(f"- {line}")
    return "\n".join(lines)


def storyteller_system(
    locale: str,
    *,
    opening: bool,
    prior: list[str],
    success: bool | None = None,
    notes: str = "",
    kind: str = "",
) -> str:
    lang = "Turkish" if locale == "tr" else "English"
    body = (
        f"You are the Storyteller at this table. Write 4 to 6 vivid literary sentences in {lang} only. "
        "Stay inside the world brief. Name the people at the table. Continue from the player's deed. "
        "Never invent dice, HP, turn order, or any number. Never mix languages. "
        "Never paste the player's deed. Name the concrete result in the scene. "
        "Never use a table, lobby, or product title as a place name. Place comes from the opening and the world brief. "
        "Write in third person. Do not write as the player. "
        'Return only JSON {"prose":"...","npc_lines":[]}.'
    )
    if opening:
        body += " Open the tale. If a host opening is given, keep it; you may add at most two sentences after it."
        if prior:
            clipped = " | ".join(p[:120] for p in prior if p)
            if clipped:
                body += f" Do not repeat: {clipped}"
        return body
    if kind == "say":
        body += (
            " The player asked a question. Answer in the world, in character. "
            "Do not change the scene, move anyone, or grant a new deed. No dice."
        )
        if prior:
            clipped = " | ".join(p[:120] for p in prior if p)
            if clipped:
                body += f" Do not repeat: {clipped}"
        return body
    body += (
        " Continue the scene with the people already present. Do not restart the tale. "
        "Do not quote the player's deed as a header. Narrate this deed. "
        "The rules engine already judged the roll. You narrate; you do not overrule it. "
        "Never freeze the table."
    )
    if success is True:
        body += (
            " This attempt succeeded. Narrate the deed happening. "
            "Do not refuse it, delay it, or replace it with a different action."
        )
    elif success is False:
        body += (
            " This attempt failed. Do not grant the goal. "
            "Add a new complication the table has not heard yet and keep the scene playable. "
            "Never write the attempt as a success. Never stall."
        )
    if stay_put_deed(notes):
        body += " The player stays put: they do not walk, step, or enter a new place."
    if prior:
        clipped = " | ".join(p[:120] for p in prior if p)
        if clipped:
            body += f" Do not repeat: {clipped}"
    return body


def storyteller_user(payload: dict[str, Any], *, opening: bool, locale: str) -> str:
    brief = redact(str(payload.get("world_brief") or "")).strip()[:2500]
    cast = format_cast(payload.get("cast"))
    prior = payload.get("prior") or []
    if not isinstance(prior, list):
        prior = []
    happened = "\n".join(redact(str(p)).strip() for p in prior[:3] if str(p).strip())
    parts = [f"language={locale}"]
    if brief:
        parts.append("WORLD\n" + brief)
    if cast:
        parts.append("PEOPLE AT THIS TABLE\n" + cast)
    if opening:
        seed = payload.get("notes") or "(none — invent a full opening scene from the world)"
        parts.append("HOST OPENING\n" + str(seed))
        parts.append("Write the opening now.")
        return "\n\n".join(parts)
    if happened:
        parts.append("WHAT ALREADY HAPPENED\n" + happened)
    facts = payload.get("facts") or []
    if isinstance(facts, list):
        traces = "\n".join(redact(str(p)).strip() for p in facts[:8] if str(p).strip())
        if traces:
            parts.append("TABLE MEMORY (short; do not invent beyond this)\n" + traces)
    actor = payload.get("actor") or "Someone"
    notes = payload.get("notes") or ""
    success = payload.get("success")
    kind = str(payload.get("kind") or "")
    if kind == "say":
        result = (
            "The player asked. Answer in the world. Do not change the scene, "
            "move anyone, or grant a deed. No dice."
        )
    elif kind in ("pass", "wait"):
        result = "They pass. One short beat. Do not change the scene."
    elif success is False:
        result = (
            "RESULT: MISS. The rules engine already ruled this attempt failed. "
            "Do not grant the deed. Fail forward: add a new sound, threat, or change "
            "the player has not seen yet, and keep the scene playable. "
            "Do not repeat the last beat. Do not write hesitation-free competence."
        )
    elif success is True:
        result = (
            "RESULT: HIT. The rules engine already ruled this attempt succeeded. "
            "Narrate the deed happening in third person. Do not refuse, delay, or replace it. "
            "Do not invent dice, HP, turn order, or any number. Do not paste the deed."
        )
    else:
        result = "Continue the scene from the world, the people, and this deed."
    parts.append(
        "THIS BEAT\n"
        f"actor={actor}\n"
        f"deed={notes}\n"
        f"kind={payload.get('kind')}\n"
        f"success={success}\n"
        + result
    )
    return "\n\n".join(parts)


def storyteller_input(req: dict[str, Any]) -> tuple[str, dict[str, Any]]:
    kind = normalize_kind(str(req.get("kind") or "action"))
    actor = str(req.get("actor_name") or "Someone")
    table_title = str(req.get("room_name") or "")
    notes = strip_engine_leak(redact(str(req.get("notes") or req.get("opening") or "")))
    opening_text = redact(str(req.get("opening") or ""))
    locale = tale_locale(req.get("locale"), opening_text, notes)
    dice = str(req.get("dice_system") or "d20")
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
    model_kind = kind

    facts = req.get("facts") or []
    if not isinstance(facts, list):
        facts = []
    facts = [redact(str(row)).strip() for row in facts if str(row).strip()][:8]

    payload = {
        "actor": actor,
        "room": scene_place(locale),
        "table_title": table_title,
        "kind": model_kind,
        "dice": dice,
        "rolls": rolls,
        "total": total,
        "success": success,
        "notes": notes,
        "presence": presence,
        "opening": opening_text,
        "world_brief": redact(str(req.get("world_brief") or ""))[:2500],
        "cast": req.get("cast") if isinstance(req.get("cast"), list) else [],
        "prior": prior_list(req),
        "facts": facts,
    }
    return locale, payload


def fallback_opening(locale: str, payload: dict[str, Any]) -> str:
    notes = str(payload.get("notes") or "").strip()
    if notes:
        return notes
    if locale == "tr":
        return "Fener yanar. Eşikte bir duraklama var. Uzak bir ses gelir. Anlatıcı sözü alır."
    return "A hush. Lanternlight. The storyteller takes the floor."


def scene_place(locale: str) -> str:
    return "salon" if locale == "tr" else "the hall"


def beat_locale(locale: str, notes: str) -> str:
    text = (notes or "").strip()
    if locale != "tr" or not text:
        return "en" if locale != "tr" else "tr"
    if any(ch in TR_LETTERS for ch in text):
        return "tr"
    if locale_matches(text, "en"):
        return "en"
    return "tr"


def first_person_notes(text: str) -> bool:
    low = (text or "").casefold()
    if low.startswith(("i ", "i'm ", "i’m ", "ben ")):
        return True
    return bool(
        re.search(r"(yorum|mekteyim|maktayım|arım|erim|ırım|urum|ürüm)(\s|[.!?:]|$)", low)
    )


def player_deed(notes: str) -> str:
    text = (notes or "").strip()
    low = text.casefold()
    marker = "deed:"
    if marker in low:
        text = text[low.index(marker) + len(marker) :].strip()
        text = text.split("Narrate this")[0].strip().rstrip(".")
    return text.strip()


def voice_deed(actor: str, notes: str) -> str:
    deed = player_deed(notes)
    if not deed or not first_person_notes(deed):
        return ""
    text = re.sub(r"(?i)^ben\s+", "", deed)
    text = re.sub(r"(?i)^i['’]m\s+", "", text)
    text = re.sub(r"(?i)^i\s+", "", text)
    text = re.sub(r"yorum\b", "yor", text)
    text = re.sub(r"mekteyim\b", "mekte", text)
    text = re.sub(r"maktayım\b", "makta", text)
    text = re.sub(r"(arım|erim|ırım|urum|ürüm)\b", lambda m: m.group(0)[:-2], text)
    text = text.strip().rstrip(".")
    if not text:
        return ""
    if text[0].isupper() and (len(text) == 1 or not text[1].isupper()):
        text = text[0].lower() + text[1:]
    return f"{actor} {text}"


def strip_engine_leak(notes: str) -> str:
    text = redact(notes or "").strip()
    if not text:
        return ""
    low = text.casefold()
    packed = "deed:" in low or "player count " in low or "attempts a deed" in low
    if not packed:
        return text
    deed = player_deed(text)
    if deed and deed != text:
        return redact(deed).strip()
    cleaned = re.sub(r"(?i)player count \d+\.\s*", "", text)
    cleaned = re.sub(r"\b(?:HIT|MISS)\.\s*", "", cleaned)
    cleaned = re.sub(r"(?i)^[^.]+ attempts a deed\.\s*", "", cleaned)
    cleaned = re.sub(r"(?i)narrate this outcome\.[^.]*", "", cleaned)
    cleaned = cleaned.strip(" .")
    return redact(cleaned) if cleaned else text


def is_salad(prose: str) -> bool:
    text = (prose or "").strip()
    if not text:
        return False
    low = text.casefold()
    if re.search(r"\bsayı\s+\d", low) or re.search(r"\bthe count is\s+\d", low):
        return True
    tells = (
        "taş susar",
        "taş cevap verir",
        "yol açılır",
        "hamleyi kaçırır:",
        "hamleyi tamamlar:",
        "uzaktan bir ses sahneyi",
        "motorun sayısı",
    )
    return any(tell in low for tell in tells)


def echoes_deed(prose: str, notes: str) -> bool:
    lowered = (prose or "").casefold()
    for chunk in ((notes or "").strip(), player_deed(notes)):
        if len(chunk) >= 16 and chunk.casefold() in lowered:
            return True
    return False


def table_deed(notes: str) -> str:
    text = (notes or "").strip()
    if not text or first_person_notes(text):
        return ""
    return text


def miss_rewrite_user(user: str) -> str:
    return beat_rewrite_user(user, success=False, notes="")


def beat_rewrite_user(user: str, *, success: bool | None, notes: str) -> str:
    extra = "The first draft failed the table. Rewrite. Narrate this deed in third person."
    if success is False:
        extra += (
            " The attempt does not succeed. Add a new complication and keep play moving. "
            "No competence, no recognition, no unhesitating action."
        )
    elif success is True:
        extra += " The deed happens. Do not refuse it or substitute another action."
    if stay_put_deed(notes):
        extra += " The actor does not walk, step, or enter a new place."
    return user + "\n\n" + extra


HIT_TR = (
    "Fener hareketin üstüne düşer; görünen şey sahneyi değiştirir.",
    "Oda boyun eğer. Bir sonraki seçenek açık kalır.",
    "Parmakların altındaki dünya cevap verir; masa kapanmaz.",
)
MISS_TR = (
    "Koridordaki metal bir karış daha yaklaşır; kaynak hâlâ görünmez.",
    "Oymaların uğultusu düşer, sonra daha alçak bir tondan döner.",
    "Karanlıkta bir nefes tutulur. Cevap gelmez; sahne kapanmaz.",
    "Madalyon soğur. Kazınmış yazı yerinde kalır.",
    "Tavandaki mavi ışık bir an kesilir, çatlak yine nefes alır.",
    "Taşın altından kısa bir tık gelir. Sıra yine oyuncuda.",
)
HIT_EN = (
    "Light finds the motion; the room shifts to match what is now seen.",
    "The world yields. Another choice stays open.",
    "The thing under the hand answers; the table does not close.",
)
MISS_EN = (
    "Down the corridor, metal drags one pace closer.",
    "The carvings drop a note, then return lower.",
    "A breath holds in the dark. No answer yet.",
    "The medallion goes cold. The scratched word stays.",
    "Pale ceiling light dies for a beat, then returns.",
    "Stone ticks once underfoot. The next move is still yours.",
)


def unused_line(lines: tuple[str, ...], prior: list[str] | None, total: int) -> str:
    blob = " ".join(prior or []).casefold()
    unused = [ln for ln in lines if ln.casefold()[:28] not in blob]
    pool = unused or list(lines)
    return pool[total % len(pool)]


def follow_through(locale: str, notes: str, success: bool, prior: list[str] | None, total: int) -> str:
    low = (notes or "").casefold()
    bag = any(k in low for k in ("torba", "çanta", "kese", "pouch", "satchel", "bag"))
    if locale == "tr":
        if success:
            if bag:
                return unused_line(
                    (
                        "Torbanın ağzı açılır. Fener kumaşa, toza ve soğuk bir kenara düşer; adları henüz yok, ama artık görünürler.",
                        "Ağız gevşer. İçeride kayış, katlanmış kumaş, parmak ucunda metal. Hiçbiri masayı kapatmaz.",
                    ),
                    prior,
                    total,
                )
            return unused_line(HIT_TR, prior, total)
        if bag:
            return unused_line(
                (
                    "Torba kayar; ağız kapanır. İçeride bir şey kumaşa takılır ve karanlıkta kalır.",
                    "Parmaklar kumaşı bulur ama ağız bir an önce kapanır. Torba sırrını tutar.",
                ),
                prior,
                total,
            )
        return unused_line(MISS_TR, prior, total)
    if success:
        if bag:
            return unused_line(
                (
                    "The bag's mouth opens. Cloth, dust, and a cold edge take the light; unnamed yet, but seen.",
                    "The drawstring yields. A strap, folded cloth, metal under a fingertip. None of it closes the table.",
                ),
                prior,
                total,
            )
        return unused_line(HIT_EN, prior, total)
    if bag:
        return unused_line(
            (
                "The bag slips; the mouth pinches shut. Something snags in the cloth and stays dark.",
                "Fingers find fabric, then the mouth closes first. The bag keeps its secret.",
            ),
            prior,
            total,
        )
    return unused_line(MISS_EN, prior, total)


def literary_action(
    locale: str,
    kind: str,
    actor: str,
    room: str,
    notes: str,
    success: bool | None,
    total: int,
    prior: list[str] | None = None,
) -> str:
    loc = beat_locale(locale, notes)
    _ = room
    raw_notes = (notes or "").strip()
    if kind == "say":
        if loc == "tr":
            return f"{actor} sözü salona bırakır: {raw_notes} Fener sönmez." if raw_notes else f"{actor} sessizliği kırar. Fener sönmez."
        return f'{actor} speaks. "{raw_notes}" The lantern holds.' if raw_notes else f"{actor} breaks the hush. The lantern holds."
    if kind == "pass":
        if loc == "tr":
            return f"{actor} bu eli bırakır. Salon bekler. Fener sönmez."
        return f"{actor} lets the beat pass. The lantern holds."
    if kind == "wait":
        if loc == "tr":
            return f"{actor} nefesini tutar. Henüz hamle yok. Fener sönmez."
        return f"{actor} holds still. Breath only. The lantern holds."
    shift = follow_through(loc, raw_notes, success is True, prior, total)
    held = " Gitmediği yer açık kalmaz." if loc == "tr" else " The place they refused stays unentered."
    if stay_put_deed(raw_notes):
        if loc == "tr":
            lead = f"{actor} yerinde kalır" if success is True else f"{actor} hamleyi kaçırır"
        else:
            lead = f"{actor} holds the beat" if success is True else f"{actor} misses"
        return f"{lead}. {shift}{held}"
    voiced = voice_deed(actor, raw_notes)
    if success is True:
        if voiced:
            return f"{voiced}. {shift}"
        return f"{actor}. {shift}"
    if loc == "tr":
        if voiced:
            return f"{voiced}. {actor} hamleyi kaçırır. {shift}"
        return f"{actor} hamleyi kaçırır. {shift}"
    if voiced:
        return f"{voiced}. {actor} misses. {shift}"
    return f"{actor} misses. {shift}"


def fallback_storyteller(locale: str, payload: dict[str, Any], *, say: bool) -> dict[str, Any]:
    actor = payload["actor"]
    room = payload["room"]
    notes = payload["notes"]
    kind = payload["kind"]
    total = int(payload.get("total") or 0)
    success = payload.get("success")
    prior = payload.get("prior") if isinstance(payload.get("prior"), list) else []

    if kind == "story":
        return {"locale": locale, "prose": redact(fallback_opening(locale, payload)), "npc_lines": []}

    use_kind = "say" if say else kind
    prose = literary_action(
        locale,
        use_kind,
        actor,
        room,
        notes,
        success if use_kind == "action" else None,
        total,
        prior,
    )
    return {"locale": locale, "prose": redact(prose), "npc_lines": []}


def parse_storyteller_response(
    raw: str,
    locale: str = "en",
    prior: list[str] | None = None,
    *,
    opening: bool = False,
    table_title: str = "",
    host: str = "",
    success: bool | None = None,
    notes: str = "",
) -> dict[str, Any] | None:
    parsed = extract_json_object(raw)
    npc_raw: Any = []
    if parsed:
        npc_raw = parsed.get("npc_lines")
    prose = recover_prose(raw)
    if not prose_looks_valid(
        prose,
        locale,
        prior,
        opening=opening,
        table_title=table_title,
        host=host,
        success=success,
        notes=notes,
    ):
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


def load_pipeline(model_id: str, token: str | None, *, apply_adapter: bool = True):
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
    if hub_has_adapter_weights(files):
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
        if apply_adapter:
            try:
                patch_torchao_for_peft()
                # Pass config so PEFT skips _get_peft_type (0.15.2 raises
                # "Can't find adapter_config.json" after a hub download hiccup).
                model = PeftModel.from_pretrained(base, model_id, config=cfg, token=token)
                print(f"lora attached from {model_id}", flush=True)
            except Exception as exc:
                print(f"lora attach failed ({exc}); using base instruct", flush=True)
                model = base
        else:
            model = base
        return pipeline("text-generation", model=model, tokenizer=tok, return_full_text=False)

    if "adapter_config.json" in {path.rsplit("/", 1)[-1] for path in files}:
        if hub_has_full_weights(files):
            print(f"{model_id}: merged checkpoint; ignoring leftover adapter_config.json", flush=True)
        else:
            from peft import PeftConfig

            cfg = PeftConfig.from_pretrained(model_id, token=token)
            tok_id = cfg.base_model_name_or_path or model_id
            print(f"{model_id}: no LoRA weights; loading base {tok_id}", flush=True)

    tok = AutoTokenizer.from_pretrained(tok_id, token=token, trust_remote_code=True)
    if tok.pad_token is None:
        tok.pad_token = tok.eos_token
    model = AutoModelForCausalLM.from_pretrained(
        tok_id,
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
        prior = prior_list(req)
        opening = kind == "story"

        global _pipe
        tokenizer = getattr(_pipe, "tokenizer", None) if _pipe is not None else None
        parsed = None
        if tokenizer is not None:
            system = storyteller_system(
                locale,
                opening=opening,
                prior=prior,
                success=payload.get("success"),
                notes=str(payload.get("notes") or ""),
                kind=kind,
            )
            user = storyteller_user(payload, opening=opening, locale=locale)
            prompt = chat_prompt(tokenizer, system, user)
            raw = self.generate(prompt, max_new_tokens=400 if opening else 320)
            host = " ".join(
                str(part) for part in (req.get("opening"), payload.get("notes")) if part
            )
            parsed = parse_storyteller_response(
                raw,
                locale,
                prior,
                opening=opening,
                table_title=str(payload.get("table_title") or ""),
                host=host,
                success=payload.get("success"),
                notes=str(payload.get("notes") or ""),
            )
            if parsed is None and payload.get("kind") == "action" and not opening:
                raw = self.generate(
                    chat_prompt(
                        tokenizer,
                        system,
                        beat_rewrite_user(
                            user,
                            success=payload.get("success"),
                            notes=str(payload.get("notes") or ""),
                        ),
                    ),
                    max_new_tokens=320,
                )
                parsed = parse_storyteller_response(
                    raw,
                    locale,
                    prior,
                    opening=opening,
                    table_title=str(payload.get("table_title") or ""),
                    host=host,
                    success=payload.get("success"),
                    notes=str(payload.get("notes") or ""),
                )

        if parsed and is_salad(str(parsed.get("prose") or "")):
            print("rejected salad", flush=True)
            parsed = None
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
            _pipe = load_pipeline(args.hf_model, token, apply_adapter=args.role != "storyteller" or apply_storyteller_adapter())
        if args.role == "storyteller" and not apply_storyteller_adapter():
            print("storyteller: base instruct (LoRA skipped)", flush=True)
    httpd = ReusableHTTPServer((args.host, args.port), Handler)
    print(f"llm-runner {args.role} hub={args.hf_model} on {args.host}:{args.port}", flush=True)
    httpd.serve_forever()


if __name__ == "__main__":
    main()
