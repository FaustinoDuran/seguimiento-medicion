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

Sprint 1 en curso. Con `DATABASE_URL` la API expone proyectos (HU-001),
integrantes (HU-003), historias de backlog (HU-005), sprints (HU-007),
asignación de historias (HU-008) y finalización de historias (HU-009).

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

Con Postgres levantado, crear y consultar un proyecto:

```bash
curl -s -X POST http://localhost:8080/projects \
  -H 'Content-Type: application/json' \
  -d '{"name":"TP Integrador","start_date":"2026-10-26","end_date":"2026-11-23"}'

curl -s http://localhost:8080/projects/1

curl -s -X POST http://localhost:8080/projects/1/members \
  -H 'Content-Type: application/json' \
  -d '{"name":"Valentino"}'

curl -s http://localhost:8080/projects/1/members

curl -s -X POST http://localhost:8080/projects/1/sprints \
  -H 'Content-Type: application/json' \
  -d '{"goal":"Entregar el alta de proyectos"}'

curl -s http://localhost:8080/sprints/1
```

Crear y consultar una historia (reemplazar los IDs por los devueltos por la API):

```bash
curl -s -X POST http://localhost:8080/projects/1/stories \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"HU-005","title":"Crear historia","priority":"high","story_points":5,"acceptance_criteria":["La historia se puede consultar"]}'

curl -s http://localhost:8080/projects/1/stories/1
```

El identificador se normaliza a mayúsculas y es único por proyecto. Toda historia
nueva comienza en `todo`. Se puede omitir prioridad (`unassigned`), story points
(`null`), descripción (`""`) y criterios de aceptación (`[]`). La escala de
prioridad en Go es `unassigned = 0`, `low = 1`, `medium = 2`, `high = 3`.

Asignar una historia a un Sprint abierto, completarla y consultar el avance
(reemplazar los IDs por los devueltos por la API):

```bash
curl -s -X POST http://localhost:8080/sprints/1/stories \
  -H 'Content-Type: application/json' \
  -d '{"story_id":1}'

curl -s -X POST http://localhost:8080/sprints/1/stories/1/complete

curl -s http://localhost:8080/sprints/1/stories
curl -s http://localhost:8080/projects/1/stories/1
```

Completar devuelve `200` con estado `done`; ese mismo estado aparece en ambas
consultas. Repetir la finalización devuelve éxito mientras el Sprint siga
abierto. No se exige estimación ni criterios de aceptación registrados.
Sin asignación al Sprint indicado o con Sprint cerrado, devuelve `409`;
si no existe el Sprint o la historia en su proyecto, devuelve `404`.
La reapertura corresponde a HU-006.

La API devuelve `400` para datos inválidos, `404` para recursos inexistentes y
`409` para identificadores duplicados o proyectos cerrados al crear.

## Verificación del backend

Desde `backend/`, ejecutar `go test ./...`. Para incluir persistencia PostgreSQL,
HTTP con base real y unicidad concurrente:

```bash
DATABASE_URL='postgres://postgres:postgres@localhost:5433/seguimiento_medicion?sslmode=disable' \
BACKLOG_TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5433/seguimiento_medicion?sslmode=disable' \
SPRINTS_TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5433/seguimiento_medicion?sslmode=disable' \
  go test -race ./...
```

Los tests de backlog y de sprints usan esquemas temporales aislados que se
eliminan al terminar. Sin la variable de cada paquete, sus tests de PostgreSQL
se omiten explícitamente.

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
backend/migrations/          SQL de dominio (projects, backlog y sprints)
backend/docker-compose.yml   PostgreSQL 16
backend/internal/<feature>/   dominio Go (TDD)
frontend/src/features/       UI (Sprint 2)
docs/sdd/                    especificaciones
docs/adr/                    decisiones
features/                    escenarios BDD
```
