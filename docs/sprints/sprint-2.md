# Sprint 2 — La interfaz

- Estado: en curso
- Fecha de inicio: 2026-10-31
- Equipo: Faustino (Agile Enabler), Valentino e Ignacio (Product Builders)
- Alcance: [ADR 0004](../adr/0004-alcance-sprint-2.md)
- Tablero: [Seguimiento y medición](https://github.com/users/FaustinoDuran/projects/3)
- Sprint Goal: repetir la demo del Sprint 1 desde la UI.

## Precondición (Sprint 1)

Cumplida. HU-009 ([#18](https://github.com/FaustinoDuran/seguimiento-medicion/issues/18), PR [#51](https://github.com/FaustinoDuran/seguimiento-medicion/pull/51)) y HU-010 ([#19](https://github.com/FaustinoDuran/seguimiento-medicion/issues/19), PR [#52](https://github.com/FaustinoDuran/seguimiento-medicion/pull/52)) están en `main`.

## Planning

Estimación = esfuerzo relativo (Fibonacci, campo `SP` del Project). **SP
vacío** hasta acordarlos en planning. Valor de negocio sigue vacío.

Dueños:

| Ítem | Issue | Prioridad | SP | Responsable | Orden |
|---|---|---|---|---|---|
| TASK-010 Alcance Sprint 2 | [#43](https://github.com/FaustinoDuran/seguimiento-medicion/issues/43) | alta | — | Faustino | Done |
| TASK-011 Shell frontend | [#44](https://github.com/FaustinoDuran/seguimiento-medicion/issues/44) | alta | — | Ignacio | Ahora, en paralelo con HU-023 |
| HU-023 Listar proyectos | [#45](https://github.com/FaustinoDuran/seguimiento-medicion/issues/45) | alta | — | Faustino | Ahora, API, en paralelo con TASK-011 |
| HU-024 UI proyectos | [#46](https://github.com/FaustinoDuran/seguimiento-medicion/issues/46) | alta | — | Faustino | Después de TASK-011 y HU-023 |
| HU-025 UI integrantes | [#47](https://github.com/FaustinoDuran/seguimiento-medicion/issues/47) | alta | — | Faustino | Después de HU-024 |
| HU-026 UI backlog | [#48](https://github.com/FaustinoDuran/seguimiento-medicion/issues/48) | alta | — | Ignacio | Después de HU-024 |
| HU-011 Consultar sprints anteriores | [#20](https://github.com/FaustinoDuran/seguimiento-medicion/issues/20) | media | — | Valentino | Ahora SDD; código Go (HU-010 ya está en `main`) |
| HU-027 UI sprints | [#49](https://github.com/FaustinoDuran/seguimiento-medicion/issues/49) | media | — | Valentino | Último, para la demo |

## Olas

1. TASK-010: Done.
2. TASK-011 (Nacho) y HU-023 (Faustino) en paralelo. Tino escribe SDD de HU-011.
3. HU-024 (UI proyectos).
4. HU-025 y HU-026 en paralelo; HU-011 en Go.
5. HU-027: crear sprint, asignar, completar, cerrar y ver anteriores.

## Cómo tomar una HU

1. Confirmar que está asignada y en Sprint 2.
2. Mover a `In progress`.
3. Escribir SDD **antes** de implementar.
4. BDD Gherkin. Si es API (HU-023, HU-011): TDD en Go. Si es UI: React de presentación.
5. Un PR por HU, con `Closes #N` (sin espacio). Otra persona **Aprueba** el PR.

TASK-011 no lleva SDD/BDD.

## Demo de la Review

Un proyecto, al menos dos integrantes, dos historias de backlog, un Sprint con
goal, asignar, completar al menos una y cerrar el Sprint. Todo operado desde
la UI.

## Fuera de este incremento

HU-002, HU-004, HU-006 (si sobra capacidad: 002 y 006). HU-012 a HU-022.
