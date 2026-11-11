# HU-021 — Reporte de proyecto o Sprint

- Issue: [#30](https://github.com/FaustinoDuran/seguimiento-medicion/issues/30)
- Módulo Go: `backend/internal/reports/`
- Presentación: `frontend/src/features/reports/`
- Estado: lista para implementar; entrevista confirmada
- Depende de HU-019.

## Objetivo

Como responsable de seguimiento, generar un reporte con historias,
estimaciones, esfuerzo, métricas y defectos para revisar un proyecto o Sprint.

## Entradas

Proyecto o Sprint cuyo reporte se desea consultar. El reporte se abre desde
el detalle del proyecto y muestra inicialmente el alcance del proyecto.
Un selector permite elegir el proyecto completo o uno de sus Sprints,
abiertos o cerrados (Q7). El contrato confirmado se detalla más adelante.

## Salidas esperadas

- API en Go y pantalla para consultar el reporte por proyecto o Sprint (Q1).
- Datos actuales al momento de la consulta, incluso para Sprints cerrados (Q2).
- Todas las historias del proyecto, incluidas las no asignadas, o todas las
  historias asignadas al Sprint consultado (Q3).
- Cada historia aparece una vez con su estado, estimación acordada y
  asignación. Una historia completada conserva su asignación; la estimación
  ausente no se representa como cero puntos (Q3).
- En un proyecto, sus defectos. En un Sprint, los defectos detectados o
  resueltos allí, aunque su historia esté asignada a otro Sprint (Q4).

- Historias: identificador, título, estado, Sprint asignado y story points (Q6).
- Esfuerzo: cada registro con historia, integrante, fecha, actividad y horas,
  además del total de horas de las métricas (Q5).
- Defectos: historia afectada con identificador y título, descripción,
  severidad, estado y Sprints de detección y resolución (Q6).
- Secciones de métricas, historias, esfuerzo y defectos; cada listado vacío
  tiene su mensaje. Un reporte sin datos sigue siendo válido (Q7).

## Reglas de negocio

1. El reporte refleja datos vigentes y puede cambiar entre consultas: cerrar
   el Sprint no congela el informe. Por ejemplo, horas registradas después
   del cierre se reflejan en consultas posteriores.
2. Una historia completada aparece una sola vez, con su estado actual.
3. Los defectos del reporte de Sprint se seleccionan por la unión de sus
   Sprints de detección y resolución. Un defecto detectado y resuelto en el
   mismo Sprint aparece una sola vez.
4. Las métricas se reutilizan de HU-019, según ADR 0006; las fórmulas
   permanecen en Go.
5. El esfuerzo del Sprint comprende todos los registros de sus historias,
   independientemente de la fecha de actividad. No hay filtro por calendario
   ni se deduce un Sprint propio para el registro (Q5).
6. Cuando un defecto se refiere a una historia fuera del listado principal
   del Sprint, se muestra la referencia de esa historia sin agregarla al
   listado principal (Q6).
7. Si falla una sección, no se presenta un reporte incompleto como válido.
   La pantalla informa el error y permite reintentar; una falla de carga no
   equivale a un listado vacío (Q8).

## Restricciones

- La exportación PDF corresponde a HU-022 (ADR 0006).
- React presenta datos; no calcula métricas ni aplica reglas de alcance.
- La entrevista debe concluir con confirmación del entendimiento compartido
  antes de escribir tests o implementar.

## Casos límite

- Proyecto con historias sin Sprint asignado: se incluyen en el reporte.
- Historia sin estimación: se mantiene la ausencia de story points.
- Sprint cerrado con esfuerzo registrado después del cierre: el reporte
  posterior incluye los datos vigentes.
- Defecto detectado en Sprint 2 y resuelto en Sprint 3, sobre una historia
  asignada al Sprint 1: aparece en los reportes de Sprint 2 y Sprint 3.
- Defecto detectado y resuelto en el mismo Sprint: una sola aparición.
- Proyecto o Sprint sin historias, esfuerzo o defectos: reporte válido y
  mensajes por sección vacía.
- Esfuerzo anterior o posterior al Sprint: se incluye según la historia,
  sin filtrar por fecha.

## Condiciones de error

- Recurso inexistente: mensaje específico en la pantalla (Q8).
- Fallo de cualquier sección: error con reintento, sin mostrar un reporte
  parcial como válido (Q8).
- Códigos HTTP y tratamiento de identificadores: contrato confirmado abajo.

## Criterios de aceptación

- [ ] AC-021-01 — Consultar un reporte de proyecto por API y pantalla con
  todas sus historias, incluidas las no asignadas a un Sprint.
- [ ] AC-021-02 — Consultar un reporte de Sprint por API y pantalla con
  todas sus historias asignadas, sin duplicar las completadas.
- [ ] AC-021-03 — Mostrar estado, estimación acordada y asignación de las
  historias, conservando la ausencia de estimación.
- [ ] AC-021-04 — Reflejar los datos vigentes en cada consulta, incluso
  después del cierre del Sprint.
- [ ] AC-021-05 — Incluir los defectos detectados o resueltos en el Sprint,
  una sola vez por defecto, aunque su historia esté asignada a otro Sprint.

- [ ] AC-021-06 — Incluir todos los registros de esfuerzo del alcance con
  historia, integrante, fecha, actividad y horas, sin filtro por calendario,
  y mostrar el total de horas provisto por las métricas.
- [ ] AC-021-07 — Mostrar identificador y título de cada historia; mostrar
  descripción, severidad, estado e historia de cada defecto y sus Sprints
  de detección y resolución, sin ampliar el listado principal de historias.
- [ ] AC-021-08 — Mostrar las métricas de HU-019 sin recalcular fórmulas
  en React ni inventar horas estimadas.
- [ ] AC-021-09 — Acceder desde el detalle del proyecto, mostrar inicialmente
  el proyecto y permitir seleccionar cualquiera de sus Sprints abiertos o cerrados.
- [ ] AC-021-10 — Mostrar mensajes por sección vacía y conservar un reporte
  sin datos como consulta válida.
- [ ] AC-021-11 — Informar recursos inexistentes y fallos de consulta con
  reintento, sin presentar fallos como vacíos ni reportes parciales como válidos.

- [ ] AC-021-12 — Respetar rutas, campos, orden y códigos del contrato; rechazar un Sprint ajeno al proyecto sin exponer datos.
- [ ] AC-021-13 — Obtener detalle y métricas desde una misma lectura coherente.
- [ ] AC-021-14 — Retirar el reporte anterior durante la carga de otro alcance e ignorar respuestas obsoletas.

## Decisiones confirmadas

El usuario confirmó Q1 a Q9: API y pantalla, datos actuales, historias
completas del alcance, defectos detectados o resueltos en el Sprint,
detalle y alcance del esfuerzo, campos de historias y defectos, navegación,
vacíos, errores y corrección separada de métricas. No se requiere un ADR
nuevo para estas reglas de la HU.

## Hallazgo en la dependencia HU-019

`backend/internal/metrics/catalog_postgres.go`, en `loadDefects`, no carga
`resolution_sprint_id`. El servicio de métricas contempla ese dato, pero
la lectura PostgreSQL no lo entrega, afectando el conteo de defectos
resueltos por Sprint. Debe corregirse y verificarse antes de validar la
consistencia entre reporte y métricas. Se acordó en Q9 una issue BUG y un
PR separados, con test RED por API HTTP y PostgreSQL real antes de corregir.
Todavía no se modificó código. La creación de la issue fue rechazada por
GitHub con `403 Resource not accessible by integration`; `gh` no está
instalado. El cuerpo de la issue está preparado en
`/private/tmp/HU-021-metrics-bug.md`, fuera del repositorio.

## Contrato confirmado (Q10)

- `GET /projects/{project_id}/report`: reporte de proyecto.
- `GET /projects/{project_id}/sprints/{sprint_id}/report`: reporte de Sprint
  perteneciente al proyecto indicado; un Sprint de otro proyecto da `404`.
- `200`: informe completo. Solo lectura; no se persiste el reporte.
- `400`: identificadores no numéricos, no positivos o fuera de `int64`.
- `404`: proyecto o Sprint inexistente, o Sprint ajeno al proyecto.
- `500`: fallo de lectura de cualquier sección, sin detalles internos.
- Errores con formato `{"error":"..."}`.

Respuesta con estas secciones:

- `project`: `id`, `name`, `status`.
- `sprint`: `null` en proyecto; `id`, `goal`, `status` en Sprint.
- `available_sprints`: todos los Sprints del proyecto con `id`, `goal` y
  `status`, ordenados por creación ascendente y, ante empate, por ID.
  Permite poblar el selector sin añadir un listado general al módulo sprints.
- `metrics`: los campos y números del contrato de HU-019, con `project_id`
  y `sprint_id` del alcance solicitado. `estimated_hours` permanece en `0`.
- `stories`: `id`, `identifier`, `title`, `status`, `sprint_id` nullable y
  `story_points` nullable; orden ascendente de ID.
- `effort`: registros con `id`, `story_id`, `story_identifier`, `story_title`,
  `member_id`, `member_name`, `date` (`YYYY-MM-DD`), `activity` y `hours`;
  orden ascendente por fecha y, ante empate, por ID.
- `defects`: `id`, `story_id`, `story_identifier`, `story_title`,
  `description`, `severity`, `status`, `detection_sprint_id` y
  `resolution_sprint_id` nullable; orden ascendente de ID.

Los listados vacíos son `[]`. Las horas son números JSON con hasta dos
decimales, según HU-016. El informe contiene el alcance completo, sin
paginación ni filtros adicionales. Detalle y métricas corresponden a una
misma lectura coherente de los datos, incluso ante cambios concurrentes;
no se garantiza que dos consultas sucesivas coincidan.

La pantalla se ubica en `/projects/:projectId/report`. Al cambiar el alcance
se muestra carga y se retira el informe anterior; una respuesta obsoleta no
reemplaza la selección actual. Si falla la carga del reporte o del selector,
se informa error con reintento. La ausencia de Sprints deja disponible el
alcance proyecto. El encabezado identifica proyecto y, si corresponde,
Sprint con su objetivo y estado. Los estados se presentan en español.

## Interfaces de prueba confirmadas (Q11)

Se aplica la skill `tdd` en las interfaces confirmadas por el usuario:

- Servicio público Go: selección de historias, esfuerzo y defectos,
  deduplicación, ausencia de estimación, aislamiento y propagación de fallos
  sin HTTP ni PostgreSQL; no demuestra JSON ni SQL.
- API HTTP con memoria: rutas, formato JSON, identificadores y estados
  HTTP; no demuestra persistencia real.
- API HTTP con PostgreSQL real y esquema aislado: datos persistidos,
  correspondencia detalle/métricas, selección y aislamiento reales. Incluye
  en el BUG separado el test de regresión del Sprint de resolución; no
  demuestra interacción del navegador.
- Pantalla con Testing Library y MSW: navegación, selector, campos visibles,
  vacíos, carga, errores, reintento y respuestas obsoletas; no demuestra una
  conexión real entre React y Go.

El usuario confirmó Q10 y Q11 y el entendimiento compartido antes del código.
La entrevista queda cerrada; no hay decisiones de producto pendientes.

## Trazabilidad

- BDD: `features/HU-021-reporte-proyecto-sprint.feature`
  (decisiones confirmadas en la entrevista).
- Tests: pendientes; aún no se inició TDD.
- Código previsto: `backend/internal/reports/` y
  `frontend/src/features/reports/`.
