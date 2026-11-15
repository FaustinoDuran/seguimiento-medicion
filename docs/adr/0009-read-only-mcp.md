# ADR 0009 — MCP de consultas de solo lectura

- Estado: propuesta
- Fecha: 2026-11-14
- Issue: [#105](https://github.com/FaustinoDuran/seguimiento-medicion/issues/105)

## Contexto

La TASK solicita consultar información existente desde un cliente de IA.
El backend ya tiene servicios de proyectos, backlog, sprints, métricas e
informes. No existen sesiones ni roles de autorización: un integrante es un
registro del proyecto, no una identidad autenticada.

El ejecutable HTTP actual ejecuta EnsureSchema al iniciar. Reutilizar ese
arranque impediría garantizar una conexión PostgreSQL con permisos de lectura.

## Decisión propuesta

Crear un ejecutable Go independiente en backend/cmd/mcp, con transporte stdio
para una demo local controlada. El cliente inicia el proceso y administra sus
credenciales. Por pedido posterior del responsable, agregar también HTTP para
una demo académica remota con un único conjunto de datos compartido, sin cuentas
ni separación por persona. Ese endpoint es público: usar exclusivamente datos
preparados para exhibición. El despliegue concreto queda pendiente de elegir
hosting y cliente, y debe servir HTTPS mediante el proxy del proveedor.

El modo HTTP usa Streamable HTTP stateless en /mcp, valida Origin cuando esté
presente contra MCP_PUBLIC_URL y conserva las mismas herramientas de lectura.
No habilita CORS abierto ni interpreta cabeceras de proxy como autorización.
El modo stdio sigue siendo el predeterminado; HTTP se habilita explícitamente.
Registrar la versión del protocolo y la dependencia MCP elegidas durante la
implementación, verificando compatibilidad con Go 1.24.

El adaptador en backend/internal/mcp recibe interfaces con métodos de lectura
de los servicios existentes. No importa handlers ni implementaciones SQL de
otros módulos, no calcula métricas y no ejecuta migraciones o EnsureSchema.
El ensamblado de dependencias vive en cmd/mcp.

Usar una conexión PostgreSQL exclusiva con SELECT sobre las tablas necesarias,
sin permisos de escritura ni creación de esquema. La preparación de permisos
se ejecuta por separado con una cuenta administradora; no al arrancar el MCP.

Catálogo inicial propuesto:

| Herramienta | Fuente existente | Alcance |
|---|---|---|
| list_projects | projects.Service.List | búsqueda por nombre y paginación |
| get_project | projects.Service.Get | detalle por ID |
| list_stories | backlog.Service.List | proyecto, estado, prioridad y paginación |
| get_story | backlog.Service.Get | ID dentro del proyecto |
| list_sprints | sprints.Service.List | proyecto, estado y paginación |
| get_sprint | sprints.Service.Get / ListStories | detalle e historias del sprint |
| get_metrics | metrics.Service.ForProject / ForSprint | métricas existentes |
| get_report | reports.Service.ForProject / ForSprint | informe estructurado existente |

Usar IDs numéricos como referencias inequívocas. La búsqueda devuelve IDs y
nombres para elegir; nunca selecciona automáticamente entre nombres repetidos.
No agregar rangos temporales que los servicios de métricas no soportan.
Validar pertenencia al proyecto al consultar un sprint por proyecto.

Los listados tendrán orden estable por ID, límite predeterminado de 20 y máximo
de 100, y cursor validado. Los filtros se aplican antes de paginar. En esta
demo los servicios cargan colecciones completas: el límite acota la respuesta,
no garantiza todavía un límite de trabajo en PostgreSQL.

Las respuestas identifican entidad, proyecto y alcance de métricas. Los errores
de argumentos, entidad ausente e infraestructura se distinguen sin exponer SQL,
credenciales ni trazas. El contenido de historias se devuelve como datos, sin
interpretarlo como instrucciones ni ejecutar acciones incluidas en su texto.

## Evaluación y cierre

Versionar casos y fixtures junto a backend/internal/mcp. Cada caso declara
entradas, herramientas esperadas, resultado verificable y criticidad. Cubrir
descubrimiento, parámetros inválidos, entidades inexistentes, referencias
ambiguas, filtros combinados, listas vacías, paginación completa sin duplicados,
pertenencia de sprint, errores internos y herramientas de escritura inexistentes.

Comparar métricas e informes con los servicios Go sobre los mismos fixtures,
incluyendo ceros y valores no disponibles. Complementar con resultados esperados
fijos para evitar que una comparación entre servicios oculte datos incorrectos.

Probar el proceso stdio desde un cliente MCP real. En PostgreSQL de prueba,
comparar datos antes y después de las consultas y comprobar que el rol MCP
rechaza INSERT, UPDATE, DELETE y DDL. Intentar escrituras solo en esa base de
prueba, sin datos reales.

La suite crítica debe aprobar el 100%; cualquier fallo bloquea el cierre.
Para la evaluación conversacional, versionar al menos 20 preguntas, hechos
esperados y rúbrica. Registrar cliente, modelo, configuración y herramientas
disponibles. Una respuesta aprueba si selecciona las entidades correctas,
presenta hechos sustentados, conserva unidades y valores no disponibles, y
no inventa datos. Exigir al menos 90% de aprobación, reportando numerador y
denominador. Registrar el resultado del cliente de IA por separado del contrato
determinista del servidor.

El PR debe incluir comandos reproducibles, resultados y limitaciones. El cierre
requiere revisión y merge según la guía; este ADR no acredita implementación.

## Consecuencias

La entrega sirve para demo local o remota con datos explícitamente preparados
para exhibición. No proporciona aislamiento por persona ni autenticación. El
endpoint remoto comparte el mismo conjunto de datos con todos los visitantes.
Agregar cuentas o reglas nuevas de autorización requiere una decisión posterior
y la clasificación de trabajo correspondiente.

Se conserva el stack y las reglas existentes. No se incluyen operaciones de
Planning Poker, registros individuales de esfuerzo o defectos ni exportación
PDF; los agregados que ya integran métricas e informes conservan su significado.
