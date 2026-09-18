# Seguimiento y medición

Aplicación de estimación, seguimiento y calidad de proyectos de software.

Trabajo práctico integrador — Ingeniería y Calidad de Software, UTN FRSR, 2026.

## Stack

| Capa | Tecnología |
|---|---|
| Backend (núcleo y reglas) | Go 1.24 + chi + pgx |
| Base de datos | PostgreSQL 16 |
| Frontend (presentación, Sprint 2) | TypeScript + React + Tailwind + Vite |

Las reglas de negocio **no** se implementan en React.

## Estado actual

Sprint 0 en preparación. El bootstrap HTTP de TASK-006 está disponible: la API
responde `GET /health`. TASK-007 deja PostgreSQL 16 y el pool pgx listos; las
tablas de dominio llegan con la primera HU de persistencia.

## Cómo levantar PostgreSQL

Desde `backend/`:

```bash
cp .env.example .env
docker compose up -d
```

La base escucha en `localhost:5433` (no usa 5432 para no chocar con otros
Postgres locales). Variables: `backend/.env.example`.

## Cómo ejecutar la API

Desde `backend/`:

```bash
go run ./cmd/api
```

Si `DATABASE_URL` está definida (en el entorno o en `backend/.env`), la API
abre un pool pgx y hace ping al arrancar. Si falta, arranca igual y solo
expone HTTP.

Luego consultar el estado:

```bash
curl -i http://localhost:8080/health
```

La respuesta esperada es `200 OK` con `{"status":"ok"}`.

## Documentación del equipo

- [Guía de trabajo y trazabilidad](docs/guia-de-trabajo.md) — cómo trabaja el equipo y el agente de Cursor
- [Registro del Sprint 0](docs/sprints/sprint-0.md)
- [Trazabilidad](docs/traceability.md)
- [ADRs](docs/adr/)
- [Plantilla SDD](docs/sdd/_plantilla.md)

## Estructura

```text
backend/cmd/api               servidor HTTP (chi)
backend/internal/httpx/      router y GET /health
backend/internal/config/     DATABASE_URL y .env
backend/internal/platform/postgres/  pool pgx
backend/migrations/          SQL de dominio (vacío hasta la primera HU)
backend/docker-compose.yml   PostgreSQL 16
backend/internal/<feature>/   dominio Go (TDD)
frontend/src/features/       UI (Sprint 2)
docs/sdd/                    especificaciones
docs/adr/                    decisiones
features/                    escenarios BDD
```
