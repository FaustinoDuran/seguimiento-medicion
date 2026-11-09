# ADR 0006 — Alcance del Sprint 4 (medición)

- Estado: aceptada
- Fecha: 2026-11-09
- Autores: Faustino, Valentino, Ignacio
- Issue: [#84 TASK-019](https://github.com/FaustinoDuran/seguimiento-medicion/issues/84)

## Contexto

El Sprint 3 deja Planning Poker operable desde la UI cuando entran a `main`
el PR [#81](https://github.com/FaustinoDuran/seguimiento-medicion/pull/81) y,
como precondición de equipo, el chequeo de backend del PR
[#83](https://github.com/FaustinoDuran/seguimiento-medicion/pull/83).

Lo que sigue del backlog es esfuerzo, defectos, métricas, dashboard, reporte
y PDF. HU-019 necesita las horas reales (HU-016) y, para los defectos
detectados y resueltos, HU-017 y HU-018. El dashboard, el reporte y el PDF
leen esas métricas; no las calculan de nuevo.

## Decisión

Entran al Sprint 4 estos ítems, en este orden de dependencia:

| Ítem | Issue | Por qué entra |
|---|---|---|
| TASK-019 Alcance Sprint 4 | [#84](https://github.com/FaustinoDuran/seguimiento-medicion/issues/84) | Fija el recorte. |
| HU-016 Registrar esfuerzo | [#25](https://github.com/FaustinoDuran/seguimiento-medicion/issues/25) | Guarda las horas reales. Sin esto no hay desviación. |
| HU-017 Registrar defecto | [#26](https://github.com/FaustinoDuran/seguimiento-medicion/issues/26) | Abre el seguimiento de calidad. |
| HU-018 Resolver defecto | [#27](https://github.com/FaustinoDuran/seguimiento-medicion/issues/27) | Cierra el defecto y deja el sprint de resolución. |
| HU-019 Consultar métricas | [#28](https://github.com/FaustinoDuran/seguimiento-medicion/issues/28) | Calcula los números en Go. |
| HU-030 UI esfuerzo y defectos | [#85](https://github.com/FaustinoDuran/seguimiento-medicion/issues/85) | Carga horas y defectos desde la pantalla. |
| HU-020 Dashboard | [#29](https://github.com/FaustinoDuran/seguimiento-medicion/issues/29) | Muestra el estado y un gráfico con métricas ya calculadas. |
| HU-021 Reporte | [#30](https://github.com/FaustinoDuran/seguimiento-medicion/issues/30) | Arma el informe de proyecto o sprint. |
| HU-022 PDF | [#31](https://github.com/FaustinoDuran/seguimiento-medicion/issues/31) | Exporta ese informe. |

**Sprint Goal / demo de la Review:** sobre un proyecto que ya tiene historias
estimadas y un sprint, cargar horas y un defecto desde la UI, ver las métricas
en el dashboard, generar el reporte y exportarlo a PDF.

Las reglas nuevas viven en Go: `effort`, `defects`, `metrics` y `reports`.
React solo presenta. Cada fórmula de HU-019 se escribe en su SDD antes del
código. No se inventa otro campo de estimación: el story point acordado ya
está en la historia.

### Fuera de este incremento

- HU-002 modificar proyecto.
- HU-004 consultar estado: ya se muestra en el detalle.
- HU-006 actualizar historia.
- TASK-012 skills de agente.

### Cómo no pisarse

HU-016 abre el paquete `effort`. HU-017 y HU-018 comparten `defects`, en ese
orden y con la misma persona. HU-019 abre `metrics` cuando esas tres están en
`main`. Las pantallas esperan la API que muestran. El reporte espera las
métricas. El PDF espera el reporte.

Dos pull requests no editan el mismo paquete a la vez. `App.tsx` se toca de a
un PR: se mergea uno y el siguiente parte de `main`.

## Consecuencias

- El Sprint 4 se planifica contra estas issues.
- TASK-019 no tiene SDD/BDD. HU-016 a HU-022 y HU-030 sí.
- Valor y estimación (SP) se acuerdan en el planning; no en este ADR.
- El dueño de cada ítem queda en `docs/sprints/sprint-4.md`.
