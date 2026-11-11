# Sprint 4 — Medición

- Estado: cerrado
- Fecha de inicio: al cerrar el Sprint 3
- Fecha de fin: 2026-11-11
- Equipo: Faustino (Agile Enabler), Valentino e Ignacio (Product Builders)
- Alcance: [ADR 0006](../adr/0006-alcance-sprint-4.md)
- Tablero: [Seguimiento y medición](https://github.com/users/FaustinoDuran/projects/3)
- Sprint Goal: cargar horas y un defecto, ver las métricas, y exportar el informe a PDF.

## Precondición (Sprint 3)

Cumplida: el Planning Poker de la UI
([#81](https://github.com/FaustinoDuran/seguimiento-medicion/pull/81)) y el
chequeo de backend
([#83](https://github.com/FaustinoDuran/seguimiento-medicion/pull/83))
están en `main`.

## Planning

Estimación = esfuerzo relativo (Fibonacci, campo `SP` del Project). **SP
vacío** hasta acordarlos en planning. Valor de negocio sigue vacío.

Dueños:

| Ítem | Issue | Prioridad | SP | Responsable | Orden |
|---|---|---|---|---|---|
| TASK-019 Alcance Sprint 4 | [#84](https://github.com/FaustinoDuran/seguimiento-medicion/issues/84) | alta | — | Faustino | Done. PR [#86](https://github.com/FaustinoDuran/seguimiento-medicion/pull/86) |
| HU-016 Registrar esfuerzo | [#25](https://github.com/FaustinoDuran/seguimiento-medicion/issues/25) | media | — | Valentino | Done. PR [#87](https://github.com/FaustinoDuran/seguimiento-medicion/pull/87) |
| HU-017 Registrar defecto | [#26](https://github.com/FaustinoDuran/seguimiento-medicion/issues/26) | media | — | Ignacio | Done. PR [#88](https://github.com/FaustinoDuran/seguimiento-medicion/pull/88) |
| HU-018 Resolver defecto | [#27](https://github.com/FaustinoDuran/seguimiento-medicion/issues/27) | media | — | Ignacio | Done. PR [#89](https://github.com/FaustinoDuran/seguimiento-medicion/pull/89) recuperado en `main` por el PR [#92](https://github.com/FaustinoDuran/seguimiento-medicion/pull/92) |
| HU-019 Consultar métricas | [#28](https://github.com/FaustinoDuran/seguimiento-medicion/issues/28) | media | — | Faustino | Done. PR [#90](https://github.com/FaustinoDuran/seguimiento-medicion/pull/90) |
| HU-030 UI esfuerzo y defectos | [#85](https://github.com/FaustinoDuran/seguimiento-medicion/issues/85) | alta | — | Valentino | Done. PR [#93](https://github.com/FaustinoDuran/seguimiento-medicion/pull/93) |
| HU-020 Dashboard | [#29](https://github.com/FaustinoDuran/seguimiento-medicion/issues/29) | media | — | Faustino | Done. PR [#91](https://github.com/FaustinoDuran/seguimiento-medicion/pull/91) |
| HU-021 Reporte | [#30](https://github.com/FaustinoDuran/seguimiento-medicion/issues/30) | media | — | Ignacio | Done. PR [#96](https://github.com/FaustinoDuran/seguimiento-medicion/pull/96) |
| HU-022 PDF | [#31](https://github.com/FaustinoDuran/seguimiento-medicion/issues/31) | baja | — | Ignacio | Done. PR [#97](https://github.com/FaustinoDuran/seguimiento-medicion/pull/97) |
| BUG-001 Conteos de resolución | [#94](https://github.com/FaustinoDuran/seguimiento-medicion/issues/94) | alta | — | Ignacio | Done. PR [#95](https://github.com/FaustinoDuran/seguimiento-medicion/pull/95) |

## Olas

1. TASK-019: este alcance.
2. HU-016 y, en otro paquete, HU-017 seguida de HU-018.
3. HU-019 cuando esas tres están en `main`.
4. Pantallas: HU-030 y HU-020. HU-021 después de las métricas, y HU-022 al final.

## Cómo tomar una HU

1. Confirmar que está asignada y en Sprint 4, y que la precondición ya está en `main`.
2. Mover a `In progress`.
3. Escribir SDD **antes** de implementar.
4. BDD Gherkin. Si es API: TDD en Go. Si es UI (HU-030 y HU-020): React de presentación.
5. Un PR por HU, con `Closes #N` (sin espacio). Otra persona **Aprueba** el PR.

TASK-019 no lleva SDD/BDD.

## Demo de la Review

Un proyecto con historias estimadas y un sprint. Desde la UI se cargan horas
y un defecto, se resuelve el defecto, el dashboard muestra las métricas, y el
reporte se exporta a PDF.

## Fuera de este incremento

HU-002, HU-004, HU-006 y TASK-012.
