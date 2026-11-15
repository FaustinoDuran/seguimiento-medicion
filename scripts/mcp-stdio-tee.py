#!/usr/bin/env python3
"""Envuelve el servidor MCP stdio y registra el JSON-RPC real.

Registra cada mensaje con dirección y hora para conservar la evidencia exacta
de argumentos y resultados, sin depender de cómo el cliente los muestre. No
interpreta ni valida el protocolo: solo reenvía bytes.

Uso (desde la configuración MCP del cliente):
  command = ["python3", "scripts/mcp-stdio-tee.py", "/ruta/al/binario-mcp"]

Variables:
  MCP_TEE_LOG  archivo JSONL donde se apendiza la traza.
"""
import json
import os
from pathlib import Path
import subprocess
import sys
import threading
import time


def log_path():
    configured = os.environ.get("MCP_TEE_LOG")
    if configured:
        return Path(configured)
    default = Path.cwd() / ".mcp-eval" / "transcript" / "jsonrpc.jsonl"
    default.parent.mkdir(parents=True, exist_ok=True)
    return default


def main():
    if len(sys.argv) < 2:
        sys.exit("uso: mcp-stdio-tee.py <comando-mcp> [args...]")

    target = log_path()
    target.parent.mkdir(parents=True, exist_ok=True)
    lock = threading.Lock()

    process = subprocess.Popen(
        sys.argv[1:],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=None,
    )

    def record(direction, payload):
        try:
            parsed = json.loads(payload)
        except json.JSONDecodeError:
            parsed = {"raw": payload[:2000]}
        line = {"ts": round(time.time(), 3), "dir": direction, "message": parsed}
        with lock, open(target, "a", encoding="utf-8") as handle:
            handle.write(json.dumps(line, ensure_ascii=False) + "\n")

    def pump_client_to_server():
        for raw in iter(sys.stdin.buffer.readline, b""):
            record("client->server", raw.decode("utf-8", "replace"))
            try:
                process.stdin.write(raw)
                process.stdin.flush()
            except (BrokenPipeError, ValueError):
                return
        try:
            process.stdin.close()
        except (BrokenPipeError, ValueError):
            pass

    def pump_server_to_client():
        for raw in iter(process.stdout.readline, b""):
            record("server->client", raw.decode("utf-8", "replace"))
            try:
                sys.stdout.buffer.write(raw)
                sys.stdout.buffer.flush()
            except (BrokenPipeError, ValueError):
                return

    upstream = threading.Thread(target=pump_client_to_server, daemon=True)
    downstream = threading.Thread(target=pump_server_to_client, daemon=True)
    upstream.start()
    downstream.start()
    downstream.join()
    code = process.wait()
    sys.exit(code)


if __name__ == "__main__":
    main()
