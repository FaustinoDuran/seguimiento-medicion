#!/usr/bin/env python3
"""Recalifica una corrida ya capturada sin volver a llamar al modelo.

Cambiar el verificador no debe obligar a gastar 90 conversaciones de nuevo: la
evidencia cruda ya está guardada. Este script reaplica solo los controles
deterministas y deja constancia de que los dictámenes semánticos no se tocan.
"""
import json
from pathlib import Path
import sys


def load_runner():
    import importlib.util

    spec = importlib.util.spec_from_file_location(
        "eval_runner", Path(__file__).resolve().parent / "mcp-eval-run.py",
    )
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def main():
    if len(sys.argv) < 2:
        sys.exit("uso: mcp-eval-recheck.py <corrida.json> [casos.json]")

    runner = load_runner()
    run_path = Path(sys.argv[1])
    cases_path = Path(sys.argv[2]) if len(sys.argv) > 2 else runner.CASES

    run = json.loads(run_path.read_text())
    frozen = json.loads(cases_path.read_text())
    by_id = {case["id"]: case for case in frozen[run["set"]]}

    changed = []
    for case_id, record in run["results"].items():
        case = by_id.get(case_id)
        if not case:
            continue
        for index, repetition in enumerate(record["repetitions"], 1):
            before = list(repetition.get("flags", []))
            after = runner.summarize(repetition, case)
            if before != after:
                changed.append((case_id, index, before, after))
            repetition["flags"] = after

    run["recheck"] = {
        "note": (
            "Controles deterministas reaplicados sobre la evidencia ya capturada; "
            "no se volvieron a ejecutar conversaciones ni se alteraron respuestas."
        ),
        "changed": len(changed),
    }
    run_path.write_text(json.dumps(run, indent=1, ensure_ascii=False))
    for case_id, index, before, after in changed:
        print(f"{case_id} rep{index}: {before} -> {after}")
    print(f"recalificadas {len(changed)} repeticiones en {run_path}")


if __name__ == "__main__":
    main()
