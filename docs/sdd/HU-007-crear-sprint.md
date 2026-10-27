# HU-007 — Crear Sprint

- Issue: [#16](https://github.com/FaustinoDuran/seguimiento-medicion/issues/16)
- Módulo Go: `backend/internal/sprints/`
- Estado: lista para implementar

## Objetivo

Dar de alta un Sprint con su objetivo sobre un proyecto existente, y poder
consultarlo después. Sin iteración no hay seguimiento.

## Entradas

Alta (`CreateSprint`):

| Campo | Obligatorio | Descripción |
|---|---|---|
| `project_id` | sí | Proyecto al que pertenece. Debe existir (HU-001). |
| `goal` | sí | Sprint Goal visible. |

Consulta: identificador del Sprint generado en el alta.

## Salidas esperadas

Un Sprint persistido con:

- `id` generado por el sistema;
- `project_id` y `goal`;
- `status` en `open`;
- `created_at`.

La consulta por `id` devuelve el mismo Sprint.

## Reglas de negocio

1. El proyecto debe existir. Si no, no se crea el Sprint.
2. El objetivo no puede estar vacío ni ser solo espacios; se guarda recortado.
3. Al crearlo el estado queda `open`. El cliente no elige el estado.
4. El identificador lo asigna el sistema.
5. Un Sprint creado se puede consultar por su `id`.

## Restricciones

- Las reglas viven en `backend/internal/sprints/`, no en React ni en el handler.
- El módulo no importa el handler ni el SQL de `projects`. Comprueba que el
  proyecto exista a través de un puerto.
- Persistencia: PostgreSQL (tabla `sprints` de esta HU).
- API JSON. Sin autenticación.
- No se asignan historias (HU-008), no se completan (HU-009) y no se cierra
  el Sprint (HU-010).

## Casos límite

- Objetivo con espacios alrededor: se acepta y se guarda recortado.
- `status` no viaja en el alta: queda `open`.

## Condiciones de error

- Objetivo ausente o en blanco: no se crea.
- Proyecto inexistente: no se crea.
- Consulta de un `id` inexistente: no hay Sprint.

## Criterios de aceptación

- [ ] AC-007-01 — Con un proyecto existente y un objetivo, el Sprint queda creado en estado `open` y se puede consultar por su `id`.
- [ ] AC-007-02 — Si falta el objetivo (o queda en blanco), no se crea y se informa el error.
- [ ] AC-007-03 — Si el proyecto no existe, no se crea y se informa el error.

## Trazabilidad

- BDD: `features/HU-007-crear-sprint.feature`
- Tests: `backend/internal/sprints/service_test.go`
- Código: `backend/internal/sprints/service.go`
