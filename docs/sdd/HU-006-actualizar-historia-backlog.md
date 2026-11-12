# HU-006 — Actualizar historia de backlog

- Issue: [#15](https://github.com/FaustinoDuran/seguimiento-medicion/issues/15)
- Módulo Go: `backend/internal/backlog/`
- Estado: lista para implementar; entendimiento final confirmado por el usuario

## Objetivo

Permitir a un integrante corregir prioridad, estado o story points de una
historia para reflejar lo acordado en la planificación, desde la UI del backlog.

## Entradas

Historia del proyecto y un subconjunto no vacío de `priority`, `status` y
`story_points`. Los campos omitidos o con valor JSON `null` conservan su valor.
Debe existir al menos un campo editable con valor no nulo. Los IDs y los puntos
son enteros positivos dentro del rango de `int64`.

## Salidas esperadas

Los cambios válidos se guardan y pueden consultarse posteriormente.
Se informa si la historia no existe.

## Reglas de negocio

Decisiones confirmadas durante las tres rondas de la entrevista:

1. Se permiten las transiciones `todo → done` y `done → todo`, ambas dentro
   de un Sprint abierto. Completar conserva las condiciones de HU-009.
2. Reabrir conserva la asignación de la historia al Sprint.
3. La corrección manual de story points exige una estimación existente.
   HU-006 no permite registrar la primera estimación; el acuerdo se registra
   mediante Planning Poker (HU-015).
4. El nuevo puntaje debe ser un entero positivo. No se permite borrar la
   estimación ni reemplazarla por cero.
5. Una historia completada permite corregir su prioridad. Para cambiar sus
   story points debe reabrirse primero.
6. Las historias asignadas a Sprints cerrados rechazan todos estos cambios,
   conforme a la restricción de HU-009.
7. Un proyecto cerrado rechaza toda edición de sus historias, aunque no
   pertenezcan a un Sprint cerrado. La consulta sigue permitida.
8. Solo se editan prioridad, estado y story points. Las prioridades admitidas
   son `unassigned`, `low`, `medium` y `high`; se permite volver a `unassigned`.
9. La actualización es parcial y atómica: los campos omitidos se conservan y
   cualquier error impide guardar todos los cambios de la solicitud.
10. Reabrir y corregir puntos exige dos operaciones: primero se guarda la
    reapertura. Si la historia está completada al recibir una solicitud con
    story points, se rechaza aunque también solicite pasar a `todo`.
11. Repetir valores actuales devuelve éxito cuando la operación sigue siendo
    válida. Una solicitud sin ningún campo editable se rechaza. Los bloqueos
    por proyecto o Sprint cerrado se verifican también en reintentos.
12. Cualquier estimación existente es corregible, incluida una ingresada al
    crear la historia. No se exige evidencia de una ronda previa.
13. Solo se guarda el puntaje vigente, sin historial ni autor de la corrección.
    Se conservan votos y rondas; métricas y reportes reflejan el valor vigente.
14. Repetir `todo` en una historia pendiente sin Sprint es válido. Solo una
    transición efectiva exige pertenecer a un Sprint abierto.
15. Una historia inicialmente pendiente, estimada y asignada a un Sprint abierto
    permite corregir puntos y completar en la misma solicitud, de forma atómica.
16. Una ronda abierta o posterior no bloquea la corrección. Un acuerdo posterior
    válido de HU-015 puede volver a reemplazar el puntaje corregido.
17. Informar puntos no nulos aplica las condiciones de corrección incluso si
    coinciden con el valor actual: una historia inicialmente completada los
    rechaza. Omitirlos o enviarlos como `null` conserva su valor.

## Contrato de la API

`PATCH /projects/{project_id}/stories/{id}` con un único objeto JSON parcial:

```json
{"priority":"high","story_points":8}
```

- `200`: historia completa con el mismo formato que su consulta individual.
- `400`: IDs inválidos; JSON inválido, no objeto o seguido por otro valor;
  campos desconocidos o no editables; cuerpo vacío o sin campos editables no
  nulos; prioridad o estado vacíos o desconocidos; puntos no enteros, no positivos
  o fuera de rango. No se aplican los valores predeterminados del alta.
- `404`: proyecto inexistente, historia inexistente o historia ajena al proyecto.
- `409`: proyecto o Sprint cerrado, transición sin Sprint abierto, primera
  estimación o puntos informados para una historia inicialmente completada.
- `500`: fallo de infraestructura sin exponer detalles internos.

Los errores usan `{"error":"..."}`. Se validan primero formato y valores,
luego existencia del proyecto y de la historia dentro de él, cierre del proyecto,
cierre del Sprint y condiciones de transición y corrección. Cualquier fallo
conserva todos los datos anteriores. Campos no editables como título, identificador,
descripción, criterios y asignación al Sprint no se modifican.

## Comportamiento de la UI

- Acción «Editar» por historia, con formulario precargado de prioridad, estado y
  puntos, y botones «Guardar» y «Cancelar». Se conserva el diseño del backlog.
- Solo se envían campos modificados. Sin cambios, no se envía una solicitud vacía.
- Tras el éxito se actualiza el listado. Cancelar descarta la edición.
- Ante un error se conserva lo ingresado y se muestra el motivo en español;
  una historia inexistente tiene un aviso específico.
- Un proyecto cerrado permite solo consultar. Para historias sin estimar o
  completadas se explica por qué no se pueden editar los puntos. Reabrir debe
  guardarse antes de habilitar su corrección.
- Go valida las reglas y persiste los cambios. React no calcula métricas.

## Restricciones

- Reglas de negocio únicamente en Go; React presenta y envía los cambios.
- Alcance del Sprint 5 conforme a ADR 0008; no reemplaza Planning Poker.
- Entrevista y entendimiento compartido confirmados por el usuario antes de implementar.

## Casos límite

- Historia sin estimar: no admite una primera estimación desde esta operación.
- Historia completada: admite corregir prioridad, salvo en un Sprint cerrado.
- Historia asignada a un Sprint cerrado: no admite cambios.
- Proyecto cerrado: no admite cambios en sus historias.
- Reapertura y corrección de puntos en la misma solicitud: se rechaza todo.
- Prioridad válida junto con puntaje inválido: no se guarda ninguno.
- Estimación ingresada en el alta: admite corrección si se cumplen las demás reglas.
- `null` conserva el valor; una petición con solo valores nulos se rechaza.
- Repetir `todo` sin Sprint es válido; pasar a `done` sin Sprint se rechaza.
- Corregir puntos y completar una historia inicialmente pendiente es válido.
- Informar los mismos puntos en una historia completada se rechaza.
- Corregir con una ronda abierta conserva la ronda y sus votos.

## Condiciones de error

- Historia inexistente.
- Intento de primera estimación, eliminación de estimación o puntaje no positivo.
- Cambio de estado sin Sprint abierto.
- Corrección de story points en una historia completada.
- Modificación de una historia asignada a un Sprint cerrado.
- Modificación de una historia de un proyecto cerrado.
- Solicitud sin campos editables.

## Criterios de aceptación

- [ ] AC-006-01 — Persistir cambios válidos y reflejarlos en consultas posteriores.
- [ ] AC-006-02 — Informar que la historia no existe.
- [ ] AC-006-03 — Completar y reabrir historias dentro de un Sprint abierto,
  conservando su asignación y las condiciones de finalización de HU-009.
- [ ] AC-006-04 — Corregir únicamente estimaciones existentes mediante enteros
  positivos, sin permitir la primera estimación ni su eliminación.
- [ ] AC-006-05 — Permitir corregir prioridad en historias completadas y exigir
  su reapertura antes de corregir story points.
- [ ] AC-006-06 — Rechazar todo cambio en historias de Sprints cerrados.
- [ ] AC-006-07 — Rechazar toda edición en proyectos cerrados y mantener la consulta.
- [ ] AC-006-08 — Actualizar únicamente los campos editables presentes de forma
  atómica; admitir las cuatro prioridades, incluida `unassigned`.
- [ ] AC-006-09 — Rechazar la corrección de puntos en historias inicialmente
  completadas, aunque se solicite su reapertura en el mismo guardado.
- [ ] AC-006-10 — Admitir reintentos válidos y rechazar solicitudes sin campos editables.
- [ ] AC-006-11 — Admitir corregir cualquier estimación existente, sin exigir
  procedencia de Planning Poker.
- [ ] AC-006-12 — Conservar votos y rondas, guardar solo el puntaje vigente y
  reflejarlo en métricas y reportes, sin agregar historial ni autor.
- [ ] AC-006-13 — Respetar el contrato HTTP, valores nulos, campos desconocidos,
  códigos y precedencia de errores, sin modificar datos ante fallos.
- [ ] AC-006-14 — Admitir repetir `todo` sin Sprint y rechazar transiciones
  efectivas sin un Sprint abierto.
- [ ] AC-006-15 — Corregir puntos y completar una historia inicialmente pendiente
  en un único guardado válido.
- [ ] AC-006-16 — Permitir corregir con rondas abiertas o posteriores y registrar
  después un acuerdo válido de HU-015 que reemplace el puntaje.
- [ ] AC-006-17 — Editar desde el backlog, guardar solo campos modificados,
  cancelar y conservar lo ingresado ante errores, mostrando el resultado.
- [ ] AC-006-18 — Presentar el proyecto cerrado como consulta y explicar las
  restricciones de puntos en historias sin estimar o completadas.

## Trazabilidad

- BDD: `features/HU-006-actualizar-historia-backlog.feature`.
- Tests previstos: servicio, HTTP y persistencia PostgreSQL en
  `backend/internal/backlog/`; interacción de UI en `frontend/src/features/backlog/`.
- Código previsto: `backend/internal/backlog/`, integración mediante puertos
  públicos de `sprints` y registro en `backend/cmd/api/main.go`.
- UI prevista: `frontend/src/features/backlog/{BacklogPage,StoryList}.tsx`,
  formulario `StoryEditForm.tsx` y cliente `api.ts`.
- Matriz: `docs/traceability.md`. Tests, commits TDD y PR todavía pendientes.

## Plan de implementación tras la confirmación final

1. Rama `feat/HU-006-update-backlog-story`; preservar los cambios ajenos existentes.
2. TDD por reglas: tests RED de servicio, implementación GREEN y refactor cuando
   aporte claridad, con commits separados que referencien `#15`.
3. Persistencia parcial y atómica: conservar campos omitidos incluso ante
   escrituras concurrentes y comprobar los bloqueos al guardar. Usar puertos
   públicos entre módulos, sin importar handlers ni SQL ajeno.
4. TDD del contrato HTTP y pruebas PostgreSQL de persistencia, aislamiento entre
   proyectos y conservación de datos ante errores; verificar la interacción con
   completar, reapertura, Planning Poker, métricas y reportes existentes.
5. Formulario de edición en el backlog y pruebas del flujo observable de UI.
6. Verificar Go con PostgreSQL local y frontend con tests, lint y build;
   completar trazabilidad y abrir PR contra `main` con `Closes #15`.

Commits previstos: `docs(#15): especificar actualización de historias`,
`test(#15): red - ...`, `feat(#15): green - ...` y `refactor(#15): ...` cuando
haya cambios de diseño. Sin nuevas métricas, autenticación, GORM ni cambio de stack.
