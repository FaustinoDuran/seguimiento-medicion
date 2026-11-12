# Seguimiento y medición

Aplicación de estimación, seguimiento y calidad de proyectos de software.

Trabajo práctico integrador — Ingeniería y Calidad de Software, UTN FRSR, 2026.

## Stack

| Capa | Tecnología |
|---|---|
| Backend (núcleo y reglas) | Go 1.24 + chi + pgx |
| Base de datos | PostgreSQL 16 |
| Frontend (presentación, Sprint 2) | TypeScript + React + Tailwind + Vite |

Las reglas de negocio **no** se implementan en React.

## Estado actual

Sprint 1 en curso. Con `DATABASE_URL` la API expone proyectos (HU-001),
integrantes (HU-003), historias de backlog (HU-005), sprints (HU-007),
asignación de historias (HU-008) y finalización de historias (HU-009).

## Cómo levantar PostgreSQL

Desde `backend/`:

```bash
cp .env.example .env
docker compose up -d
```

La base escucha en `localhost:5433` (no usa 5432 para no chocar con otros
Postgres locales). Variables: `backend/.env.example`.

## Cómo ejecutar la API

Desde `backend/`:

```bash
go run ./cmd/api
```

Si `DATABASE_URL` está definida (en el entorno o en `backend/.env`), la API
abre un pool pgx y hace ping al arrancar. Si falta, arranca igual y solo
expone HTTP.

Luego consultar el estado:

```bash
curl -i http://localhost:8080/health
```

La respuesta esperada es `200 OK` con `{"status":"ok"}`.

Con Postgres levantado, crear y consultar un proyecto:

```bash
curl -s -X POST http://localhost:8080/projects \
  -H 'Content-Type: application/json' \
  -d '{"name":"TP Integrador","start_date":"2026-10-26","end_date":"2026-11-23"}'

curl -s http://localhost:8080/projects/1

curl -s -X PATCH http://localhost:8080/projects/1 \
  -H 'Content-Type: application/json' \
  -d '{"name":"Seguimiento"}'

curl -s -X POST http://localhost:8080/projects/1/members \
  -H 'Content-Type: application/json' \
  -d '{"name":"Valentino"}'

curl -s http://localhost:8080/projects/1/members

curl -s -X POST http://localhost:8080/projects/1/sprints \
  -H 'Content-Type: application/json' \
  -d '{"goal":"Entregar el alta de proyectos"}'

curl -s http://localhost:8080/sprints/1

curl -s -X POST http://localhost:8080/sprints/1/close
```

Crear y consultar una historia (reemplazar los IDs por los devueltos por la API):

```bash
curl -s -X POST http://localhost:8080/projects/1/stories \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"HU-005","title":"Crear historia","priority":"high","story_points":5,"acceptance_criteria":["La historia se puede consultar"]}'

curl -s http://localhost:8080/projects/1/stories/1
```

El identificador se normaliza a mayúsculas y es único por proyecto. Toda historia
nueva comienza en `todo`. Se puede omitir prioridad (`unassigned`), story points
(`null`), descripción (`""`) y criterios de aceptación (`[]`). La escala de
prioridad en Go es `unassigned = 0`, `low = 1`, `medium = 2`, `high = 3`.

Asignar una historia a un Sprint abierto, completarla y consultar el avance
(reemplazar los IDs por los devueltos por la API):

```bash
curl -s -X POST http://localhost:8080/sprints/1/stories \
  -H 'Content-Type: application/json' \
  -d '{"story_id":1}'

curl -s -X POST http://localhost:8080/sprints/1/stories/1/complete

curl -s http://localhost:8080/sprints/1/stories
curl -s http://localhost:8080/projects/1/stories/1
```

Completar devuelve `200` con estado `done`; ese mismo estado aparece en ambas
consultas. Repetir la finalización devuelve éxito mientras el Sprint siga
abierto. No se exige estimación ni criterios de aceptación registrados.
Sin asignación al Sprint indicado o con Sprint cerrado, devuelve `409`;
si no existe el Sprint o la historia en su proyecto, devuelve `404`.
HU-006 permite editar prioridad, estado y estimaciones existentes desde el
backlog o mediante una actualización parcial:

```bash
curl -s -X PATCH http://localhost:8080/projects/1/stories/1 \
  -H 'Content-Type: application/json' \
  -d '{"status":"todo"}'

curl -s -X PATCH http://localhost:8080/projects/1/stories/1 \
  -H 'Content-Type: application/json' \
  -d '{"priority":"high","story_points":8}'
```

Reabrir conserva la asignación y requiere un Sprint abierto. Primero se guarda
la reapertura y después se corrigen los puntos. La corrección exige una
estimación existente y un entero positivo; no permite borrar la estimación.
Los campos omitidos o `null` se conservan; se rechazan campos desconocidos y
solicitudes sin campos editables no nulos. Un error impide guardar todos los
cambios. Proyectos o Sprints cerrados rechazan la edición con `409`.

