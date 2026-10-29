# Sprint 2 — La interfaz

- Estado: planificado
- Fecha de inicio: a confirmar en planning
- Equipo: Faustino (Agile Enabler), Valentino e Ignacio (Product Builders)
- Alcance: [ADR 0004](../adr/0004-alcance-sprint-2.md)
- Tablero: [Seguimiento y medición](https://github.com/users/FaustinoDuran/projects/3)
- Sprint Goal: repetir la demo del Sprint 1 desde la UI.

## Precondición (Sprint 1)

Siguen abiertas y **no** se mueven a Sprint 2:

- [#18 HU-009](https://github.com/FaustinoDuran/seguimiento-medicion/issues/18) Completar historia
- [#19 HU-010](https://github.com/FaustinoDuran/seguimiento-medicion/issues/19) Cerrar Sprint

HU-011 y HU-027 esperan a que esas dos estén en `main`.

## Planning

Estimación = esfuerzo relativo (Fibonacci, campo `SP` del Project). **SP
vacío** hasta la reunión de planning. Valor de negocio sigue vacío.

Dueños tentativos (a confirmar en planning):

| Ítem | Issue | Prioridad | SP | Responsable | Orden |
|---|---|---|---|---|---|
| TASK-010 Alcance Sprint 2 | [#43](https://github.com/FaustinoDuran/seguimiento-medicion/issues/43) | alta | — | Faustino | 1.º |
| TASK-011 Shell frontend | [#44](https://github.com/FaustinoDuran/seguimiento-medicion/issues/44) | alta | — | Ignacio | En paralelo con HU-023 |
| HU-023 Listar proyectos | [#45](https://github.com/FaustinoDuran/seguimiento-medicion/issues/45) | alta | — | Faustino | API, en paralelo con TASK-011 |
| HU-024 UI proyectos | [#46](https://github.com/FaustinoDuran/seguimiento-medicion/issues/46) | alta | — | Faustino | Después de TASK-011 y HU-023 |
| HU-025 UI integrantes | [#47](https://github.com/FaustinoDuran/seguimiento-medicion/issues/47) | alta | — | Faustino | Después de HU-024 |
| HU-026 UI backlog | [#48](https://github.com/FaustinoDuran/seguimiento-medicion/issues/48) | alta | — | Ignacio | Después de HU-024 |
| HU-011 Consultar sprints anteriores | [#20](https://github.com/FaustinoDuran/seguimiento-medicion/issues/20) | media | — | Valentino | Después de HU-010 |
| HU-027 UI sprints | [#49](https://github.com/FaustinoDuran/seguimiento-medicion/issues/49) | media | — | Valentino | Último, para la demo |

## Olas

1. TASK-010 (este recorte).
2. TASK-011 y HU-023 en paralelo. Nacho y Faustino cierran HU-009 y HU-010 del Sprint 1.
3. HU-024 (UI proyectos).
4. HU-025 y HU-026 en paralelo; HU-011 cuando exista HU-010.
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
