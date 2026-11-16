#!/usr/bin/env python3
"""Comprueba el protocolo y consultas reales de la demo pública, sin secretos."""
import json
import os
import urllib.parse
import sys
import urllib.request

endpoint = sys.argv[1] if len(sys.argv) > 1 else "https://seguimiento-medicion.aedestec.com/mcp"
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
projects_result = call("tools/call", {"name": "list_projects", "arguments": {"query": "Demo UI y MCP"}})
projects_data = projects_result["structuredContent"]["data"]
if not projects_data:
    projects_data = call("tools/call", {"name": "list_projects", "arguments": {}})["structuredContent"]["data"]
if not projects_data:
    raise RuntimeError("No hay proyectos para verificar la conexión compartida")
project_id = projects_data[0]["id"]
for name, arguments in [
    ("list_projects", {}),
    ("get_project", {"project_id": project_id}),
    ("list_stories", {"project_id": project_id, "status": "todo"}),
    ("get_metrics", {"project_id": project_id}),
    ("get_report", {"project_id": project_id}),
]:
    result = call("tools/call", {"name": name, "arguments": arguments})
    if result.get("isError"):
        raise RuntimeError(f"La consulta {name} falló")
    print(f"{name}: OK")
print("HTTPS, negociación MCP, ocho herramientas y consultas: OK")

public = urllib.parse.urlsplit(endpoint)
api_base = os.environ.get("MCP_SMOKE_API_URL", f"{public.scheme}://{public.netloc}/api")

def api_get(path):
    with urllib.request.urlopen(api_base + path, timeout=60) as response:
        return json.load(response)

def tool_data(name, args):
    result = call("tools/call", {"name": name, "arguments": args})
    if result.get("isError"):
        raise RuntimeError(f"{name} falló")
    return result["structuredContent"]["data"]

api_project = api_get(f"/projects/{project_id}")
mcp_project = tool_data("get_project", {"project_id": project_id})
for field in ("id", "name", "status"):
    if api_project[field] != mcp_project[field]:
        raise RuntimeError(f"API y MCP difieren en proyecto: {field}")
api_metrics = api_get(f"/projects/{project_id}/metrics")
if api_metrics != tool_data("get_metrics", {"project_id": project_id}):
    raise RuntimeError("API y MCP devuelven métricas diferentes")
print(f"API/UI y MCP: proyecto {project_id} y métricas coinciden")
