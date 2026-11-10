# HU-016 — Registrar esfuerzo

- Issue: [#25](https://github.com/FaustinoDuran/seguimiento-medicion/issues/25)
- Módulo Go: `backend/internal/effort/`
- Estado: implementada en `feat/HU-016-registrar-esfuerzo`, pendiente de PR

## Objetivo

Guardar las horas reales que un integrante dedicó a una actividad sobre una
historia. Cada alta agrega un registro. La comparación entre el esfuerzo
estimado y el real corresponde a HU-019. La pantalla corresponde a HU-030.

## Entradas

Proyecto e historia en la ruta. En el cuerpo JSON: `member_id`, `date`,
`activity` y `hours`.

Los identificadores son enteros positivos dentro del rango de `int64`.
`date` es un día calendario `YYYY-MM-DD`. `activity` es texto. `hours` es un
número JSON.

## Salidas esperadas

El registro queda persistido y la respuesta lo devuelve con su `id`.

## Contrato de la API

`POST /projects/{project_id}/stories/{story_id}/effort`

```json
{"member_id": 7, "date": "2026-11-10", "activity": "implementar el alta", "hours": 1.5}
```

Devuelve `201`:

```json
{
  "id": 1,
  "project_id": 1,
  "story_id": 4,
  "member_id": 7,
  "date": "2026-11-10",
  "activity": "implementar el alta",
  "hours": 1.5
}
```

`hours` sale como número JSON. `1`, `1.5`, `1.50` y `0.01` se aceptan y se
normalizan a `1`, `1.5`, `1.5` y `0.01`.

- `400`: JSON inválido; identificadores ausentes o no positivos; fecha
  ausente o inválida; actividad vacía o de más de 200 caracteres; horas
  ausentes, no numéricas, cero, negativas, con más de dos decimales o mayores
  que `99999999.99`.
- `404`: historia o integrante inexistentes, o pertenecientes a otro proyecto.
- `500`: fallo de infraestructura, sin exponer detalles internos.

Errores con formato `{"error":"..."}`. Se validan primero formato y rangos,
después la historia en el proyecto y después el integrante. Un rechazo no
agrega ni modifica registros.

## Reglas de negocio

1. Cada alta crea un registro nuevo. El mismo integrante puede cargar varias
   actividades el mismo día sobre la misma historia.
2. El integrante y la historia tienen que pertenecer al proyecto de la ruta.
3. La historia se acepta en cualquier estado y con cualquier sprint, incluido
   el sprint cerrado y la historia completada.
4. La actividad se recorta. Queda vacía si solo tenía espacios. El máximo es
   200 caracteres.
5. La fecha es el día calendario informado. Se acepta cualquier fecha válida,
   pasada o futura.
6. Las horas son mayores que cero y tienen como máximo dos decimales. El
   máximo representable es `99999999.99`.
7. No se calcula la desviación ni se modifican los story points de la historia.

## Restricciones

- Depende de HU-003 y HU-005.
- Reglas en Go. Sin autenticación.
- La UI corresponde a HU-030.
- La existencia de la historia se consulta por un puerto de `backlog` y la del
  integrante por un puerto de `projects`. No se importan handlers ni
  repositorios PostgreSQL de otros módulos.
- La tabla `effort_entries` referencia `projects`, `backlog_stories` y
  `project_members`. La columna de la fecha es `entry_date`.

## Casos límite

- `0.01` horas es el mínimo aceptado.
- `1.50` se guarda como `1.5`.
- Una actividad de 200 caracteres se acepta; una de 201 se rechaza.
- Una historia completada acepta el registro.
- Repetir la misma actividad, fecha e integrante crea otro registro.

## Condiciones de error

- Horas ausentes, cero, negativas o con más de dos decimales.
- Actividad vacía o demasiado larga.
- Fecha ausente o que no es un día calendario.
- Integrante o historia inexistentes o ajenos al proyecto.
- Fallo de persistencia: no queda un registro parcial.

## Criterios de aceptación

- [x] AC-016-01 — Guardar integrante, fecha, actividad y horas sobre una historia del mismo proyecto, y devolver el registro con id.
- [x] AC-016-02 — Agregar otro registro del mismo integrante, fecha e historia.
- [x] AC-016-03 — Aceptar `1`, `1.5`, `1.50` y `0.01`, normalizando los ceros de cola.
- [x] AC-016-04 — Rechazar horas ausentes, cero, negativas, con más de dos decimales o por encima de `99999999.99`, sin persistir.
- [x] AC-016-05 — Rechazar actividad vacía o de más de 200 caracteres, y fecha ausente o inválida, sin persistir.
- [x] AC-016-06 — Rechazar integrante o historia inexistentes o de otro proyecto.
- [x] AC-016-07 — Aceptar el registro sobre una historia completada.
- [x] AC-016-08 — Respetar el contrato HTTP. Un error de infraestructura responde `500` sin detalles internos y sin agregar un registro.

## Trazabilidad

- BDD: `features/HU-016-registrar-esfuerzo.feature`
- Servicio: `backend/internal/effort/service.go` y `service_test.go` (AC-016-01 a AC-016-06).
- HTTP: `handler.go`, `routes.go` y `handler_test.go` (AC-016-01, AC-016-04 a AC-016-06 y AC-016-08).
- PostgreSQL: `repository_postgres.go`, `repository_postgres_test.go` y migración `008` (AC-016-01, AC-016-02, AC-016-04, AC-016-06 y AC-016-07).
- Adaptadores: `story_lookup.go` y `member_lookup.go`. Registro en `backend/cmd/api/main.go`.
- Verificación: `go test -count=1 -race ./internal/effort/`, `go vet ./...` y `go test -count=1 -race ./...` con `BACKLOG_TEST_DATABASE_URL` y `SPRINTS_TEST_DATABASE_URL` contra PostgreSQL local. Las pruebas crean esquemas aislados y los eliminan al terminar.
- Evidencia de commits RED/GREEN: `docs/traceability.md`.
