# HU-002 — Modificar proyecto

- Issue: [#11](https://github.com/FaustinoDuran/seguimiento-medicion/issues/11)
- Módulo Go: `backend/internal/projects/`
- Estado: implementada en `feat/HU-002-modificar-proyecto`; falta el PR

## Objetivo

Cambiar el nombre, las fechas o el estado de un proyecto ya creado, para
mantener los datos al día. El alta y la consulta por `id` siguen siendo
HU-001. Esta HU no tiene UI.

## Entradas

Modificación (`Update`):

| Campo | Obligatorio | Descripción |
|---|---|---|
| `id` | sí | Proyecto existente. Lo asignó el sistema en el alta. |
| `name` | no | Si viene, reemplaza el nombre. Si se omite, se conserva. |
| `start_date` | no | Día calendario `YYYY-MM-DD`. Si se omite, se conserva. |
| `end_date` | no | Día calendario `YYYY-MM-DD`. Si se omite, se conserva. |
| `status` | no | Si viene, reemplaza el estado. Si se omite, se conserva. |

Un campo JSON `null` equivale a omitirlo: no borra el dato guardado.

## Salidas esperadas

El mismo proyecto, con los campos enviados ya aplicados y el resto igual:

- `id` y `created_at` no cambian;
- `name`, `start_date`, `end_date` y `status` quedan en el valor nuevo o en el
  que ya tenía;
- la consulta por `id` devuelve ese resultado.

Si la modificación se rechaza, el proyecto persistido queda como estaba.

## Reglas de negocio

1. Solo se modifica un proyecto que ya existe.
2. Se puede cambiar el nombre, una fecha, las dos fechas, el estado, o
   cualquier combinación. Lo que no viene en la petición se conserva.
3. El nombre, si viene, no puede quedar vacío ni ser solo espacios; se guarda
   recortado.
4. El par de fechas resultante no puede tener `end_date` anterior a
   `start_date`. Si se cambia una sola fecha, se compara contra la otra que
   ya estaba guardada.
5. `start_date` igual a `end_date` es válido.
6. `status` admitidos: `planned`, `active`, `closed`. Cualquier otro se
   rechaza. Omitirlo no vuelve el estado a `active`: eso solo ocurre en el
   alta (HU-001).
7. `id` y `created_at` no se editan.
8. Una petición sin ningún campo modificable devuelve el proyecto tal como
   está.
9. Integrantes, historias, sprints, esfuerzo y defectos del proyecto no se
   tocan. Cambiar el estado del proyecto no cierra sus sprints.

## Restricciones

- Las reglas viven en `backend/internal/projects/`, no en React ni en el handler.
- Persistencia: tabla `projects` de HU-001. Sin migración nueva.
- API JSON. Sin autenticación.
- Sin React. La pantalla de proyectos (HU-024) no se modifica en esta HU.

## Casos límite

- Nombre con espacios alrededor: se acepta y se guarda recortado.
- Cambiar solo el nombre: fechas y estado siguen iguales.
- Cambiar solo la fecha de fin a un día igual al inicio: válido.
- Cuerpo sin campos modificables: el proyecto no cambia y se devuelve.
- Proyecto `closed`: se puede seguir modificando. Esta HU no lo congela.

## Condiciones de error

- Proyecto inexistente: no se modifica y se informa que no existe.
- Nombre presente pero en blanco: no se modifica y se informa que es obligatorio.
- Fecha mal formada, o `end_date` resultante anterior a `start_date`: no se
  modifica y se informa que las fechas son inválidas.
- `status` distinto de `planned`, `active` o `closed`: no se modifica y se
  informa que el estado es inválido.

## Contrato de la API

- `PATCH /projects/{id}`: cuerpo JSON con cualquiera de `name`, `start_date`,
  `end_date` y `status`.
- `200` con el proyecto (`id`, `name`, `start_date`, `end_date`, `status`,
  `created_at`).
- `400` si el `id` no es válido, el nombre queda en blanco, las fechas son
  inválidas o el estado no es uno de los tres admitidos.
- `404` si el proyecto no existe.

El verbo es `PATCH` porque la petición puede traer un subconjunto de campos.
El resto de la API sigue usando `POST` para comandos que no editan así un
recurso ya existente.

## Criterios de aceptación

- [x] AC-002-01 — Con cambios válidos de nombre, fechas o estado, el proyecto queda modificado y se puede consultar por su `id`. El `id` y `created_at` no cambian.
- [x] AC-002-02 — Un campo omitido conserva el valor que ya tenía el proyecto.
- [x] AC-002-03 — Si la fecha de fin resultante es anterior a la de inicio, no se modifica y se informa el error.
- [x] AC-002-04 — Si el nombre queda en blanco, no se modifica y se informa el error.
- [x] AC-002-05 — Si el estado no es `planned`, `active` ni `closed`, no se modifica y se informa el error.
- [x] AC-002-06 — Si el proyecto no existe, no se modifica y se informa el error.

## Trazabilidad

- BDD: `features/HU-002-modificar-proyecto.feature`
- Tests: `backend/internal/projects/{service,handler,repository_postgres}_test.go`
- Código: `backend/internal/projects/`
