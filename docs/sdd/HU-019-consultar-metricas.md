# HU-019 — Consultar métricas

- Issue: [#28](https://github.com/FaustinoDuran/seguimiento-medicion/issues/28)
- Módulo Go: `backend/internal/metrics/`
- Estado: mergeada en `main` ([PR #90](https://github.com/FaustinoDuran/seguimiento-medicion/pull/90))

## Objetivo

Calcular, en el backend, los números de seguimiento de un proyecto o de un
Sprint: story points planificados y completados, velocidad, horas, desviación,
porcentaje de historias completadas y defectos detectados y resueltos. React
solo muestra el resultado (HU-020).

## Entradas

Identificador del proyecto, o identificador del Sprint. Son enteros positivos
dentro del rango de `int64`.

## Salidas esperadas

Un informe con los campos del contrato. Los valores salen de historias,
asignaciones, horas reales y defectos ya persistidos. Esta HU no crea ni
modifica esas entidades.

## Contrato de la API

- `GET /projects/{project_id}/metrics`: métricas del proyecto. `200`.
- `GET /sprints/{sprint_id}/metrics`: métricas de ese Sprint. `200`.

```json
{
  "project_id": 1,
  "sprint_id": null,
  "planned_story_points": 13,
  "completed_story_points": 8,
  "velocity": 8,
  "estimated_hours": 0,
  "actual_hours": 4.5,
  "deviation": 5,
  "completed_story_percentage": 50,
  "defects_detected": 1,
  "defects_resolved": 0
}
```

En la consulta por Sprint, `sprint_id` es el Sprint pedido. En la de proyecto
es `null`. `actual_hours`, `estimated_hours`, `velocity` y
`completed_story_percentage` salen como número JSON, con el mismo criterio de
ceros de cola que HU-016 (`8`, `8.5`, `33.33`).

- `400`: identificador ausente, no numérico o no positivo.
- `404`: proyecto inexistente; Sprint inexistente.
- `500`: fallo de infraestructura, sin detalles internos.

Errores con formato `{"error":"..."}`.

## Fórmulas

Las historias sin `story_points` no suman planificado ni completado en puntos.
Su ausencia no equivale a cero puntos (ver glosario). Sí cuentan para el
porcentaje de historias.

1. **Alcance de un Sprint:** historias asignadas a ese Sprint.
2. **Alcance de un proyecto (historias):** todas las historias del proyecto.
3. **Story points planificados (Sprint):** suma de `story_points` de las
   historias del alcance que tienen estimación.
4. **Story points planificados (proyecto):** suma de `story_points` de las
   historias del proyecto que tienen estimación y están asignadas a algún
   Sprint.
5. **Story points completados:** suma de `story_points` de las historias del
   alcance con estado `done` y estimación.
6. **Velocidad (Sprint):** story points completados de ese Sprint, aunque
   siga abierto.
7. **Velocidad (proyecto):** promedio de los story points completados de los
   Sprints **cerrados** del proyecto. Si no hay Sprints cerrados, es `0`.
   El promedio se redondea a dos decimales (0.5 hacia arriba).
8. **Horas estimadas:** siempre `0`. No existe un campo de horas estimadas;
   la estimación del producto son los story points acordados (ADR 0006).
9. **Horas reales:** suma de las horas de los registros de esfuerzo cuyas
   historias están en el alcance (proyecto: todos los registros del proyecto;
   Sprint: registros de historias asignadas a ese Sprint).
10. **Desviación:** story points planificados menos story points completados.
11. **Porcentaje de historias completadas:** historias `done` del alcance
    dividido las historias del alcance, por 100, con dos decimales. Si no hay
    historias, es `0`.
12. **Defectos detectados (proyecto):** cantidad de defectos del proyecto.
    **(Sprint):** defectos con `detection_sprint_id` igual a ese Sprint.
13. **Defectos resueltos (proyecto):** defectos del proyecto con estado
    `resolved`. **(Sprint):** defectos con `resolution_sprint_id` igual a ese
    Sprint. Mientras no haya resolución persistida, el valor es `0`.

Un proyecto o Sprint vacío devuelve ceros. Los datos de otro proyecto no
entran en el informe.

## Reglas de negocio

1. El cálculo vive solo en `metrics`. El handler no aplica fórmulas.
2. No se inventa un campo de horas estimadas ni se convierte story point a hora.
3. No se persiste el informe: se calcula en cada consulta.
4. Un Sprint de otro proyecto no se consulta por el `project_id`; la consulta
   por Sprint usa el id del Sprint.

## Restricciones

- Depende de HU-010, HU-016, HU-017 y HU-018 (esta última puede devolver
  resueltos en `0` hasta que su esquema esté en `main`).
- Sin autenticación. Sin UI (HU-020).
- Los catálogos leen datos por puertos; no se importan handlers ajenos.

## Casos límite

- Proyecto o Sprint sin historias: todo en cero.
- Historia sin estimar asignada: no suma puntos; sí cuenta en el porcentaje.
- Horas `1.5` y `0.01` se suman a `1.51`.
- Un Sprint abierto aporta velocidad de Sprint, no velocidad de proyecto.
- Defectos abiertos no suman resueltos.

## Condiciones de error

- Identificador inválido: no se calcula.
- Proyecto o Sprint inexistente: no se calcula.
- Fallo de lectura: `500` genérico.

## Criterios de aceptación

- [x] AC-019-01 — Devolver las métricas de un proyecto con historias estimadas, horas y defectos, calculadas en Go.
- [x] AC-019-02 — Devolver las métricas de un Sprint, sin mezclar historias ni defectos de otro Sprint.
- [x] AC-019-03 — Excluir del planificado y del completado en puntos a las historias sin estimar.
- [x] AC-019-04 — Calcular velocidad de proyecto como promedio de Sprints cerrados, y `0` si no hay ninguno cerrado.
- [x] AC-019-05 — Informar `estimated_hours` en `0` y `actual_hours` como la suma de esfuerzo del alcance.
- [x] AC-019-06 — Informar desviación como planificados menos completados, y el porcentaje de historias `done`.
- [x] AC-019-07 — Contar defectos detectados y resueltos según las fórmulas.
- [x] AC-019-08 — Aislar un proyecto de los datos de otro.
- [x] AC-019-09 — Rechazar identificadores inválidos con `400` y recursos inexistentes con `404`.
- [x] AC-019-10 — Respetar el contrato HTTP. Un error de infraestructura responde `500` sin detalles internos.

## Trazabilidad

- BDD: `features/HU-019-consultar-metricas.feature`
- Tests: `backend/internal/metrics/{service,handler,repository_postgres}_test.go`
- Código: `backend/internal/metrics/`

## Corrección BUG-001 — Sprint de resolución

Issue [#94](https://github.com/FaustinoDuran/seguimiento-medicion/issues/94),
vinculada a AC-019-07. La lectura PostgreSQL omitía `resolution_sprint_id`
y entregaba al servicio una referencia vacía; los resueltos por Sprint
quedaban en cero. Se incorpora la referencia nullable sin modificar fórmulas.

`TestPostgresSprintMetricsCountResolvedDefects` reproduce el fallo por HTTP
y PostgreSQL reales. Cubre detección y resolución en Sprints distintos, un
defecto abierto y otro proyecto con un defecto resuelto. El test falló con
`detected/resolved = 1/0, want 1/1` antes de corregir la lectura.

RED `cfa872d` → GREEN `9ee5655`; verificación: `go test -count=1 -race
./internal/metrics/` con PostgreSQL local. PR [#95](https://github.com/FaustinoDuran/seguimiento-medicion/pull/95),
independiente de HU-021.
