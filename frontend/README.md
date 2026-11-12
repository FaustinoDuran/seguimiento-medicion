Frontend de presentación. El README del monorepo y `docs/guia-de-trabajo.md` mandan.

```bash
npm install
npm run dev
```

Levantar la API Go en otra terminal con `go run ./cmd/api` desde `backend/`.
Abrir la dirección que informa Vite (por defecto `http://localhost:5173`).

- `/`: portada del shell.
- `/connection`: comprueba `GET /health`, con carga, error y reintento.
- `/projects/:projectId/backlog`: alta y listado (HU-026), edición de prioridad,
  estado y estimaciones existentes (HU-006).
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
- **Navegación** desde el detalle de proyecto al backlog y a sprints (`HU-024`):
  hoy solo se prueba el acceso por URL directa.
- Mensajes de error de `invalid json` e `initial story status must be todo` no
  tienen prueba todavía; son rechazos de la API sin flujo propio en la UI.
- Los **snapshots** no se usan como evidencia; se comprueba comportamiento
  visible y asociación accesible.

## Matriz de escenarios: UI de sprints

Corresponde a TASK-016 (#69) y a los escenarios de
`features/HU-027-ui-sprints.feature`. Igual que en backlog, HTTP está simulado
con MSW: estas pruebas **no** demuestran la aceptación integrada con Go y
PostgreSQL.

| AC | Escenario | Test | Archivo |
| --- | --- | --- | --- |
| AC-027-06 | Sprints anteriores con goal, estado y fecha | lista los sprints cerrados del proyecto con goal, estado y fecha | `SprintListPage.test.tsx` |
| AC-027-06 | Proyecto sin sprints anteriores | muestra el vacío de sprints anteriores tras una consulta exitosa | ídem |
| AC-027-07 | Ruta de sprints por URL directa | abre la ruta de sprints por URL directa con formulario y listado | ídem |
| AC-027-10 | Fallo de consulta de la lista y reintento | falla la consulta de sprints, no muestra vacío falso y se recupera al reintentar | ídem |
| AC-027-10 | Proyecto con formato inválido | muestra el error cuando el proyecto llega con un formato inválido | ídem |
| AC-027-10 | Fecha de creación ilegible | muestra una fecha desconocida cuando el sprint no trae una fecha válida | ídem |
| AC-027-09 | Proyecto inexistente | muestra el proyecto inexistente sin formulario de alta | ídem |
| AC-027-01 | Crear un Sprint con goal y quedar en su detalle | crea un sprint con goal y queda en el detalle del sprint creado | ídem |
| AC-027-01 | Alta deshabilitada mientras se procesa | deshabilita el alta y no repite el envío mientras está en curso | ídem |
| AC-027-02 | Rechazo de goal vacío junto al campo | muestra el rechazo de goal vacío junto al campo | ídem |
| AC-027-07 | Detalle por URL directa | abre el detalle por URL directa con goal, estado e historias | `SprintDetailPage.test.tsx` |
| AC-027-07 | Recargar el detalle | mantiene el sprint y sus historias al recargar la pantalla | ídem |
| AC-027-03 | Asignar una historia al Sprint | asigna una historia al sprint y deja de ofrecerla en el selector | ídem |
| AC-027-04 | Completar una historia asignada | completa una historia asignada y la mantiene en el sprint | ídem |
| AC-027-09 | Historia ya asignada a otro Sprint | muestra el rechazo cuando la historia ya está asignada a un sprint | ídem |
| AC-027-09 | Sprint cerrado durante la asignación | muestra el rechazo cuando el sprint se cerró durante la asignación | ídem |
| AC-027-05 | Cerrar el Sprint con confirmación | cierra el sprint tras confirmar y quita las acciones | ídem |
| AC-027-05 | Cancelar el cierre | cancelar el cierre no llama al servicio | ídem |
| AC-027-08 | Sprint cerrado de solo lectura | muestra un sprint cerrado sin asignación ni acciones | ídem |
| AC-027-09 | Sprint inexistente o de otro proyecto | muestra el sprint no encontrado cuando pertenece a otro proyecto | ídem |
| AC-027-10 | Sprint con formato inválido | muestra el error cuando el sprint llega con un formato inválido | ídem |
| AC-027-10 | Conexión cortada al asignar | advierte que no pudo confirmar el resultado cuando se corta la conexión | ídem |
| AC-027-10 | Fallo de consulta del detalle y reintento | muestra el error de consulta y se recupera al reintentar | ídem |
| AC-027-10 | Proyecto sin historias | avisa que el proyecto no tiene historias en lugar de mostrar un selector vacío | ídem |
| AC-027-10 | Todas las historias ya asignadas | avisa cuando todas las historias del proyecto ya están en el sprint | ídem |
| AC-027-02, 03 | Tabla de historias del Sprint | lista las historias con identificador, título y estado | `SprintStories.test.tsx` |
| AC-027-03, 04 | Acción de completar solo en pendientes | completa solo las historias pendientes y avisa cuál se está completando | ídem |
| AC-027-04 | Acción en curso deshabilitada | deshabilita la acción de la historia que se está completando | ídem |
| AC-027-08 | Sprint cerrado sin acciones | no ofrece acciones cuando el sprint no lo permite | ídem |
| AC-027-03, 04, 05, 06 | Ciclo Scrum completo desde la UI | recorre el ciclo Scrum completo desde la interfaz | `sprintCycle.test.tsx` |

### Pendiente de la UI de sprints

- Aceptación integrada con Go y PostgreSQL: el ciclo completo se prueba contra un
  backend simulado en memoria, no contra la API real.
- Navegación hacia sprints desde otras pantallas (`HU-024`).



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

## Matriz de escenarios: edición de backlog

HU-006 (#15): pruebas con HTTP simulado por MSW. La persistencia y las reglas
se prueban en Go con PostgreSQL real; no se declara validación manual en navegador.

| Criterios | Comportamiento | Archivo en `src/features/backlog/` |
| --- | --- | --- |
| AC-006-01, 08, 17 | Formulario precargado, envío parcial y listado confirmado | `BacklogPage.test.tsx` |
| AC-006-17 | Recarga, otras ediciones y alta bloqueadas mientras guarda | `BacklogPage.test.tsx` |
| AC-006-17 | Cancelar, guardar sin cambios, evitar duplicados y descartar respuesta tras desmontaje | `StoryEditForm.test.tsx` |
| AC-006-02, 04, 05, 06, 07, 14, 17 | Avisos en español y borrador conservado ante errores de API o conexión | `StoryEditForm.test.tsx` |
| AC-006-15 | Enviar corrección de puntos y finalización juntas | `StoryEditForm.test.tsx` |
| AC-006-07, 18 | Proyecto cerrado de consulta y explicación de restricciones de puntos | `BacklogPage.test.tsx`, `StoryEditForm.test.tsx` |

## Matriz de escenarios: UI de esfuerzo y defectos

Corresponde a HU-030 (#85) y a los escenarios de
`features/HU-030-ui-esfuerzo-defectos.feature`. HTTP está simulado con MSW:
estas pruebas **no** demuestran la aceptación integrada con Go y PostgreSQL.

| AC | Escenario | Test | Archivo |
| --- | --- | --- | --- |
| AC-030-01 | Alta de esfuerzo con integrante, fecha, actividad y horas | registra esfuerzo con integrante, fecha, actividad y horas | `Measurement.test.tsx` |
| AC-030-04 | Horas rechazadas por Go | muestra en español el rechazo de las horas y conserva los datos | ídem |
| AC-030-02 | Alta de defecto abierto y visible en la lista | registra un defecto y lo muestra abierto en la lista | ídem |
| AC-030-04 | Descripción rechazada por Go | muestra en español el rechazo de la descripción | ídem |
| AC-030-03 | Resolver un defecto abierto | resuelve un defecto con su sprint de resolución | ídem |
| AC-030-04 | Conflicto al resolver | muestra en español el conflicto al resolver | ídem |
| AC-030-05 | Sprints abiertos y cerrados para elegir | ofrece los sprints abiertos y cerrados del proyecto | ídem |
| AC-030-06 | Proyecto sin defectos | muestra el estado vacío cuando no hay defectos | ídem |
| AC-030-07 | Fallo de carga de defectos y reintento | un fallo al cargar defectos se avisa y no se muestra como vacío | ídem |
| AC-030-07 | Faltan integrante o historia | no envía el esfuerzo sin integrante ni historia | ídem |
| AC-030-08 | Sin cálculo de métricas en React | no calcula horas ni desviaciones en la pantalla | ídem |
| — | Entrada desde el detalle del proyecto | muestra el enlace desde el detalle del proyecto | ídem |

### Pendiente de la UI de esfuerzo y defectos

- Aceptación integrada con Go y PostgreSQL: las pruebas usan HTTP simulado.
- `GET /projects/{id}/sprints` y `GET /projects/{id}/defects` los agrega esta HU
  en Go; el resto de la API viene de HU-016, HU-017 y HU-018.
