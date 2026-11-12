# HU-022 — Exportar informe a PDF

- Issue: [#31](https://github.com/FaustinoDuran/seguimiento-medicion/issues/31)
- Módulo Go: `backend/internal/reports/`
- Presentación: `frontend/src/features/reports/`
- Estado: mergeada en `main` ([PR #97](https://github.com/FaustinoDuran/seguimiento-medicion/pull/97))
- Depende de HU-021, mergeada en `main` por el PR #96.

## Objetivo

Como responsable de seguimiento, exportar a PDF el informe del proyecto o del
Sprint que ya está consultando, para entregárselo al Cliente.

## Entradas

El mismo alcance de HU-021: el proyecto completo o uno de sus Sprints, abiertos
o cerrados. En la pantalla, el alcance es el que está seleccionado cuando el
reporte ya se mostró.

## Salidas esperadas

- Un archivo PDF generado en Go a partir del reporte ya armado. React no
  compone el archivo ni recalcula métricas.
- El archivo se obtiene por API y se descarga desde la pantalla del reporte.
- El contenido corresponde a ese reporte: encabezado, métricas, historias,
  esfuerzo y defectos, con los mismos valores y los mismos rótulos en español
  que muestra la pantalla.
- Un reporte sin historias, esfuerzo o defectos sigue siendo un PDF válido.

## Reglas de negocio

1. El PDF renderiza el `Report` que ya devuelven `ForProject` y `ForSprint`.
   No vuelve a seleccionar historias, esfuerzo ni defectos, y no ejecuta las
   fórmulas de HU-019.
2. El alcance del archivo es el de esa consulta. El PDF de un Sprint no
   incluye datos de otro alcance. El listado `available_sprints` no forma una
   sección del documento: alimenta el selector de la pantalla.
3. Los valores ausentes se conservan como en la pantalla: una historia sin
   story points dice «Sin estimar» y no aparece como cero. Un defecto sin
   sprint de resolución dice «Sin resolución».
4. Cada sección vacía incluye su mensaje: «No hay historias del alcance.»,
   «No hay registros de esfuerzo.» y «No hay defectos del alcance.».
5. Si el reporte no se puede obtener, no se entrega un PDF. Un fallo al
   renderizar tampoco publica un archivo parcial.

## Restricciones

- Las rutas JSON de HU-021 quedan igual.
- Los rótulos siguen a la pantalla: estado del proyecto («Planificado»,
  «Activo», «Cerrado»); Sprint «Abierto» o «Cerrado»; historia «Por hacer» o
  «Completada»; severidad «Baja», «Media», «Alta», «Crítica»; defecto
  «Abierto» o «Resuelto».
- La librería que produce el PDF es `github.com/go-pdf/fpdf` v0.9.0
  ([ADR 0007](../adr/0007-generacion-pdf.md)).
- El documento tiene que poder leerse como texto en los tests, para comprobar
  que coincide con el reporte.

## Casos límite

- Proyecto con historias sin Sprint, sin estimación, con esfuerzo y defectos:
  el PDF incluye esos datos del reporte de proyecto.
- Sprint con una historia sin estimar y un defecto resuelto en otro Sprint:
  el PDF muestra «Sin estimar» y el sprint de resolución, sin agregar esa
  historia al listado si el reporte de HU-021 no la agrega.
- Alcance sin historias, sin esfuerzo y sin defectos: PDF válido con los tres
  mensajes de sección vacía y las métricas de ese reporte.
- Identificador inválido, proyecto inexistente o Sprint de otro proyecto: no
  hay archivo PDF.
- La pantalla está cargando, muestra error o el recurso no existe: no ofrece
  la descarga.

## Condiciones de error

- `400`: identificador no numérico, no positivo o fuera de `int64`. Cuerpo
  JSON `{"error":"project and sprint ids must be positive"}`.
- `404`: proyecto o Sprint inexistente, o Sprint ajeno al proyecto. Cuerpo
  JSON con `project not found` o `sprint not found`.
- `500`: fallo de lectura del reporte o fallo al renderizar el PDF. Cuerpo
  JSON `{"error":"internal error"}`, sin cuerpo PDF.
- Esos errores usan `Content-Type: application/json`.

## Criterios de aceptación

- [x] AC-022-01 — Exportar el reporte de proyecto a un archivo PDF con el
  nombre del proyecto, su estado, las métricas y las secciones de historias,
  esfuerzo y defectos de ese reporte.
- [x] AC-022-02 — Exportar el reporte del Sprint seleccionado, con su objetivo
  y estado, limitado al contenido que HU-021 ya devuelve para ese alcance.
- [x] AC-022-03 — Conservar la ausencia de estimación y de resolución con los
  textos «Sin estimar» y «Sin resolución».
- [x] AC-022-04 — Entregar un PDF válido cuando una sección no tiene filas,
  incluyendo el mensaje de sección vacía.
- [x] AC-022-05 — Rechazar identificadores inválidos y recursos inexistentes
  sin devolver un PDF, y responder `500` si la renderización falla, sin un
  archivo parcial.
- [x] AC-022-06 — Mantener sin cambios las respuestas JSON de
  `GET /projects/{project_id}/report` y
  `GET /projects/{project_id}/sprints/{sprint_id}/report`.
- [x] AC-022-07 — Mostrar «Exportar PDF» solo cuando el reporte está en
  pantalla, y descargar el archivo del alcance seleccionado. Ocultar la
  acción durante la carga, el error y el recurso inexistente.

## Decisiones confirmadas

El usuario confirmó el 2026-11-11 que Go genera el PDF del mismo alcance que
muestra la pantalla, reutilizando el reporte de HU-021 sin recalcular
métricas, y que la pantalla agrega la descarga. No se requiere otro ADR para
esta regla. El ADR de una dependencia de PDF, si hace falta, queda aparte.

## Contrato confirmado

- `GET /projects/{project_id}/report.pdf`: PDF del proyecto.
- `GET /projects/{project_id}/sprints/{sprint_id}/report.pdf`: PDF del Sprint
  perteneciente a ese proyecto.
- `200`: cuerpo PDF. `Content-Type: application/pdf`. El archivo empieza con
  `%PDF`. `Content-Disposition: attachment` con el nombre
  `reporte-proyecto-{project_id}.pdf` o
  `reporte-proyecto-{project_id}-sprint-{sprint_id}.pdf`.
- No se persiste el archivo. Cada exportación vuelve a leer el reporte
  vigente, igual que una consulta de HU-021.
- El texto del PDF incluye, cuando el dato existe en el reporte: nombre y
  estado del proyecto; objetivo y estado del Sprint si el alcance es un
  Sprint; los nueve rótulos de métricas y sus valores; identificador, título,
  estado, asignación y estimación de cada historia; integrante, fecha,
  actividad y horas de cada registro de esfuerzo; descripción, severidad,
  estado e historia de cada defecto, con sus sprints de detección y
  resolución.

## Interfaces de prueba

- Renderizado en Go, sin HTTP ni PostgreSQL: un `Report` de memoria produce
  un PDF cuyo texto contiene los valores y los mensajes de sección vacía.
  También cubre el fallo de renderizado, sin demostrar rutas ni SQL.
- API HTTP con memoria: rutas nuevas, `Content-Type`, nombre de archivo,
  prefijo `%PDF`, códigos `400`, `404` y `500`, y que las rutas JSON siguen
  respondiendo el reporte. No demuestra persistencia real.
- API HTTP con PostgreSQL real y esquema aislado: un dato persistido aparece
  en el texto del PDF del alcance correcto y no en el del otro alcance.
- Pantalla con Testing Library: el botón, la URL del alcance seleccionado, y
  su ausencia en carga, error y recurso inexistente. No demuestra la conexión
  real entre React y Go ni el aspecto visual del PDF.

## Trazabilidad

- BDD: `features/HU-022-exportar-informe-pdf.feature`
- Render, HTTP y PostgreSQL: `backend/internal/reports/pdf.go`,
  `pdf_test.go`, `pdf_handler_test.go` y `repository_postgres_test.go`.
- Pantalla: `frontend/src/features/reports/ReportPage.tsx` y
  `Report.test.tsx`.
