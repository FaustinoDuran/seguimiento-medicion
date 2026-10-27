# HU-005 — Crear historia de backlog

- Issue: [#14](https://github.com/FaustinoDuran/seguimiento-medicion/issues/14)
- Módulo Go: `backend/internal/backlog/`
- Estado: implementada en rama; pendiente de revisión y merge

## Objetivo

Permitir a un integrante crear una historia de backlog y consultarla individualmente mediante la API, con persistencia PostgreSQL.

## Entradas

Proyecto existente, identificador de historia ingresado por el integrante y título obligatorios.
Descripción, prioridad, story points y criterios de aceptación opcionales.
El estado inicial lo establece el sistema como «To do» (`todo`).

## Salidas esperadas

Historia persistida con un ID interno generado por el sistema, además del identificador de negocio.
Consulta individual de la historia creada.

## Reglas de negocio

- La historia pertenece a un proyecto existente, según los comentarios de la Issue.
- El identificador de negocio es obligatorio y único dentro del proyecto.
- El mismo identificador puede utilizarse en proyectos distintos.
- Sin título no se crea la historia, según la validación de la Issue.
- Toda historia nueva comienza en «To do» (`todo`); no se elige otro estado al crearla.
- Informar prioridad es opcional. Cuando se omite, la historia queda con prioridad «sin asignar» (`unassigned`). La escala definida en Go es `unassigned = 0`, `low = 1`, `medium = 2`, `high = 3`: mayor valor implica mayor prioridad. La API recibe y devuelve el nombre de la prioridad.
- Story points puede omitirse: significa «sin estimar». Cuando se informa debe ser un entero positivo; cero no es válido.
- La descripción puede estar vacía.
- Los criterios de aceptación son una lista de textos no vacíos; se admite una lista vacía.
- No se permite crear historias en proyectos `closed`; sí en proyectos `planned` o `active`.

## Restricciones

- Alcance confirmado: creación y consulta individual por API con PostgreSQL.
- La interfaz visual y el listado quedan fuera de esta historia.
- Reglas de negocio exclusivamente en Go.
- Registrar story points no incluye implementar Planning Poker.

## Casos límite

- Quitar espacios exteriores del identificador y convertirlo a mayúsculas antes de comprobar unicidad y persistirlo. No se exige el prefijo `HU-`.
- `hu-005`, `HU-005` y ` HU-005 ` identifican la misma historia dentro de un proyecto.
- Quitar espacios exteriores del título, descripción y de cada criterio de aceptación.
- Rechazar títulos y criterios que queden vacíos después de quitar espacios exteriores.
- La descripción y la lista de criterios pueden quedar vacías.

## Contrato de la API

- `POST /projects/{project_id}/stories`: crea una historia y devuelve `201` con sus datos.
- `GET /projects/{project_id}/stories/{id}`: consulta por ID interno y devuelve `200`.
- Campos de creación: `identifier`, `title`, `description`, `priority`, `story_points`, `acceptance_criteria`; `status` solo admite `todo` cuando se informa.
- La respuesta incluye `id`, `project_id`, `identifier`, `title`, `description`, `priority`, `status`, `story_points` y `acceptance_criteria`.
- Sin estimación se devuelve `story_points: null`; sin criterios se devuelve `acceptance_criteria: []`.
- Sin prioridad se devuelve `priority: "unassigned"`.
- Errores con formato `{"error":"..."}`.
- `400`: datos inválidos, prioridad desconocida o intento de crear con otro estado.
- `404`: proyecto o historia inexistentes, o historia que pertenece a otro proyecto.
- `409`: identificador duplicado dentro del proyecto o proyecto cerrado al crear.
- Consultar una historia existente sigue permitido aunque el proyecto esté cerrado.

## Condiciones de error

- Identificador ausente o duplicado dentro del proyecto.
- Título ausente.
- Proyecto inexistente.
- Proyecto cerrado.
- Story points informado que no sea un entero positivo.
- Criterios de aceptación con elementos vacíos.

Las validaciones corresponden al servicio Go; el handler decodifica la petición y traduce los errores al contrato HTTP.

## Criterios de aceptación

- [x] AC-005-01 — Crear y persistir una historia con sus campos válidos en un proyecto existente.
- [x] AC-005-02 — Consultar individualmente la historia creada.
- [x] AC-005-03 — Rechazar la creación sin título.
- [x] AC-005-04 — Rechazar identificadores ausentes o repetidos dentro del proyecto y permitir el mismo identificador en proyectos diferentes.
- [x] AC-005-05 — Establecer «To do» como estado inicial de toda historia nueva.
- [x] AC-005-06 — Permitir crear historias sin prioridad ni story points; rechazar story points informado que no sea un entero positivo.
- [x] AC-005-07 — Permitir descripción vacía y lista de criterios vacía; rechazar elementos vacíos dentro de la lista.
- [x] AC-005-08 — Rechazar nuevas historias en proyectos cerrados y permitirlas en proyectos planificados o activos.
- [x] AC-005-09 — Respetar el contrato HTTP y rechazar consultas de historias bajo un proyecto distinto.

## Trazabilidad

- BDD: `features/HU-005-crear-historia-backlog.feature`.
- Tests de servicio: `backend/internal/backlog/service_test.go` (AC-005-01 a AC-005-08).
- Tests HTTP: `backend/internal/backlog/handler_test.go` (AC-005-01 a AC-005-09).
- Tests PostgreSQL: `backend/internal/backlog/repository_postgres_test.go` (persistencia, unicidad concurrente e integración HTTP con base real).
- Código: `backend/internal/backlog/` y montaje en `backend/cmd/api/main.go`.
- Migración: `backend/migrations/003_create_backlog_stories.sql`.
- Verificación: suite completa con `go test -race ./...` y `BACKLOG_TEST_DATABASE_URL` configurada; pruebas PostgreSQL realizadas en esquemas aislados.
- Decisiones de alcance e identificador confirmadas por el usuario durante la entrevista `/grill-with-docs`.
