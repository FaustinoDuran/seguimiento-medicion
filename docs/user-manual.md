# Manual breve de usuario

Aplicación web para crear un proyecto, estimar historias, seguir un Sprint y
consultar métricas. Las cuentas y los permisos no forman parte de esta versión:
quien abre la aplicación puede usar todas las pantallas.

## Requisitos

- API Go en `http://localhost:8080`.
- PostgreSQL de desarrollo, como describe el [README](../README.md).
- Frontend en `http://localhost:5173`.

Desde `backend/`:

```bash
docker compose up -d
go run ./cmd/api
```

Desde `frontend/`:

```bash
npm install
npm run dev
```

Vite reenvía `/api` a la API. Abrí la dirección que informa Vite.

## Recorrido inicial

1. Abrí la aplicación: el inicio lleva directamente a **Proyectos**.
2. Entrá a **Proyectos** y creá el primero o abrí uno existente.
3. Si una operación falla, **Conexión** permite comprobar el servicio y volver
   a intentarlo. Es una herramienta de diagnóstico opcional.

Dentro de un proyecto, la navegación **Resumen → Backlog → Planning Poker →
Sprints → Esfuerzo y defectos → Dashboard → Reporte** permite cambiar de etapa
sin volver al inicio. La etapa actual está marcada y se conserva el proyecto.
Al final de cada pantalla hay un acceso al siguiente paso; podés seguir otro orden.
La interfaz usa un tema oscuro con tipografía Inter y conserva los datos de los proyectos.

## Proyectos e integrantes

En **Proyectos**, completá **Nuevo proyecto**:

- **Nombre**.
- **Estado**: Planificado, Activo o Cerrado. El valor inicial es Activo.
- **Inicio** y **Fin**.

**Crear proyecto** muestra «Proyecto creado correctamente» y agrega la fila al
listado. El nombre de la fila abre el detalle.

El detalle muestra nombre, fechas y estado. En **Integrantes**, escribí un nombre
y usá **Agregar integrante**. Un nombre vacío o repetido muestra el error que
devuelve la API.

Desde el detalle se entra a Backlog, Sprints, Planning Poker, Esfuerzo y
defectos, Dashboard y Reporte.

**Editar proyecto** abre los datos actuales en el mismo detalle. Cambiá nombre,
fechas o estado y usá **Guardar cambios**. **Cancelar edición** descarta la
edición. Ante un rechazo, los datos permanecen y el campo afectado muestra el
motivo. Si se corta la conexión, revisá el detalle antes de repetir el guardado:
el servidor podría haber recibido el cambio.

## Backlog

En **Backlog**, **Nueva historia** pide identificador y título. La historia
nace en **Por hacer**. Descripción, prioridad, story points y criterios de
aceptación son opcionales y están en «Más datos».

**Crear historia** la agrega al listado, con las altas más recientes primero.
**Editar** permite cambiar prioridad, estado (**Por hacer** o **Completada**) y
story points. **Guardar** aplica el cambio; **Cancelar** lo descarta.

Si el proyecto está cerrado, el alta queda deshabilitada. Un proyecto o Sprint
cerrado también rechaza la edición.

## Sprints

En **Sprints**, **Nuevo sprint** pide un goal. **Crear sprint** lo deja abierto.
**Sprints abiertos** permite retomar uno existente después de visitar otra
pantalla. **Sprints anteriores** conserva el historial cerrado.

En el detalle de un Sprint abierto:

1. Elegí una historia y usá **Asignar al sprint**.
2. Usá **Completar** en una historia en estado Por hacer.
3. Para terminar, usá **Cerrar sprint** y después **Confirmar cierre**.
   **Cancelar** vuelve atrás. El cierre no se deshace desde la pantalla.

Un Sprint cerrado se puede consultar. La pantalla avisa que no permite asignar,
completar ni volver a cerrar. **Sprints anteriores** lista los Sprints ya
cerrados.
Si otra persona cambia el estado mientras operás, el rechazo ofrece
**Actualizar sprint** para consultar nuevamente la información del servidor.

## Planning Poker

En **Planning Poker**, cada historia tiene **Estimar**. Al abrirlo:

1. Elegí **Quién vota**.
2. Escribí **Tu story point**, un entero positivo, y usá **Votar**.
3. Usá **Revelar votos** cuando el equipo haya votado.

Si hay puntajes distintos, la pantalla dice «Hay diferencias» y ofrece **Abrir
otra ronda**. Si coinciden, dice que no hace falta otra ronda. Después de
revelar, **Story points acordados** y **Guardar estimación** registran el número
que elige el equipo. La pantalla no calcula ese acuerdo.

Mientras la ronda está abierta, los votos de las otras personas no se muestran.
Cambiar **Quién vota** descarta la entrada privada anterior y vuelve a consultar
la ronda. La tabla de historias se puede desplazar horizontalmente en móvil;
también admite foco y desplazamiento con teclado.

## Esfuerzo y defectos

En **Esfuerzo y defectos**:

- **Cargar esfuerzo** pide integrante, historia, fecha, actividad y horas.
  **Registrar esfuerzo** guarda el registro. Hacen falta al menos un integrante
  y una historia.
- **Defectos** pide historia afectada, descripción, severidad y Sprint de
  detección. **Registrar defecto** lo guarda abierto. Hacen falta una historia
  y un Sprint.
- En un defecto abierto, elegí **Sprint de resolución** y usá **Marcar
  resuelto**. El Sprint tiene que ser el de detección o uno posterior.

Cuando faltan integrantes, historias o sprints, los avisos incluyen accesos para
crearlos en el mismo proyecto. Si falla una carga parcial, las secciones que
tienen información siguen disponibles. **Reintentar sprints**, por ejemplo,
recupera ese catálogo sin borrar una actividad que estés escribiendo.

## Dashboard y reporte

**Dashboard** muestra el estado del proyecto y las métricas que calcula la API:
story points, velocidad, horas, desviación, historias completadas y defectos.
Si no hay datos, dice «Todavía no hay métricas para mostrar». El gráfico repite
los story points planificados y completados; no los recalcula.

**Reporte** permite elegir **Todo el proyecto** o un Sprint. Muestra métricas,
historias, esfuerzo y defectos de ese alcance. **Exportar PDF** descarga el
mismo alcance que está en pantalla.
La selección del sprint queda en la URL (`?sprint=ID`), por lo que se conserva
al recargar, compartir la dirección o usar Atrás/Adelante.

## Errores frecuentes

- «No pudimos conectar» o «No pudimos cargar…»: la API o la base no están
  disponibles. Levantá PostgreSQL y la API, y usá el botón de reintento.
- «Proyecto no encontrado», «Sprint no encontrado» o «Ronda no encontrada»: la
  dirección no corresponde a un recurso existente. Volvé al listado.
- Una página distinta de las anteriores muestra «Página no encontrada».
- Un dato inválido permanece en el formulario y muestra el mensaje devuelto por
  la API. Corregí el campo señalado antes de reenviar.

## Límites de esta versión

- No hay inicio de sesión ni permisos por persona. «Quién vota» es una elección
  en la pantalla, no una cuenta.
- Cerrar un Sprint no se deshace desde la pantalla.
- La velocidad, la desviación y los porcentajes salen de la API. La pantalla
  solo los muestra.
