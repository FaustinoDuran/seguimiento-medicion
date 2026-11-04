# HU-024 — Usar proyectos desde la UI

- Issue: [#46](https://github.com/FaustinoDuran/seguimiento-medicion/issues/46)
- Módulo Go existente: `backend/internal/projects/`
- Presentación: `frontend/src/features/projects/`
- Estado: implementada

## Objetivo

Crear y ver proyectos desde la interfaz, sin curl. Las reglas siguen en Go
(HU-001 y HU-023). React solo presenta.

## Entradas

Alta, los mismos campos que `POST /projects` y la tabla `projects`:

| Campo | Obligatorio | Descripción |
|---|---|---|
| `name` | sí | Nombre visible. |
| `start_date` | sí | Día calendario `YYYY-MM-DD`. |
| `end_date` | sí | Día calendario `YYYY-MM-DD`. |
| `status` | no | `planned`, `active` o `closed`. Si la pantalla no lo cambia, envía `active`. |

La pantalla no modifica un proyecto existente (eso es HU-002).

## Salidas esperadas

- Listado en `/projects`, en el orden que devuelve `GET /projects` (`id` ascendente).
- Cada fila muestra nombre, inicio, fin y estado, y abre el detalle.
- Detalle en `/projects/:projectId` con nombre, fechas y estado.
- Desde el detalle se llega al backlog y a los sprints. Esas pantallas ya
  vuelven a `/projects/:projectId`.
- Formulario sobre el listado. Durante el envío se deshabilita el botón.
- Alta correcta: confirmación, formulario limpio y listado actualizado.
- Estado vacío solo después de un listado exitoso sin proyectos.
- El nombre en blanco y las fechas inválidas se muestran en español, junto al
  campo, y el formulario conserva lo ingresado.

## Reglas de negocio

No hay reglas nuevas. Go recorta el nombre, rechaza el blanco, rechaza
`end_date` anterior a `start_date` y solo admite `planned`, `active` y
`closed`. React traduce esos errores y no los recalcula.

## Restricciones

- Depende de TASK-011, HU-001 y HU-023, ya en `main`.
- Sin migración y sin cambios de reglas en Go.
- React no calcula métricas ni decide si las fechas son válidas.

## Contrato que usa la pantalla

- `GET /projects`: `200` y un array. Sin proyectos: `[]`.
- `POST /projects`: `201` con el proyecto creado.
- `GET /projects/{id}`: `200` con `id`, `name`, `start_date`, `end_date`,
  `status`, `created_at`.
- Errores `{ "error": "..." }`:
  - `project name is required` → `400`
  - `project dates are invalid` → `400`
  - `project status is invalid` → `400`
  - `project not found` → `404`

## Casos límite

- Listado vacío: «Todavía no hay proyectos», solo tras una consulta exitosa.
- Proyecto inexistente o id inválido en el detalle: «Proyecto no encontrado»,
  con vuelta al listado.
- Fallo de red al listar: aviso y reintento. No se muestra como lista vacía.

## Condiciones de error

- Nombre en blanco: «Ingresá el nombre del proyecto.»
- Fechas inválidas: «Revisá las fechas: las dos son obligatorias y el fin no
  puede ser anterior al inicio.»
- Los demás rechazos de alta: mensaje general en español, con los datos
  conservados.

## Criterios de aceptación

- [x] AC-024-01 — Crear un proyecto con nombre y fechas válidas desde la UI.
- [x] AC-024-02 — Ver el detalle con nombre, fechas y estado.
- [x] AC-024-03 — Ver el listado en el orden que entrega la API.
- [x] AC-024-04 — Mostrar el vacío solo si el listado exitoso no tiene proyectos.
- [x] AC-024-05 — Mostrar el error si falta el nombre, sin perder lo ingresado.
- [x] AC-024-06 — Mostrar el error si las fechas son inválidas.
- [x] AC-024-07 — Abrir backlog y sprints desde el detalle, y volver al listado.
  Un proyecto inexistente se informa.

## Trazabilidad

- BDD: `features/HU-024-ui-proyectos.feature`
- Tests: `frontend/src/features/projects/*.test.tsx`
- Código: `frontend/src/features/projects/`; rutas en `frontend/src/App.tsx`
