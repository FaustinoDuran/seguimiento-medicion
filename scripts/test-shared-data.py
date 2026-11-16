#!/usr/bin/env python3
"""Verifica la carga y API/MCP reales en PostgreSQL descartable (Docker + Go)."""
import json
import os
from pathlib import Path
import socket
import subprocess
import tempfile
import time
import urllib.error
import urllib.request
import uuid

root = Path(__file__).resolve().parent.parent
container = "shared-data-test-" + uuid.uuid4().hex[:10]
processes = []


def run(*args, **kwargs):
    return subprocess.run(args, text=True, capture_output=True, check=True, **kwargs)


def sql(statement, check=True):
    return subprocess.run(
        ["docker", "exec", "-i", container, "psql", "-X", "-qAt", "-U", "postgres",
         "-d", "app", "-v", "ON_ERROR_STOP=1"],
        input=statement, text=True, capture_output=True, check=check,
    )


def wait_http(url, process):
    for _ in range(100):
        if process.poll() is not None:
            raise RuntimeError("El servidor terminó antes de estar disponible")
        try:
            with urllib.request.urlopen(url, timeout=1) as response:
                if response.status == 200:
                    return
        except (urllib.error.URLError, TimeoutError):
            time.sleep(0.2)
    raise RuntimeError("El servidor no respondió a tiempo")


def api(path, body=None, method=None):
    request = urllib.request.Request(
        "http://127.0.0.1:8080" + path,
        data=json.dumps(body).encode() if body is not None else None,
        headers={"Content-Type": "application/json"}, method=method,
    )
    with urllib.request.urlopen(request, timeout=10) as response:
        return json.load(response)


try:
    # La API del proyecto escucha en 8080; no consultar una instancia ajena.
    with socket.socket() as check_port:
        check_port.bind(("127.0.0.1", 8080))
    run("docker", "run", "-d", "--name", container, "-e", "POSTGRES_PASSWORD=testpass",
        "-e", "POSTGRES_DB=app", "-p", "127.0.0.1::5432", "postgres:16")
    for _ in range(100):
        if sql("SELECT 1", check=False).returncode == 0:
            break
        time.sleep(0.2)
    else:
        raise RuntimeError("PostgreSQL no inició")
    port = run("docker", "port", container, "5432/tcp").stdout.strip().rsplit(":", 1)[1]
    address = f"postgres://postgres:testpass@127.0.0.1:{port}/app?sslmode=disable"
    with tempfile.TemporaryDirectory() as directory:
        directory = Path(directory)
        for executable in ("api", "mcp"):
            run("go", "build", "-o", str(directory / executable), f"./cmd/{executable}", cwd=root / "backend")
        with (directory / "servers.log").open("w") as log:
            server = subprocess.Popen([str(directory / "api")], cwd=directory,
                                      env={**os.environ, "DATABASE_URL": address}, stdout=log, stderr=log)
            processes.append(server)
            wait_http("http://127.0.0.1:8080/health", server)
            original = api("/projects", {"name": "Proyecto existente", "start_date": "2026-01-01",
                                        "end_date": "2026-12-31", "status": "active"})
            seed = (root / "scripts/seed-project.sql").read_text()
            sql("""CREATE FUNCTION reject_seed() RETURNS trigger LANGUAGE plpgsql AS $$
                BEGIN IF NEW.identifier = 'DEMO-003' THEN RAISE EXCEPTION 'forced seed failure';
                END IF; RETURN NEW; END $$;
                CREATE TRIGGER reject_seed BEFORE INSERT ON backlog_stories
                FOR EACH ROW EXECUTE FUNCTION reject_seed();""")
            assert sql(seed, check=False).returncode != 0, "La carga debe abortar ante un error"
            assert sql("SELECT count(*) FROM projects").stdout.strip() == "1", "Debe revertir el proyecto parcial"
            sql("DROP TRIGGER reject_seed ON backlog_stories; DROP FUNCTION reject_seed();")
            sql(seed)
            project_id = int(sql("SELECT project_id FROM demo_seed.runs WHERE name='ui-mcp-v1'").stdout)
            assert project_id != original["id"], "La carga no debe asumir IDs libres"
            snapshot_query = "SELECT json_build_array(" + ",".join(
                f"(SELECT count(*) FROM {table})" for table in
                ("projects", "project_members", "backlog_stories", "sprints", "sprint_stories", "effort_entries", "defects")
            ) + ")"
            counts = sql(snapshot_query).stdout
            assert json.loads(counts) == [2, 2, 4, 2, 3, 3, 2]
            sql(seed)
            assert sql(snapshot_query).stdout == counts, "Repetir no debe duplicar registros"
            assert api(f"/projects/{original['id']}")["name"] == "Proyecto existente"
            metrics = api(f"/projects/{project_id}/metrics")
            for field, value in {"planned_story_points": 16, "completed_story_points": 13,
                                 "velocity": 13, "actual_hours": 6.5, "deviation": 3,
                                 "completed_story_percentage": 50, "defects_detected": 2,
                                 "defects_resolved": 1}.items():
                assert metrics[field] == value, (field, metrics[field], value)
            sql("\\set reader_password abcdef123456\n" + (root / "deploy/mcp-reader.sql").read_text())
            sql("\\set reader_password abcdef123456\n" + (root / "deploy/mcp-reader.sql").read_text())
            denied = sql("SET ROLE mcp_app_reader; UPDATE projects SET name='forbidden';", check=False)
            assert denied.returncode != 0 and "permission denied" in denied.stderr
            with socket.socket() as sock:
                sock.bind(("127.0.0.1", 0))
                mcp_port = sock.getsockname()[1]
            mcp_address = address.replace("postgres:testpass@", "mcp_app_reader:abcdef123456@")
            mcp = subprocess.Popen([str(directory / "mcp")], cwd=directory,
                                   env={**os.environ, "MCP_DATABASE_URL": mcp_address, "MCP_TRANSPORT": "http",
                                        "MCP_PUBLIC_URL": "https://demo.example", "PORT": str(mcp_port)},
                                   stdout=log, stderr=log)
            processes.append(mcp)
            wait_http(f"http://127.0.0.1:{mcp_port}/healthz", mcp)
            renamed = "Demo UI y MCP - editada desde UI"
            api(f"/projects/{project_id}", {"name": renamed}, "PATCH")
            sql(seed)
            assert api(f"/projects/{project_id}")["name"] == renamed, "La carga debe conservar ediciones"
            checked = run("python3", str(root / "scripts/mcp-http-smoke.py"), f"http://127.0.0.1:{mcp_port}/mcp",
                          env={**os.environ, "MCP_SMOKE_API_URL": "http://127.0.0.1:8080"})
            print(checked.stdout, end="")
            print("Carga transaccional, repetición, conservación, permisos y lectura compartida: OK")
finally:
    for process in reversed(processes):
        process.terminate()
        try:
            process.wait(timeout=10)
        except subprocess.TimeoutExpired:
            process.kill()
            process.wait()
    subprocess.run(["docker", "rm", "-fv", container], capture_output=True)
