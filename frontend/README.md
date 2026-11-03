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
  ordenado por `id DESC`; el endpoint está implementado por separado en HU-028
  ([PR #64](https://github.com/FaustinoDuran/seguimiento-medicion/pull/64)),
  que debe integrarse antes de usar el listado desde `main`.
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

## Tests automatizados

```bash
npm test           # ejecución única; falla con código distinto de cero
npm run test:watch # modo interactivo durante el desarrollo
npm run test:coverage
```

- Los tests usan Vitest con entorno `jsdom`, React Testing Library, user-event y
  jest-dom. No se ejecutan contra Go ni PostgreSQL; las pruebas de aceptación
  integrada siguen siendo manuales.
- HTTP se simula con MSW. La aplicación conserva su cliente real (`/api/*`), así
  que las pruebas ejercitan el mismo camino que el navegador.
- Una solicitud HTTP sin handler falla la prueba: no se toleran llamadas
  silenciosas a servicios reales.
- Los handlers compartidos viven en `frontend/src/test/handlers.ts`; cada prueba
  puede sobrescribirlos con `server.use(...)`.
- Las pruebas se guardan junto a la feature como `<archivo>.test.tsx`
  (por ejemplo `src/app/ConnectionPage.test.tsx`).
- Al terminar cada prueba se limpian el DOM y los handlers, así que el orden de
  ejecución no importa.
- GitHub Actions (`.github/workflows/frontend.yml`) corre `npm ci`, lint, build y
  tests en cada push y pull request.

Ejemplo mínimo:

```tsx
import { render, screen } from '@testing-library/react'
import { expect, test } from 'vitest'
import { ConnectionPage } from './ConnectionPage'

test('muestra la respuesta exitosa del servicio', async () => {
  render(<ConnectionPage />)
  expect(await screen.findByText('El servicio responde')).toBeInTheDocument()
})
```

## Matriz de escenarios cubiertos

Corresponde a TASK-015 (#62). HTTP siempre está simulado con MSW: estas pruebas
demuestran el comportamiento de la interfaz, **no** la aceptación integrada con
Go y PostgreSQL.

| Grupo (issue #62) | Escenario | Test | Archivo |
| --- | --- | --- | --- |
| 1 | Navegación de la portada a conexión | navega de la portada a la pantalla de conexión | `src/App.test.tsx` |
| 1 | Conexión por URL directa, con carga | abre la pantalla de conexión por URL directa | `src/App.test.tsx` |
| 1 | Ruta desconocida | muestra la página de dirección no encontrada | `src/App.test.tsx` |
| 1 | Respuesta exitosa y reintento tras error | (ya cubierto por TASK-013, sin duplicar) | `src/app/ConnectionPage.test.tsx` |
| 2 | Proyecto e historias, orden recibido, completadas y sin estimar | carga el proyecto y las historias en el orden entregado por la API | `src/features/backlog/BacklogPage.test.tsx` |
| 2 | Respuesta exitosa vacía | muestra el vacío solo cuando la consulta fue exitosa | ídem |
| 2 | Fallo de consulta y recuperación sin vacío falso | falla la consulta, no muestra vacío falso y se recupera al reintentar | ídem |
| 3 | Alta mínima: payload, confirmación, limpieza y recarga | crea una historia con los datos mínimos y refresca el listado | ídem |
| 3 | Bloqueo de envíos repetidos | no repite el alta mientras la primera está en curso | ídem |
| 4 | Opcionales: descripción, prioridad, story points y criterios | envía los campos opcionales con el alta | `src/features/backlog/StoryForm.test.tsx` |
| 4 | Agregar y quitar criterios | permite agregar y quitar criterios de aceptación | ídem |
| 5 | Título faltante, asociación accesible, foco y datos conservados | muestra el error de título junto al campo, enfoca y conserva los datos | ídem |
| 5 | Identificador faltante, asociación accesible, foco y datos conservados | muestra el error de identificador junto al campo, enfoca y conserva los datos | ídem |
| 5 | Identificador duplicado | muestra el error de identificador duplicado en su campo | ídem |
| 5 | Story points rechazados y detalles desplegados | muestra el rechazo de story points y despliega los detalles | ídem |
| 5 | Prioridad rechazada | muestra el rechazo de prioridad y despliega los detalles | ídem |
| 5 | Criterios vacíos rechazados | muestra el rechazo de criterios de aceptación vacíos | ídem |
| 5 | Error desconocido con mensaje general | muestra un mensaje general ante un error desconocido | ídem |
| 6 | Alta confirmada con recarga fallida y reintento solo de consulta | conserva la confirmación si falla la recarga y reintenta solo la consulta | `BacklogPage.test.tsx` |
| 7 | Conexión interrumpida durante el alta | advierte que el alta no está confirmada si se corta la conexión | `StoryForm.test.tsx` |
| 8 | Proyecto inexistente | muestra el proyecto inexistente sin formulario | `BacklogPage.test.tsx` |
| 8 | Proyecto cerrado de solo lectura | muestra un proyecto cerrado sin formulario de alta | ídem |
| 8 | Cierre concurrente durante el alta | conserva los datos cuando el proyecto se cierra durante el alta | ídem |
| 9 | Cambio de proyecto con consulta pendiente | no conserva borradores ni datos del proyecto anterior | ídem |
| 9 | Desmontaje con consulta pendiente | descarta la respuesta de una consulta pendiente tras el desmontaje | ídem |

### Pendiente y fuera de alcance

- **Aceptación integrada** con Go y PostgreSQL, y el contrato real de listado
  (`HU-028`): todas las pruebas usan MSW.
- **Navegación** desde el detalle de proyecto al backlog (`HU-024`): hoy solo se
  prueba el acceso por URL directa.
- **UI de sprints** (`HU-027`): queda para una TASK propia, porque la issue #62
  la excluye explícitamente.
- Mensajes de error de `invalid json` e `initial story status must be todo` no
  tienen prueba todavía; son rechazos de la API sin flujo propio en la UI.
- Los **snapshots** no se usan como evidencia; se comprueba comportamiento
  visible y asociación accesible.



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
