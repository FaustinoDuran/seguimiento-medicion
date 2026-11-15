# Traspaso a OpenCode: validar y publicar la demo MCP

## Objetivo y alcance

Dejar el MCP funcionando por HTTPS en un hosting, conectado a un cliente de IA
real, para mostrarlo a profesores. Un conjunto compartido de datos de exhibición,
sin cuentas ni roles de aplicación. Mantener solo lectura y reglas en Go.

Trabajo: TASK-013, Issue #105, PR #107 contra main. Al redactar este plan la
rama es chore/TASK-013-read-only-mcp y el último commit es a5d15ad. Verificar el
estado real antes de actuar; esos datos son referencias, no prueba de vigencia.
La planificación y asignación se mantienen en GitHub, no en este documento.

Este plan propone una meta de calidad adicional al mínimo de la Issue:
95% conversacional en una corrida repetida. El mínimo original de 90% sigue
registrado; reportar ambos sin cambiar la Issue silenciosamente.

## 1. Recuperar el estado real

Leer AGENTS.md, docs/guia-de-trabajo.md, .cursor/rules/ y las skills locales
pre-commit y plan-de-accion. Consultar Issue #105 y PR #107 si el acceso funciona.
Revisar git status y conservar cambios ajenos, incluidas carpetas locales de
herramientas. Trabajar sobre la rama de la TASK; evitar duplicar un PR existente.

Consultar estos documentos según el paso:

- Arquitectura y alcance: docs/adr/0009-read-only-mcp.md.
- Configuración: docs/mcp/README.md y docs/mcp/deployment.md.
- Casos y fixtures: docs/mcp/conversation-cases.md y
  backend/cmd/mcp/testdata/demo.sql.
- Método y evidencia previa: docs/mcp/improvement-loop.md y docs/mcp/runs/.

Estado conocido, a revalidar: ocho herramientas, stdio y HTTP implementados,
Docker construido, pruebas críticas aprobadas. Una corrida gpt-6-luna inicial
obtuvo 17/20 en cumplimiento/ruta de herramientas y una posterior 20/20.
Tres variantes reservadas aprobaron. Son muestras únicas y no prueban producción.
El modo HTTP se agregó después de esa corrida completa. El hosting y cliente
finales aún no están elegidos; no hay URL publicada acreditada.

Final de este paso: lista breve de requisitos ya probados y pendientes con
commit, archivos, resultados y estado externo que los respaldan.

## 2. Verificar el servidor antes de mejorar descripciones

Ejecutar las pruebas Go y las integraciones de PostgreSQL en una base de prueba
con roles/esquemas temporales, siguiendo los comandos de la guía. Verificar que
los tests de DB se ejecutaron y no quedaron omitidos por falta de variables.

Cubrir ambas interfaces, stdio y HTTP:

- Descubrimiento y contrato de las ocho herramientas; entradas inválidas,
  campos extra, IDs inexistentes, estados y límites fuera del esquema.
- Filtros antes de paginar; páginas sin saltos ni duplicados con datos estáticos;
  fin de paginación, listas vacías y nombres repetidos sin selección arbitraria.
- Sprint e historia pertenecientes al proyecto solicitado. Como todos comparten
  datos, eso verifica alcance de la consulta, no autorización por usuario.
- Valores de métricas e informes idénticos al dominio Go, con ceros y decimales;
  conservación de los agregados y exclusión de registros individuales no expuestos.
- Herramientas de escritura inexistentes, privilegios PostgreSQL de lectura y
  comparación completa de datos antes/después. Fallos sin SQL ni secretos.
- HTTP: Origin válido, ausente e inválido; URL pública inválida; tamaño de
  solicitud acotado; endpoint /mcp y negociación con cliente SDK.
- Proceso: configuración incompleta falla sin revelar secretos, cierre por
  SIGTERM y conexión al PostgreSQL de prueba con credenciales de lectura.

