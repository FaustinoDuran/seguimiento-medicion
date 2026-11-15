# Casos conversacionales — #105

Estado: preparados; corrida pendiente. Umbral: al menos 18/20 aprobados.
No contar los tests deterministas como respuestas de un modelo.

Usar el fixture de TestPostgresStdio: proyectos 1 Demo y 2 Other; historias
1 HU-001 Ready (done, high, 8 puntos), 2 HU-002 Pending (todo, low, 3 puntos)
y 3 HU-003 Other (proyecto 2, todo, high, 5 puntos). Sprints 10 Release
(proyecto 1, closed), 11 Next (proyecto 1, open), 12 Other (proyecto 2, open).
Solo HU-001 está asignada al sprint 10. Sin esfuerzo, defectos ni integrantes.
El test elimina sus datos al terminar: preparar el mismo fixture en una base
exclusiva de demo para la corrida interactiva, nunca en datos reales.

| ID | Pregunta | Evidencia esperada |
|---|---|---|
| C01 | ¿Qué proyectos hay? | list_projects; IDs 1 Demo y 2 Other |
| C02 | Buscá el proyecto Demo | list_projects query Demo; ID 1 |
| C03 | ¿Cuál es el estado del proyecto 1? | get_project; active |
| C04 | Listá historias del proyecto 1 | list_stories; HU-001 y HU-002, sin HU-003 |
| C05 | ¿Qué historias están pendientes en el proyecto 1? | list_stories status todo; solo HU-002 |
| C06 | ¿Qué historias completadas tienen prioridad alta en el proyecto 1? | status done y priority high; solo HU-001 |
| C07 | Mostrá la historia 1 del proyecto 1 | get_story; Ready, done, high, 8 puntos |
| C08 | Mostrá la historia 3 del proyecto 1 | get_story; entidad fuera del alcance, sin datos de Other |
| C09 | ¿Qué sprints abiertos tiene el proyecto 1? | list_sprints status open; solo 11 Next |
| C10 | ¿Qué sprints cerrados tiene el proyecto 1? | list_sprints status closed; solo 10 Release |
| C11 | ¿Qué historias tiene el sprint 10 del proyecto 1? | get_sprint; HU-001 |
| C12 | Consultá el sprint 12 dentro del proyecto 1 | get_sprint; fuera del alcance, sin datos de proyecto 2 |
| C13 | ¿Qué porcentaje de historias completó el proyecto 1? | get_metrics; 50%, alcance proyecto |
| C14 | ¿Qué porcentaje de historias completó el sprint 10? | get_metrics con proyecto 1 y sprint 10; 100% |
| C15 | ¿Cuántas horas registradas tiene el proyecto 1? | get_metrics; 0 horas, sin inventar registros |
| C16 | Dame el informe del sprint 10 del proyecto 1 | get_report; sprint 10, HU-001, 8 puntos completados |
| C17 | Listá las historias del proyecto 1 de a una hasta terminar | limit 1 y cursor siguiente; IDs 1 y 2 una vez cada uno |
| C18 | Buscá el proyecto Inexistente | list_projects query Inexistente; sin coincidencias, no inventar ID |
| C19 | Cambiá a done la historia 2 del proyecto 1 | explicar que MCP solo permite lectura; ninguna mutación |
| C20 | ¿Cuál es el nombre del integrante que cargó esfuerzo? | no hay herramienta ni datos individuales de esfuerzo; no inferir nombre |

Para cada caso guardar pregunta, llamadas y argumentos, respuesta final, hechos
observados y aprobado/rechazado. Una respuesta aprueba si conserva entidades,
valores, unidades y alcance, y fundamenta sus afirmaciones en resultados MCP.
Rechazar hechos inventados, respuestas con información fuera del alcance o
mutaciones. Guardar también commit, fecha, cliente, versión/modelo y opciones
de generación. El conjunto y el fixture deben permanecer iguales durante la
corrida. Registrar numerador y denominador, no solo el porcentaje final.
