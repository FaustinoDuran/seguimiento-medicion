# HU-018 — Resolver defecto

- Issue: [#27](https://github.com/FaustinoDuran/seguimiento-medicion/issues/27)
- Módulo Go: `backend/internal/defects/`
- Estado: lista para implementar; contrato e interfaces de prueba confirmados
- Dependencia: HU-017, disponible en la base de `feat/HU-018-resolve-defect`.

## Objetivo

Registrar que el equipo considera corregido el problema de un defecto e
identificar el Sprint en el que se resolvió.

## Entradas

- Proyecto y defecto a resolver.
- Sprint de resolución obligatorio, informado explícitamente.

No se solicitan autor, fecha ni comentario de resolución.
El cliente no elige un estado arbitrario: la operación asigna `resolved`.

## Salidas esperadas

El defecto queda resuelto y conserva el Sprint de resolución, consultable
posteriormente junto con los datos registrados en HU-017.

## Contrato HTTP

`POST /projects/{project_id}/defects/{id}/resolve`

```json
{"resolution_sprint_id": 3}
```

Devuelve `200` con el defecto completo:

```json
{
  "id": 17,
  "project_id": 1,
  "story_id": 4,
  "description": "El archivo exportado sale vacío",
  "severity": "high",
  "status": "resolved",
  "detection_sprint_id": 2,
  "resolution_sprint_id": 3
}
```

`GET /projects/{project_id}/defects/{id}` incorpora `resolution_sprint_id`:
es `null` para defectos abiertos y el ID guardado para los resueltos. El alta
de HU-017 mantiene estado inicial `open` y no registra una resolución.

Todos los identificadores son enteros positivos dentro del rango de `int64`.
Los errores usan `{"error":"..."}`:

- `400`: IDs ausentes, inválidos o fuera de rango; JSON inválido o más de
  un valor JSON.
- `404`: defecto o Sprint de resolución inexistentes o ajenos al proyecto.
- `409`: Sprint de resolución anterior al de detección, o intento de cambiar
  el Sprint de un defecto ya resuelto.
- `500`: fallo de infraestructura, sin exponer detalles internos.

Primero se validan formato y rangos, después el defecto dentro del proyecto,
el Sprint de resolución y el orden respecto del Sprint de detección. Se
comprueba la resolución existente al guardar para impedir que solicitudes
concurrentes sobrescriban una resolución ya registrada.

## Reglas de negocio

Decisiones Q1 a Q8 confirmadas durante la entrevista `grill-with-docs`:

1. Esta HU agrega la transición de abierto a resuelto, sin estados
   intermedios ni reapertura. Resuelto significa que el equipo considera
   corregido el problema.
2. El Sprint de resolución es obligatorio y pertenece al mismo proyecto
   que el defecto. Se indica explícitamente; no se deduce un Sprint actual.
3. El Sprint de resolución puede coincidir o diferir del Sprint de detección
   y del Sprint asignado a la historia.
4. Se permite resolver defectos en proyectos cerrados, con Sprints de
   resolución cerrados y sobre historias completadas. No se reabren esas
   entidades.
5. Solo cambian el estado y el Sprint de resolución. Se conservan descripción,
   severidad, historia y Sprint de detección.
6. Estado y Sprint de resolución se guardan en una única operación atómica.
   Ante un fallo no cambia ninguno de los dos.
7. Decisión Q7: un defecto ya resuelto no permite cambiar el Sprint de
   resolución mediante esta operación. Si se indica otro Sprint, se devuelve
   conflicto (`409`) y se conserva la resolución existente.
8. Repetir la resolución con el mismo Sprint devuelve éxito con los mismos
   datos, sin nuevas modificaciones. Si se pierde la primera respuesta, la
   solicitud se puede reintentar.
9. La referencia para impedir una resolución anterior es el Sprint de
   detección registrado en HU-017. La resolución puede ocurrir en ese mismo
   Sprint o en uno posterior en orden de creación: `created_at` ascendente y,
   ante empate, `id` ascendente. Se reutiliza el criterio de HU-011, no se
   compara solo el ID.
10. Ese orden representa el alta de los Sprints, no un calendario de ejecución.
    Esta HU no agrega fechas de inicio/fin ni cambia la creación de Sprints.

## Restricciones

- Depende de HU-017 y comparte el módulo `defects`.
- Reglas de negocio únicamente en Go; sin autenticación.
- No se calculan métricas ni se agrega UI; corresponden a otras HUs.
- No se implementan modificaciones de descripción o severidad ni reapertura.

## Casos límite

- Resolver en el mismo Sprint de detección o en otro del proyecto.
- Proyecto y Sprint de resolución cerrados.
- Historia completada.
- Historia sin Sprint asignado: la resolución se refiere al defecto y no
  requiere asignar la historia.
- Misma resolución repetida, incluso después de perder la primera respuesta.
- Dos Sprints con la misma fecha de creación: desempata su ID.
- Solicitudes concurrentes para resolver en Sprints distintos: la primera
  resolución guardada se conserva y la otra se rechaza.

## Condiciones de error

- Defecto inexistente o ajeno al proyecto indicado.
- Sprint de resolución ausente, inexistente o de otro proyecto.
- Sprint de resolución anterior al de detección en el orden acordado.
- Defecto ya resuelto con otro Sprint de resolución.
- Fallo de persistencia: se conserva el estado previo y no queda una
  resolución parcial.

## Criterios de aceptación

- [ ] AC-018-01 — Resolver un defecto abierto con un Sprint de resolución
  explícito del mismo proyecto y consultar posteriormente ambos datos.
- [ ] AC-018-02 — Informar que el defecto no existe cuando no se encuentra
  en el proyecto indicado, sin modificar datos.
- [ ] AC-018-03 — Exigir un Sprint de resolución existente y del mismo proyecto,
  permitiendo que coincida o difiera del de detección y del asignado a la historia.
- [ ] AC-018-04 — Admitir proyectos cerrados, Sprints de resolución cerrados
  e historias completadas, sin reabrirlos.
- [ ] AC-018-05 — Conservar descripción, severidad, historia y Sprint de
  detección; guardar estado y Sprint de resolución juntos o conservar ambos
  sin cambios ante un fallo.
- [ ] AC-018-06 — Repetir una resolución con el mismo Sprint devuelve éxito
  conservando los datos existentes.
- [ ] AC-018-07 — Rechazar el cambio a otro Sprint de resolución con `409`,
  incluso ante solicitudes concurrentes, sin sobrescribir la primera resolución.
- [ ] AC-018-08 — Permitir el mismo Sprint de detección o uno posterior según
  `created_at` y luego `id`; rechazar uno anterior sin modificar el defecto.
- [ ] AC-018-09 — Respetar el contrato HTTP de resolución y consulta, validar
  identificadores/JSON y responder errores de infraestructura sin detalles internos.

## Confirmación final

Las reglas de negocio, el contrato HTTP y las tres interfaces de prueba
fueron confirmados por el usuario antes de escribir tests e implementar.

## Interfaces de prueba confirmadas

- Servicio público Go: transición, orden de Sprints, repetición y conflictos,
  sin HTTP ni base real; no demuestra JSON ni SQL.
- API HTTP con memoria: rutas, payloads, códigos de respuesta y consulta del
  resultado; no demuestra persistencia PostgreSQL.
- API HTTP con PostgreSQL real y esquemas aislados: persistencia, atomicidad,
  solicitudes concurrentes y conservación de defectos de HU-017 al actualizar
  el esquema; no incluye UI.

El esquema actual solo admite defectos abiertos. La implementación deberá
ampliarlo para guardar resolución y conservar los registros existentes;
`CREATE TABLE IF NOT EXISTS` por sí solo no actualiza ese constraint.

## Trazabilidad

- BDD: `features/HU-018-resolver-defecto.feature`.
- Tests previstos: `backend/internal/defects/` (pendientes).
- Código previsto: `backend/internal/defects/` (pendiente).
- PR y evidencia TDD: pendientes de implementación.