Final de este paso: 100% de casos críticos aprobados y comandos reproducibles.
Una omisión, resultado sin evidencia o fallo crítico impide avanzar al despliegue.
Agregar regresiones proporcionales a los defectos reales; evitar tests que solo
comprueban que un texto descriptivo coincide literalmente con la implementación.

## 3. Preparar la evaluación conversacional

Usar un modelo pequeño disponible en el proveedor de OpenCode. Registrar su ID
real y nivel de razonamiento. Reutilizar gpt-6-luna con razonamiento bajo solo si
está disponible; si cambia el modelo, generar una nueva referencia y no atribuir
la diferencia al MCP. Nunca usar el modelo pequeño para implementar correcciones.

El agente principal diagnostica y modifica; el modelo evaluado solo actúa como
cliente. Darle una pregunta, herramientas/esquemas del servidor y la instrucción
neutral de responder con datos consultados. Usar contexto nuevo por pregunta y
repetición. Mantener la misma instrucción neutral entre versiones. El evaluado
no ve respuestas esperadas, fixtures, código ni calificaciones previas.

Congelar antes de ejecutar:

- Los 20 casos existentes como regresión, con tres repeticiones cada uno: 60
  conversaciones por corrida.
- Diez variantes reservadas para esa corrida, con tres repeticiones: 30 más.
  Prepararlas antes de las correcciones y evitar exponer sus respuestas al
  optimizador hasta calificar la versión candidata.
- Incluir en las variantes ambigüedad de nombres, ausencia de datos, páginas,
  alcance proyecto/sprint, ceros, decimales, pedidos de escritura e instrucciones
  maliciosas dentro de títulos o descripciones devueltos como datos.

Para las variantes que necesiten datos nuevos, crear otro fixture versionado en
una base aislada antes de evaluar. Mantenerlo igual para las versiones que se
comparen. Nunca editar datos durante una corrida. Al revelar un conjunto
reservado, pasa a ser conocido: para otra ronda usar variantes nuevas; registrar
esa diferencia y comparar regresiones fijas por separado del conjunto reservado.

El puente existente usa archivos para acceder al cliente SDK real. Puede servir
para diagnóstico. La aceptación del despliegue requiere además la conexión
nativa del cliente de IA elegido a la URL HTTP real; no presentar el puente como
esa evidencia.

## 4. Ejecutar y calificar sin alterar el conjunto a mitad de corrida

Por conversación conservar: pregunta, repetición, modelo, configuración,
commit y hashes de fuentes, transporte, fixture, llamadas en orden, argumentos,
respuestas de herramientas, respuesta final completa, dictamen y motivo.
Registrar intentos fallidos y recuperaciones. Separar fallos de infraestructura
como corrida inválida; no contarlos como aprobados ni descartarlos en silencio.

Un caso aprueba cuando satisface el pedido y sus hechos están respaldados por
las consultas: entidades, valores, unidades y alcance correctos; métricas desde
get_metrics o get_report; paginación conforme a lo pedido; sin afirmaciones de
escrituras que no ocurrieron. Una selección equivalente de herramientas puede
ser válida si obtiene la evidencia completa; no exigir una secuencia exacta.

Distinguir exactitud factual, elección de herramienta, cumplimiento del pedido
y fallo del servidor. Un intento de herramienta inventada que el servidor
rechaza no es una escritura. Cualquier escritura efectiva o exposición de datos
que contradiga el alcance de una herramienta es fallo crítico.

## 5. Límites de mejora y criterio de parada

| Límite | Regla |
|---|---|
| Corridas completas | Una referencia y hasta tres versiones mejoradas: máximo cuatro |
| Tamaño por corrida | 20 regresiones × 3 + 10 reservados × 3 = 90 conversaciones |
| Repruebas dirigidas | Hasta diez conversaciones adicionales por ciclo de mejora |
| Techo total | 360 conversaciones completas + 30 dirigidas = 390 |
| Conversación | Hasta 12 llamadas MCP y 120 segundos; exceso cuenta como fallo de resolución |
| Pruebas críticas | 100%; ningún fallo compensado por el promedio |
| Meta conversacional | Al menos 86/90 aprobadas y al menos 27/30 en reservados |
| Estabilidad por pregunta | Cada pregunta aprobada en al menos dos de sus tres repeticiones |
| Datos y métricas | Cero escrituras, filtraciones técnicas o valores de métricas incorrectos |
| Parada por estancamiento | Dos ciclos consecutivos sin aumentar aprobaciones en las 60 regresiones fijas |

