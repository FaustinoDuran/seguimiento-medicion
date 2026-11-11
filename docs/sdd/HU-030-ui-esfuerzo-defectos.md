# HU-030 — Cargar esfuerzo y defectos desde la UI

- Issue: [#85](https://github.com/FaustinoDuran/seguimiento-medicion/issues/85)
- Módulos Go: `backend/internal/effort/` y `backend/internal/defects/`
- Presentación: `frontend/src/features/measurement/`, `features/effort/` y `features/defects/`
- Estado: en desarrollo
- Depende de: HU-016, HU-017 y HU-018

## Objetivo

Cargar horas y defectos desde la interfaz y resolver defectos con su Sprint,
sin usar la API a mano. Las reglas siguen en Go (HU-016, HU-017 y HU-018).
Esta HU agrega únicamente dos lecturas que la pantalla necesita y no existían.

## Entradas

El proyecto es el de `/projects/:projectId`.

| Campo | Obligatorio | Descripción |
|---|---|---|
| Integrante | sí, para el esfuerzo | Se elige en la pantalla; no hay sesión. Se envía como `member_id`. |
| Historia | sí | Historia del proyecto. Define sobre qué se cargan horas y defectos. |
| Fecha | sí, para el esfuerzo | Día calendario `YYYY-MM-DD`. La pantalla propone la fecha de hoy. |
| Actividad | sí, para el esfuerzo | Texto de hasta 200 caracteres. |
| Horas | sí, para el esfuerzo | Número mayor que cero, con hasta dos decimales. |
| Descripción | sí, para el defecto | Texto de hasta 2000 caracteres. |
| Severidad | sí, para el defecto | Baja, media, alta o crítica. Sin valor predeterminado. |
| Sprint de detección | sí, para el defecto | Sprint del proyecto. Puede ser distinto del asignado a la historia. |
| Sprint de resolución | sí, para resolver | Sprint del proyecto, no anterior al de detección. |

## Salidas esperadas

- Desde el detalle del proyecto, el enlace «Esfuerzo y defectos» abre
  `/projects/:projectId/measurement`.
- La pantalla muestra el formulario de esfuerzo, el de defecto y la lista de
  defectos del proyecto.
- Registrar esfuerzo confirma el alta con un aviso y limpia actividad y horas.
- Registrar defecto confirma el alta, deja el defecto abierto y lo agrega a la
  lista.
- Un defecto abierto ofrece elegir sprint de resolución y marcarlo resuelto. Al
  resolverlo, la lista muestra el sprint de resolución y deja de ofrecer la acción.
- Un rechazo de Go se ve en español, junto al formulario que lo provocó.

## Reglas de negocio

No hay reglas nuevas de esfuerzo, defectos ni resolución: se reutilizan las de
HU-016, HU-017 y HU-018. React no valida horas, no decide severidades, no
compara sprints por fecha y no calcula métricas ni desviaciones.

Las dos lecturas nuevas solo ordenan y aíslan por proyecto:

1. `GET /projects/{project_id}/sprints` devuelve todos los Sprints del proyecto,
   abiertos y cerrados, ordenados por `id` ascendente. No filtra por estado.
2. `GET /projects/{project_id}/defects` devuelve los defectos del proyecto,
   ordenados por `id` descendente. Incluye `resolution_sprint_id`, que es `null`
   cuando el defecto está abierto.
3. Ambas lecturas responden `404` si el proyecto no existe, `400` si el
   identificador de la ruta no es un entero positivo y `500` genérico ante un
   fallo de infraestructura.

## Restricciones

- Depende de HU-016, HU-017 y HU-018 en `main`.
- Sin cambios en las reglas de esfuerzo, defectos y resolución.
- Sin autenticación: el integrante se elige en la pantalla.
- React solo presenta; no calcula horas, métricas ni desviaciones.
- Un solo PR toca `App.tsx` y `ProjectDetailPage.tsx`.

## Contrato que usa la pantalla

Existentes:

- `GET /projects/{id}`, `GET /projects/{id}/members`, `GET /projects/{id}/stories`.
- `POST /projects/{id}/stories/{story_id}/effort` con
  `{ "member_id", "date", "activity", "hours" }`: `201`.
- `POST /projects/{id}/stories/{story_id}/defects` con
  `{ "description", "severity", "detection_sprint_id" }`: `201`, estado `open`.
- `POST /projects/{id}/defects/{id}/resolve` con `{ "resolution_sprint_id" }`:
  `200` con el defecto resuelto.

Nuevas en esta HU:

- `GET /projects/{id}/sprints`: `200` con un arreglo de
  `{ "id", "project_id", "goal", "status", "created_at" }` ordenado por `id`.
- `GET /projects/{id}/defects`: `200` con `{ "defects": [...] }` ordenado por
  `id` descendente, cada defecto igual al de `GET /projects/{id}/defects/{id}`.

Errores que la pantalla traduce:

| Error de Go | Texto en la pantalla |
|---|---|
| `activity is required` | Escribí la actividad realizada. |
| `activity must be at most 200 characters` | La actividad no puede superar los 200 caracteres. |
| `effort date is required` | Elegí la fecha del esfuerzo. |
| `hours are required` | Ingresá las horas dedicadas. |
| `hours must be greater than zero` | Las horas deben ser mayores que cero. |
| `hours are invalid` | Ingresá las horas con hasta dos decimales. |
| `description is required` | Escribí la descripción del defecto. |
| `description must be at most 2000 characters` | La descripción no puede superar los 2000 caracteres. |
| `severity must be low, medium, high or critical` | Elegí una severidad válida. |
| `initial defect status must be open` | El defecto se registra abierto. |
| `story not found` | La historia no pertenece a este proyecto. |
| `member not found` | El integrante no pertenece a este proyecto. |
| `detection sprint not found` | El sprint no pertenece a este proyecto. |
| `defect not found` | El defecto no existe en este proyecto. |
| `resolution sprint precedes detection sprint` | El sprint de resolución no puede ser anterior al de detección. |
| `defect is already resolved in another sprint` | El defecto ya está resuelto en otro sprint. |
| Otro rechazo | No pudimos completar la operación. Volvé a intentarlo. |

## Casos límite

- Proyecto sin integrantes: el formulario de esfuerzo avisa que falta un
  integrante y no envía.
- Proyecto sin historias: los dos formularios avisan que falta una historia y no
  envían.
- Proyecto sin sprints: el alta de defecto y la resolución avisan que falta un
  sprint y no envían.
- Proyecto sin defectos: «Todavía no hay defectos».
- Defecto ya resuelto: la lista no ofrece la acción de resolver.
- Fallo al cargar el proyecto, los catálogos o los defectos: aviso y reintento,
  nunca se muestra como vacío.
- Proyecto cerrado: se puede seguir cargando esfuerzo y defectos, igual que la
  API de HU-016 y HU-017.

## Condiciones de error

- Alta de esfuerzo rechazada: alerta en español y el formulario conserva los datos.
- Alta de defecto rechazada: alerta en español y el formulario conserva los datos.
- Resolución rechazada: alerta en español y el defecto sigue abierto en la lista.

## Criterios de aceptación

- [ ] AC-030-01 — Registrar esfuerzo con integrante, fecha, actividad y horas
  sobre una historia del proyecto y ver la confirmación.
- [ ] AC-030-02 — Registrar un defecto con descripción, severidad, historia y
  sprint de detección, y verlo abierto en la lista.
- [ ] AC-030-03 — Resolver un defecto abierto indicando el sprint de resolución y
  ver el sprint en la lista.
- [ ] AC-030-04 — Mostrar en español el rechazo de Go en los tres formularios.
- [ ] AC-030-05 — Listar los sprints del proyecto, abiertos y cerrados, para
  elegir detección y resolución.
- [ ] AC-030-06 — Listar los defectos del proyecto, más recientes primero, con su
  estado y sprint de resolución.
- [ ] AC-030-07 — Mostrar estados vacíos y no enviar cuando falta integrante,
  historia o sprint; un fallo de carga no se muestra como vacío.
- [ ] AC-030-08 — React no calcula horas ni métricas; solo envía valores y
  muestra la respuesta del servidor.

## Trazabilidad

- BDD: `features/HU-030-ui-esfuerzo-defectos.feature`
- Backend nuevo: `backend/internal/sprints/` (listado completo) y
  `backend/internal/defects/` (listado por proyecto), con tests de servicio, HTTP
  y PostgreSQL.
- UI: `frontend/src/features/measurement/`, `features/effort/` y
  `features/defects/`; enlace en `ProjectDetailPage.tsx`; ruta en `App.tsx`.
- Tests: `frontend/src/features/measurement/Measurement.test.tsx`.
- Fila en `docs/traceability.md`.