La API devuelve `400` para datos inválidos, `404` para recursos inexistentes y
`409` para identificadores duplicados, proyectos cerrados al crear o un Sprint
ya cerrado.

## Planning Poker

HU-012 permite abrir rondas y votar; HU-013 revela los votos y cierra la ronda.
HU-014 agrega `has_differences` a la respuesta de revelar y a las consultas de
rondas cerradas: `true` si hay dos puntajes distintos; `false` con todos iguales,
un voto o ninguno. Las rondas abiertas omiten el campo para preservar el secreto.

Para repetir, usar `POST /projects/{project_id}/stories/{story_id}/rounds`
después de revelar. Devuelve otra ronda abierta sin votos y conserva las rondas,
votos y estimación anteriores. Repetir es voluntario, incluso si hay coincidencia.
Solo puede existir una ronda abierta por historia. HU-015 registra manualmente
el acuerdo desde la última ronda revelada con votos; la UI corresponde a HU-029.

## Registro de defectos

HU-017 agrega alta y consulta individual de defectos. Reemplazar los IDs por
los del proyecto, la historia y el Sprint de detección:

```bash
curl -s -X POST http://localhost:8080/projects/1/stories/4/defects \
  -H 'Content-Type: application/json' \
  -d '{"description":"El archivo exportado sale vacío","severity":"high","detection_sprint_id":3}'

curl -s http://localhost:8080/projects/1/defects/17
```

El alta devuelve `201` con el ID generado y estado `open`. La descripción se
recorta y admite de 1 a 2000 caracteres; la severidad es obligatoria (`low`,
`medium`, `high` o `critical`). Historia y Sprint de detección deben pertenecer
al proyecto. Se admiten proyectos cerrados, historias completadas o sin Sprint
y Sprints de detección cerrados o distintos del asignado a la historia.
Repetir el alta crea otro defecto. La consulta devuelve `200`; los datos
inválidos devuelven `400`, los recursos inexistentes o ajenos `404`, y los
fallos internos `500` genérico. La resolución corresponde a HU-018 y la UI a HU-030.

## Resolución de defectos

HU-018 agrega resolución y consulta del Sprint donde se corrigió el defecto:

```bash
curl -s -X POST http://localhost:8080/projects/1/defects/17/resolve \
  -H 'Content-Type: application/json' \
  -d '{"resolution_sprint_id":3}'

curl -s http://localhost:8080/projects/1/defects/17
```

Devuelve `200` con estado `resolved` y `resolution_sprint_id`. En defectos
abiertos ese campo es `null`. El Sprint debe pertenecer al proyecto y ser el
de detección o uno posterior según orden de creación (`created_at`, luego
`id`), siguiendo HU-011; ese orden no es un calendario de ejecución.
Se admiten proyectos/Sprints cerrados e historias completadas o sin Sprint.

Repetir la misma resolución devuelve éxito. Elegir un Sprint anterior o
cambiar el Sprint de un defecto ya resuelto devuelve `409`, sin modificarlo.
Las solicitudes concurrentes conservan la primera resolución guardada;
si ambas indican el mismo Sprint, ambas pueden devolver éxito. La descripción,
severidad, historia y Sprint de detección se conservan. No se agrega reapertura.

## Verificación del backend

HU-028 agrega `GET /projects/{project_id}/stories`: devuelve
`{ "stories": [...] }` con todas las historias del proyecto en `id DESC`,
incluidas las completadas y asignadas a Sprint. Un proyecto vacío devuelve
`{ "stories": [] }`; uno cerrado sigue siendo consultable. ID inválido: `400`;
proyecto inexistente: `404`; fallo de persistencia: `500` JSON genérico.

Desde `backend/`, ejecutar `go test ./...`. Para incluir persistencia PostgreSQL,
HTTP con base real y unicidad concurrente:

```bash
DATABASE_URL='postgres://postgres:postgres@localhost:5433/seguimiento_medicion?sslmode=disable' \
BACKLOG_TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5433/seguimiento_medicion?sslmode=disable' \
SPRINTS_TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5433/seguimiento_medicion?sslmode=disable' \
  go test -race ./...
```

Los tests de backlog y de sprints usan esquemas temporales aislados que se
eliminan al terminar. Sin la variable de cada paquete, sus tests de PostgreSQL
se omiten explícitamente.

## Hook pre-push

TASK-014 agrega un hook Git versionado que ejecuta la suite completa del backend
antes de enviar cambios al remoto. Si un test falla o falta un requisito, el
push se cancela y el remoto no recibe nada.

### Instalación (una vez por clon)

```bash
./scripts/install-hooks.sh
```

