# HU-025 — Gestionar integrantes desde la UI

- Issue: [#47](https://github.com/FaustinoDuran/seguimiento-medicion/issues/47)
- Módulo Go existente: `backend/internal/projects/`
- Presentación: `frontend/src/features/projects/`
- Estado: implementada

## Objetivo

Agregar y ver los integrantes de un proyecto desde su detalle, sin usar la API
a mano. Las reglas siguen en Go (HU-003).

## Entradas

| Campo | Obligatorio | Descripción |
|---|---|---|
| `name` | sí | Nombre del integrante. Go lo recorta y rechaza duplicados sin distinguir mayúsculas. |

El proyecto es el que está abierto en `/projects/:projectId`.

## Salidas esperadas

- En el detalle, el listado de integrantes de `GET /projects/{id}/members`.
- Cada uno muestra su nombre, en el orden de la API (`id` ascendente).
- Formulario de alta. Durante el envío se deshabilita el botón.
- Alta correcta: confirmación, formulario limpio y listado actualizado.
- Nombre en blanco: error en español y el formulario conserva lo demás.
- Nombre repetido: aviso en español y el nombre queda en el campo.
- Sin integrantes: vacío solo después de una consulta exitosa.

## Reglas de negocio

No hay reglas nuevas. Go rechaza el nombre en blanco y el duplicado
(comparación sin distinguir mayúsculas). React traduce el error.

## Restricciones

- Depende de HU-003 y de la pantalla de detalle de HU-024.
- Sin cambios de reglas en Go.
- No se editan ni se borran integrantes.

## Contrato que usa la pantalla

- `GET /projects/{id}/members`: `200` y `{ "members": [...] }`. Sin integrantes: `"members": []`.
- Cada integrante: `id`, `project_id`, `name`, `created_at`.
- `POST /projects/{id}/members` con `{ "name": "..." }`: `201`.
- `member name is required` → `400`.
- `member already belongs to the project` → `409`.
- `project not found` → `404`.

## Casos límite

- Proyecto sin integrantes: «Todavía no hay integrantes».
- Fallo al cargar: aviso y reintento, sin tratarlo como lista vacía.

## Condiciones de error

- Nombre en blanco: «Ingresá el nombre del integrante.»
- Duplicado: «Ese integrante ya está en el proyecto.»
- Otro rechazo: mensaje general en español.

## Criterios de aceptación

- [x] AC-025-01 — Agregar un integrante con nombre válido y verlo en el listado.
- [x] AC-025-02 — Mostrar el error si el nombre está en blanco.
- [x] AC-025-03 — Avisar si el integrante ya existe y conservar el nombre.
- [x] AC-025-04 — Mostrar el vacío solo cuando el listado exitoso no tiene integrantes.

## Trazabilidad

- BDD: `features/HU-025-ui-integrantes.feature`
- Tests: `frontend/src/features/projects/MemberSection.test.tsx`
- Código: `frontend/src/features/projects/MemberSection.tsx`
