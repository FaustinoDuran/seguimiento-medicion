Frontend de presentación. El README del monorepo y `docs/guia-de-trabajo.md` mandan.

```bash
npm install
npm run dev
```

Levantar la API Go en otra terminal con `go run ./cmd/api` desde `backend/`.
Abrir la dirección que informa Vite (por defecto `http://localhost:5173`).

- `/`: portada del shell.
- `/connection`: comprueba `GET /health`, con carga, error y reintento.
- `/projects/:projectId/backlog`: formulario y listado del backlog (HU-026).
  Requiere `GET /projects/:projectId` y `POST /projects/:projectId/stories`.
  El listado espera `GET /projects/:projectId/stories` con `{ "stories": [...] }`,
  ordenado por `id DESC`; este endpoint todavía no está implementado en Go.
  Las rutas de detalle y listado de proyectos y su enlace de entrada al backlog
  se integran con HU-024; mientras tanto el backlog admite URL directa.
- Cualquier otra ruta muestra una página de dirección no encontrada.

El navegador llama a `/api/*` en el mismo origen. El proxy de desarrollo de
Vite reenvía a `http://localhost:8080/*`, quitando `/api`; no hace falta CORS.
Por ejemplo, `/api/health` se convierte en `/health` en Go.

Verificación local:

```bash
npm run lint
npm run build
curl -i http://localhost:5173/api/health
```

Con la API levantada, el curl devuelve `200` y `{"status":"ok"}`. Para
comprobar la interfaz, navegar de Inicio a Conexión, recargar `/connection`,
usar Atrás/Adelante y abrir una ruta inexistente. Detener la API y volver a
comprobar permite verificar el estado de error y luego recuperar la conexión.

El proxy es de desarrollo. Al desplegar, el servidor debe reenviar `/api/*`
a Go quitando el prefijo y servir `index.html` para las rutas de la SPA.
`npm run preview` solo previsualiza el build y no reemplaza esa configuración.

Verificación de HU-026, cuando esté disponible el contrato de listado:

1. Abrir `/projects/1/backlog` con un proyecto existente. Comprobar carga,
   listado y vacío; un error de consulta debe ofrecer reintento sin mostrar vacío.
2. Crear con identificador y título; comprobar confirmación, limpieza y recarga.
   Enviar sin título debe mostrar el error en español junto al campo y conservar
   el identificador. Probar también identificador duplicado.
3. Desplegar «Más detalles»: descripción, prioridad, story points y criterios
   que se pueden agregar o quitar. Sin estimación se muestra «Sin estimar».
4. Simular alta exitosa y recarga fallida: el reintento solo consulta, no crea
   otra historia. Durante el alta no se puede repetir el envío.
5. Comprobar proyecto inexistente, cerrado y cierre durante el alta. El cierre
   concurrente conserva el formulario y bloquea nuevos envíos.
6. Comprobar fallos de conexión: el alta no confirmada sugiere revisar el listado
   antes de reenviar. Cambiar de proyecto no conserva el borrador anterior.
7. Revisar la pantalla en móvil y escritorio y navegar con teclado.
