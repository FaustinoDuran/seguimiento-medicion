# HU-009 — Completar historia

- Issue: [#18](https://github.com/FaustinoDuran/seguimiento-medicion/issues/18)
- Módulo Go: `backend/internal/sprints/`
- Estado: lista para implementar

## Objetivo

Permitir a un integrante marcar como completada una historia asignada a un
Sprint abierto y consultar su avance real por API, con persistencia PostgreSQL.

## Entradas

ID interno del Sprint e ID interno de la historia asignada. La operación no
requiere cuerpo JSON ni permite elegir un estado arbitrario.

## Salidas esperadas

Historia con estado `done`, conservando sus datos y su asignación. El listado
de historias del Sprint y la consulta individual del backlog muestran el mismo
estado. Las historias nuevas y las asignadas sin completar siguen en `todo`.

## Reglas de negocio

1. El Sprint debe existir y estar `open`.
2. La historia debe existir dentro del proyecto del Sprint.
3. La historia debe estar asignada al Sprint indicado. No basta con que esté
   asignada a otro Sprint.
4. Completar cambia el único estado de la historia de `todo` a `done`.
5. Repetir la solicitud para una historia `done` devuelve éxito sin cambiar sus
   datos ni su asignación, siempre que el Sprint siga abierto.
6. No se exige estimación ni criterios de aceptación registrados para completar.
7. Si la operación falla, se conserva el estado y la asignación anteriores.

## Restricciones

- Reglas en el servicio Go de `sprints`; handler sin reglas y sin UI en esta HU.
- La escritura de la historia usa un puerto público del módulo `backlog`, sin
  importar handlers ni SQL de otro módulo.
- El estado no se duplica en la asignación al Sprint.
- La migración debe admitir `todo` y `done` en bases nuevas y existentes,
  conservando los datos y el estado inicial `todo`.
- Reabrir (`done` a `todo`) queda para HU-006 ([#15](https://github.com/FaustinoDuran/seguimiento-medicion/issues/15)),
  cuyo contrato deberá respetar que no se cambian historias de Sprints cerrados.
- Cerrar un Sprint corresponde a HU-010. No se agrega ese comportamiento aquí.
- Alcance y reglas confirmados con el usuario mediante `grill-with-docs`.

## Casos límite

- Historia sin story points y sin criterios: se permite completar.
- Historia ya completada: éxito si sigue asignada al Sprint abierto indicado.
- Sprint cerrado: se rechaza incluso si la historia ya está completada.
- Historia asignada a otro Sprint del mismo proyecto: se rechaza.
- Completar una historia no cambia las otras historias del Sprint.

## Contrato de la API

- `POST /sprints/{id}/stories/{story_id}/complete`, sin cuerpo: devuelve `200`
  con `sprint_id`, `story_id`, `identifier`, `title` y `status: "done"`.
- `GET /sprints/{id}/stories` agrega `status` a cada historia, tanto `todo` como
  `done`. La respuesta de asignación también incluye el estado actual.
- `GET /projects/{project_id}/stories/{id}` conserva su contrato y refleja `done`.
- Errores con formato `{"error":"..."}`.
- `400`: ID de Sprint o historia inválido, no entero positivo o fuera de rango.
- `404`: Sprint inexistente, historia inexistente o historia de otro proyecto.
- `409`: Sprint cerrado, historia sin asignación o asignada a otro Sprint.
- `500`: error de infraestructura; no se expone su detalle al cliente.
- Si coinciden errores, se valida primero la existencia y apertura del Sprint,
  luego la historia en el proyecto y finalmente su asignación.

## Condiciones de error

Sprint inexistente o cerrado, historia inexistente o ajena al proyecto,
historia sin asignación al Sprint indicado o fallo de persistencia.

## Criterios de aceptación

- [ ] AC-009-01 — Completar y persistir una historia `todo` de un Sprint abierto,
  sin exigir story points ni criterios, conservando sus datos y su asignación.
- [ ] AC-009-02 — Consultar el mismo estado en el backlog y el listado del Sprint,
  sin alterar otras historias.
- [ ] AC-009-03 — Rechazar historias sin asignación o asignadas a otro Sprint.
- [ ] AC-009-04 — Rechazar un Sprint cerrado, incluso en solicitudes repetidas.
- [ ] AC-009-05 — Rechazar Sprint o historia inexistentes e historias ajenas al proyecto.
- [ ] AC-009-06 — Repetir la finalización en un Sprint abierto devuelve éxito y mantiene `done`.
- [ ] AC-009-07 — Respetar el contrato HTTP, validar IDs y conservar el estado ante errores.

## Trazabilidad

- BDD: `features/HU-009-completar-historia.feature`.
- Interfaces de prueba: servicio público Go (reglas), HTTP (contrato) y
  repositorios PostgreSQL (persistencia y actualización de bases existentes).
- Tests: `backend/internal/sprints/{service,handler,repository_postgres}_test.go`.
- Código: `backend/internal/sprints/`, puerto y persistencia de `backend/internal/backlog/`.
- Migración: `backend/migrations/006_allow_completed_stories.sql`.
