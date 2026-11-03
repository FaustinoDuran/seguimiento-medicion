# HU-028 — Listar historias de backlog de un proyecto

- Issue: [#63](https://github.com/FaustinoDuran/seguimiento-medicion/issues/63)
- Módulo: `backend/internal/backlog/`
- Estado: implementada; pendiente revisión y merge

## Objetivo

Como integrante, consultar todas las historias de un proyecto para ver su
trabajo disponible y realizado. Desbloquea HU-026 sin mezclar backend y UI.

## Entradas

`GET /projects/{project_id}/stories`, con identificador interno de proyecto
entero positivo representable en int64.

## Salidas esperadas

`200` con `{ "stories": [...] }`. Cada elemento conserva todos los campos de
HU-005, incluidos estado, story points nulos y criterios de aceptación.

## Reglas de negocio

- AC-028-01 — Devolver exclusivamente las historias del proyecto, en orden
  `id DESC`, incluidas las completadas y las asignadas a Sprint. No filtrar por
  estado ni asignación, ni modificar las historias durante la consulta.
- AC-028-02 — Para un proyecto existente sin historias, devolver `200` y
  `{ "stories": [] }`, nunca `null`.
- AC-028-03 — Permitir consultar proyectos planificados, activos y cerrados.
- AC-028-04 — Proyecto inexistente: `404` JSON. ID inválido, cero, negativo o
  fuera de int64: `400` JSON.
- AC-028-05 — Propagar fallos de consulta/persistencia sin convertirlos en un
  listado vacío; HTTP devuelve `500` con error genérico sin detalles internos.

## Restricciones

Go 1.24, chi, pgx y PostgreSQL existentes. Handler fino; servicio comprueba
existencia; repositorio filtra por proyecto y ordena. Sin paginación, fechas
nuevas ni cambios de esquema. El ID ordena altas, no representa una fecha ni
garantiza el orden de confirmación de transacciones concurrentes.

## Casos límite y errores

Proyecto vacío, cerrado, historias de otro proyecto, completadas o asignadas,
fallos de lectura de proyecto y de repositorio. La asignación no elimina la
historia de backlog y se verifica en integración con sprints.

## Trazabilidad

- BDD: `features/HU-028-list-project-stories.feature`.
- Tests: `backend/internal/backlog/{service,handler,repository_postgres}_test.go`.
- Código: servicio, repositorio y rutas de `backend/internal/backlog/`.
- Rama: `feat/HU-028-list-project-stories`. Commits RED/GREEN separados.
- Servicio: RED `c22760d` → GREEN `7a30c3e`.
- PostgreSQL: RED `5d67b72` → GREEN `2d2d806`.
- HTTP: RED `67c86f2` → GREEN `ce6c375`.
