# Evidencia de evaluación — #105

Estado: pruebas críticas y corrida conversacional ejecutadas; pendiente revisión y merge del PR.

Se ejecutó `go test -count=1 -race ./...` con DATABASE_URL,
BACKLOG_TEST_DATABASE_URL, SPRINTS_TEST_DATABASE_URL y MCP_TEST_DATABASE_URL
apuntando a PostgreSQL local: todos los paquetes aprobaron. CI define también
MCP_TEST_DATABASE_URL para ejecutar las pruebas MCP de base de datos.
El fixture SQL del ejecutable y de los casos conversacionales es compartido;
la prueba del ejecutable se repitió después de extraerlo y aprobó.

| Evidencia ejecutada | Resultado |
|---|---|
| Descubrimiento de ocho herramientas por cliente SDK | aprobado |
| Rechazo de herramienta de escritura inexistente | aprobado |
| Filtros de proyectos, paginación y errores de consulta | aprobado |
| Subproceso stdio con cliente SDK oficial | aprobado |
| cmd/mcp compilado, ocho herramientas contra PostgreSQL con rol de lectura | aprobado |
| Filtros combinados y paginación de historias y sprints con PostgreSQL | aprobado |
| Comparación completa de las siete tablas fixture antes/después | sin cambios |
| Esquemas de entrada por herramienta, con campos obligatorios y rechazo de extras | aprobado |
| Métricas de proyecto vacío y errores internos de los cinco servicios | aprobado |
| Métricas proyecto/sprint contra servicios y valores fijos | aprobado |
| Métricas de informe contra servicio de informes | aprobado |
| Exclusión de datos individuales de esfuerzo y defectos | aprobado |
| Rechazo de sprint ajeno al proyecto | aprobado |
| PostgreSQL: lectura permitida, DML y DDL rechazados | aprobado |
| PostgreSQL: arranque con rol que tiene UPDATE | rechazado como se esperaba |

Fixtures actuales: proyecto 1, sprint cerrado 10, historia completada de 8 puntos,
2,5 horas registradas. El servicio conserva los cálculos y la serialización.
La prueba de permisos utiliza un rol y esquema PostgreSQL temporales y verifica
que el registro inicial siga presente después de intentar escrituras.

Pendiente para acreditar el conjunto completo:

- Revisión de la evidencia conversacional por otra persona. La corrida inicial de gpt-6-luna dio 17/20 en cumplimiento/ruta; la reprueba dirigida y variantes dieron 3/3 cada una; la corrida completa posterior dio **20/20**. Ver [ciclo de mejora](improvement-loop.md) y [respuestas/llamadas](runs/candidate-luna-full.json).
- Revisión y merge del PR.

Para cada corrida conversacional guardar pregunta, datos fixture, herramientas
invocadas y argumentos, respuesta final, hechos esperados y dictamen. Registrar
cliente, modelo, configuración, fecha y commit. Aprobar una respuesta solamente
si usa entidades correctas, preserva unidades y valores, justifica sus hechos
en datos consultados y evita inventar información. Requerir al menos 18/20
casos aprobados (90%). La corrida completa registrada aprueba el umbral; falta revisión por otra persona.

Modo remoto preparado después de la corrida conversacional: cliente SDK por HTTPS
con descubrimiento y consulta aprobados; Origin ajeno rechazado. Falta validar
cliente y hosting reales según [el plan de despliegue](deployment.md).
