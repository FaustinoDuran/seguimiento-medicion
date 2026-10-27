Las migraciones de dominio viven acá. La API aplica el esquema de `projects`
(`001` y `002`) al arrancar cuando hay `DATABASE_URL` (`projects.EnsureSchema`).

Luego aplica `backlog` (`003`, `backlog.EnsureSchema`) para las historias de
HU-005. El esquema incluye unicidad de identificador por proyecto.
