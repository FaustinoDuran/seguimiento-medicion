# HU-013 — Revelar votos

- Issue: [#22](https://github.com/FaustinoDuran/seguimiento-medicion/issues/22)
- Módulo Go: `backend/internal/planningpoker/`
- Estado: implementada

## Objetivo

El facilitador revela los votos de una ronda abierta para poder discutir las
estimaciones. Revelar cierra la ronda y deja visibles todos los puntajes.
Abrir otra ronda es HU-014 y guardar el puntaje acordado en la historia es
HU-015.

## Dependencia

HU-012 ([#21](https://github.com/FaustinoDuran/seguimiento-medicion/issues/21),
PR [#77](https://github.com/FaustinoDuran/seguimiento-medicion/pull/77)) define
la ronda, el voto y la consulta que oculta el puntaje ajeno. Esta HU agrega el
estado revelado sobre ese modelo; no repite sus reglas.

## Entradas

Revelar (`Reveal`):

| Campo | Obligatorio | Descripción |
|---|---|---|
| `project_id` | sí | Proyecto de la ronda. |
| `round_id` | sí | Ronda abierta que se quiere revelar. |
| `member_id` | sí | Integrante del mismo proyecto que revela. |

Consultar (`View`, definida en HU-012): `project_id`, `round_id` y `member_id`
de quien mira. El `member_id` indica de quién es la vista; no hay sesión.

## Salidas esperadas

Revelar (`POST /projects/{project_id}/rounds/{round_id}/reveal` con
`{ "member_id": N }`) devuelve `200` con la ronda en `status = closed` y la
lista de votos; cada voto trae `member_id` y `story_points`.

La consulta (`GET /projects/{project_id}/rounds/{round_id}?member_id=N`) de una
ronda `closed` devuelve todos los `story_points`. La consulta de una ronda
`open` sigue devolviendo solo el puntaje de quien mira.

## Reglas de negocio

1. Solo se revela una ronda abierta.
2. Revelar cambia el estado de la ronda de `open` a `closed`; no crea otra ronda.
3. Al revelar, todos los votos de la ronda quedan visibles para cualquier
   integrante del proyecto.
4. Antes de revelar, la consulta sigue ocultando el puntaje ajeno (regla 8 de
   HU-012).
5. Una ronda cerrada no acepta votos ni cambios de voto (regla 5 de HU-012);
   esta HU no la reabre.
6. Quien revela debe pertenecer al proyecto de la ronda.
7. La ronda debe existir en el proyecto.
8. Revelar no escribe `story_points` en la historia ni calcula diferencias entre
   votos: eso es HU-015 y HU-014.

## Restricciones

- Las reglas viven en `backend/internal/planningpoker/`.
- Persistencia: PostgreSQL. El cambio de estado se guarda en `planning_rounds`.
- API JSON. Sin autenticación.
- El estado `closed` ya existe en el esquema de HU-012: no hace falta una
  migración nueva.

## Casos límite

- Ronda abierta sin votos: se puede revelar y devuelve la lista vacía.
- Ronda abierta con un solo voto: se revela igual.
- Votaron solo algunos integrantes: la lista trae solo los votos emitidos, sin
  entradas para quien no votó.
- Todos los votos coinciden: se revelan igual; abrir otra ronda no es
  obligatorio (HU-014).

## Condiciones de error

- Ronda inexistente: no se revela.
- Ronda de otro proyecto: no se revela.
- Integrante inexistente en el proyecto: no se revela.
- Ronda ya cerrada: no se revela de nuevo.
- `member_id` ausente o inválido: `400`.

## Criterios de aceptación

- [x] AC-013-01 — Con una ronda abierta y votos de dos integrantes, al revelar se ven todos los puntajes.
- [x] AC-013-02 — Antes de revelar, la consulta de la ronda no incluye el puntaje ajeno.
- [x] AC-013-03 — Al revelar, la ronda queda cerrada.
- [x] AC-013-04 — Después de revelada, cualquier integrante del proyecto consulta y ve todos los puntajes.
- [x] AC-013-05 — Si quien revela no pertenece al proyecto, no se revela.
- [x] AC-013-06 — Si la ronda no existe o es de otro proyecto, no se revela.
- [x] AC-013-07 — Si la ronda ya está cerrada, no se revela de nuevo.
- [x] AC-013-08 — Revelar no modifica el `story_points` de la historia.
- [x] AC-013-09 — Una ronda cerrada sigue rechazando votos nuevos.

## Trazabilidad

- BDD: `features/HU-013-revelar-votos.feature`
- Tests: `backend/internal/planningpoker/{service,handler,repository_postgres}_test.go`
- Código: `backend/internal/planningpoker/`
- Depende de: HU-012
