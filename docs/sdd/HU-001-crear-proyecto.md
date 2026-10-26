# HU-001 — Crear proyecto

- Issue: [#10](https://github.com/FaustinoDuran/seguimiento-medicion/issues/10)
- Módulo Go: `backend/internal/projects/`
- Estado: lista para implementar

## Objetivo

Dar de alta un proyecto con nombre, fechas de inicio y fin y estado, y poder
consultarlo después. Es la raíz del resto de las historias del Sprint 1.

## Entradas

Alta (`CreateProject`):

| Campo | Obligatorio | Descripción |
|---|---|---|
| `name` | sí | Nombre visible del proyecto. |
| `start_date` | sí | Fecha de inicio (día calendario, `YYYY-MM-DD`). |
| `end_date` | sí | Fecha de fin (día calendario, `YYYY-MM-DD`). |
| `status` | no | Si se omite, el sistema usa `active`. |

Consulta: identificador del proyecto generado en el alta.

## Salidas esperadas

Un proyecto persistido con:

- `id` generado por el sistema;
- `name`, `start_date`, `end_date`, `status`;
- `created_at`.

La consulta por `id` devuelve el mismo proyecto.

## Reglas de negocio

1. El nombre no puede estar vacío ni ser solo espacios; se guarda recortado.
2. `end_date` no puede ser anterior a `start_date`.
3. `status` admitidos: `planned`, `active`, `closed`. Cualquier otro se rechaza.
4. Si `status` no viene, queda `active`.
5. El identificador lo asigna el sistema; el cliente no lo elige.
6. Un proyecto creado se puede consultar por su `id`.

## Restricciones

- Las reglas viven en `backend/internal/projects/`, no en React ni en el handler.
- Persistencia: PostgreSQL (tabla de dominio de esta HU).
- API JSON. Sin autenticación (no hay HU de auth).
- No se modifican proyectos acá (HU-002) ni se listan todos.

## Casos límite

- Nombre con espacios alrededor: se acepta y se guarda recortado.
- `start_date` igual a `end_date`: válido (proyecto de un día).
- `status` omitido: queda `active`.

## Condiciones de error

- Nombre ausente o en blanco: no se crea.
- Fecha ausente, mal formada o `end_date` anterior a `start_date`: no se crea.
- `status` distinto de `planned`, `active` o `closed`: no se crea.
- Consulta de un `id` inexistente: no hay proyecto.

## Criterios de aceptación

- [ ] AC-001-01 — Con nombre y fechas válidas, el proyecto queda creado y se puede consultar por su `id`.
- [ ] AC-001-02 — Si falta el nombre (o queda en blanco), no se crea y se informa el error.
- [ ] AC-001-03 — El estado inicial queda `active` cuando no se envía `status`.
- [ ] AC-001-04 — Si `end_date` es anterior a `start_date`, no se crea y se informa el error.

## Trazabilidad

- BDD: `features/HU-001-crear-proyecto.feature`
- Tests: `backend/internal/projects/service_test.go`
- Código: `backend/internal/projects/service.go`
