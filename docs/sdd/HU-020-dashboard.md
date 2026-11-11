# HU-020 — Dashboard

- Issue: [#29](https://github.com/FaustinoDuran/seguimiento-medicion/issues/29)
- Módulo Go existente: `backend/internal/metrics/` y `backend/internal/projects/`
- Presentación: `frontend/src/features/metrics/`
- Estado: implementada; en revisión ([PR #91](https://github.com/FaustinoDuran/seguimiento-medicion/pull/91))

## Objetivo

Ver el estado del proyecto y las métricas ya calculadas, con al menos un
gráfico, sin curl. Las fórmulas siguen en Go (HU-019). React solo presenta.

## Entradas

El proyecto es el de `/projects/:projectId`. No hay formulario: la pantalla
consulta y muestra.

## Salidas esperadas

- Desde el detalle del proyecto, el enlace Dashboard abre
  `/projects/:projectId/dashboard`.
- Se ve el nombre y el estado del proyecto (`planned`, `active` o `closed`)
  con las mismas etiquetas que HU-024.
- Se ven los números que devuelve `GET /projects/{id}/metrics`: story points
  planificados y completados, velocidad, horas estimadas y reales, desviación,
  porcentaje de historias completadas y defectos detectados y resueltos.
- Hay un gráfico de barras de story points planificados frente a completados.
  La longitud de cada barra usa esos dos valores de la API; no se calcula una
  métrica nueva.
- Si la API devuelve todos esos indicadores en cero, se muestra un vacío
  entendible: «Todavía no hay métricas para mostrar.»
- Proyecto inexistente o id inválido: «Proyecto no encontrado».
- Fallo de red: aviso y reintento. No se muestra como vacío.

## Reglas de negocio

No hay reglas nuevas. Go calcula. React no deriva velocidad, desviación ni
porcentaje.

## Restricciones

- Depende de HU-019, HU-001 y HU-024, ya en `main`.
- Sin cambios de reglas en Go.
- Un PR toca `App.tsx` para la ruta del dashboard. HU-030 no se implementa acá.

## Contrato que usa la pantalla

- `GET /projects/{id}`: `200` con `id`, `name`, `status`, fechas.
- `GET /projects/{id}/metrics`: `200` con el informe de HU-019.
- `404`: proyecto no encontrado.

Errores `{ "error": "..." }` se muestran en español genérico si no son 404.

## Casos límite

- Informe con ceros: vacío, no un gráfico en blanco sin texto.
- `estimated_hours` en `0` se muestra tal cual.
- Números con decimales (`4.5`, `33.33`) se muestran como los envía la API.

## Condiciones de error

- Proyecto no encontrado.
- Falla al pedir el proyecto o las métricas.

## Criterios de aceptación

- [x] AC-020-01 — Desde el detalle, Dashboard abre la pantalla del proyecto.
- [x] AC-020-02 — Mostrar el estado del proyecto que devuelve la API.
- [x] AC-020-03 — Mostrar las métricas calculadas en Go, sin recalcularlas.
- [x] AC-020-04 — Mostrar un gráfico de story points planificados y completados.
- [x] AC-020-05 — Mostrar un vacío entendible cuando todos los indicadores vienen en cero.
- [x] AC-020-06 — Informar proyecto inexistente y permitir reintentar un fallo de carga.

## Trazabilidad

- BDD: `features/HU-020-dashboard.feature`
- Tests: `frontend/src/features/metrics/Dashboard.test.tsx`
- Código: `frontend/src/features/metrics/`; enlace en `ProjectDetailPage.tsx`; ruta en `App.tsx`
