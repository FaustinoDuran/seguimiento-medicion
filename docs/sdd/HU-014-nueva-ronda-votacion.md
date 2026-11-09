# HU-014 — Nueva ronda de votación

- Issue: [#23](https://github.com/FaustinoDuran/seguimiento-medicion/issues/23)
- Módulo Go: `backend/internal/planningpoker/`
- Estado: implementada y mergeada en `main` (PR #80, merge `ca374a2`)

## Objetivo

Informar diferencias entre votos revelados y permitir otra ronda para discutir
la estimación. Decisiones confirmadas mediante `grill-with-docs`.

## Entradas

Se reutilizan las operaciones de HU-012 y HU-013: abrir con `project_id` y
`story_id`; revelar o consultar con `project_id`, `round_id` y `member_id`.
Abrir no exige `member_id` ni un rol de facilitador.

## Salidas esperadas

La respuesta de revelar y las consultas de una ronda cerrada informan si hay
diferencias. Las respuestas de rondas abiertas omiten ese indicador para
preservar el secreto. El campo HTTP es `has_differences`: booleano en rondas cerradas y ausente
en rondas abiertas (incluida la respuesta de apertura).
Abrir otra ronda devuelve `201`, otro ID, estado `open` y `votes: []`.

## Reglas de negocio

1. Hay diferencias cuando al menos dos votos emitidos tienen puntajes distintos.
2. Cero votos, uno solo o todos iguales dan ausencia de diferencias, sin afirmar
   que exista un acuerdo. No se exige que voten todos los integrantes.
3. Solo se informa la comparación después de revelar, tanto en esa respuesta
   como en consultas posteriores. El cálculo pertenece al servicio Go.
4. Se reutiliza `POST /projects/{project_id}/stories/{story_id}/rounds` para
   crear otra ronda. No se reabre ni modifica la ronda anterior.
5. La nueva ronda empieza sin votos. Conserva los votos y las rondas anteriores,
   y no modifica los story points actuales de la historia.
6. Solo puede existir una ronda abierta por historia, también ante concurrencia.
7. Repetir es voluntario: se permite con diferencias, coincidencia, un voto o
   ninguno una vez revelada la ronda anterior. No se crea automáticamente.
8. La ronda nueva vuelve a ocultar puntajes ajenos conforme a HU-012.
9. No se agregan restricciones sobre el estado de historia o Sprint a la apertura.
   Las condiciones para guardar un acuerdo siguen siendo las de HU-015.

## Restricciones

- Depende de HU-013. La UI pertenece a HU-029; sin cambios React ni autenticación.
- No se calcula una estimación acordada ni se exige repetir para guardarla.
- HU-015 solo permite registrar desde la última ronda revelada con votos: abrir
  otra invalida la anterior para registrar acuerdos, sin borrar el puntaje actual.
- El esquema existente admite varias rondas y garantiza una sola abierta por
  historia con un índice único parcial. No se necesita migración.

## Casos límite

- Votos 5, 5 y 8: hay diferencias; 5 y 5: no hay diferencias.
- Sin votos o con uno solo: no hay diferencias y se permite repetir.
- Cambiar un voto abierto: se compara su último valor al revelar.
- Nueva ronda: los votos anteriores siguen consultables, aunque se vote de nuevo.

## Condiciones de error

Se conservan los contratos existentes: IDs de ruta inválidos o consulta sin
integrante válido: `400`; historia, ronda o integrante inexistentes o ajenos
al proyecto: `404`; otra ronda abierta o voto en ronda cerrada: `409`;
fallo de infraestructura: `500` con error genérico.

## Criterios de aceptación

- [x] AC-014-01 — Revelar y consultar votos distintos informa diferencias.
- [x] AC-014-02 — Rondas abiertas no informan diferencias ni puntajes ajenos.
- [x] AC-014-03 — Votos iguales, un voto o ninguno informan ausencia de diferencias sin exigir otra ronda.
- [x] AC-014-04 — Se abre voluntariamente otra ronda vacía con nuevo ID conservando la anterior y su estimación.
- [x] AC-014-05 — Se puede repetir con diferencias, coincidencia, un voto o ninguno.
- [x] AC-014-06 — La nueva ronda oculta puntajes ajenos y la anterior sigue cerrada y consultable.
- [x] AC-014-07 — Se rechaza otra apertura mientras exista una ronda abierta y se conserva el aislamiento entre proyectos.

## Trazabilidad

- BDD: `features/HU-014-nueva-ronda-votacion.feature`.
- Tests: servicio, HTTP e integración PostgreSQL en `backend/internal/planningpoker/`.
- Código: `model.go`, `service.go`, `handler.go`; apertura y persistencia existentes.
- PR y evidencia TDD: `docs/traceability.md`.

## Verificación acordada

- Servicio público Go: comparación después de revelar, secreto antes de revelar,
  apertura voluntaria y conservación de la ronda anterior. Sin HTTP ni base.
- HTTP público: presencia y valor de `has_differences` al revelar y consultar;
  ausencia durante la votación y errores de apertura.
- API con PostgreSQL: rondas y votos independientes, estimación conservada,
  consulta posterior de los votos anteriores y unicidad concurrente.
- Sin promedio ni nuevas restricciones de estado. No hace falta un ADR: son
  reglas de esta HU, sin cambios de arquitectura, stack o alcance del MVP.

## Evidencia de implementación

- Servicio: `differences_service_test.go`, comparación al revelar/consultar,
  casos sin votos, un voto, coincidencia, diferencias, reemplazo y nuevas rondas.
- HTTP: `differences_handler_test.go`, presencia/ausencia del campo, booleano
  falso explícito, apertura voluntaria y contratos de error.
- PostgreSQL: `differences_postgres_test.go`, conservación de historia y votos,
  aislamiento entre rondas, secreto y unicidad ante ocho aperturas concurrentes.
- Verificado con `go test -race -count=1 ./...` y las tres variables de conexión
  contra PostgreSQL local; `go vet ./...` y `git diff --check`.
- SDD y BDD precedieron al código; dos ciclos RED/GREEN separados. La apertura
  reutiliza comportamiento existente y sus pruebas de regresión pasan sin cambios SQL.
