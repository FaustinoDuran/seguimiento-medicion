# ADR 0005 — Alcance del Sprint 3 (Planning Poker)

- Estado: aceptada
- Fecha: 2026-11-04
- Autores: Faustino, Valentino, Ignacio
- Issue: [#73 TASK-017](https://github.com/FaustinoDuran/seguimiento-medicion/issues/73)

## Contexto

El Sprint 2 dejó el ciclo Scrum operable desde la UI. El backlog que sigue
es Planning Poker, esfuerzo, defectos, métricas y el informe. Esas historias
no entran todas en un incremento.

HU-012 es la única que queda con prioridad alta. HU-013, HU-014 y HU-015
dependen de ella. Las métricas (HU-019) necesitan además las horas reales
(HU-016).

## Decisión

Entran al Sprint 3 estos ítems, en este orden de dependencia:

| Ítem | Issue | Por qué entra |
|---|---|---|
| TASK-017 Alcance Sprint 3 | [#73](https://github.com/FaustinoDuran/seguimiento-medicion/issues/73) | Fija el recorte. |
| HU-012 Votar en secreto | [#21](https://github.com/FaustinoDuran/seguimiento-medicion/issues/21) | Abre la ronda. Sin esto no hay estimación. |
| HU-013 Revelar votos | [#22](https://github.com/FaustinoDuran/seguimiento-medicion/issues/22) | Muestra los votos al cerrar la ronda. |
| HU-014 Nueva ronda de votación | [#23](https://github.com/FaustinoDuran/seguimiento-medicion/issues/23) | Si los votos difieren, se puede votar de nuevo. |
| HU-015 Registrar estimación acordada | [#24](https://github.com/FaustinoDuran/seguimiento-medicion/issues/24) | Guarda el story point en la historia. |
| HU-029 UI Planning Poker | [#74](https://github.com/FaustinoDuran/seguimiento-medicion/issues/74) | Hacer todo lo anterior desde la pantalla. |

**Sprint Goal / demo de la Review:** con un proyecto, al menos dos integrantes
y una historia, cada uno vota sin ver al otro, se revelan los votos, si no
coinciden se abre otra ronda, y el story point acordado queda en la historia.
Todo operado desde la UI.

Las reglas viven en Go, módulo `planningpoker`. React solo presenta.

### Fuera de este incremento

- HU-016 (esfuerzo): si sobra capacidad, solo la API.
- HU-002, HU-004 y HU-006.
- HU-017 y HU-018 (defectos).
- HU-019 a HU-022 (métricas, dashboard, reporte, PDF).

HU-004 no se reabre: el estado del proyecto ya se consulta y se muestra en el
detalle. El story point de la historia ya existe; HU-015 lo escribe con el
valor acordado.

## Consecuencias

- El Sprint 3 se planifica contra estas issues.
- TASK-017 no tiene SDD/BDD. HU-012 a HU-015 sí, con TDD en Go. HU-029 sí, en React.
- Valor y estimación (SP) se acuerdan en el planning; no en este ADR.
- El dueño de cada ítem se confirma en `docs/sprints/sprint-3.md`.
