Las migraciones de dominio viven acá. Con `DATABASE_URL`, al arrancar la API
aplica el esquema de `projects` (`001` y `002`, `projects.EnsureSchema`), el
de `backlog` (`003` y `006`, `backlog.EnsureSchema`) y el de `sprints` (`004` y `005`,
`sprints.EnsureSchema`). `005` referencia `backlog_stories`, así que el esquema
de backlog tiene que existir antes.

`006` amplía los estados de historia a `todo` y `done`. El arranque también
aplica este cambio a tablas existentes: conserva las historias y sus datos,
y mantiene `todo` como estado inicial. La operación puede repetirse después
de completar historias. No volver a aplicar `003` para actualizar el constraint:
`CREATE TABLE IF NOT EXISTS` no modifica una tabla existente.
