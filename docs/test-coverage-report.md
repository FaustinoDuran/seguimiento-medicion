# Informe de cobertura y calidad de pruebas

Medición de cierre del TPI. El porcentaje de sentencias cubiertas describe qué
código ejecutaron las pruebas; no reemplaza la revisión de escenarios, casos
límite ni la trazabilidad de cada historia.

## Medición

| Dato | Valor |
|---|---|
| Producto medido | `bcb7bdbe8e7c1cba219a328ff1535c553d58608e` (`main`) |
| Fecha | 2026-11-15 |
| Cambio de esta tarea | workflows, `.gitignore` y este informe. No cambia código de producto. |
| Backend | `go test -count=1 -race -covermode=atomic -coverprofile=coverage.out ./...` desde `backend/` |
| Frontend | `npm run test:coverage` desde `frontend/` |

La medición local usó PostgreSQL 14.18. Para que el rol de lectura del MCP
coincida con PostgreSQL 16, que es el que usa CI y ya no otorga `CREATE` sobre
`public` a `PUBLIC`, se ejecutó `REVOKE CREATE ON SCHEMA public FROM PUBLIC`
en esa base temporal. Con eso, la suite completa quedó en verde. CI sigue
usando la imagen `postgres:16` y no necesita ese ajuste.

Los archivos generados no se versionan: `backend/coverage.out`,
`backend/coverage.txt` y `frontend/coverage/`.

## Backend

Total del perfil: **82,9% de sentencias**.

| Paquete | Sentencias |
|---|---|
| `internal/httpx` | 100,0% |
| `internal/backlog` | 96,0% |
| `internal/defects` | 95,5% |
| `internal/effort` | 92,7% |
| `internal/planningpoker` | 89,6% |
| `internal/reports` | 89,2% |
| `internal/metrics` | 86,9% |
| `internal/platform/postgres` | 86,4% |
| `internal/sprints` | 84,5% |
| `internal/mcp` | 81,2% |
| `internal/projects` | 68,7% |
| `cmd/api` | 19,7% |
| `internal/config` | 0,0% |
| `cmd/mcp` | 0,0% en este perfil |

`cmd/mcp` figura en 0% porque su prueba levanta el binario como otro proceso.
La cobertura de ese proceso no entra en el perfil del proceso de `go test`.
`cmd/api` y `internal/config` cubren el arranque del servidor; sus reglas de
negocio no viven ahí.

La suite incluye pruebas de reglas en memoria, HTTP con `httptest` y
persistencia PostgreSQL en esquemas temporales. `-race` quedó activo.

## Frontend

Vitest 5 con cobertura v8, 16 archivos de prueba y 116 pruebas en verde.

| Medida | Resultado |
|---|---|
| Sentencias | 88,35% (994/1125) |
| Líneas | 93,08% (862/926) |
| Ramas | 78,90% (722/915) |
| Funciones | 92,70% (343/370) |

Quedan fuera `src/main.tsx` y `src/test/**`, como define `frontend/vite.config.ts`.
Las pruebas usan jsdom, Testing Library y MSW. No calculan métricas ni reglas
de Planning Poker: ejercitan la presentación y los estados de carga, vacío y
error.

Las áreas con menos ramas cubiertas son Planning Poker, la página de medición
y el detalle de proyecto. Son estados de interfaz, no reglas de negocio.

## Qué no dice este número

- No dice que todos los escenarios Gherkin estén automatizados. Godog se
  aborda aparte, sobre una muestra de HU-019.
- No fija un umbral mínimo. Un porcentaje más alto en `cmd/api` no mejora el
  dominio, y bajar `projects` sí deja más comportamiento sin ejecutar.
- No incluye una corrida manual de la interfaz contra la API.

## CI

`.github/workflows/backend.yml` conserva `go test` con `-race` y agrega el
perfil de cobertura, el total en el resumen del job y el artefacto
`backend-coverage`.

`.github/workflows/frontend.yml` conserva lint, build y `npm test`, y agrega
`npm run test:coverage` más el artefacto `frontend-coverage`.

Ningún workflow falla por un porcentaje. Una prueba fallida sigue fallando el
job.

## Acciones siguientes

1. Mantener esta medición como línea base, sin un umbral elegido al azar.
2. Si se amplían pruebas de dominio, mirar primero `internal/projects` (68,7%).
3. No perseguir el 0% de `cmd/mcp` dentro de este perfil: esa ejecución ya
   existe como proceso separado.
4. Completar la muestra automatizada de HU-019 sin presentar ese trabajo como
   cobertura de todos los `.feature`.
