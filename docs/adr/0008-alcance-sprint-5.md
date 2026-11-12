# ADR 0008 — Alcance del Sprint 5 (backlog residual)

- Estado: aceptada
- Fecha: 2026-11-11
- Autores: Faustino, Valentino, Ignacio
- Issue: [#98 TASK-020](https://github.com/FaustinoDuran/seguimiento-medicion/issues/98)

## Contexto

El Sprint 4 dejó en `main` el ciclo de medición: esfuerzo, defectos (alta y
resolución), métricas, dashboard, reporte y PDF. No hay PRs abiertos. El
Product Backlog que sigue son tres historias de mantenimiento (HU-002, HU-004,
HU-006) y TASK-012, que los sprints anteriores dejaron afuera a propósito.

HU-004 ya se cubre: `GET /projects/{id}` y el detalle de la UI muestran el
estado y avisan si el proyecto no existe (ADR 0005). No se reabre.

## Decisión

Entran al Sprint 5 estos ítems, en este orden de dependencia:

| Ítem | Issue | Por qué entra |
|---|---|---|
| TASK-020 Alcance Sprint 5 | [#98](https://github.com/FaustinoDuran/seguimiento-medicion/issues/98) | Fija el recorte y cierra el acta del Sprint 4. |
| HU-002 Modificar proyecto | [#11](https://github.com/FaustinoDuran/seguimiento-medicion/issues/11) | Permite corregir nombre, fechas o estado. |
| HU-006 Actualizar historia | [#15](https://github.com/FaustinoDuran/seguimiento-medicion/issues/15) | Permite corregir prioridad, estado o story points de una historia. |
| TASK-012 Skills de agente | [#54](https://github.com/FaustinoDuran/seguimiento-medicion/issues/54) | Infraestructura de proceso; no toca producto. |

**Sprint Goal / demo de la Review:** sobre un proyecto existente, cambiar
datos del proyecto y de una historia desde la UI, y dejar las skills de
agente en el repo para el flujo Issue → SDD → BDD → TDD.

Las reglas nuevas viven en Go (`projects` y `backlog`). React solo presenta.
HU-006 no reemplaza Planning Poker: el story point acordado sigue saliendo
de HU-015; esta historia solo permite corregir campos de la historia.

### Fuera de este incremento

- HU-004 consultar estado: ya se muestra en el detalle.
- Auth, GORM, cambio de stack.
- Nuevas métricas o informes.

### Cómo no pisarse

HU-002 abre cambios en `projects`. HU-006 abre cambios en `backlog`. TASK-012
solo toca `.cursor/skills/` (y la guía). Dos pull requests no editan el mismo
paquete a la vez. `App.tsx` se toca de a un PR.

## Consecuencias

- El Sprint 5 se planifica contra estas issues.
- TASK-020 y TASK-012 no tienen SDD/BDD. HU-002 y HU-006 sí.
- Valor y estimación (SP) se acuerdan en el planning; no en este ADR.
- El dueño de cada ítem queda en `docs/sprints/sprint-5.md`.
