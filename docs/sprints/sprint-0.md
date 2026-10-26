# Sprint 0 — La preparación

- Estado: cerrado
- Fecha de inicio: 2026-09-14
- Fecha de fin: 2026-10-26
- Equipo: Faustino (Agile Enabler), Valentino e Ignacio (Product Builders)
- Objetivo: preparar la forma de trabajo, el repositorio y las decisiones
  técnicas antes de implementar historias de usuario.

## Decisiones tomadas

- Aplicación web.
- Backend y reglas de negocio: Go 1.24 + chi.
- Persistencia: PostgreSQL mediante pgx.
- Frontend de presentación: TypeScript + React + Tailwind + Vite. Arranca en el
  Sprint 2.
- Arquitectura: monolito modular feature-based.
- Fuente de verdad del proceso: GitHub Issues, Projects, Pull Requests y repo.
- Alcance del Sprint 1: ver ADR 0003.

Detalle:

- `docs/adr/0001-stack-tecnologico.md`
- `docs/adr/0002-arquitectura-feature-based.md`
- `docs/adr/0003-alcance-mvp.md`

## Trabajo realizado

- Repositorio en GitHub, tablero Scrum y roles.
- Product Backlog inicial: HU-001 a HU-022 (issues #10–#31).
- Guía SDD → BDD → TDD y plantillas.
- API mínima: `go run ./cmd/api` y `GET /health` (TASK-006).
- PostgreSQL 16 local: Docker Compose, `DATABASE_URL` y pool pgx (TASK-007).
- MVP del Sprint 1 acordado (TASK-008): HU-001, HU-003, HU-005, HU-007,
  HU-008, HU-009 y HU-010.

## Corrección de alcance

Durante el armado inicial se adelantaron una API de health, middlewares,
conexión PostgreSQL, migración, test y scripts de ejecución sin una Issue
técnica que los justificara.

Ese trabajo fue retirado. TASK-006 y TASK-007 lo reintrodujeron con Issue y PR.

## Tablero

- Project: [Seguimiento y medición](https://github.com/users/FaustinoDuran/projects/3)
- Columnas: Backlog → Sprint ready → In progress → In review → Done
- Campo Sprint: Sprint 0–4

## Issues del Sprint 0

Hecho:

- [#2 TASK-001](https://github.com/FaustinoDuran/seguimiento-medicion/issues/2) estructura
- [#3 TASK-002](https://github.com/FaustinoDuran/seguimiento-medicion/issues/3) guía y trazabilidad
- [#4 TASK-003](https://github.com/FaustinoDuran/seguimiento-medicion/issues/4) frontend
- [#6 TASK-005](https://github.com/FaustinoDuran/seguimiento-medicion/issues/6) Product Backlog
- [#7 TASK-006](https://github.com/FaustinoDuran/seguimiento-medicion/issues/7) API Go + chi
- [#8 TASK-007](https://github.com/FaustinoDuran/seguimiento-medicion/issues/8) PostgreSQL
- [#9 TASK-008](https://github.com/FaustinoDuran/seguimiento-medicion/issues/9) ADR del MVP

Pendiente de comentario en GitHub (no bloquea el código del Sprint 1):

- [#5 TASK-004](https://github.com/FaustinoDuran/seguimiento-medicion/issues/5) los tres confirman ADR 0001, 0002 y 0003 en la issue

## Evidencias

- Guía: `docs/guia-de-trabajo.md`
- ADRs: `docs/adr/`
- Retro: `docs/retrospectives/sprint-0.md`
- Plantillas: `.github/`, `docs/sdd/_plantilla.md`, `features/_plantilla.feature`

## Criterio de cierre

Cumplido: tablero, backlog inicial, entorno (API + Postgres) y recorte del
Sprint 1 documentado. No se implementan HUs en este sprint.
