Las migraciones de dominio viven acá. Con `DATABASE_URL`, al arrancar la API
aplica el esquema de `projects` (`001` y `002`, `projects.EnsureSchema`), el
de `backlog` (`003`, `backlog.EnsureSchema`) y el de `sprints` (`004` y `005`,
`sprints.EnsureSchema`). `005` referencia `backlog_stories`, así que el esquema
de backlog tiene que existir antes.
