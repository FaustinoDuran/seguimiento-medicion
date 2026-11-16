# Datos compartidos entre UI y MCP — TASK-025 / #116

Esta guía describe el despliegue que reemplaza la base separada de TASK-024.
La transición se aplica al mergear el PR y ejecutar CI/CD. El MCP sigue siendo
de solo lectura y comparte la base que muestra la UI, por pedido del responsable.

## TASK-025: datos compartidos entre UI y MCP

URL: **https://seguimiento-medicion.aedestec.com/mcp**.
La base utilizada es ahora la misma que consulta la API y muestra la UI.
El usuario solicitó este cambio para comparar los datos desde ambos clientes.
El endpoint público puede consultar todos los proyectos de esa base; no limita
la lectura al proyecto cargado por el script. Usar esta instalación con datos
adecuados para exhibición. No se incorpora autenticación ni aislamiento.

El flujo de main primero actualiza la aplicación (y su esquema), luego ejecuta
deploy/deploy-mcp.sh con los mismos secrets SSH existentes:

1. Genera MCP_READER_PASSWORD hexadecimal en .env.mcp solo si el archivo no existe.
   Conserva las credenciales y corrige el dominio antiguo con cero si corresponde.
2. Consulta el nombre real de la base usando las variables del contenedor
   PostgreSQL de producción; no asume IDs ni usa su contraseña desde el runner.
3. Ejecuta scripts/seed-project.sh sobre esa base.
4. Prepara mcp_app_reader con SELECT sobre las siete tablas necesarias.
   Envía su contraseña por stdin, sin argumentos del proceso ni logs.
5. Despliega MCP conectado a application-database:5432, alias del PostgreSQL
   principal en seguimiento-medicion-prod_default, con el nombre real de la base.
6. Comprueba salud y guarda .deploy-mcp-sha; CI compara proyecto y métricas entre
   la API pública y el MCP mediante scripts/mcp-http-smoke.py.

No se vuelven a levantar una API auxiliar ni una base de demo. El contenedor y
volumen de la demo anterior se conservan; pueden aparecer como huérfanos en
Compose. No ejecutar down -v ni borrar esos recursos para hacer esta transición.
MCP_ADMIN_PASSWORD puede permanecer en .env.mcp para una reversión histórica,
pero el despliegue actual no lo utiliza. No editar MCP_READER_PASSWORD fuera
del despliegue: el rol se aprovisiona con esa credencial en cada ejecución.
El nombre de la base debe usar letras, números y guion bajo para la URL.

## Script de carga

Desde el checkout del VPS, después de desplegar la API:

```bash
sh scripts/seed-project.sh SHA_DESPLEGADO
```

Para desarrollo local, con psql instalado y el esquema inicializado por la API:

```bash
DATABASE_URL='postgres://postgres:postgres@localhost:5433/seguimiento_medicion?sslmode=disable' \
  sh scripts/seed-project.sh
```

Con un SHA explícito el script siempre usa el contenedor de producción, incluso
si la sesión exporta DATABASE_URL. Sin SHA, usa DATABASE_URL y el cliente psql.
CI ejecuta la carga automáticamente; no hace falta una ejecución manual inicial.

scripts/seed-project.sql usa una transacción y bloqueo asesor para serializar
las cargas. Inserta con IDs generados por PostgreSQL y registra ui-mcp-v1 en
demo_seed.runs. Repetirlo conserva los datos existentes y las ediciones de la
UI, incluso si se renombra el proyecto. No hace TRUNCATE, DELETE ni UPDATE
sobre tablas del producto. El rol MCP no tiene acceso al registro interno.
Si falla una inserción, se revierte todo el lote. No se recuperan números de
secuencia consumidos en intentos fallidos: no se debe asumir un ID concreto.

Buscar en la UI el proyecto **Demo UI y MCP - Tienda**. El script imprime su ID
real. Datos iniciales:

| Dato | Contenido |
|---|---|
| Proyecto | Activo, 01/01/2026–31/12/2026 |
| Integrantes | Ana Demo y Luis Demo |
| DEMO-001 | Consultar catálogo, done, alta, 8 puntos |
| DEMO-002 | Administrar carrito, done, alta, 5 puntos |
| DEMO-003 | Confirmar compra, todo, media, 3 puntos |
| DEMO-004 | Notificar estado del pedido, todo, baja, 2 puntos |
| Sprint cerrado | Catálogo y carrito disponibles; DEMO-001 y DEMO-002 |
| Sprint abierto | Completar el flujo de compra; DEMO-003 |
| Esfuerzo | Tres registros que suman 6,5 horas |
| Defectos | Dos detectados; uno resuelto y uno abierto |

Métricas de proyecto esperadas antes de editar: 16 puntos planificados,
13 completados, velocidad 13, desviación de 3 puntos, 50 % de historias
completadas y 6,5 horas reales. Las horas estimadas permanecen en cero según
el servicio existente. Estos valores no son cálculos nuevos del script.

## Ruta HTTPS y operación

El Nginx HTTPS del VPS reenvía / al frontend; frontend/nginx.conf envía /mcp
y /mcp/healthz al servicio mcp:8080 por la red de Docker. No cambia el dominio
ni el certificado. El puerto 3008 solo escucha en loopback; PostgreSQL no
publica puertos.

```bash
curl --fail https://seguimiento-medicion.aedestec.com/mcp/healthz
python3 scripts/mcp-http-smoke.py
```

/healthz solo comprueba el proceso. El smoke verifica negociación, ocho
herramientas y consultas, y contrasta proyecto y métricas con /api. El proyecto
se descubre dinámicamente, sin depender del ID 1. Usa el nombre de la carga
si está disponible y, si fue renombrado, un proyecto de los listados.
La prueba no sustituye la validación manual de UI y ChatGPT.

El despliegue vuelve al SHA anterior si falla Compose o el healthcheck local.
Un fallo de permisos o carga aborta antes de actualizar MCP. El rollback de
imagen conserva la base compartida y los datos agregados; no restaura la base
separada ni cambia el esquema. Para revertir también el origen, recuperar el
Compose y el script anteriores a TASK-025 con su configuración de la demo.
El volumen anterior queda preservado para esa operación.

No se garantiza que el primer despliegue tenga una imagen previa. Si falla el
smoke público, CI queda en rojo y se revisa conectividad/proxy; no se revierte
una configuración externa al contenedor automáticamente.

## Prueba del cliente en ChatGPT

En ChatGPT web, usar Plugins → + → Add custom MCP server; si la cuenta todavía
muestra la interfaz anterior, usar Configuración → Apps → Create. Configurar:

- Nombre: Seguimiento y medición.
- URL: https://seguimiento-medicion.aedestec.com/mcp
- Autenticación: ninguna.

Revisar las herramientas descubiertas, crear e instalar el plugin.
La disponibilidad depende del plan y permisos del espacio de trabajo; verificar
la [guía oficial vigente](https://help.openai.com/en/articles/12584461-connecting-custom-mcp-servers-in-chatgpt).

En un chat nuevo seleccionar el plugin con @ y pedir:

1. Listá los proyectos usando Seguimiento y medición.
2. Mostrá las historias pendientes de Demo UI y MCP - Tienda.
3. Consultá las métricas del proyecto y de su sprint cerrado.
4. Mostrá su informe y explicá en qué datos te basaste.

Comprobar que descubre ocho herramientas y que los valores coinciden con el
fixture. Registrar URL, fecha, SHA, cliente/modelo, llamadas y resultados en
evaluation.md. Una prueba del SDK no acredita una prueba de ChatGPT.
No incluir credenciales en la evidencia.
