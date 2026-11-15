# HU-004 — Consultar estado del proyecto

- Issue: [#13](https://github.com/FaustinoDuran/seguimiento-medicion/issues/13)
- Módulo Go: `backend/internal/projects/`
- Estado: implementada

## Objetivo

Que un integrante vea el estado de un proyecto para saber si sigue activo o
cerrado, y que reciba un aviso cuando el proyecto no existe.

El comportamiento ya está en `main` desde HU-001 (consulta por `id`) y HU-024
(detalle en la UI). Esta HU no agrega reglas ni endpoints: documenta ese
comportamiento, lo cubre con BDD y lo verifica con tests. Los ADR 0005, 0006 y
0008 ya la declaraban cubierta y fuera de alcance.

## Entradas

Consulta: identificador del proyecto.

- `GET /projects/{id}`: el `id` en la ruta.
- Detalle en la UI: el `id` de la ruta `/projects/:id`.

## Salidas esperadas

Un proyecto con `id`, `name`, `start_date`, `end_date`, `status` y
`created_at`. El campo que mira esta historia es `status`, con etiqueta en la UI:

| `status` persistido | Etiqueta en pantalla |
|---|---|
| `planned` | Planificado |
| `active` | Activo |
| `closed` | Cerrado |

Si el proyecto no existe, la consulta responde `404` y el detalle muestra el
aviso "Proyecto no encontrado".

## Reglas de negocio

1. La consulta devuelve el estado persistido del proyecto; no se calcula a
   partir de las fechas ni de otro dato.
2. Los únicos estados posibles son `planned`, `active` y `closed` (HU-001).
3. Cambiar el estado es de HU-002; esta historia solo lo consulta.
4. Un `id` inexistente no devuelve un proyecto: responde `404`.
5. La UI muestra el estado con su etiqueta en español y no lo recalcula.

## Restricciones

- Las reglas viven en `backend/internal/projects/`. Handler sin reglas.
- Sin migración nueva: la columna `status` existe desde la migración `001`.
- React solo presenta: no deriva ni valida el estado.
- Sin autenticación (no hay HU de auth).

## Casos límite

- Proyecto en estado `planned`: se muestra "Planificado".
- Proyecto en estado `closed`: se muestra "Cerrado".
- `id` con formato inválido (no numérico o menor que 1): se rechaza con `400`.

## Condiciones de error

- Proyecto inexistente: `404` con `{"error":"project not found"}`.
- Fallo de persistencia: `500`; no se devuelve un proyecto vacío ni un estado
  inventado.

## Criterios de aceptación

- [x] AC-004-01 — La consulta de un proyecto existente devuelve su estado persistido (`planned`, `active` o `closed`).
- [x] AC-004-02 — El detalle del proyecto muestra el estado con su etiqueta en español ("Planificado", "Activo" o "Cerrado").
- [x] AC-004-03 — Si el proyecto no existe, la consulta responde `404` y el detalle avisa "Proyecto no encontrado".

## Trazabilidad

- BDD: `features/HU-004-consultar-estado-proyecto.feature`
- Tests: `backend/internal/projects/{service,handler}_test.go`; `frontend/src/features/projects/ProjectDetailPage.test.tsx`
- Código: `backend/internal/projects/{handler,model}.go`; `frontend/src/features/projects/ProjectDetailPage.tsx`
- Historias de origen: HU-001 (consulta por `id`), HU-024 (detalle en la UI)

## Evidencia de verificación

El comportamiento ya existía en `main`, así que no hubo RED: las comprobaciones
se agregaron sobre tests existentes y pasan en verde. La excepción de TDD se
justifica porque la historia no introduce comportamiento nuevo.
