# HU-026 — Gestionar backlog desde la UI

- Issue: [#48](https://github.com/FaustinoDuran/seguimiento-medicion/issues/48)
- Módulo Go existente: `backend/internal/backlog/`
- Presentación prevista: `frontend/src/features/backlog/`
- Estado: diseño confirmado; implementación pendiente de resolución de dependencias

## Objetivo

Como integrante, crear y ver historias de backlog desde la interfaz para
organizar el trabajo del proyecto sin curl.

## Entradas

- Proyecto al que pertenecen las historias.
- Identificador de historia y título obligatorios.
- Descripción, prioridad, story points y criterios de aceptación opcionales.

## Salidas esperadas

- Listado de todas las historias del proyecto, incluidas las asignadas a un
  Sprint y las completadas.
- Listado completo ordenado por `id DESC` desde el backend; React conserva
  ese orden. Esta versión no tiene búsqueda, filtros ni paginación.
- Cada historia muestra identificador, título, prioridad, estado y estimación.
- Formulario para crear historias con los campos indicados.
- Campos opcionales agrupados en «Más detalles»: prioridad mediante selector,
  story points que puede quedar vacío, y criterios de aceptación que pueden
  agregarse o quitarse individualmente, comenzando con una lista vacía.
- Ruta `/projects/:projectId/backlog`, accesible desde el detalle del proyecto,
  con enlace para volver y acceso directo por URL.
- Formulario sobre el listado. Durante el envío se deshabilita el botón.
- Al crear correctamente, confirmación de alta, formulario limpio y listado
  actualizado. Si falla solamente la actualización del listado, se conserva
  la confirmación de alta y se ofrece reintentar la carga sin reenviar el alta.

## Reglas de negocio

Se reutilizan las reglas de HU-005 y el vocabulario de `GLOSSARY.md`.
Go valida los datos y asigna el estado inicial `todo`; React presenta los
resultados y errores de la API.

## Restricciones

- Completar la especificación ahora; resolver las dependencias antes de
  implementar la pantalla.
- Depende de HU-024 ([#46](https://github.com/FaustinoDuran/seguimiento-medicion/issues/46)).
- Falta una HU de backend para listar historias por proyecto, con contrato y
  criterios propios, según ADR 0004. No implementar ese comportamiento dentro
  de esta HU de presentación.
- Creación disponible: `POST /projects/{project_id}/stories`.
- Consulta individual disponible: `GET /projects/{project_id}/stories/{id}`.
- El endpoint de listado está pendiente de definición en la HU de backend.
  Debe devolver todas las historias del proyecto ordenadas por `id DESC`.
  Las historias actuales no tienen `created_at`; no se incorpora ese campo
  para esta pantalla. El ID sirve como orden de altas, no como fecha real de
  creación ni garantía del orden de confirmación de transacciones concurrentes.

## Casos límite

- Proyecto sin historias: mostrar «Todavía no hay historias» únicamente tras
  una consulta exitosa sin resultados.
- Historia sin estimación: mostrar «Sin estimar».
- Proyecto cerrado: permitir consultar historias; reemplazar el formulario
  por un aviso de que no se pueden crear nuevas.
- Proyecto cerrado durante la edición: mostrar el rechazo de Go y conservar
  los datos ingresados.

## Condiciones de error

- Mostrar mensajes en español; traducir los errores conocidos de la API.
- Conservar los datos del formulario cuando falla la creación, incluidos
  campos inválidos, identificador duplicado y cierre concurrente del proyecto.
- Mostrar carga visible y ofrecer reintento ante fallos de consulta. Un error
  de conexión o servidor no se presenta como backlog vacío.
- No confirmar un alta cuya respuesta fue fallida.
- Mostrar errores de campos junto al campo correspondiente y los demás en
  un aviso general. Los errores desconocidos tienen un mensaje general en español.
- Si falla la conexión durante el alta, indicar que no se pudo confirmar el
  resultado y sugerir revisar el listado antes de reenviar.
- Proyecto inexistente: mostrar «Proyecto no encontrado», sin formulario y
  con enlace al listado de proyectos. Diferenciarlo de un fallo temporal que
  permite reintentar.

## Criterios de aceptación

- [ ] AC-026-01 — Crear una historia con identificador y título desde la UI,
  mediante la API existente.
- [ ] AC-026-02 — Permitir ingresar descripción, prioridad, story points y
  criterios de aceptación opcionales.
- [ ] AC-026-03 — Ver todas las historias del proyecto con identificador,
  título, prioridad, estado y estimación, incluidas las asignadas y completadas.

- [ ] AC-026-04 — Acceder al backlog desde el detalle del proyecto o por URL
  directa, y disponer de un enlace para volver al detalle.
- [ ] AC-026-05 — Mostrar el formulario sobre el listado, deshabilitar el envío
  mientras está en curso y, tras el éxito, confirmar, limpiar y actualizar.
- [ ] AC-026-06 — Si el alta fue exitosa pero falla recargar el listado,
  informar ambos resultados y permitir reintentar solo la consulta.
- [ ] AC-026-07 — Presentar errores conocidos en español y conservar los datos
  ingresados si falla el alta, incluido el caso de título faltante.
- [ ] AC-026-08 — Mostrar carga y reintento ante errores de consulta; mostrar
  «Todavía no hay historias» solo ante un listado exitoso vacío.
- [ ] AC-026-09 — Mostrar «Sin estimar» cuando no hay story points.
- [ ] AC-026-10 — Permitir consultar proyectos cerrados sin formulario de alta;
  mostrar un aviso explicativo. Si el cierre ocurre durante la edición,
  mostrar el rechazo de la API y conservar los datos ingresados.
- [ ] AC-026-11 — Agrupar los opcionales en «Más detalles» y permitir agregar
  y quitar criterios de aceptación individualmente.
- [ ] AC-026-12 — Mostrar los errores por campo junto a su entrada, los demás
  en un aviso general y un mensaje español para errores desconocidos. Ante un
  fallo de conexión durante el alta, advertir que no se pudo confirmar el
  resultado y sugerir revisar el listado antes de reenviar.
- [ ] AC-026-13 — Ante un proyecto inexistente, mostrar «Proyecto no encontrado»,
  sin formulario y con enlace al listado de proyectos.
- [ ] AC-026-14 — Mostrar el listado completo en el orden `id DESC` entregado
  por el backend, sin búsqueda, filtros ni paginación.

## Cierre de diseño y dependencias

- Decisiones de la entrevista acordadas: navegación por proyecto, formulario
  mínimo y opcionales, listado con historias asignadas y completadas, flujo de
  alta y recarga, carga/vacío/errores, proyectos cerrados o inexistentes y orden.
- Entrevista cerrada con confirmación final del usuario el 2026-11-02.
- Pendiente HU de backend para listar historias, con Issue, SDD, BDD y TDD
  propios; el contrato debe satisfacer el orden y alcance acordados aquí.
- Pendiente resolver HU-024 antes de implementar la pantalla.

El diseño está confirmado. La implementación espera la resolución de las
dependencias. No se requieren nuevas decisiones de
arquitectura: se mantiene ADR 0004.

## Trazabilidad

- BDD: `features/HU-026-ui-backlog.feature` (diseño confirmado).
- Código previsto: `frontend/src/features/backlog/` (pendiente).
- Reglas existentes: `backend/internal/backlog/service.go` y sus tests.
