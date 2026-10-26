# ADR 0003 — Alcance del MVP (Sprint 1)

- Estado: aceptada
- Fecha: 2026-10-26
- Autores: Faustino, Valentino, Ignacio
- Issue: [#9 TASK-008](https://github.com/FaustinoDuran/seguimiento-medicion/issues/9)

## Contexto

La consigna pide en el Sprint 1 una versión funcional básica, no el producto
completo. El Product Backlog ya tiene HU-001 a HU-022. Hay que fijar qué entra
en este incremento para poder planificar y demostrar un ciclo mínimo, sin
adelantar Planning Poker, métricas ni UI.

Este recorte es el **incremento del Sprint 1**, no el cumplimiento total del TP.

## Decisión

Entran al Sprint 1 estas historias, en este orden de dependencia:

| HU | Issue | Por qué entra |
|---|---|---|
| HU-001 Crear proyecto | [#10](https://github.com/FaustinoDuran/seguimiento-medicion/issues/10) | Raíz del resto. |
| HU-003 Registrar integrantes | [#12](https://github.com/FaustinoDuran/seguimiento-medicion/issues/12) | Sin personas no hay quién vote ni cargue horas después. |
| HU-005 Crear historia de backlog | [#14](https://github.com/FaustinoDuran/seguimiento-medicion/issues/14) | Unidad de estimación, sprint, esfuerzo y defectos. |
| HU-007 Crear Sprint | [#16](https://github.com/FaustinoDuran/seguimiento-medicion/issues/16) | Sin iteración no hay seguimiento. |
| HU-008 Asignar historia a un Sprint | [#17](https://github.com/FaustinoDuran/seguimiento-medicion/issues/17) | Un sprint sin historias no se puede demostrar. |
| HU-009 Completar historia | [#18](https://github.com/FaustinoDuran/seguimiento-medicion/issues/18) | Cierra el avance real del sprint. |
| HU-010 Cerrar Sprint | [#19](https://github.com/FaustinoDuran/seguimiento-medicion/issues/19) | Congela planificado vs completado; deja el ciclo Scrum demostrable. |

**Demo de la Review:** un proyecto, al menos dos integrantes, dos historias de
backlog, un Sprint con goal, asignar las historias, completar al menos una y
cerrar el Sprint. API/JSON alcanza; la UI usable es Sprint 2.

### Fuera de este incremento

- HU-002 y HU-004 (modificar / consultar estado): se toman si sobra capacidad.
- HU-006, HU-011.
- HU-012 a HU-015 (Planning Poker y SP acordado).
- HU-016 (esfuerzo).
- HU-017 y HU-018 (defectos).
- HU-019 a HU-022 (métricas, dashboard, reporte, PDF).

Valor y estimación de cada HU se acuerdan en el planning del Sprint 1, no en
este ADR.

## Consecuencias

- El Sprint 1 se planifica solo contra estas siete issues.
- Cada HU recorre SDD → BDD → TDD → código Go antes del merge.
- React no entra en este incremento.
- Tablas y repositorios de `projects`, `backlog` y `sprints` nacen con estas HUs.
- Planning Poker, esfuerzo y métricas quedan para sprints siguientes, encima de
  este modelo.
