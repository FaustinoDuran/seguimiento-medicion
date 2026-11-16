# Demo académica remota — #105

Alcance pedido por el responsable: mostrar el MCP a profesores usando una URL
HTTPS y un único conjunto de datos. No hay cuentas ni separación por usuario.
El endpoint no tiene autenticación: cualquier persona que pueda acceder a la
URL puede consultar los datos de demo. Preparar una base exclusiva con datos
para exhibición. No conectar una base con información privada a este modo.

## Paquete preparado

- Ejecutable Go con MCP_TRANSPORT=http y Streamable HTTP stateless en /mcp.
- Mismas ocho herramientas y credenciales PostgreSQL de lectura que stdio.
- Puerto configurable por PORT, predeterminado 8080.
- /healthz confirma que el proceso HTTP atiende. No mide disponibilidad de DB.
- MCP_PUBLIC_URL define el origen HTTPS público, sin /mcp ni rutas adicionales.
  Origin ajeno se rechaza; clientes nativos sin Origin pueden conectarse.
- El proveedor debe terminar TLS y enviar /mcp al contenedor sin reescribirlo.
- Contenedor sin usuario root, sin código fuente ni archivos .env en runtime.
- Cierre por SIGTERM con hasta diez segundos para atender solicitudes en curso.

## Construcción y ejecución

Desde la raíz del repo:

```bash
docker build -f backend/Dockerfile.mcp -t seguimiento-mcp:demo .
```

Preparar PostgreSQL y el rol mcp_reader según la [guía](README.md). El MCP no
inicializa esquemas ni carga datos. En el hosting configurar como secretos o
variables del servicio:

```text
MCP_TRANSPORT=http
PORT=8080
MCP_PUBLIC_URL=https://mcp-demo.example
MCP_DATABASE_URL=postgres://mcp_reader:PASSWORD@DB_HOST:5432/seguimiento_medicion?sslmode=verify-full
```

La URL anterior es un ejemplo; usar la cadena TLS y certificados que corresponda
al proveedor PostgreSQL. No usar el usuario administrador. El arranque rechaza
permisos efectivos de escritura o creación de tablas/esquemas.

Prueba con un proxy HTTPS propio o del proveedor:

```bash
docker run --rm -p 8080:8080 --env-file /ruta/privada/mcp-demo.env seguimiento-mcp:demo
```

Configurar en el cliente un servidor MCP remoto **Streamable HTTP** con URL:

```text
https://mcp-demo.example/mcp
```

El comando exacto depende del cliente de IA elegido. No es una API REST de
herramientas ni un transporte SSE legado. El cliente negocia el protocolo MCP
con el SDK oficial del servidor.

## Validación en el hosting

1. Confirmar HTTPS y GET /healthz con estado 200.
2. Conectar el cliente real, descubrir ocho herramientas y leer el proyecto Demo.
3. Consultar get_metrics y get_report y contrastar con los datos preparados.
4. Comprobar que una herramienta de escritura inventada es rechazada y que los
   datos permanecen iguales. Mantener las pruebas destructivas de permisos en
   PostgreSQL de prueba; no ejecutar la suite contra datos de producción.
5. Registrar URL de servicio, commit de imagen, cliente, fecha y resultado en
   la evidencia de la demo. Mantener los secretos fuera del reporte.

Estado: TASK-024 / #111 incorpora publicación de imagen y despliegue automático
de demo separado de la aplicación. Falta ejecutar CI/CD desde main y registrar
la prueba pública y de ChatGPT. No hay una URL pública acreditada.
El Dockerfile de la API sigue en backend/Dockerfile; el del MCP está en
backend/Dockerfile.mcp y usa el contexto de la raíz del repositorio.
No hay evidencia de un despliegue publicado ni de una prueba contra esa URL.

## TASK-024: despliegue automático desde main

URL configurada: **https://seguimiento-medicion.aedestec.com/mcp**.
Es la URL prevista por la configuración; no se considera publicada hasta que
el workflow y la prueba de ChatGPT la validen.

CI publica las tres imágenes con el SHA del commit. Después del despliegue
habitual de la aplicación ejecuta deploy/deploy-mcp.sh en el VPS con los mismos
secrets SSH existentes. No necesita un nuevo dominio ni secretos adicionales.
El script requiere Docker Compose, openssl y curl, disponibles en el servidor.

En el primer despliegue:

