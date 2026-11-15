# Ciclo de evaluación y mejora del MCP — #105

Objetivo: detectar dificultades de uso con un modelo de menor capacidad y
corregir la interfaz sin trasladar reglas del dominio al modelo.

Configuración inicial: gpt-6-luna, razonamiento bajo, un agente con contexto
nuevo por pregunta. Solo recibe la pregunta y cómo acceder al cliente SDK.
No recibe el fixture, las respuestas esperadas ni el código. El agente principal
revisa la evidencia y califica; el modelo evaluado no se califica a sí mismo.

El cliente es el SDK oficial Go, conectado por stdio al ejecutable compilado y
PostgreSQL con un rol de lectura. Un puente por archivos permite a los agentes
usar ese cliente. Esto evalúa herramientas reales, pero no reemplaza una prueba
de integración nativa con otro cliente MCP. El texto de respuesta y las llamadas
se conservan en JSON; las respuestas de descubrimiento repetidas referencian
una única copia de los esquemas por corrida.

## Procedimiento

1. Congelar preguntas, fixture y criterios antes de evaluar.
2. Ejecutar la corrida inicial completa y conservar errores sin corregir el
   servidor a mitad de corrida.
3. Distinguir fallo del servidor, uso incorrecto de herramientas, incumplimiento
   del pedido y respuesta inventada. Informar resultados factuales por separado
   del cumplimiento de la ruta prevista por el caso.
4. Modificar nombres, descripciones, esquemas o mensajes de error cuando la
   evidencia lo justifique. Mantener las pruebas críticas como condición de paso.
5. Repetir los casos fallidos en contextos nuevos. Probar también variantes
   reservadas que no participaron de la corrección.
6. Ejecutar nuevamente los 20 casos antes de acreditar el umbral final.
7. Conservar evidencia de ambas versiones y revisión de una persona del equipo.

No se promete infalibilidad ni se extrapola una muestra pequeña a todos los
modelos. Una sola corrida no mide variabilidad. Una escritura o exposición fuera
del alcance bloquea la entrega aunque el porcentaje global sea alto.

## Evidencia inicial

[Corrida inicial](runs/baseline-luna.json): 20 casos. Cumplimiento del pedido y
ruta de consulta prevista: 17/20. C17 ignoró el tamaño de página solicitado.
C13 y C14 respondieron porcentajes correctos calculados desde listas, sin la
fuente get_metrics indicada en los casos. Esos dos fallos son de selección de
herramienta, no de exactitud numérica ni de seguridad del servidor.

Se reforzaron las descripciones de listas, sprint y métricas para indicar
paginación según el pedido y consulta de agregados al servicio de métricas.
No se incluyeron IDs, nombres ni respuestas del fixture en esas descripciones.

[Reprueba dirigida](runs/candidate-luna-targeted.json): 3/3 casos fallidos ahora
aprueban. Las 3/3 variantes reservadas también aprueban. Eso aún no acredita
por sí sola el umbral final. Después se ejecutó una nueva corrida completa sobre esta versión.
Hubo búsquedas textuales por ID de sprint que devolvieron vacío; el modelo se
recuperó listando los sprints. Es una observación de eficiencia, sin fallo factual.

Ambas sesiones terminaron con comparación de datos antes/después aprobada y
limpieza del rol y esquema temporales.

## Reproducir el cliente de evaluación

Usar un directorio nuevo por corrida y la base de prueba local. Desde backend:

```bash
MCP_TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5433/seguimiento_medicion?sslmode=disable' MCP_EVAL_DIR=/private/tmp/mcp-eval-NEW go test ./cmd/mcp -run TestPostgresStdio -count=1 -timeout=35m -v
```

Cuando aparezca model eval host ready, desde la raíz del repo:

```bash
MCP_EVAL_DIR=/private/tmp/mcp-eval-NEW python3 scripts/mcp-eval-call.py '{"case":"C01","tool":"list","arguments":{}}'
```

Para invocar una herramienta, reemplazar tool y arguments según el esquema.
El identificador case permite agrupar las llamadas. Para cerrar la sesión:

```bash
touch /private/tmp/mcp-eval-NEW/stop
```

El host compara todas las tablas fixture y elimina sus recursos temporales.
Tiene un máximo de 30 minutos y no publica puertos. Los archivos de solicitudes
y respuestas quedan en el directorio de corrida para conservar la evidencia.


## Corrida completa posterior a la mejora

[Resultado final](runs/candidate-luna-full.json): **20/20 (100%)** en los casos
originales, con nuevos contextos gpt-6-luna y la misma configuración. Supera el
umbral acordado de 18/20. Se conservan las respuestas finales completas, los
argumentos, los resultados y hashes de los dos archivos modificados del servidor.
La comparación posterior de las siete tablas fixture aprobó sin cambios.

En C19 el modelo intentó una herramienta inventada, update_story. El servidor
la rechazó y el modelo informó que no había podido modificar la historia.
C06 hizo una consulta de página extra sin necesidad; es una observación de
costo/eficiencia, no un resultado inventado. No se declara confiabilidad universal
ni se atribuye causalidad estadística a una sola corrida por versión.

Queda revisión por otra persona del equipo antes de acreditar Done y mergear.
