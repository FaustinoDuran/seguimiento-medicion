# HU-023 — Listar proyectos

- Issue: [#45](https://github.com/FaustinoDuran/seguimiento-medicion/issues/45)
- Módulo Go: `backend/internal/projects/`
- Estado: lista para implementar

## Objetivo

Listar los proyectos ya persistidos para poder elegir uno y trabajar sobre él.
El alta y la consulta por `id` siguen siendo HU-001. Esta HU no tiene UI.

## Entradas

Listado (`List`): ninguna. Se devuelven todos los proyectos persistidos.

## Salidas esperadas

Una lista de proyectos, cada uno con los mismos datos que la consulta por `id`:

- `id`;
- `name`;
- `start_date`;
- `end_date`;
- `status`;
- `created_at`.

Si no hay proyectos, la lista queda vacía.

## Reglas de negocio

1. Se listan todos los proyectos persistidos, sin filtrar por estado.
2. Un proyecto recién creado aparece en el listado con los mismos datos que devuelve su alta.
3. Si no hay proyectos, el resultado es una lista vacía, no un error.
4. El orden es por `id` ascendente.

## Restricciones

- Las reglas viven en `backend/internal/projects/`. Handler sin reglas.
- Persistencia: tabla `projects` (HU-001). Sin migración nueva.
- API JSON. Sin autenticación. Sin React (UI en HU-024).
- No se modifican proyectos (HU-002) ni se calcula estado del proyecto (HU-004).

## Casos límite

- Ningún proyecto persistido: lista vacía.
- Varios proyectos: todos aparecen, en orden de `id`.

## Contrato de la API

- `GET /projects`: `200` con un array JSON de proyectos
  (`id`, `name`, `start_date`, `end_date`, `status`, `created_at`).
  Sin proyectos: `[]`.

## Condiciones de error

Esta consulta no rechaza la petición. Un fallo de persistencia no se traduce
en una lista vacía.

## Criterios de aceptación

- [ ] AC-023-01 — Un proyecto recién creado aparece en el listado con sus datos.
- [ ] AC-023-02 — Si no hay proyectos, la lista queda vacía.
- [ ] AC-023-03 — Varios proyectos aparecen todos, ordenados por `id`.

## Trazabilidad

- BDD: `features/HU-023-listar-proyectos.feature`
- Tests: `backend/internal/projects/{service,handler}_test.go`
- Código: `backend/internal/projects/`
