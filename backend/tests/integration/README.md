Tests de integración contra PostgreSQL (Testcontainers o docker compose).

TASK-007 deja el pool en `internal/platform/postgres`. Las tablas de dominio
llegan con la primera HU de persistencia. Para ping real:

```bash
DATABASE_URL=postgres://postgres:postgres@localhost:5433/seguimiento_medicion?sslmode=disable \
  go test ./internal/platform/postgres -run TestConnectSucceedsWhenDatabaseURLIsReachable
```
