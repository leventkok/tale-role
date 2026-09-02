#!/usr/bin/env python3
"""Local Storyteller / mechanics runner. No paid APIs. Weights stay on disk."""

from __future__ import annotations

import argparse
import json
import os
import re
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import urlparse

EMAIL = re.compile(r"[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}", re.I)


def redact(text: str) -> str:
    return EMAIL.sub("[redacted]", text or "")


def has_weights(root: Path) -> bool:
    if not root.is_dir():
        return False
    for p in root.rglob("*"):
        if p.suffix.lower() in {".safetensors", ".gguf"} or p.name == "adapter_config.json":
            return True
    return False


class Handler(BaseHTTPRequestHandler):
    role = "storyteller"
    adapter_dir = Path(".")
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

    def do_GET(self) -> None:  # noqa: N802
        if urlparse(self.path).path == "/health/live":
            self._send(200, {"status": "alive", "role": self.role, "weights": has_weights(self.adapter_dir)})
            return
        self._send(404, {"error": "not found"})

    def do_POST(self) -> None:  # noqa: N802
        length = int(self.headers.get("Content-Length") or 0)
        body = json.loads(self.rfile.read(length) or b"{}")
        path = urlparse(self.path).path
        if path == "/v1/narrate":
            self._send(200, self.narrate(body))
            return
        if path == "/v1/intent":
            self._send(200, self.intent(body))
            return
        self._send(404, {"error": "not found"})

    def narrate(self, req: dict) -> dict:
        if not has_weights(self.adapter_dir) and not self.allow_unloaded:
            return {"locale": req.get("locale") or "en", "prose": "", "error": "weights missing"}
        notes = redact(str(req.get("notes") or ""))
        actor = req.get("actor_name") or "Someone"
        total = req.get("total") or 0
        dice = req.get("dice_system") or "d20"
        locale = "tr" if req.get("locale") == "tr" else "en"
        if locale == "tr":
            prose = f"[local] {actor} hareket eder. Motorun zarı {dice} toplam {total}. {notes}"
        else:
            prose = f"[local] {actor} acts. The engine's dice {dice} total {total}. {notes}"
        return {"locale": locale, "prose": redact(prose), "npc_lines": []}

    def intent(self, req: dict) -> dict:
        if not has_weights(self.adapter_dir) and not self.allow_unloaded:
            return {"kind": req.get("kind") or "action", "error": "weights missing"}
        kind = req.get("kind") or "action"
        skill = req.get("skill") or ("str" if kind == "action" else "")
        out = {"kind": kind, "notes": redact(str(req.get("notes") or ""))}
        if skill:
            out["skill"] = skill
        if kind == "action":
            out["dc"] = 12
        return out


def main() -> None:
    p = argparse.ArgumentParser(description="Tale Role local runner")
    p.add_argument("--role", choices=("storyteller", "mechanics"), required=True)
    p.add_argument("--host", default="127.0.0.1")
    p.add_argument("--port", type=int, default=8091)
    p.add_argument("--adapter-dir", default=os.environ.get("TALEROLE_ADAPTER_DIR", ""))
    p.add_argument("--allow-unloaded", action="store_true", help="serve bound templates without GPU weights")
    args = p.parse_args()
    Handler.role = args.role
    Handler.adapter_dir = Path(args.adapter_dir or ".")
    Handler.allow_unloaded = args.allow_unloaded
    httpd = ThreadingHTTPServer((args.host, args.port), Handler)
    print(f"llm-runner {args.role} on {args.host}:{args.port} dir={Handler.adapter_dir}", flush=True)
    httpd.serve_forever()


if __name__ == "__main__":
    main()
