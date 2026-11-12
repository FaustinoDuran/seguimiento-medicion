# Sprint 5 — Backlog residual

- Estado: en curso
- Fecha de inicio: al cerrar el Sprint 4
- Equipo: Faustino (Agile Enabler), Valentino e Ignacio (Product Builders)
- Alcance: [ADR 0008](../adr/0008-alcance-sprint-5.md)
- Tablero: [Seguimiento y medición](https://github.com/users/FaustinoDuran/projects/3)
- Sprint Goal: corregir un proyecto y una historia desde la UI, y dejar las skills de agente en el repo.

## Precondición (Sprint 4)

Cumplida: esfuerzo, defectos, métricas, dashboard, reporte y PDF están en
`main` (PRs [#87](https://github.com/FaustinoDuran/seguimiento-medicion/pull/87)
a [#97](https://github.com/FaustinoDuran/seguimiento-medicion/pull/97), más
[#92](https://github.com/FaustinoDuran/seguimiento-medicion/pull/92) y
[#95](https://github.com/FaustinoDuran/seguimiento-medicion/pull/95)).

## Planning

Estimación = esfuerzo relativo (Fibonacci, campo `SP` del Project). **SP
vacío** hasta acordarlos en planning. Valor de negocio sigue vacío.

Dueños:

| Ítem | Issue | Prioridad | SP | Responsable | Orden |
|---|---|---|---|---|---|
| TASK-020 Alcance Sprint 5 | [#98](https://github.com/FaustinoDuran/seguimiento-medicion/issues/98) | alta | — | Faustino | In progress |
| HU-002 Modificar proyecto | [#11](https://github.com/FaustinoDuran/seguimiento-medicion/issues/11) | media | — | Valentino | Sprint ready |
| HU-006 Actualizar historia | [#15](https://github.com/FaustinoDuran/seguimiento-medicion/issues/15) | media | — | Ignacio | Sprint ready |
| TASK-012 Skills de agente | [#54](https://github.com/FaustinoDuran/seguimiento-medicion/issues/54) | alta | — | Faustino | Sprint ready |

## Olas

1. TASK-020: este alcance.
2. HU-002 en `projects` y, en otro paquete, HU-006 en `backlog`.
3. TASK-012 en paralelo: solo skills, sin código de producto.

## Cómo tomar una HU

1. Confirmar que está asignada y en Sprint 5, y que la precondición ya está en `main`.
2. Mover a `In progress`.
3. Escribir SDD **antes** de implementar.
4. BDD Gherkin. API: TDD en Go. UI: React de presentación.
5. Un PR por HU, con `Closes #N` (sin espacio). Otra persona **Aprueba** el PR.

TASK-020 y TASK-012 no llevan SDD/BDD.

## Demo de la Review

Un proyecto existente. Se cambia nombre o fechas, se actualiza una historia
del backlog, y las skills de agente quedan en el repositorio.

## Fuera de este incremento

HU-004 (el estado ya se consulta en el detalle). Auth, GORM y cambio de stack.
