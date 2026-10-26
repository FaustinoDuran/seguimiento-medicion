# HU-003 — Registrar integrantes

- Issue: [#12](https://github.com/FaustinoDuran/seguimiento-medicion/issues/12)
- Módulo Go: `backend/internal/projects/`
- Estado: lista para implementar

## Objetivo

Asociar integrantes a un proyecto existente y poder listarlos. Sin personas no
hay quién estime, vote ni cargue horas en sprints siguientes.

## Entradas

Alta (`AddMember`):

| Campo | Obligatorio | Descripción |
|---|---|---|
| `project_id` | sí | Proyecto al que se asocia. Debe existir (HU-001). |
| `name` | sí | Nombre visible del integrante. |

Consulta: `project_id` del proyecto.

## Salidas esperadas

Un integrante persistido con `id` generado, `project_id`, `name` y `created_at`.

La consulta por proyecto devuelve todos sus integrantes.

## Reglas de negocio

1. El proyecto debe existir. Si no, no se agrega el integrante.
2. El nombre no puede estar vacío ni ser solo espacios; se guarda recortado.
3. No se duplica el mismo integrante en el mismo proyecto. La comparación es
   sin distinguir mayúsculas (`Valentino` y `valentino` son el mismo).
4. El mismo nombre puede existir en otro proyecto.
5. El identificador lo asigna el sistema.

## Restricciones

- Las reglas viven en `backend/internal/projects/`.
- Persistencia: PostgreSQL.
- API JSON. Sin autenticación.
- No se editan ni eliminan integrantes en esta HU.

## Casos límite

- Nombre con espacios alrededor: se acepta y se guarda recortado.
- Proyecto sin integrantes: la consulta devuelve lista vacía.

## Condiciones de error

- Proyecto inexistente: no se crea el integrante.
- Nombre ausente o en blanco: no se crea.
- Nombre ya asociado a ese proyecto: no se crea.

## Criterios de aceptación

- [ ] AC-003-01 — Con un proyecto existente y un nombre válido, el integrante queda asociado y se puede listar.
- [ ] AC-003-02 — Si el nombre está en blanco, no se crea y se informa el error.
- [ ] AC-003-03 — Si el mismo nombre ya está en el proyecto, no se duplica y se informa el error.
- [ ] AC-003-04 — Si el proyecto no existe, no se crea el integrante.

## Trazabilidad

- BDD: `features/HU-003-registrar-integrantes.feature`
- Tests: `backend/internal/projects/service_test.go`
- Código: `backend/internal/projects/service.go`
