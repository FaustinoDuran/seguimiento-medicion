# Consultas MCP — TASK-013 / #105

Servidor de solo lectura para demo controlada; modo local por defecto y modo HTTP opcional. No hay sesiones, roles de
aplicación ni aislamiento por persona. Usar datos preparados para esa demo.
Decisión arquitectónica: [ADR 0009](../adr/0009-read-only-mcp.md).

TASK-025 conecta el despliegue del VPS a la misma base que usa la UI. El script
`scripts/seed-project.sh` agrega datos de exhibición sin reemplazar los existentes;
ver [carga compartida, datos y despliegue](deployment.md). La preparación manual
de abajo sigue siendo válida para una instancia local independiente.

## Preparación

Inicializar el esquema con el arranque habitual de la API antes de usar MCP.
El proceso MCP no crea tablas ni ejecuta migraciones. Con una cuenta
administradora, crear un rol exclusivo en la base de demo:

```sql
CREATE ROLE mcp_reader LOGIN PASSWORD 'REEMPLAZAR_POR_PASSWORD_LOCAL';
GRANT CONNECT ON DATABASE seguimiento_medicion TO mcp_reader;
GRANT USAGE ON SCHEMA public TO mcp_reader;
GRANT SELECT ON projects, project_members, backlog_stories, sprints,
  sprint_stories, effort_entries, defects TO mcp_reader;
ALTER ROLE mcp_reader SET default_transaction_read_only = on;
```

El rol debe carecer de permisos de escritura heredados o concedidos a PUBLIC.
El arranque rechaza privilegios de escritura sobre relaciones y secuencias,
creación de esquemas/tablas y privilegios administrativos. No usar la cuenta
postgres para ejecutar el servidor. Funciones SQL instaladas por administradores
siguen siendo parte del entorno confiable; estas comprobaciones no sustituyen
una auditoría de permisos del servidor PostgreSQL completo.

Desde backend:

```bash
go build -o /tmp/seguimiento-mcp ./cmd/mcp
```

En el cliente que soporte servidores MCP locales, configurar el ejecutable y
la variable de entorno (reemplazar contraseña y ruta):

```json
{
  "mcpServers": {
    "seguimiento-medicion": {
      "command": "/tmp/seguimiento-mcp",
      "env": {
        "MCP_DATABASE_URL": "postgres://mcp_reader:PASSWORD@localhost:5433/seguimiento_medicion?sslmode=disable"
      }
    }
  }
}
```

Los mensajes del protocolo usan stdout; los errores de arranque van a stderr.
Para exponer /mcp por HTTP ver [despliegue de demo remota](deployment.md).
El servidor usa el SDK oficial Go MCP v1.4.1 y su
negociación de protocolo. Se conserva Go 1.25.

## Herramientas y parámetros

| Herramienta | Parámetros admitidos |
|---|---|
| list_projects | query, status, limit, cursor |
| get_project | project_id obligatorio |
| list_stories | project_id obligatorio; query, status, priority, limit, cursor |
| get_story | project_id e id obligatorios |
| list_sprints | project_id obligatorio; query, status, limit, cursor |
| get_sprint | project_id y sprint_id obligatorios |
| get_metrics | project_id obligatorio; sprint_id opcional |
| get_report | project_id obligatorio; sprint_id opcional |

Las referencias son IDs numéricos positivos. Para nombres repetidos, listar y
pedir al usuario que seleccione el ID; el servidor no elige una coincidencia.
La búsqueda es una coincidencia parcial sin distinguir mayúsculas: por nombre
de proyecto, identificador/título de historia u objetivo del sprint.

Estado de proyecto: planned, active, closed. Estado de sprint: open, closed.
Prioridad: unassigned, low, medium, high. Estado de historia: todo o done, según el dominio existente. No se admiten rangos temporales arbitrarios.

Los listados responden con data y next_cursor cuando quedan resultados.
Orden por ID ascendente; limit predeterminado 20, máximo 100. Repetir filtros
con el cursor siguiente. No hay snapshot entre páginas ante cambios concurrentes.
Los servicios actuales cargan la colección antes de paginar: no es paginación
SQL ni una garantía de costo acotado en la base.

El detalle responde en data. Se conserva la serialización del dominio:
algunas entidades usan campos Go en PascalCase; historias e informes usan
snake_case. Amount conserva su representación JSON decimal, no centésimos
enteros. Las métricas se calculan exclusivamente en los servicios existentes.
El informe conserva proyecto, sprint, métricas, historias y sprints disponibles;
excluye entradas individuales de esfuerzo y defectos. No exporta PDF.

Errores de argumentos: invalid arguments. Entidad ausente o fuera del proyecto:
entity not found within requested scope. Fallos internos: query failed.
El SDK devuelve errores MCP sin SQL ni credenciales del almacenamiento.

## Evaluación

Desde backend:

```bash
go test ./internal/mcp -count=1 -v
go test ./cmd/mcp ./internal/platform/postgres
MCP_TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5433/seguimiento_medicion?sslmode=disable' go test ./internal/platform/postgres -run TestReadOnlyRole -count=1 -v
```

La última prueba requiere una base PostgreSQL de prueba y una cuenta que pueda
crear roles/esquemas. Crea y elimina solamente su rol y esquema temporales.
No usar una base de producción. Sin la variable, la prueba se omite: eso no
acredita el criterio de PostgreSQL.

El test stdio de internal/mcp usa servicios Go con repositorios fixture.
La prueba completa compila cmd/mcp y lo ejecuta con PostgreSQL y un rol de
lectura temporales:

```bash
MCP_TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5433/seguimiento_medicion?sslmode=disable' go test ./cmd/mcp -run TestPostgresStdio -count=1 -v
```

Los [20 casos conversacionales](conversation-cases.md) tienen hechos esperados
y una rúbrica; su ejecución requiere un cliente de IA y debe registrarse.

Ver [reporte de evidencia](evaluation.md). La aprobación crítica es obligatoria
en todos los casos; el resultado conversacional no compensa fallos críticos.

Ciclo de evaluación con modelo de menor capacidad: [procedimiento y evidencia](improvement-loop.md).

Evaluación posterior con modelo de menor capacidad, casos congelados y runner
de OpenCode: [TASK-022](task-022-improvement.md).
