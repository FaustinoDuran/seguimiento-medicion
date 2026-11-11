# HU-029 — Usar Planning Poker desde la UI

- Issue: [#74](https://github.com/FaustinoDuran/seguimiento-medicion/issues/74)
- Módulo Go existente: `backend/internal/planningpoker/`
- Presentación: `frontend/src/features/planningpoker/`
- Estado: mergeada en `main` ([PR #81](https://github.com/FaustinoDuran/seguimiento-medicion/pull/81))

## Objetivo

Votar, revelar y guardar la estimación de una historia desde la interfaz, sin
usar la API a mano. Las reglas siguen en Go (HU-012 a HU-015).

## Entradas

El proyecto es el de `/projects/:projectId`.

| Campo | Obligatorio | Descripción |
|---|---|---|
| Integrante | sí, para operar una ronda | Quién mira y actúa. No hay sesión: se elige en la pantalla y se envía como `member_id`. |
| Historia | sí, para abrir | Historia del backlog de ese proyecto. |
| Tu story point | sí, para votar | Entero que Go valida. |
| Story points acordados | sí, para guardar | Entero elegido en la discusión. Go decide si la ronda lo admite. |

## Salidas esperadas

- Desde el detalle del proyecto, el enlace Planning Poker abre `/projects/:projectId/planning`.
- Ahí se ven las historias y la estimación que ya tiene cada una (`story_points` o «Sin estimar»).
- Abrir una ronda lleva a `/projects/:projectId/planning/rounds/:roundId`.
- Con la ronda abierta, quien vota ve su puntaje. El puntaje de otro integrante no aparece: la pantalla muestra «Oculto».
- Revelar muestra el puntaje de cada voto y el indicador que devuelve la API.
- Si `has_differences` es verdadero, se puede abrir otra ronda. La nueva queda en la URL, vacía y otra vez en secreto.
- Si `has_differences` es falso, no se ofrece otra ronda y se puede guardar el acuerdo.
- Guardar muestra la estimación nueva en la historia.
- Un rechazo de Go se ve en español.

## Reglas de negocio

No hay reglas nuevas. React no compara votos, no decide si hay diferencias y no
calcula el story point acordado.

- Ocultar el puntaje ajeno, reemplazar el voto propio y rechazar un puntaje que no sea un entero mayor que cero: HU-012.
- Revelar y cerrar la ronda: HU-013.
- `has_differences` y abrir otra ronda: HU-014. La pantalla solo muestra el booleano cuando la API lo envía.
- Guardar el acuerdo en la historia: HU-015. El número lo escribe quien opera; puede no coincidir con los votos.

## Restricciones

- Depende de HU-012, HU-013, HU-014 y HU-015, ya en `main`.
- Sin cambios de reglas en Go.
- Sin autenticación. Dos personas usan la misma URL de ronda y eligen integrantes distintos.

## Contrato que usa la pantalla

- `GET /projects/{id}`, `GET /projects/{id}/members`, `GET /projects/{id}/stories`.
- `POST /projects/{id}/stories/{story_id}/rounds`: `201`. Abre la primera ronda o la siguiente.
- `GET /projects/{id}/rounds/{round_id}?member_id=`: la vista de quien mira.
- `POST /projects/{id}/rounds/{round_id}/votes` con `{ "member_id", "story_points" }`.
- `POST /projects/{id}/rounds/{round_id}/reveal` con `{ "member_id" }`.
- `POST /projects/{id}/rounds/{round_id}/estimate` con `{ "member_id", "story_points" }`: `200`.

En una ronda abierta la respuesta no trae `story_points` ajenos ni `has_differences`.
En una ronda cerrada trae todos los puntajes y `has_differences`.

Errores que la pantalla traduce:

| Error de Go | Texto en la pantalla |
|---|---|
| `story points must be a positive integer` | Ingresá un story point entero mayor que cero. |
| `round is closed` | Esta ronda ya está cerrada. |
| `an open round already exists` | Ya hay una ronda abierta para esta historia. |
| `round must be revealed` | Revelá la ronda antes de guardar la estimación. |
| `round has no votes` | No se puede guardar una estimación sin votos. |
| `round is not the latest for this story` | Esta ronda ya no es la última. |
| `completed story cannot be estimated` | No se puede estimar una historia completada. |
| `story belongs to a closed sprint` | La historia está en un sprint cerrado. |
| Otro rechazo | No pudimos completar la votación. Volvé a intentarlo. |

## Casos límite

- Sin integrante elegido no se consulta la ronda: la vista depende de `member_id`.
- Cambiar de integrante vuelve a consultar. El puntaje que solo veía el anterior deja de mostrarse.
- Ronda sin votos: «Todavía no hay votos».
- Proyecto sin historias: «Todavía no hay historias».

## Condiciones de error

- Puntaje rechazado al votar: alerta en español y el formulario sigue.
- Fallo al cargar el proyecto, las historias o la ronda: aviso y reintento, sin tratarlo como vacío.

## Criterios de aceptación

- [x] AC-029-01 — Votar y ver el puntaje propio; el de otro integrante queda oculto.
- [x] AC-029-02 — Revelar y ver el puntaje de todos los votos.
- [x] AC-029-03 — Si la API informa diferencias, abrir otra ronda vacía.
- [x] AC-029-04 — Si la API informa que no hay diferencias, no se exige otra ronda.
- [x] AC-029-05 — Guardar el story point acordado y verlo en la historia.
- [x] AC-029-06 — Mostrar un error en español si el voto no se puede registrar.

## Trazabilidad

- BDD: `features/HU-029-ui-planning-poker.feature`
- Tests: `frontend/src/features/planningpoker/PlanningPoker.test.tsx`
- Código: `frontend/src/features/planningpoker/`
