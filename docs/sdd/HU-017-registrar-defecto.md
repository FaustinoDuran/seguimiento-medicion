# HU-017 — Registrar defecto

- Issue: [#26](https://github.com/FaustinoDuran/seguimiento-medicion/issues/26)
- Módulo Go: `backend/internal/defects/`
- Estado: mergeada en `main` ([PR #88](https://github.com/FaustinoDuran/seguimiento-medicion/pull/88))

## Objetivo

Registrar un problema detectado en el comportamiento de una historia para
seguir la calidad del trabajo y poder consultar posteriormente ese defecto.

## Entradas

- Proyecto e historia afectados.
- Descripción obligatoria de 1 a 2000 caracteres después de recortar espacios
  al principio y al final.
- Severidad obligatoria: baja, media, alta o crítica, sin valor predeterminado.
- Sprint de detección obligatorio, informado explícitamente.

No se solicita informante ni una fecha de detección independiente del Sprint.
El estado inicial es abierto (`open`), asignado automáticamente.

## Salidas esperadas

El defecto queda persistido y puede consultarse individualmente por su ID.
Se devuelve su ID generado, proyecto, historia, Sprint de detección,
descripción recortada, severidad y estado abierto.

## Contrato de la API

Contrato confirmado en la entrevista, siguiendo las convenciones de las APIs
existentes:

- `POST /projects/{project_id}/stories/{story_id}/defects`: registra el defecto
  y devuelve `201` con el registro completo.
- `GET /projects/{project_id}/defects/{id}`: devuelve `200` con el defecto
  perteneciente al proyecto indicado. No exige conocer previamente su historia.

Cuerpo del alta:

```json
{
  "description": "El archivo exportado sale vacío",
  "severity": "high",
  "detection_sprint_id": 3
}
```

Respuesta de alta y consulta:

```json
{
  "id": 17,
  "project_id": 1,
  "story_id": 4,
  "description": "El archivo exportado sale vacío",
  "severity": "high",
  "status": "open",
  "detection_sprint_id": 3
}
```

Las severidades del contrato son `low`, `medium`, `high` y `critical`,
correspondientes a baja, media, alta y crítica. Se exige uno de esos valores
exactos. El alta asigna `open` si se omite `status`; si se informa ese campo,
solo se acepta `open` (no acepta vacío ni `null`). No permite registrar un
defecto ya resuelto.

Todos los identificadores son enteros positivos dentro del rango de `int64`.
Los errores tienen formato `{"error":"..."}`:

- `400`: JSON inválido o más de un valor JSON; IDs inválidos o ausentes;
  descripción inválida; severidad ausente o inválida; estado inicial inválido.
- `404`: historia o Sprint inexistentes o ajenos al proyecto al registrar;
  defecto inexistente o ajeno al proyecto al consultar.
- `500`: fallo de infraestructura, con mensaje genérico sin detalles internos.

Se validan primero formato y rangos; luego se consulta la historia dentro del
proyecto y después el Sprint de detección. Un rechazo no crea un defecto ni
modifica otros datos. La consulta individual siempre incluye el proyecto en
la búsqueda para preservar el aislamiento entre proyectos.

## Reglas de negocio

Decisiones confirmadas durante la entrevista `grill-with-docs`:

1. Cada alta crea un defecto abierto. Resolverlo corresponde a HU-018.
2. La descripción es obligatoria; no se acepta una descripción compuesta
   únicamente por espacios ni de más de 2000 caracteres después del recorte.
3. La severidad se elige explícitamente entre baja, media, alta y crítica.
   Expresa impacto, no prioridad de atención.
4. La historia y el Sprint de detección pertenecen al proyecto indicado.
5. El Sprint de detección puede ser distinto del asignado a la historia.
   No se deduce automáticamente a partir de un supuesto Sprint actual.
6. Se aceptan historias completadas y Sprints de detección cerrados.
   Registrar el defecto no reabre la historia ni el Sprint.
7. Se permite registrar defectos en proyectos cerrados y consultar los existentes.
8. Cada alta crea un registro nuevo, aunque coincidan descripción, historia
   y Sprint de detección. No hay detección automática de duplicados.
9. Se aceptan historias sin Sprint asignado, siempre que se informe un Sprint
   de detección del mismo proyecto.

## Restricciones

- Depende de HU-005 y HU-007.
- Reglas de negocio únicamente en Go; sin autenticación.
- Esta HU incluye alta y consulta individual. El listado queda fuera del alcance.
- La resolución corresponde a HU-018; la pantalla, a HU-030.
- No se calculan métricas ni se modifican historias o Sprints.

## Casos límite

- Historia completada cuyo defecto se detecta en un Sprint posterior.
- Sprint de detección cerrado.
- Proyecto cerrado.
- Dos altas con los mismos datos producen dos defectos distintos.
- Descripción de exactamente 2000 caracteres después del recorte.
- Historia sin Sprint asignado con un Sprint de detección informado.
- Los caracteres de la descripción se cuentan como puntos de código Unicode,
  no como bytes.

## Condiciones de error

- Descripción ausente, vacía después del recorte o demasiado larga.
- Severidad ausente o fuera de las cuatro opciones permitidas.
- Historia o Sprint de detección ausentes, inexistentes o de otro proyecto.

- Estado inicial distinto de abierto.
- Defecto inexistente o consultado bajo otro proyecto.
- Fallo de persistencia: no se crea un registro parcial ni se exponen detalles
  internos en la respuesta.

## Criterios de aceptación

- [x] AC-017-01 — Registrar un defecto con descripción, severidad, historia y
  Sprint de detección del mismo proyecto, con estado inicial abierto.
- [x] AC-017-02 — Consultar individualmente el defecto guardado por su ID.
- [x] AC-017-03 — Recortar la descripción y rechazarla si queda vacía o supera
  2000 caracteres, sin crear un defecto.
- [x] AC-017-04 — Exigir severidad baja, media, alta o crítica.
- [x] AC-017-05 — Exigir un Sprint de detección explícito del mismo proyecto,
  permitiendo que difiera del Sprint asignado a la historia.
- [x] AC-017-06 — Permitir historias completadas, Sprints de detección cerrados
  y proyectos cerrados, sin reabrirlos.
- [x] AC-017-07 — Crear registros distintos ante altas con los mismos datos.
- [x] AC-017-08 — Permitir historias sin Sprint asignado cuando se indica un
  Sprint de detección del mismo proyecto.
- [x] AC-017-09 — Rechazar historia o Sprint inexistentes o ajenos al proyecto;
  no devolver un defecto al consultarlo bajo un proyecto distinto.
- [x] AC-017-10 — Respetar el contrato HTTP, rechazar datos inválidos sin crear
  registros y responder fallos de infraestructura sin detalles internos.

## Confirmación de la entrevista

Las reglas de negocio Q1 a Q11 y el contrato completo fueron confirmados por
el usuario antes de implementar. No se requiere un ADR: estas decisiones pertenecen a la HU y
no cambian el stack, la arquitectura ni el alcance transversal del producto.

## Interfaces de prueba

Confirmadas por el usuario antes de escribir los tests, según la skill `tdd`:

- Servicio público Go (`Register` y `Get`): reglas, validaciones y aislamiento
  de proyecto sin HTTP ni base real; no demuestra serialización ni SQL.
- API HTTP con almacenamiento en memoria: rutas, JSON y estados HTTP;
  no demuestra persistencia PostgreSQL.
- API HTTP con PostgreSQL real y esquema aislado: alta y consulta persistida,
  adaptadores entre módulos y aislamiento entre proyectos; no incluye UI.

## Trazabilidad

- BDD: `features/HU-017-registrar-defecto.feature`.
- Servicio: `backend/internal/defects/service_test.go`, alta/consulta,
  descripción Unicode, severidad, estado inicial y referencias al proyecto.
- HTTP: `backend/internal/defects/handler_test.go`, contrato JSON, códigos de
  respuesta, duplicados, aislamiento y errores genéricos de infraestructura.
- PostgreSQL: `backend/internal/defects/repository_postgres_test.go`, consulta
  desde otra instancia del servicio, estados cerrados, historias sin Sprint,
  detección en otro Sprint, duplicados y aislamiento real entre proyectos.
- Código: `backend/internal/defects/`; migración `009`; registro de esquema y
  rutas en `backend/cmd/api/main.go`.
- Evidencia TDD y PR: `docs/traceability.md`.

## Verificación

Se ejecutaron `go test -count=1 -race ./internal/defects/`,
`go test -count=1 -race ./...` y `go vet ./...` desde `backend/`, con
PostgreSQL local para las pruebas de persistencia.

La suite completa usa `DATABASE_URL`, `BACKLOG_TEST_DATABASE_URL` y
`SPRINTS_TEST_DATABASE_URL`. Las pruebas de defectos usan
`DEFECTS_TEST_DATABASE_URL` o, si no está definida, `BACKLOG_TEST_DATABASE_URL`;
sin ambas, se omiten explícitamente. Cada prueba de integración crea y elimina
un esquema temporal aislado. No hay ejecución automatizada de Gherkin con Godog.