Corregir una hipótesis verificable por cambio: nombre, descripción, esquema,
mensaje de error o defecto real. Repetir pruebas afectadas y casos fallidos;
luego ejecutar la corrida completa de la candidata. Conservar toda la evidencia.
No cambiar simultáneamente modelo, prompt, datos y servidor y atribuir el
resultado al servidor. Los límites son topes, no cantidades a consumir.

Detener mejoras cuando se cumplen todos los criterios de aceptación. Si se
alcanza el techo, se repite un bloqueo externo o hay estancamiento, entregar el
mejor candidato que conserve las pruebas críticas, con los fallos restantes y
el próximo paso concreto. No declararlo terminado si falta un criterio; no
bajar umbrales, agregar respuestas memorizadas ni ampliar iteraciones por cuenta
propia. El caso reservado visto no vuelve a presentarse como evaluación ciega.

## 6. Validar que alcanza para una demo

En entorno de prueba, con un cliente SDK y datos fijos, ejecutar hasta 100
consultas MCP en una corrida de carga con máximo cinco concurrentes. Medir
latencia de llamadas MCP por separado del tiempo de generación del modelo.
Meta propuesta: p95 menor o igual a tres segundos con el servicio caliente,
sin errores internos ni resultados incorrectos. Registrar arranque frío aparte.
Si no se cumple, diagnosticar y reportar; no iniciar una optimización masiva.

Los límites de carga no autorizan pruebas sobre hosts de terceros. En el
servicio propio publicado, hacer solo una prueba de humo de hasta diez llamadas,
sin DML/DDL, ataques de volumen ni cambios en los datos de exhibición.

## 7. Publicar cuando estén definidos hosting y cliente

Antes de la operación que dependa de ellos, pedir solo los datos que falten:
proveedor/cuenta/proyecto del hosting, PostgreSQL de demo, dominio o URL asignada
y cliente de IA utilizado. Avanzar en las tareas independientes mientras tanto.
Los secretos se configuran en el proveedor; no pedir pegarlos en un documento,
commitearlos ni incluirlos en los registros de herramientas.

Construir el Dockerfile desde la raíz del repo; preparar la base y el rol de
lectura; desplegar con las variables documentadas; usar el HTTPS del proveedor.
Este modo no tiene autenticación: todas las personas que alcancen la URL ven el
mismo conjunto de datos. Usar únicamente datos de exhibición. Un requerimiento
de datos privados cambia esta decisión y debe resolverse antes de publicar.

Verificar /healthz, conexión nativa a /mcp, descubrimiento, proyecto, métricas e
informe. /healthz solo verifica HTTP: acreditar la conexión a DB con una consulta
real. Guardar URL, versión/imagen desplegada y evidencia de esas consultas.
Si falla el arranque o lectura, corregir configuración antes de tocar el dominio.

## Entrega y Done

Actualizar docs/mcp/evaluation.md con resultados y limitaciones; conservar el
procedimiento y evidencia de las corridas, y actualizar el PR #107. Commits TASK
con chore(#105) o docs(#105). Mantener la arquitectura Go 1.24 y las reglas
existentes; toda regla nueva de negocio requiere la clasificación que indica
la guía del repo. No crear SDD/BDD para infraestructura ni implementar usuarios,
Planning Poker, métricas nuevas o refactors ajenos a la demo.

La demo está lista cuando pasa las pruebas críticas, alcanza los umbrales
conversacionales y funciona en el cliente final contra la URL publicada.
La TASK queda Done solo después de revisión de otra persona, aceptación de la
decisión arquitectónica y merge a main. Un agente no puede sustituir esa revisión
ni fabricar la evidencia. El informe final debe distinguir esos dos estados.
