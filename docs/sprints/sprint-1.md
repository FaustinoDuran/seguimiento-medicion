# Sprint 1 — El MVP

- Estado: en curso
- Fecha de inicio: 2026-10-26
- Equipo: Faustino (Agile Enabler), Valentino e Ignacio (Product Builders)
- Alcance: [ADR 0003](../adr/0003-alcance-mvp.md)
- Tablero: [Seguimiento y medición](https://github.com/users/FaustinoDuran/projects/3)
- Sprint Goal: demostrar un ciclo Scrum mínimo persistido (API/JSON). La UI usable queda para el Sprint 2.

## Planning

Estimación = esfuerzo relativo (Fibonacci, campo `SP` del Project). Valor de
negocio sigue vacío: no bloquea este incremento.

Capacidad planificada: **24 SP**.

| HU | Issue | Prioridad | SP | Responsable | Módulo | Orden |
|---|---|---|---|---|---|---|
| HU-001 Crear proyecto | [#10](https://github.com/FaustinoDuran/seguimiento-medicion/issues/10) | alta | 5 | Faustino | `projects` | 1.º — desbloquea el resto |
| HU-003 Registrar integrantes | [#12](https://github.com/FaustinoDuran/seguimiento-medicion/issues/12) | alta | 3 | Faustino | `projects` | Después de HU-001 |
| HU-005 Crear historia de backlog | [#14](https://github.com/FaustinoDuran/seguimiento-medicion/issues/14) | alta | 5 | Ignacio | `backlog` | En paralelo tras HU-001 |
| HU-007 Crear Sprint | [#16](https://github.com/FaustinoDuran/seguimiento-medicion/issues/16) | alta | 3 | Valentino | `sprints` | En paralelo tras HU-001 |
| HU-008 Asignar historia a un Sprint | [#17](https://github.com/FaustinoDuran/seguimiento-medicion/issues/17) | media | 3 | Valentino | `sprints` | Después de HU-005 y HU-007 |
| HU-009 Completar historia | [#18](https://github.com/FaustinoDuran/seguimiento-medicion/issues/18) | media | 2 | Ignacio | `sprints` | Después de HU-008 |
| HU-010 Cerrar Sprint | [#19](https://github.com/FaustinoDuran/seguimiento-medicion/issues/19) | media | 3 | Faustino | `sprints` | Último, para la demo |

## Cómo tomar una HU

1. Confirmar que está asignada y en Sprint 1.
2. Mover la tarjeta a `In progress`.
3. Escribir SDD (`docs/sdd/HU-XXX-*.md`) **antes** de tests y código.
4. Escenarios BDD → TDD en Go (commits RED / GREEN / REFACTOR separados).
5. Un PR por HU, con `Closes #N` (sin espacio). React no entra en este sprint.

## Demo de la Review

Un proyecto, al menos dos integrantes, dos historias de backlog, un Sprint con
goal, asignar las historias, completar al menos una y cerrar el Sprint.

## Fuera de este incremento

HU-002, HU-004 (si sobra capacidad), HU-006, HU-011 a HU-022.
