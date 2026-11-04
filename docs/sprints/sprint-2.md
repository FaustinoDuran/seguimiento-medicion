# Sprint 2 — La interfaz

- Estado: en curso
- Fecha de inicio: 2026-10-31
- Equipo: Faustino (Agile Enabler), Valentino e Ignacio (Product Builders)
- Alcance: [ADR 0004](../adr/0004-alcance-sprint-2.md)
- Tablero: [Seguimiento y medición](https://github.com/users/FaustinoDuran/projects/3)
- Sprint Goal: repetir la demo del Sprint 1 desde la UI.

## Precondición (Sprint 1)

Cumplida. HU-009 ([#18](https://github.com/FaustinoDuran/seguimiento-medicion/issues/18), PR [#51](https://github.com/FaustinoDuran/seguimiento-medicion/pull/51)) y HU-010 ([#19](https://github.com/FaustinoDuran/seguimiento-medicion/issues/19), PR [#52](https://github.com/FaustinoDuran/seguimiento-medicion/pull/52)) están en `main`.

## Avance

Coincide con el código en `main`: TASK-010, TASK-011, HU-011, HU-023 ([#58](https://github.com/FaustinoDuran/seguimiento-medicion/pull/58)), HU-026 y HU-027 están hechas. Sigue HU-024, la pantalla de proyectos.

## Planning

Estimación = esfuerzo relativo (Fibonacci, campo `SP` del Project). **SP
vacío** hasta acordarlos en planning. Valor de negocio sigue vacío.

Dueños:

| Ítem | Issue | Prioridad | SP | Responsable | Orden |
|---|---|---|---|---|---|
| TASK-010 Alcance Sprint 2 | [#43](https://github.com/FaustinoDuran/seguimiento-medicion/issues/43) | alta | — | Faustino | Done. PR [#50](https://github.com/FaustinoDuran/seguimiento-medicion/pull/50) |
| TASK-011 Shell frontend | [#44](https://github.com/FaustinoDuran/seguimiento-medicion/issues/44) | alta | — | Ignacio | Done. PR [#55](https://github.com/FaustinoDuran/seguimiento-medicion/pull/55) |
| HU-011 Consultar sprints anteriores | [#20](https://github.com/FaustinoDuran/seguimiento-medicion/issues/20) | media | — | Valentino | Done. PR [#56](https://github.com/FaustinoDuran/seguimiento-medicion/pull/56) |
| HU-023 Listar proyectos | [#45](https://github.com/FaustinoDuran/seguimiento-medicion/issues/45) | alta | — | Faustino | Done. PR [#58](https://github.com/FaustinoDuran/seguimiento-medicion/pull/58) |
| HU-024 UI proyectos | [#46](https://github.com/FaustinoDuran/seguimiento-medicion/issues/46) | alta | — | Faustino | En revisión. PR [#71](https://github.com/FaustinoDuran/seguimiento-medicion/pull/71) |
| HU-025 UI integrantes | [#47](https://github.com/FaustinoDuran/seguimiento-medicion/issues/47) | alta | — | Faustino | En revisión |
| HU-026 UI backlog | [#48](https://github.com/FaustinoDuran/seguimiento-medicion/issues/48) | alta | — | Ignacio | Done. PR [#59](https://github.com/FaustinoDuran/seguimiento-medicion/pull/59) |
| HU-027 UI sprints | [#49](https://github.com/FaustinoDuran/seguimiento-medicion/issues/49) | media | — | Valentino | Done. PR [#66](https://github.com/FaustinoDuran/seguimiento-medicion/pull/66) |

## Olas

1. TASK-010: Done.
2. TASK-011 y HU-011: Done.
3. HU-023: Done.
4. HU-024 (UI proyectos): en revisión.
5. HU-025: en revisión.
6. HU-026 y HU-027: Done.

## Cómo tomar una HU

1. Confirmar que está asignada y en Sprint 2.
2. Mover a `In progress`.
3. Escribir SDD **antes** de implementar.
4. BDD Gherkin. Si es API (HU-023): TDD en Go. Si es UI: React de presentación.
5. Un PR por HU, con `Closes #N` (sin espacio). Otra persona **Aprueba** el PR.

TASK-011 no lleva SDD/BDD.

## Demo de la Review

Un proyecto, al menos dos integrantes, dos historias de backlog, un Sprint con
goal, asignar, completar al menos una y cerrar el Sprint. Todo operado desde
la UI.

## Fuera de este incremento

HU-002, HU-004, HU-006 (si sobra capacidad: 002 y 006). HU-012 a HU-022.
