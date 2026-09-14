# Instrucciones para agentes (Cursor)

Repo del TP Integrador: estimación, seguimiento y métricas de proyectos de software.

Antes de codear leé `docs/guia-de-trabajo.md`. Convenciones duras también están en `.cursor/rules/`.

## Stack (no cambiar sin ADR)

- Backend: Go 1.24, chi, pgx v5.7.x, PostgreSQL.
- Frontend: React + TypeScript + Tailwind + Vite. Presentación solamente.
- Núcleo y reglas de negocio: solo Go.

## Qué hacer cuando pidan una funcionalidad

1. Pedir o crear `HU-XXX` (issue).
2. Escribir o completar `docs/sdd/HU-XXX-*.md` **antes** de implementar.
3. Escribir `features/HU-XXX-*.feature`.
4. TDD: commit RED (test que falla), GREEN, REFACTOR. Commits separados. Mencionar `#HU-XXX` o `#<issue>`.
5. Código en `backend/internal/<modulo>/`. Handler sin reglas.
6. Actualizar `docs/traceability.md`.
7. No calcular métricas ni Planning Poker en React.

Para infraestructura o configuración usar una Issue `TASK-XXX`, con criterios
de finalización y PR. No inventar SDD/BDD para una tarea sin comportamiento de
negocio y no adelantarla sin Issue.

## Dónde va cada cosa

- Decisiones de arquitectura/stack/MVP: `docs/adr/`
- Spec de una HU: `docs/sdd/`
- Gherkin: `features/`
- Tests unitarios: al lado del código Go
- Retros: `docs/retrospectives/`
- Asignación de trabajo: GitHub Issues/Projects, no archivos nuevos

## Módulos

`projects`, `backlog`, `sprints`, `planningpoker`, `effort`, `defects`, `metrics`, `reports`.

No crear paquetes `shared` con reglas de negocio. No agregar auth ni GORM. No subir `go` del `go.mod` a 1.25+.

## Idioma

Código y nombres de archivos en inglés. Docs del proceso, issues y Gherkin en español.
