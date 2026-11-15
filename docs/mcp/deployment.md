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
docker build -f backend/Dockerfile -t seguimiento-mcp:demo .
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

Estado: paquete preparado; hosting, dominio y cliente aún sin seleccionar.
No hay evidencia de un despliegue publicado ni de una prueba contra esa URL.
