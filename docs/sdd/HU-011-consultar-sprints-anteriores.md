# HU-011 — Consultar sprints anteriores

- Issue: [#20](https://github.com/FaustinoDuran/seguimiento-medicion/issues/20)
- Módulo Go: `backend/internal/sprints/`
- Estado: implementada

## Objetivo

Listar los Sprints ya cerrados de un proyecto para poder elegirlos y comparar
después lo planificado con lo completado.

## Entradas

Listado (`ListClosed`):

| Campo | Obligatorio | Descripción |
|---|---|---|
| `project_id` | sí | Proyecto cuyos sprints cerrados se consultan. Debe existir (HU-001). |

## Salidas esperadas

Una lista de Sprints con `status` `closed` del proyecto, cada uno con:

- `id`;
- `project_id`;
- `goal`;
- `status` (`closed`);
- `created_at`.

No se incluyen sprints `open`. La lista no embebe historias: para ver lo
planificado y lo completado se usa la consulta ya existente de historias del
Sprint (`GET /sprints/{id}/stories`, HU-008 / HU-009).

Si el proyecto existe y no hay sprints cerrados, la lista queda vacía.

## Reglas de negocio

1. El proyecto debe existir. Si no, no se lista.
2. Solo se listan sprints con `status` `closed` del `project_id` indicado.
3. Los sprints `open` del mismo proyecto no aparecen.
4. El orden es por `created_at` ascendente y, a igualdad, por `id` ascendente.
5. Cada sprint cerrado del listado sigue consultable por `id` y conserva sus
   historias asignadas (comportamiento de HU-007, HU-008 y HU-010).

## Restricciones

- Las reglas viven en `backend/internal/sprints/`. Handler sin reglas.
- Persistencia: tabla `sprints` (filtro por `project_id` y `status`).
- API JSON. Sin autenticación. Sin React (UI en HU-027).
- No se cierran sprints (HU-010) ni se calculan métricas.

## Casos límite

- Proyecto sin sprints: lista vacía.
- Proyecto solo con sprints abiertos: lista vacía.
- Varios sprints cerrados: todos aparecen, en orden de creación.

## Contrato de la API

- `GET /projects/{id}/sprints/closed`: `200` con un array JSON de sprints
  (`id`, `project_id`, `goal`, `status`, `created_at`). Sin cerrados: `[]`.
- `404`: proyecto inexistente.
- `400`: `id` de proyecto inválido.
- Composición (sin lógica nueva en esta HU):
  - `GET /sprints/{id}`: detalle del sprint cerrado.
  - `GET /sprints/{id}/stories`: historias planificadas y su estado
    (`todo` / `done`).

## Condiciones de error

- Proyecto inexistente: no se lista.
- Identificador de proyecto inválido: no se lista.

## Criterios de aceptación

- [x] AC-011-01 — Con sprints cerrados en el proyecto, el listado los devuelve con sus datos y omite los abiertos.
- [x] AC-011-02 — Si el proyecto no tiene sprints cerrados, la lista queda vacía.
- [x] AC-011-03 — Si el proyecto no existe, no se lista y se informa el error.
- [x] AC-011-04 — Un sprint cerrado del listado se puede seguir consultando por `id` y con sus historias.

## Trazabilidad

- BDD: `features/HU-011-consultar-sprints-anteriores.feature`
- Tests: `backend/internal/sprints/{service,handler,repository_postgres}_test.go`
- Código: `backend/internal/sprints/`
