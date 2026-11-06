# HU-015 — Registrar estimación acordada

- Issue: [#24](https://github.com/FaustinoDuran/seguimiento-medicion/issues/24)
- Módulo Go: `backend/internal/planningpoker/`
- Estado: implementada en rama; pendiente revisión y merge

## Objetivo

Guardar el story point acordado en la historia para consultarlo después y
utilizarlo en planificación y métricas.

## Entradas

Proyecto y ronda en la ruta; `member_id` y `story_points` en el cuerpo JSON.
Todos son enteros positivos dentro del rango de `int64`.

## Salidas esperadas

La historia conserva la estimación acordada y permite consultarla después.

## Contrato de la API

`POST /projects/{project_id}/rounds/{round_id}/estimate`

```json
{"member_id": 7, "story_points": 8}
```

Devuelve `200` con `project_id`, `story_id`, `round_id` y `story_points`.
La historia se obtiene de la ronda, sin un identificador de historia adicional.
`GET /projects/{project_id}/stories/{id}` y el listado del backlog reflejan el
valor persistido, incluso en consultas posteriores independientes del registro.

- `400`: IDs o puntaje ausentes, inválidos, no positivos o fuera de rango;
  cuerpo JSON inválido o puntaje fraccionario.
- `404`: ronda, historia o integrante inexistentes o ajenos al proyecto indicado.
- `409`: ronda abierta, sin votos o anterior a otra de la historia; historia
  completada o perteneciente a un Sprint cerrado.
- `500`: fallo de infraestructura, sin exponer detalles internos.

Errores con formato `{"error":"..."}`. Se validan primero formato y rangos,
luego ronda en el proyecto, integrante, historia y condiciones del registro.
Las condiciones se comprueban también en reintentos. Un fallo conserva el
puntaje anterior y no modifica votos, rondas ni otros datos de la historia.

## Reglas de negocio

Decisiones confirmadas durante la entrevista `grill-with-docs`:

1. El puntaje final se ingresa manualmente y debe ser un entero positivo.
2. Puede diferir de todos los votos: no se exige unanimidad ni se calcula un promedio.
3. La ronda debe estar revelada (`closed`) y tener al menos un voto.
4. Cualquier integrante del proyecto puede registrar el acuerdo mediante su
   `member_id`, siguiendo el modelo de HU-013.
5. El acuerdo reemplaza los story points existentes. Repetir el mismo valor
   devuelve éxito si se siguen cumpliendo todas las condiciones de la operación.
6. Solo la ronda más reciente de la historia permite registrar el acuerdo.
   Si existe otra posterior, aunque esté abierta, se rechaza la ronda anterior.
7. No se permite registrar el acuerdo en una historia completada ni en una
   historia perteneciente a un Sprint cerrado. Una historia pendiente sin Sprint
   o en un Sprint abierto sí puede estimarse.
8. Solo se guarda el puntaje actual en la historia. Se conservan intactos los
   votos y las rondas; no se agrega historial del acuerdo ni autor del registro.

## Restricciones

- Depende de HU-005 y HU-013; no exige implementar HU-014.
- Reglas en Go; sin autenticación ni un nuevo rol de facilitador.
- La UI corresponde a HU-029.
- La escritura se realiza a través de un puerto público de `backlog`; las
  consultas de Sprint usan una interfaz pública de `sprints`. No se importan
  handlers ni repositorios PostgreSQL de otros módulos.
- La ronda más reciente es la última creada para la historia; debe existir un
  orden estable aunque coincidan sus fechas de creación.

## Casos límite

- Un solo voto permite registrar el acuerdo.
- Votos distintos permiten registrar un puntaje elegido en la discusión.
- Una estimación existente puede reemplazarse desde la ronda más reciente válida.
- Repetir el mismo valor devuelve éxito si la operación sigue siendo válida.
- Una ronda posterior abierta impide registrar desde la anterior.
- Una historia completada o perteneciente a un Sprint cerrado rechaza el registro.

## Condiciones de error

- Puntaje ausente, no entero o no positivo.
- Ronda abierta o revelada sin votos.
- Integrante ajeno al proyecto.
- Ronda anterior a la más reciente de la historia.
- Historia completada o perteneciente a un Sprint cerrado.
- Ronda o historia inexistente o ajena al proyecto.
- Fallo de persistencia: se conserva la estimación anterior.

## Criterios de aceptación

- [x] AC-015-01 — Guardar un entero positivo elegido explícitamente desde una ronda revelada con al menos un voto, aunque difiera de los votos.
- [x] AC-015-02 — Consultar posteriormente el puntaje guardado en la historia.
- [x] AC-015-03 — Rechazar una ronda abierta o revelada sin votos.
- [x] AC-015-04 — Rechazar a quien no sea integrante del proyecto.
- [x] AC-015-05 — Rechazar puntajes ausentes, no enteros o no positivos.
- [x] AC-015-06 — Reemplazar una estimación existente y aceptar la repetición del mismo valor cuando se cumplen las condiciones del registro.
- [x] AC-015-07 — Rechazar una ronda anterior si existe otra más reciente, incluso abierta.
- [x] AC-015-08 — Rechazar historias completadas o pertenecientes a un Sprint cerrado; permitir pendientes sin Sprint o en un Sprint abierto.
- [x] AC-015-09 — Conservar votos, rondas y demás datos de la historia; guardar solo el puntaje actual, sin historial ni autor del acuerdo.
- [x] AC-015-10 — Respetar el contrato HTTP y el aislamiento entre proyectos, conservando el puntaje anterior ante errores.

## Trazabilidad

- BDD: `features/HU-015-registrar-estimacion-acordada.feature`.
- Servicio: `backend/internal/planningpoker/estimate_service.go` y
  `estimate_service_test.go` (AC-015-01 y AC-015-03 a AC-015-09).
- HTTP: `estimate_handler.go` y `estimate_handler_test.go` (AC-015-01,
  AC-015-03 a AC-015-08 y AC-015-10); registro en `backend/cmd/api/main.go`.
- PostgreSQL: `estimate_postgres_test.go` verifica el registro por HTTP, consulta
  individual y listado del backlog, conservación de datos y estados de Sprint
  (AC-015-01, AC-015-02 y AC-015-06 a AC-015-10).
- Adaptador: `story_estimates.go`; escritura pública en
  `backend/internal/backlog/repository_postgres.go`; consulta pública de Sprint
  en `backend/internal/sprints/service.go`. Sin migraciones nuevas.
- Verificación: `go test -race -count=1 ./...` con `DATABASE_URL`,
  `BACKLOG_TEST_DATABASE_URL` y `SPRINTS_TEST_DATABASE_URL` contra PostgreSQL
  local, y `go vet ./...`. Tras el ajuste final de legibilidad y la extensión
  del recorrido HTTP se repitió `go test -race -count=1 ./internal/planningpoker`
  con PostgreSQL. Las pruebas crean esquemas aislados y los eliminan al terminar.
- Evidencia de commits RED/GREEN/REFACTOR y PR: `docs/traceability.md`.
