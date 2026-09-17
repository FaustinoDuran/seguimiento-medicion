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
responde `GET /health`. Todavía no hay historias de usuario ni PostgreSQL.

## Cómo ejecutar la API

Desde `backend/`:

```bash
go run ./cmd/api
```

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
backend/internal/<feature>/   dominio Go (TDD)
frontend/src/features/       UI (Sprint 2)
docs/sdd/                    especificaciones
docs/adr/                    decisiones
features/                    escenarios BDD
```
