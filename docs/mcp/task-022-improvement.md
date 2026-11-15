# Ciclo de evaluación TASK-022 (#108)

Objetivo: verificar y mejorar de forma acotada la facilidad de uso y la robustez
del MCP existente con un modelo de menor capacidad, conservando el contrato de
solo lectura y las reglas de negocio en Go. Esta TASK no lleva SDD ni BDD porque
no agrega comportamiento de negocio.

Continúa el trabajo de [la corrida de #105](improvement-loop.md), que queda como
historial y no se reescribe.

## Estado de la corrida

**Bloqueo externo.** Las pruebas críticas, la carga y el andamiaje están
completos y verdes. La corrida conversacional **no** se ejecutó: el proveedor de
OpenCode rechazó el modelo pago por saldo insuficiente y agotó el límite de los
modelos gratuitos.

```
Insufficient balance. This request is pay-as-you-go, billed to OpenCode Models
credits. ... Free usage limit reached.
```

Sin esa corrida no se acreditan los umbrales de 86/90 ni 27/30. Por lo tanto la
TASK **no** está terminada y no se abre PR como si lo estuviera. La Issue #108
mantiene el trabajo pendiente.

## Qué está probado

| Evidencia | Resultado |
|---|---|
| `go test -count=1 ./...` con las cuatro variables de base | aprobado, 0 omitidos |
| Pruebas críticas de contrato, alcance, métricas y errores (stdio) | aprobado |
| `TestPostgresStdio`: binario real contra PostgreSQL con rol de lectura | aprobado |
| Comparación de las siete tablas fixture antes/después | sin cambios |
| `TestReadOnlyRole`: DML y DDL rechazados; arranque con UPDATE rechazado | aprobado |
| Formato uniforme `snake_case` en las ocho herramientas | aprobado |
| `get_metrics` idéntico a `get_report.metrics` | aprobado |
| Carga: 100 llamadas, 5 concurrentes, p95 ≤ 3 s | p95 5 ms, 0 errores |

Los comandos están en [la guía](README.md). La medición de carga requiere
`MCP_EVAL_LOAD=1` y `MCP_DATABASE_URL` para no alargar la suite habitual.

## Mejoras aplicadas

Ambas se justifican por fallos verificables, no por preferencia de estilo.

1. **Un solo formato de campos.** `get_project`, `get_sprint`, `get_metrics` y
   los listados devolvían campos Go en PascalCase, mientras `get_report` usaba
   snake_case. Un mismo dato aparecía como `ActualHours` en una herramienta y
   `actual_hours` en otra. Se agregaron [vistas de
   respuesta](../../backend/internal/mcp/presentation.go) que exponen
   snake_case en todas las herramientas, sin tocar los servicios de dominio.
2. **Límite de página explícito.** Una descripción no aclaraba que `limit` por
   encima de 100 se rechaza en lugar de recortarse. Se documentó en el esquema y
   en la descripción de cada listado. Verificado: `limit: 500` y `limit: 101`
   devuelven error de validación, no una página recortada.

No se incluyeron IDs, nombres ni respuestas del fixture en las descripciones.

## Conjunto congelado

[docs/mcp/cases/TASK-022-frozen.json](cases/TASK-022-frozen.json) contiene las 20
regresiones y las 10 variantes reservadas, con sus hechos esperados y los
controles automatizables. Se congeló antes de cualquier corrida.

| Conjunto | Base | Fixture |
|---|---|---|
| Regresiones | `mcp_demo_regress` | `backend/cmd/mcp/testdata/demo.sql` |
| Reservadas | `mcp_demo_shown` | `backend/cmd/mcp/testdata/task022-reserved.sql` |

Las variantes reservadas cubren lo que la regresión no tiene: dos proyectos con
el mismo nombre, un proyecto sin historias ni sprints, una historia sin puntos,
horas con decimales exactos, paginación de a dos sobre cinco historias, un
`sprint_id` inexistente y dos textos hostiles dentro de títulos devueltos como
datos. Verificado con el servidor real: los datos llegan literales, el alcance
se respeta y las escrituras se rechazan.

## Reproducir

Preparar las dos bases y el rol de lectura según [la guía](README.md):

```bash
# Esquema, fixture y rol de cada base (una vez por corrida)
DATABASE_URL='postgres://postgres:postgres@localhost:5433/mcp_demo_regress?sslmode=disable' \
  backend/bin/api  # primer arranque crea las tablas; detenerlo después
psql "$MCP_DEMO_ADMIN_URL" -v ON_ERROR_STOP=1 -f backend/cmd/mcp/testdata/demo.sql
psql "$MCP_DEMO_ADMIN_URL" -v ON_ERROR_STOP=1 -f backend/cmd/mcp/testdata/task022-reserved.sql
```

El cliente evaluado corre en un directorio aislado, fuera del repositorio, para
que no pueda leer `AGENTS.md`, `docs/`, el fixture ni el código del servidor. Su
`opencode.json` conecta el MCP por stdio a través de
[el envoltorio de traza](../../scripts/mcp-stdio-tee.py), que registra el
JSON-RPC real: argumentos enviados y resultados devueltos.

```bash
python3 scripts/mcp-eval-run.py --set regressions --reps 3 \
  --model opencode/gpt-6-luna --variant low \
  --out docs/mcp/runs/TASK-022-baseline.json
```

Cada repetición es un proceso nuevo: contexto nuevo, sin respuestas esperadas ni
fixture. El agente principal califica; el script solo captura evidencia y marca
los hechos demostrables por máquina.

## Límites del ciclo

Tomados de la Issue, sin ampliarlos por cuenta propia:

| Concepto | Límite |
|---|---|
| Corridas completas | Una referencia y hasta tres candidatas |
| Conversaciones por corrida | 20 × 3 + 10 × 3 = 90 |
| Repruebas dirigidas | Hasta 10 por ciclo |
| Techo total | 360 + 30 = 390 conversaciones |
| Por conversación | 12 llamadas MCP y 120 segundos |
| Umbral | 86/90 regresiones y 27/30 reservadas; cada pregunta ≥ 2 de 3 |
| Parada | Dos ciclos consecutivos sin mejorar las 60 regresiones fijas |

Al revelar un conjunto reservado pasa a ser conocido: la ronda siguiente usa
variantes nuevas y las reporta separadas. No se cambian modelo, prompt, datos y
servidor a la vez.

## Pendiente para completar la TASK

1. Restablecer acceso a un modelo en el proveedor de OpenCode y registrar su ID
   real y su configuración.
2. Ejecutar la corrida de referencia (90 conversaciones) y conservarla.
3. Diagnosticar los fallos, corregir solo las hipótesis verificables y correr
   hasta tres candidatas dentro del techo.
4. Repetir las pruebas críticas después de cualquier cambio en el servidor.
5. Revisión por otra persona y PR con `Closes #108`.
