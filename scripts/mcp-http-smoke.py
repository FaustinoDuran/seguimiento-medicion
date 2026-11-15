#!/usr/bin/env python3
"""Comprueba el protocolo y consultas reales de la demo pública, sin secretos."""
import json
import sys
import urllib.request

endpoint = sys.argv[1] if len(sys.argv) > 1 else "https://seguimiento0-medicion.aedestec.com/mcp"
version = "2025-06-18"
counter = 0


def call(method, params):
    global counter
    counter += 1
    request = urllib.request.Request(
        endpoint,
        data=json.dumps({"jsonrpc": "2.0", "id": counter, "method": method, "params": params}).encode(),
        headers={"Content-Type": "application/json", "Accept": "application/json, text/event-stream",
                 "MCP-Protocol-Version": version},
    )
    with urllib.request.urlopen(request, timeout=60) as response:
        result = json.load(response)
    if "error" in result:
        raise RuntimeError(f"{method}: {result['error']}")
    return result["result"]


initialized = call("initialize", {"protocolVersion": version, "capabilities": {},
                                  "clientInfo": {"name": "deployment-smoke", "version": "1"}})
version = initialized["protocolVersion"]
# El servidor stateless no conserva sesión ni requiere notificaciones de estado.
names = {tool["name"] for tool in call("tools/list", {})["tools"]}
expected = {"list_projects", "get_project", "list_stories", "get_story",
            "list_sprints", "get_sprint", "get_metrics", "get_report"}
if names != expected:
    raise RuntimeError(f"Herramientas inesperadas: {sorted(names)}")
for name, arguments in [
    ("list_projects", {}),
    ("get_project", {"project_id": 1}),
    ("list_stories", {"project_id": 1, "status": "todo"}),
    ("get_metrics", {"project_id": 1}),
    ("get_report", {"project_id": 1}),
]:
    result = call("tools/call", {"name": name, "arguments": arguments})
    if result.get("isError"):
        raise RuntimeError(f"La consulta {name} falló")
    print(f"{name}: OK")
print("HTTPS, negociación MCP, ocho herramientas y consultas: OK")
