#!/usr/bin/env python3
"""Gate git commit/push against a few mechanical TP rules."""

from __future__ import annotations

import json
import re
import sys


def payload() -> dict:
    try:
        return json.load(sys.stdin)
    except Exception:
        return {}


def command_text(data: dict) -> str:
    for key in ("command", "commandLine", "cmd"):
        value = data.get(key)
        if isinstance(value, str) and value:
            return value
    return ""


def deny(message: str) -> None:
    json.dump(
        {
            "permission": "deny",
            "agent_message": message,
            "user_message": message,
        },
        sys.stdout,
    )
    sys.exit(0)


def allow() -> None:
    json.dump({"permission": "allow"}, sys.stdout)
    sys.exit(0)


def main() -> None:
    cmd = command_text(payload())
    if not re.search(r"\bgit\s+(commit|push)\b", cmd):
        allow()

    if re.search(r"--no-verify|--no-gpg-sign", cmd):
        deny(
            "No saltear hooks del TP (--no-verify / --no-gpg-sign) salvo que la Issue lo pida."
        )

    if re.search(r"Closes\s+#\s+\d+", cmd, re.I):
        deny("Usá `Closes #N` sin espacio entre # y el número.")

    has_red = re.search(r":\s*red\b", cmd, re.I)
    has_green = re.search(r":\s*green\b", cmd, re.I)
    if has_red and has_green:
        deny("No combines RED y GREEN en el mismo commit. Un commit `test(#N): red`, otro `feat(#N): green`.")

    allow()


if __name__ == "__main__":
    try:
        main()
    except Exception:
        json.dump({"permission": "allow"}, sys.stdout)
        sys.exit(0)
