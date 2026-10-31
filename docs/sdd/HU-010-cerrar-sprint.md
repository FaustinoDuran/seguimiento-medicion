# HU-010 — Cerrar Sprint

- Issue: [#19](https://github.com/FaustinoDuran/seguimiento-medicion/issues/19)
- Módulo Go: `backend/internal/sprints/`
- Estado: lista para implementar

## Objetivo

Cerrar un Sprint abierto para dejar fijas las historias planificadas (y las
completadas, cuando existan) y poder seguir consultándolo.

## Entradas

Cierre (`Close`):

| Campo | Obligatorio | Descripción |
|---|---|---|
| `sprint_id` | sí | Sprint a cerrar. Debe existir (HU-007) y estar `open`. |

## Salidas esperadas

El Sprint queda con `status` `closed`. La consulta por `id` lo sigue devolviendo
con el mismo `goal`, `project_id` e historias ya asignadas.

## Reglas de negocio

1. El Sprint debe existir. Si no, no se cierra.
2. Solo se cierra un Sprint `open`.
3. Un Sprint `closed` se sigue consultando por su `id`.
4. Un Sprint `closed` no acepta nuevas asignaciones (regla ya usada en HU-008).
5. Las asignaciones existentes se conservan. Esta HU no completa historias
   (HU-009) ni lista sprints anteriores (HU-011).

## Restricciones

- Las reglas viven en `backend/internal/sprints/`.
- Persistencia: el `status` de la tabla `sprints`.
- API JSON. Sin autenticación. Sin React.

## Casos límite

- Cerrar un Sprint abierto sin historias: válido.
- Cerrar un Sprint que ya tiene historias asignadas: válido; el listado no cambia.

## Contrato de la API

- `POST /sprints/{id}/close`: cierra y devuelve `200` con el Sprint.
- `GET /sprints/{id}`: sigue devolviendo `200` después del cierre.
- `POST /sprints/{id}/stories` sobre un Sprint cerrado: `409`.
- `404`: Sprint inexistente.
- `409`: Sprint ya cerrado.

## Condiciones de error

- Sprint inexistente: no se cierra.
- Sprint ya cerrado: no se vuelve a cerrar.

## Criterios de aceptación

- [ ] AC-010-01 — Un Sprint abierto queda `closed` y se puede seguir consultando.
- [ ] AC-010-02 — Un Sprint cerrado no acepta nuevas asignaciones y conserva las que ya tenía.
- [ ] AC-010-03 — Si el Sprint ya está cerrado, no se cierra de nuevo y se informa el error.
- [ ] AC-010-04 — Si el Sprint no existe, no se cierra y se informa el error.

## Trazabilidad

- BDD: `features/HU-010-cerrar-sprint.feature`
- Tests: `backend/internal/sprints/service_test.go`
- Código: `backend/internal/sprints/service.go`