1. Genera dos contraseñas aleatorias en .env.mcp con permisos privados. El archivo
   queda en el VPS y no se imprime ni se envía a GitHub. Se conserva entre releases.
2. Levanta PostgreSQL en un proyecto y volumen exclusivos de demo.
3. Ejecuta la API auxiliar sin puertos públicos para inicializar el esquema.
4. Crea mcp_reader si falta, concede lectura y carga el fixture solo si no hay
   proyectos. El fixture aborta si detecta otras tablas con datos: no borra nada.
5. Detiene la API auxiliar y despliega el MCP con healthcheck y reinicio automático.
6. Guarda el SHA en .deploy-mcp-sha, independiente del estado de la aplicación.

En actualizaciones se conservan credenciales y datos; se vuelve a inicializar el
esquema de forma idempotente. No editar las contraseñas sin rotar también los
roles de PostgreSQL. Se puede preparar .env.mcp de antemano usando
 deploy/mcp.env.example; si existe, no se reemplaza. Cambios de esquema deben
ser compatibles con el SHA anterior: el rollback no revierte tablas ni datos.

Excepción: si .env.mcp conserva el dominio original con un cero
(`seguimiento0-medicion.aedestec.com`), el script corrige únicamente ese origen
a `seguimiento-medicion.aedestec.com`. No cambia las contraseñas ni otros dominios.

## Ruta HTTPS

El Nginx HTTPS del host sigue enviando / al frontend como en el despliegue
existente. frontend/nginx.conf agrega rutas exactas /mcp y /mcp/healthz hacia
el servicio MCP, usando DNS interno de Docker y desactivando buffering.
El servicio MCP se conecta también a la red externa
seguimiento-medicion-prod_default, creada por el Compose principal.
La resolución dinámica permite que la SPA arranque aunque el MCP aún no exista.
No se modifica el certificado ni la configuración del Nginx del VPS.

Este recorrido requiere que el proxy del host reenvíe /mcp al frontend, como
hace el bloque location / documentado en README. Si el VPS tiene una ruta
especial que lo intercepta, corregirla antes de acreditar la publicación.

El puerto 3008 solo escucha en loopback. PostgreSQL no publica puertos. El
endpoint no tiene autenticación: usar exclusivamente datos de exhibición.

## Validación y reversión

CI ejecuta scripts/mcp-http-smoke.py después del despliegue. Comprueba HTTPS,
negociación JSON-RPC, el conjunto exacto de ocho herramientas y consultas de
proyectos, historias pendientes, métricas e informe. No acepta una página HTML
con estado 200 como prueba del protocolo. La prueba no sustituye ChatGPT ni
el contraste manual de los valores del fixture.

El proyecto Demo tiene ID 1, historia completada de 8 puntos e historia pendiente
de 3 puntos; sprint cerrado 10 y abierto 11. El fixture no carga esfuerzo:
no atribuirle las 2,5 h de otras pruebas que sí agregan registros.

```bash
curl --fail https://seguimiento-medicion.aedestec.com/mcp/healthz
python3 scripts/mcp-http-smoke.py
```

/healthz solo comprueba el proceso HTTP. El script genérico de despliegue vuelve
al SHA anterior si falla Compose o el healthcheck local. El bootstrap fallido
aborta antes de actualizar el MCP; la API auxiliar se detiene también en ese caso.
Si falla el smoke público, CI falla y se debe revisar proxy y conectividad: no
se revierte automáticamente una configuración externa al contenedor.

Reversión manual desde /opt/seguimiento-medicion:

```bash
sh deploy/deploy-mcp.sh SHA_ANTERIOR
```

El primer despliegue no tiene imagen anterior. Para detener la demo sin borrar
los datos, ejecutar docker compose --env-file .env.mcp -f docker-compose.mcp.yml
down (sin -v). El siguiente deploy a main vuelve a levantarla.

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
2. Mostrá las historias pendientes del proyecto 1.
3. Consultá las métricas del proyecto 1 y del sprint 10.
4. Mostrá el informe del proyecto 1 y explicá en qué datos te basaste.

Comprobar que descubre ocho herramientas y que los valores coinciden con el
fixture. Registrar URL, fecha, SHA, cliente/modelo, llamadas y resultados en
evaluation.md. Una prueba del SDK no acredita una prueba de ChatGPT.
No incluir credenciales en la evidencia.
