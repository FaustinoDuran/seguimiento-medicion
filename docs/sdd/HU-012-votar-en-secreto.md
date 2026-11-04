# HU-012 — Votar en secreto

- Issue: [#21](https://github.com/FaustinoDuran/seguimiento-medicion/issues/21)
- Módulo Go: `backend/internal/planningpoker/`
- Estado: lista para implementar

## Objetivo

Un integrante vota los story points de una historia sin ver el voto de los
demás. Revelar, abrir otra ronda y guardar el puntaje acordado en la historia
son HU-013, HU-014 y HU-015.

## Entradas

Abrir ronda (`OpenRound`):

| Campo | Obligatorio | Descripción |
|---|---|---|
| `project_id` | sí | Proyecto de la historia. |
| `story_id` | sí | Historia a estimar. Debe existir en ese proyecto (HU-005). |

Votar (`CastVote`):

| Campo | Obligatorio | Descripción |
|---|---|---|
| `project_id` | sí | Proyecto de la ronda. |
| `round_id` | sí | Ronda abierta. |
| `member_id` | sí | Integrante del mismo proyecto (HU-003). |
| `story_points` | sí | Entero mayor que cero. |

Consultar la ronda (`View`): `project_id`, `round_id` y `member_id` de quien
mira. El `member_id` indica de quién es la vista; no hay sesión.

## Salidas esperadas

Una ronda abierta con `id`, `project_id`, `story_id`, `status = open` y
`created_at`.

Un voto con `id`, `round_id`, `member_id`, `story_points` y `created_at`.

La consulta devuelve la ronda y los votos emitidos. El integrante que mira ve
su propio puntaje. Los demás votos no traen `story_points`.

## Reglas de negocio

1. La historia debe existir en el proyecto. Si no, no se abre la ronda ni se vota.
2. Solo puede haber una ronda abierta por historia.
3. El integrante debe pertenecer al proyecto de la historia.
4. El puntaje es un entero mayor que cero, la misma regla que los story points de la historia.
5. Solo se vota en una ronda abierta.
6. Cada integrante tiene un voto por ronda. Si vuelve a votar en la misma ronda abierta, se reemplaza solo el suyo.
7. Un voto no modifica el puntaje de otro integrante.
8. La consulta nunca incluye el puntaje de otro integrante. Mostrarlo es trabajo de HU-013.

## Restricciones

- Las reglas viven en `backend/internal/planningpoker/`.
- Persistencia: PostgreSQL.
- API JSON. Sin autenticación.
- Esta HU no revela votos, no abre una ronda nueva después de una diferencia y no escribe el story point de la historia.

## Casos límite

- El integrante puede cambiar su propio voto mientras la ronda sigue abierta.
- Una historia sin ronda abierta no tiene votos que consultar por esta operación: primero se abre la ronda.
- El mismo integrante puede votar en otra historia.

## Condiciones de error

- Historia inexistente en el proyecto: no se abre la ronda.
- Ya hay una ronda abierta para esa historia: no se abre otra.
- Integrante inexistente en el proyecto: no se guarda el voto.
- Puntaje ausente, cero o negativo: no se guarda.
- Ronda inexistente, de otro proyecto o cerrada: no se guarda el voto.
- Consulta sin integrante del proyecto: no se muestra la ronda.

## Criterios de aceptación

- [ ] AC-012-01 — Con una historia y un integrante del proyecto, se abre una ronda y el integrante emite un voto que puede ver.
- [ ] AC-012-02 — Al consultar, el puntaje de otro integrante no viene en la respuesta.
- [ ] AC-012-03 — El voto de un integrante no cambia el puntaje guardado de otro.
- [ ] AC-012-04 — Un integrante puede reemplazar su propio voto mientras la ronda está abierta.
- [ ] AC-012-05 — Si el puntaje no es un entero mayor que cero, no se guarda.
- [ ] AC-012-06 — Si el integrante no pertenece al proyecto, no se guarda.
- [ ] AC-012-07 — Si la historia no existe en el proyecto, no se abre la ronda.
- [ ] AC-012-08 — Si ya hay una ronda abierta para la historia, no se abre otra.
- [ ] AC-012-09 — Si la ronda está cerrada, no se acepta el voto.

## Trazabilidad

- BDD: `features/HU-012-votar-en-secreto.feature`
- Tests: `backend/internal/planningpoker/`
- Código: `backend/internal/planningpoker/`
