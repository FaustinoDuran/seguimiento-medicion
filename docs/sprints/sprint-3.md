# Sprint 3 — Planning Poker

- Estado: en curso
- Fecha de inicio: 2026-11-04
- Equipo: Faustino (Agile Enabler), Valentino e Ignacio (Product Builders)
- Alcance: [ADR 0005](../adr/0005-alcance-sprint-3.md)
- Tablero: [Seguimiento y medición](https://github.com/users/FaustinoDuran/projects/3)
- Sprint Goal: estimar una historia desde la UI y guardar el story point acordado.

## Precondición (Sprint 2)

El tablero del Sprint 2 está en Done. La pantalla de integrantes llega a
`main` con el PR [#75](https://github.com/FaustinoDuran/seguimiento-medicion/pull/75)
(el PR [#72](https://github.com/FaustinoDuran/seguimiento-medicion/pull/72) se unió a la rama de proyectos, que ya estaba en `main`).

## Planning

Estimación = esfuerzo relativo (Fibonacci, campo `SP` del Project). **SP
vacío** hasta acordarlos en planning. Valor de negocio sigue vacío.

Dueños:

| Ítem | Issue | Prioridad | SP | Responsable | Orden |
|---|---|---|---|---|---|
| TASK-017 Alcance Sprint 3 | [#73](https://github.com/FaustinoDuran/seguimiento-medicion/issues/73) | alta | — | Faustino | Done. PR #76 |
| HU-012 Votar en secreto | [#21](https://github.com/FaustinoDuran/seguimiento-medicion/issues/21) | alta | — | Faustino | Done. PR #77 |
| HU-013 Revelar votos | [#22](https://github.com/FaustinoDuran/seguimiento-medicion/issues/22) | media | — | Valentino | Done. PR #78 |
| HU-014 Nueva ronda de votación | [#23](https://github.com/FaustinoDuran/seguimiento-medicion/issues/23) | media | — | Ignacio | Done. PR #80 |
| HU-015 Registrar estimación acordada | [#24](https://github.com/FaustinoDuran/seguimiento-medicion/issues/24) | media | — | Ignacio | Done. PR #79 |
| HU-029 UI Planning Poker | [#74](https://github.com/FaustinoDuran/seguimiento-medicion/issues/74) | alta | — | Faustino | In review. PR #81 |

## Olas

1. TASK-017: este alcance.
2. HU-012: votar sin ver el voto de los demás.
3. HU-013 y, detrás, HU-014 y HU-015.
4. HU-029: la misma secuencia desde la pantalla.

## Cómo tomar una HU

1. Confirmar que está asignada y en Sprint 3.
2. Mover a `In progress`.
3. Escribir SDD **antes** de implementar.
4. BDD Gherkin. Si es API (HU-012 a HU-015): TDD en Go. Si es UI (HU-029): React de presentación.
5. Un PR por HU, con `Closes #N` (sin espacio). Otra persona **Aprueba** el PR.

TASK-017 no lleva SDD/BDD.

## Demo de la Review

Un proyecto, al menos dos integrantes y una historia. Cada uno vota sin ver
al otro, se revelan los votos, si no coinciden se abre otra ronda, y el story
point acordado queda en la historia. Todo operado desde la UI.

## Fuera de este incremento

HU-016, si sobra capacidad y solo la API. HU-002, HU-004, HU-006. HU-017 a HU-022.
