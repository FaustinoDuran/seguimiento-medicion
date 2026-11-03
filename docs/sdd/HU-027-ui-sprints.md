# HU-027 — Gestionar sprints desde la UI

- Issue: [#49](https://github.com/FaustinoDuran/seguimiento-medicion/issues/49)
- Módulo Go existente: `backend/internal/sprints/`
- Presentación prevista: `frontend/src/features/sprints/`
- Estado: UI implementada; verificación en navegador pendiente

## Objetivo

Como facilitador, crear un Sprint, asignar historias, completar al menos una y
cerrarlo desde la interfaz para demostrar el ciclo Scrum en la Review sin usar
solo JSON.

## Entradas

- Proyecto identificado por `projectId` en la URL.
- Sprint identificado por `sprintId` en la URL.
- Goal del Sprint, obligatorio, escrito en el formulario de alta.
- Historias del proyecto obtenidas de la API.
- Historia a asignar, elegida de un selector.

## Salidas esperadas

Dos rutas, ambas accesibles por URL directa:

1. `/projects/:projectId/sprints`
   - Nombre del proyecto como encabezado.
   - Formulario «Nuevo sprint» con un único campo obligatorio: goal.
   - Tras un alta exitosa, redirección a la ruta de detalle del Sprint creado.
   - Sección «Sprints anteriores» con los sprints cerrados del proyecto:
     goal, estado y fecha de creación.
   - Estado vacío cuando el proyecto no tiene sprints cerrados.
2. `/projects/:projectId/sprints/:sprintId`
   - Goal, estado y fecha de creación del Sprint.
   - Sprint abierto: listado de historias asignadas y formulario para asignar
     una historia del proyecto; botón para cerrar el Sprint.
   - Sprint cerrado: solo lectura, sin acciones disponibles.
   - Cada historia asignada muestra identificador, título y estado.
   - En las historias en estado `todo` se ofrece la acción «Completar».
   - Enlace para volver al listado de sprints del proyecto.

Etiquetas en español: Sprint `open` = «Abierto», `closed` = «Cerrado»; historia
`todo` = «Por hacer», `done` = «Completada».

## Reglas de negocio

Ninguna regla se implementa en React. La UI presenta lo que decide Go:

- El goal se recorta y no puede quedar vacío (`ErrGoalRequired`, `400`).
- El Sprint nace en estado `open`.
- Asignar exige un Sprint abierto y una historia del mismo proyecto que no
  esté ya asignada a ningún Sprint.
- Completar exige un Sprint abierto y que la historia esté asignada a ese
  Sprint; el resultado es `status: "done"`. Repetir la operación sobre una
  historia ya completada vuelve a responder éxito mientras el Sprint siga
  abierto.
- Cerrar un Sprint ya cerrado es rechazado.
- El listado de sprints anteriores devuelve únicamente los `closed`, sin
  historias embebidas.

## Contratos de la API

| Acción | Endpoint | Éxito | Respuesta |
|---|---|---|---|
| Crear | `POST /projects/{id}/sprints` | `201` | sprint |
| Consultar | `GET /sprints/{id}` | `200` | sprint |
| Asignar | `POST /sprints/{id}/stories` | `201` | historia asignada |
| Listar asignadas | `GET /sprints/{id}/stories` | `200` | array de historias asignadas |
| Completar | `POST /sprints/{id}/stories/{story_id}/complete` | `200` | historia asignada |
| Cerrar | `POST /sprints/{id}/close` | `200` | sprint |
| Anteriores | `GET /projects/{id}/sprints/closed` | `200` | array de sprints |

Sprint: `{ "id", "project_id", "goal", "status", "created_at" }`.

Historia asignada: `{ "sprint_id", "story_id", "identifier", "title", "status" }`.

Los listados son arrays JSON planos, no envueltos en un objeto. Los errores usan
`{ "error": "..." }`.

Historias del proyecto para el selector: `GET /projects/{id}/stories`, que
responde `{ "stories": [...] }` con el esquema de HU-005/HU-028.

## Restricciones

- HU de presentación: no agrega ni modifica código Go. Se mantiene ADR 0004.
- Depende de HU-007, HU-008, HU-009, HU-010, HU-011 y HU-026, todas mergeadas.
- **No existe un endpoint que liste los sprints abiertos de un proyecto.**
  HU-011 excluye explícitamente los `open`. Por eso el `sprintId` viaja en la
  URL: sin él, tras una recarga no hay forma de recuperar el Sprint en curso.
  Listar sprints abiertos queda fuera de esta HU.
- No hay navegación hacia estas rutas: el acceso es por URL directa. La
  navegación corresponde a HU-024 y se agrega después.
- La UI no agrega restricciones que Go no aplique: por ejemplo, no bloquea
  crear un Sprint en un proyecto cerrado ni filtra historias por estado.
- Un selector de historias solo puede excluir las ya asignadas a este Sprint,
  porque es lo único que la API informa. Una historia asignada a otro Sprint
  se descubre al recibir `409`.

## Casos límite

- Proyecto sin sprints cerrados: mostrar el estado vacío del listado de
  anteriores, no un error.
- Sprint sin historias asignadas: mostrar el estado vacío del listado.
- Proyecto sin historias: no ofrecer asignación; explicar que primero deben
  crearse historias en el backlog.
- Todas las historias del proyecto ya asignadas a este Sprint: selector vacío
  con aviso, sin formulario inutilizable.
- Sprint cerrado: no mostrar acciones de asignar, completar ni cerrar.
- Recarga de la página de detalle: debe seguir funcionando, porque el
  `sprintId` está en la URL.
- `sprintId` de otro proyecto: mostrar «Sprint no encontrado».
- `projectId` o `sprintId` no numérico: mostrar «Sprint no encontrado» o
  «Proyecto no encontrado» según corresponda, sin llamar a la API.
- Goal compuesto solo por espacios: mostrar el rechazo de Go en español.

## Condiciones de error

- Mensajes en español; traducir los errores conocidos de la API.
- Errores de campo junto al campo; los demás, en un aviso general.
- Conservar los datos ingresados cuando falla el alta.
- Ante `409` al asignar, explicar que la historia ya está asignada a un Sprint.
- Ante `409` al cerrar, explicar que el Sprint ya está cerrado.
- Ante `409` al completar, explicar que la historia no pertenece a este Sprint.
- Mostrar carga visible y ofrecer reintento ante fallos de consulta. Un error
  de conexión no se presenta como listado vacío.
- No confirmar una acción cuya respuesta fue fallida.
- Si falla la conexión durante una acción, indicar que no se pudo confirmar el
  resultado y sugerir recargar antes de reintentar.

## Criterios de aceptación

- [ ] AC-027-01 — Crear un Sprint con goal desde la UI y quedar en su pantalla
  de detalle.
- [ ] AC-027-02 — Rechazar el alta con goal vacío mostrando el mensaje de la
  API junto al campo, conservando el resto de los datos.
- [ ] AC-027-03 — Asignar una historia del proyecto al Sprint abierto desde la
  UI y verla aparecer en el listado con su identificador, título y estado.
- [ ] AC-027-04 — Completar una historia asignada al Sprint abierto y ver el
  estado «Completada» reflejado en el listado.
- [ ] AC-027-05 — Cerrar el Sprint desde la UI y ver el estado «Cerrado».
- [ ] AC-027-06 — Ver los sprints anteriores del proyecto, con goal, estado y
  fecha de creación.
- [ ] AC-027-07 — Acceder a ambas rutas por URL directa, sin depender de
  navegación agregada por otra HU.
- [ ] AC-027-08 — Con el Sprint cerrado, no ofrecer asignar, completar ni
  cerrar, y mostrar el aviso correspondiente.
- [ ] AC-027-09 — Presentar en español los rechazos `409` de asignación
  duplicada, Sprint cerrado e historia no asignada, y los `404` de Sprint o
  proyecto inexistente.
- [ ] AC-027-10 — Mostrar carga, estados vacíos y reintento ante errores de
  consulta, diferenciando el vacío real del fallo.
- [ ] AC-027-11 — No duplicar reglas de negocio en React: toda validación de
  dominio proviene de la API.
- [ ] AC-027-12 — Verificar con `npm run build`, `npm run lint` y un recorrido
  completo contra la API Go y PostgreSQL reales.

## Decisiones de diseño

- Dos rutas en lugar de una: el `sprintId` en la URL mantiene la pantalla de
  detalle usable tras una recarga, dado que la API no lista sprints abiertos.
- El alta redirige al detalle para encadenar crear → asignar → completar →
  cerrar sin pasos manuales.
- El cierre pide confirmación explícita en la UI porque es una acción sin
  vuelta atrás en esta HU: la reapertura no está en el alcance.
- Navegación fuera de alcance por decisión del responsable; la agrega la HU que
  corresponda.

## Trazabilidad

- BDD: `features/HU-027-ui-sprints.feature`
- Código previsto: `frontend/src/features/sprints/` y rutas en
  `frontend/src/App.tsx`
- Reglas existentes: `backend/internal/sprints/service.go` y sus tests
- Tests Go de referencia: `backend/internal/sprints/{service,handler,repository_postgres}_test.go`
- Contrato de historias: `backend/internal/backlog/handler.go` (HU-028)
