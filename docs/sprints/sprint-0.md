# Sprint 0 — La preparación

- Estado: en curso
- Fecha de inicio: 2026-09-14
- Equipo: Faustino (Agile Enabler), Valentino e Ignacio (Product Builders)
- Objetivo: preparar la forma de trabajo, el repositorio y las decisiones
  técnicas antes de implementar historias de usuario.

## Decisiones tomadas

- Aplicación web.
- Backend y reglas de negocio: Go 1.24 + chi.
- Persistencia: PostgreSQL mediante pgx.
- Frontend de presentación: TypeScript + React + Tailwind + Vite.
- Arquitectura: monolito modular feature-based.
- Fuente de verdad del proceso: GitHub Issues, Projects, Pull Requests y repo.

Detalle:

- `docs/adr/0001-stack-tecnologico.md`
- `docs/adr/0002-arquitectura-feature-based.md`

## Trabajo realizado

- Repositorio creado en GitHub y clonado localmente.
- Estructura mínima de carpetas para backend y frontend.
- Módulos previstos del backend:
  `projects`, `backlog`, `sprints`, `planningpoker`, `effort`, `defects`,
  `metrics` y `reports`.
- Frontend inicializado con Vite, React, TypeScript y Tailwind.
- Plantillas iniciales de SDD, BDD, Issue y Pull Request.
- Guía del flujo Historia → SDD → aceptación → BDD → tests → código Go.
- Reglas de Cursor y `AGENTS.md` para que los tres agentes sigan el mismo proceso.

## Corrección de alcance

Durante el armado inicial se adelantaron una API de health, middlewares,
conexión PostgreSQL, migración, test y scripts de ejecución sin una Issue
técnica que los justificara.

Ese trabajo fue retirado. El repositorio conserva estructura y documentación,
pero no implementa todavía ninguna funcionalidad ni infraestructura ejecutable.

## Tablero

- Project: [Seguimiento y medición](https://github.com/users/FaustinoDuran/projects/3)
- Columnas: Backlog → Sprint ready → In progress → In review → Done
- Campo Sprint: Sprint 0–4

## Issues del Sprint 0

Trabajo ya hecho, pendiente de commit a `main`:

- [#2 TASK-001](https://github.com/FaustinoDuran/seguimiento-medicion/issues/2) estructura del repo
- [#3 TASK-002](https://github.com/FaustinoDuran/seguimiento-medicion/issues/3) guía y trazabilidad
- [#4 TASK-003](https://github.com/FaustinoDuran/seguimiento-medicion/issues/4) scaffold frontend

Pendiente:

- [#5 TASK-004](https://github.com/FaustinoDuran/seguimiento-medicion/issues/5) aprobar ADRs en equipo
- [#6 TASK-005](https://github.com/FaustinoDuran/seguimiento-medicion/issues/6) Product Backlog inicial (HUs)
- [#7 TASK-006](https://github.com/FaustinoDuran/seguimiento-medicion/issues/7) bootstrap API Go + chi
- [#8 TASK-007](https://github.com/FaustinoDuran/seguimiento-medicion/issues/8) PostgreSQL
- [#9 TASK-008](https://github.com/FaustinoDuran/seguimiento-medicion/issues/9) ADR del MVP

## Evidencias

- Guía: `docs/guia-de-trabajo.md`
- ADRs: `docs/adr/`
- Reglas para agentes: `AGENTS.md` y `.cursor/rules/`
- Plantillas: `.github/`, `docs/sdd/_plantilla.md`,
  `features/_plantilla.feature`

## Criterio de cierre

El Sprint 0 se cerrará cuando exista el tablero, el backlog inicial, los roles y
decisiones estén confirmados por el equipo, y las primeras HUs cumplan la
Definition of Ready. No se implementarán HUs antes de ese cierre.
