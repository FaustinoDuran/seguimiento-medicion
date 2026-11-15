#!/usr/bin/env python3
"""Runner de conversaciones para TASK-022 (#108).

Ejecuta una conversación por contexto nuevo contra el cliente MCP nativo de
OpenCode, sin darle al modelo evaluado respuestas esperadas, fixtures ni código.
Registra pregunta, repetición, llamadas ordenadas, argumentos, resultados y
respuesta final en un JSON versionable.

El agente principal califica; este script solo captura evidencia.

Uso:
  python3 scripts/mcp-eval-run.py --set regressions --reps 3 \
      --model opencode/gpt-6-luna --variant low --out docs/mcp/runs/TASK-022-baseline.json

Requiere que .mcp-eval/ws/opencode.json apunte al MCP evaluado y a la base
congelada correspondiente al conjunto (regressions o reserved).
"""
import argparse
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import time

ROOT = Path(__file__).resolve().parent.parent
CASES = ROOT / "docs" / "mcp" / "cases" / "TASK-022-frozen.json"
TEE_SCRIPT = ROOT / "scripts" / "mcp-stdio-tee.py"
# The evaluated model runs in an isolated directory outside the repository on
# purpose: it must not read AGENTS.md, docs/, fixtures or the server source.
WORKSPACE = Path(os.environ.get("MCP_EVAL_WORKSPACE", "/tmp/mcp-eval/ws"))
TRANSCRIPT = Path(os.environ.get("MCP_EVAL_TRANSCRIPT", "/tmp/mcp-eval/transcript"))

# Both frozen fixtures live in isolated demo databases created for TASK-022.
DATABASES = {
    "regressions": "postgres://mcp_reader:eval-only@localhost:5433/mcp_demo_regress?sslmode=disable",
    "reserved": "postgres://mcp_reader:eval-only@localhost:5433/mcp_demo_shown?sslmode=disable",
}


def load_cases(set_name):
    frozen = json.loads(CASES.read_text())
    return frozen, frozen[set_name]


def prepare_workspace():
    """Each run starts from an empty transcript so no prior answer leaks in."""
    if TRANSCRIPT.exists():
        shutil.rmtree(TRANSCRIPT)
    TRANSCRIPT.mkdir(parents=True, exist_ok=True)


def read_tool_calls():
    """Pair JSON-RPC requests with responses from the stdio tee trace.

    Reading the wire protocol is stronger evidence than parsing client output:
    it keeps the exact arguments the model sent and the exact payload returned.
    """
    trace = TRANSCRIPT / "jsonrpc.jsonl"
    if not trace.exists():
        return []
    pending = {}
    calls = []
    for line in trace.read_text(encoding="utf-8", errors="replace").splitlines():
        try:
            entry = json.loads(line)
        except json.JSONDecodeError:
            continue
        message = entry.get("message") or {}
        method = message.get("method")
        if method == "tools/call" and entry.get("dir") == "client->server":
            pending[message.get("id")] = {
                "tool": (message.get("params") or {}).get("name"),
                "arguments": (message.get("params") or {}).get("arguments") or {},
                "ts": entry.get("ts"),
            }
        elif entry.get("dir") == "server->client" and message.get("id") in pending:
            call = pending.pop(message["id"])
            result = message.get("result") or {}
            call["is_error"] = bool(result.get("isError"))
            call["result"] = result
            call["error"] = message.get("error")
            calls.append(call)
    # Requests without a response still count as attempted calls.
    for call in pending.values():
        call["is_error"] = True
        call["result"] = None
        call["error"] = "no response recorded"
        calls.append(call)
    return calls


def run_conversation(question, model, variant, timeout, instruction, server_config):
    """Fresh context: a new client process per repetition.

    The evaluated model only gets the neutral instruction and its own MCP tools;
    it never receives the fixture, the expected answers or the server source.
    """
    if TRANSCRIPT.exists():
        shutil.rmtree(TRANSCRIPT)
    TRANSCRIPT.mkdir(parents=True, exist_ok=True)

    prompt = f"{instruction}\n\nPregunta: {question}"
    command = [
        "codex", "exec",
        "--skip-git-repo-check",
        "--json",
        "-C", str(WORKSPACE),
        "-m", model,
        "-c", f'model_reasoning_effort="{variant}"',
    ]
    command += server_config
    command.append(prompt)

    started = time.time()
    try:
        completed = subprocess.run(
            command, cwd=str(WORKSPACE), capture_output=True, text=True, timeout=timeout,
        )
        stdout, stderr, code = completed.stdout, completed.stderr, completed.returncode
    except subprocess.TimeoutExpired as expired:
        stdout = expired.stdout or ""
        if isinstance(stdout, bytes):
            stdout = stdout.decode("utf-8", "replace")
        stderr = "timeout"
        code = -1
    elapsed = round(time.time() - started, 2)

    calls = read_tool_calls()

    return {
        "elapsed_s": elapsed,
        "exit_code": code,
        "stderr_tail": stderr[-2000:] if stderr else "",
        "stdout_tail": stdout[-4000:] if stdout else "",
        "answer": extract_answer(stdout),
        "tool_calls": calls,
        "tool_call_count": len(calls),
    }


