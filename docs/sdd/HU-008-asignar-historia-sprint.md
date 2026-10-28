# HU-008 — Asignar historia a un Sprint

- Issue: [#17](https://github.com/FaustinoDuran/seguimiento-medicion/issues/17)
- Módulo Go: `backend/internal/sprints/`
- Estado: lista para implementar

## Objetivo

Vincular historias de backlog a un Sprint abierto para armar el trabajo de la
iteración, y poder consultar después cuáles quedaron en ese Sprint.

## Entradas

Asignación (`Assign`):

| Campo | Obligatorio | Descripción |
|---|---|---|
| `sprint_id` | sí | Sprint destino. Debe existir (HU-007). |
| `story_id` | sí | Historia de backlog. Debe existir (HU-005) y pertenecer al mismo proyecto que el Sprint. |

Consulta: identificador del Sprint.

## Salidas esperadas

La historia queda vinculada a ese Sprint. La consulta del Sprint devuelve cada
historia asignada con `sprint_id`, `story_id`, `identifier` y `title`.

La historia sigue en estado `todo`. Esta HU no la completa.

## Reglas de negocio

1. El Sprint debe existir y estar `open`.
2. La historia debe existir dentro del proyecto del Sprint. Una historia de otro proyecto no se asigna.
3. Una historia pertenece como máximo a un Sprint. Si ya está asignada, no se reasigna ni se duplica.
4. Un Sprint abierto puede recibir varias historias, también si no tienen story points.
5. El vínculo se puede consultar por el `id` del Sprint.
6. El estado de la historia no cambia al asignarla.

## Restricciones

- Las reglas viven en `backend/internal/sprints/`, no en React ni en el handler.
- El módulo no importa el handler ni el SQL de `backlog`. Comprueba la historia a través de un puerto, usando el proyecto del Sprint.
- Persistencia: tabla `sprint_stories` de esta HU. `story_id` es único.
- API JSON. Sin autenticación.
- No se desasigna, no se reordena, no se completa la historia (HU-009) y no se cierra el Sprint (HU-010).

## Casos límite

- Dos historias distintas del mismo proyecto quedan las dos en el Sprint.
- Volver a asignar la misma historia, al mismo Sprint o a otro, se rechaza y el vínculo anterior se conserva.
- Una historia sin story points se puede asignar.

## Contrato de la API

- `POST /sprints/{id}/stories` con `{ "story_id": 1 }`: asigna y devuelve `201`.
- `GET /sprints/{id}/stories`: devuelve `200` y la lista de historias del Sprint. Sin historias, la lista va vacía.
- Errores con formato `{"error":"..."}`.
- `400`: `story_id` ausente o inválido.
- `404`: Sprint inexistente, historia inexistente o historia de otro proyecto.
- `409`: Sprint cerrado o historia ya asignada a un Sprint.

## Condiciones de error

- Sprint inexistente: no se asigna.
- Historia inexistente o de otro proyecto: no se asigna.
- Sprint cerrado: no se asigna.
- Historia ya asignada: no se cambia el vínculo.
- Consulta de un Sprint inexistente: no hay lista.

## Criterios de aceptación

- [ ] AC-008-01 — Con un Sprint abierto y una historia del mismo proyecto, la historia queda asignada y aparece al consultar el Sprint.
- [ ] AC-008-02 — Si el Sprint o la historia no existen, o la historia es de otro proyecto, no se asigna.
- [ ] AC-008-03 — Si el Sprint está cerrado, no se asigna.
- [ ] AC-008-04 — Si la historia ya está en un Sprint, no se asigna de nuevo y el vínculo anterior se conserva.

## Trazabilidad

- BDD: `features/HU-008-asignar-historia-sprint.feature`
- Tests: `backend/internal/sprints/service_test.go`
- Código: `backend/internal/sprints/service.go`
