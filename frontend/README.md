Frontend de presentación. El README del monorepo y `docs/guia-de-trabajo.md` mandan.

```bash
npm install
npm run dev
```

Levantar la API Go en otra terminal con `go run ./cmd/api` desde `backend/`.
Abrir la dirección que informa Vite (por defecto `http://localhost:5173`).

- `/`: portada del shell.
- `/connection`: comprueba `GET /health`, con carga, error y reintento.
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