def extract_answer(stdout):
    """Collect the final agent message from Codex JSONL events."""
    texts = []
    for line in stdout.splitlines():
        line = line.strip()
        if not line.startswith("{"):
            continue
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            continue
        if event.get("type") == "error":
            texts.append("ERROR: " + json.dumps(event.get("error", {}))[:500])
            continue
        item = event.get("item") or {}
        if item.get("type") == "agent_message" and isinstance(item.get("text"), str):
            texts.append(item["text"])
    return "\n".join(texts).strip()


def mcp_server_config(database, binary, trace):
    """Pass the evaluated server as a temporary config override.

    Codex reads these overrides as TOML, so arrays and inline tables are written
    in TOML form. Credentials reach the server through the tee environment only;
    the model never sees a file with the fixture or the expected answers.
    """
    args = "[" + ", ".join(f'"{value}"' for value in (TEE_SCRIPT, binary)) + "]"
    environment = f'{{MCP_DATABASE_URL="{database}", MCP_TEE_LOG="{trace}"}}'
    return [
        "-c", 'mcp_servers.seguimiento-medicion.command="python3"',
        "-c", f"mcp_servers.seguimiento-medicion.args={args}",
        "-c", f"mcp_servers.seguimiento-medicion.env={environment}",
    ]


def tool_failed(call):
    """A rejection can arrive as a result flag or as a JSON-RPC protocol error.

    Argument validation is rejected by the SDK before the handler runs, so it
    surfaces as a -32602 protocol error rather than result.isError. Counting
    only one of the two would under-report correct rejections.
    """
    return bool(call.get("is_error")) or call.get("error") is not None


def summarize(run, case):
    """Record raw evidence plus the deterministic checks that can be automated.

    Semantic verdicts stay with the human or parent agent: this only flags the
    facts a script can prove (tool errors, missing ids, forbidden strings).
    """
    joined = json.dumps(run["tool_calls"], ensure_ascii=False)
    answer = run["answer"]
    flags = []
    check = case.get("check", {})

    if check.get("must_fail") and not any(tool_failed(c) for c in run["tool_calls"]):
        flags.append("expected a tool error and found none")
    if check.get("must_empty") and case.get("expect", "").lower().find("vac") >= 0:
        if '"data":[]' not in joined:
            flags.append("expected an empty list")
    for text in check.get("must_contain_text", []):
        if text not in joined and text not in answer:
            flags.append(f"missing expected text: {text}")
    for text in check.get("must_not_contain_text", []):
        if text in joined:
            flags.append(f"forbidden data leaked through tools: {text}")
    for identifier in check.get("must_contain_ids", []):
        if f'"id":{identifier}' not in joined.replace(" ", ""):
            flags.append(f"missing expected id {identifier}")
    for key, value in check.get("must_contain_value", {}).items():
        pattern = re.compile(r'"%s":\s*(-?\d+(?:\.\d+)?)' % re.escape(key))
        found = pattern.search(joined)
        if not found:
            flags.append(f"missing metric {key}")
        elif float(found.group(1)) != float(value):
            flags.append(f"{key} = {found.group(1)}, expected {value}")
    if run["tool_call_count"] > 12:
        flags.append("exceeded 12 MCP calls")
    if run["elapsed_s"] > 120:
        flags.append("exceeded 120 seconds")
    return flags


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--set", dest="set_name", default="regressions",
                        choices=["regressions", "reserved"])
    parser.add_argument("--reps", type=int, default=3)
    parser.add_argument("--model", default="gpt-6.1-sol")
    parser.add_argument("--variant", default="low", help="Codex reasoning effort")
    parser.add_argument("--binary", default="/tmp/mcp-eval/mcp",
                        help="MCP server build under evaluation")
    parser.add_argument("--only", default="", help="comma-separated case ids")
    parser.add_argument("--timeout", type=float, default=120)
    parser.add_argument("--out", required=True)
    args = parser.parse_args()

    frozen, cases = load_cases(args.set_name)
    selected = set(args.only.split(",")) if args.only else None
    if selected:
        cases = [c for c in cases if c["id"] in selected]

    prepare_workspace()
    git_commit = subprocess.run(
        ["git", "rev-parse", "HEAD"], cwd=str(ROOT), capture_output=True, text=True,
    ).stdout.strip()
    server_config = mcp_server_config(
        DATABASES[args.set_name], args.binary, TRANSCRIPT / "jsonrpc.jsonl",
    )

    record = {
        "task": frozen["task"],
        "issue": frozen["issue"],
        "set": args.set_name,
        "model": args.model,
        "variant": args.variant,
        "client": f"Codex CLI exec con MCP nativo por stdio. {frozen['client']}",
        "instruction": frozen["neutral_instruction"],
        "commit": git_commit,
        "started_utc": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "limitation": (
            "Automated flags only detect provable facts. Semantic verdicts "
            "require review; a run with provider errors is invalid, not a pass."
        ),
        "results": {},
    }

    for case in cases:
        repetitions = []
        for repetition in range(1, args.reps + 1):
            print(f"{case['id']} rep {repetition}/{args.reps}", file=sys.stderr, flush=True)
            run = run_conversation(
                case["question"], args.model, args.variant, args.timeout,
                frozen["neutral_instruction"], server_config,
            )
            run["flags"] = summarize(run, case)
            repetitions.append(run)
        record["results"][case["id"]] = {
            "question": case["question"],
            "expect": case["expect"],
            "repetitions": repetitions,
        }

    record["finished_utc"] = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
    output = Path(args.out)
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(record, indent=1, ensure_ascii=False))
    print(f"wrote {output}", file=sys.stderr)


if __name__ == "__main__":
    main()