El script es idempotente, se puede ejecutar desde cualquier directorio del repo
y crea un symlink `.git/hooks/pre-push → .githooks/pre-push`. Si el clon ya
tenía un `pre-push` distinto, lo conserva como `pre-push.local` y nuestro hook lo
ejecuta después de la suite. Los demás hooks (`pre-commit`, `commit-msg`, etc.)
no se tocan. Si el sistema no permite symlinks, se copia el archivo y hay que
reinstalar después de cada cambio en `.githooks/`.

### Qué ejecuta

Desde `backend/`, independientemente del directorio desde el que invoques Git:

```bash
go test -count=1 -race ./...
```

Corre en cada push normal, incluso si solo cambió el frontend, porque el objetivo
es que ningún push deje el backend roto. No ejecuta un `git push` propio ni
modifica archivos.

### Requisitos

- Go 1.24 o superior en `PATH`.
- PostgreSQL de pruebas accesible.
- `DATABASE_URL`, `BACKLOG_TEST_DATABASE_URL` y `SPRINTS_TEST_DATABASE_URL`
  exportadas en la sesión. **Sin las tres, el hook falla a propósito**: sin esas
  variables los tests de PostgreSQL se omitirían y la suite no sería completa.

```bash
docker compose -f backend/docker-compose.yml up -d postgres

export DATABASE_URL='postgres://postgres:postgres@localhost:5433/seguimiento_medicion?sslmode=disable'
export BACKLOG_TEST_DATABASE_URL="$DATABASE_URL"
export SPRINTS_TEST_DATABASE_URL="$DATABASE_URL"
```

Usá siempre la base de pruebas local, nunca una de producción. Los tests crean y
eliminan sus propios esquemas temporales.

### Tiempos esperables

Con PostgreSQL levantado y el build en caché, unos 10 a 20 segundos. La primera
ejecución después de un cambio grande compila de nuevo y puede tardar más.

### Diagnóstico de fallos

| Síntoma | Causa probable | Qué hacer |
| --- | --- | --- |
| `Faltan variables de entorno…` | Variables no exportadas en esta shell | Exportar las tres y reintentar |
| `No se encontró 'go' en PATH` | Go no instalado o fuera del `PATH` | Instalar Go 1.24+ |
| `connection refused` en los tests | PostgreSQL apagado o URL equivocada | Levantar el contenedor y revisar el puerto `5433` |
| Un test rojo real | Cambio que rompe el backend | Corregir el código y volver a hacer push |

La salida de los tests se muestra completa: el hook la conserva y propaga el
código de error.

### Límites

- El hook corre localmente. GitHub Actions, en
  [`.github/workflows/backend.yml`](.github/workflows/backend.yml), corre la
  misma suite después de recibir el push. Llega más tarde y no reemplaza este
  control.
- Se puede omitir con `git push --no-verify`. El hook no es una política
  obligatoria del servidor: no reemplaza la protección de ramas ni la revisión
  de otra persona.
- Los commits RED del TDD siguen siendo posibles localmente. Mientras la suite
  esté roja, el `pre-push` impedirá publicarlos por la vía normal; entre el
  commit RED y el GREEN, la publicación se hace cuando la suite vuelve a verde
  (o con `--no-verify`, de forma consciente y avisando al equipo).

## GitHub Actions

En cada push y en cada pull request corren dos checks:

- Frontend: `.github/workflows/frontend.yml` hace `npm ci`, lint, build y tests.
- Backend: `.github/workflows/backend.yml` levanta PostgreSQL 16 y, desde
  `backend/`, ejecuta `go test -count=1 -race ./...` con `DATABASE_URL`,
  `BACKLOG_TEST_DATABASE_URL` y `SPRINTS_TEST_DATABASE_URL`. Es el mismo
  comando que el hook. La base existe solo durante el check y usa el usuario
  de prueba del `docker-compose`, no una base del equipo.

El hook sigue corriendo antes del push en cada clon donde esté instalado.

## Documentación del equipo

- [Guía de trabajo y trazabilidad](docs/guia-de-trabajo.md) — cómo trabaja el equipo y el agente de Cursor
- [Registro del Sprint 0](docs/sprints/sprint-0.md)
- [Trazabilidad](docs/traceability.md)
- [ADRs](docs/adr/)
- [Plantilla SDD](docs/sdd/_plantilla.md)

## Estructura

```text
backend/cmd/api               servidor HTTP (chi)
backend/internal/httpx/      router y GET /health
backend/internal/config/     DATABASE_URL y .env
backend/internal/platform/postgres/  pool pgx
backend/migrations/          SQL de dominio (projects, backlog y sprints)
backend/docker-compose.yml   PostgreSQL 16
backend/internal/<feature>/   dominio Go (TDD)
frontend/src/features/       UI (Sprint 2)
docs/sdd/                    especificaciones
docs/adr/                    decisiones
features/                    escenarios BDD
```
