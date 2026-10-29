# ADR 0004 — Alcance del Sprint 2 (interfaz usable)

- Estado: aceptada
- Fecha: 2026-10-28
- Autores: Faustino, Valentino, Ignacio
- Issue: [#43 TASK-010](https://github.com/FaustinoDuran/seguimiento-medicion/issues/43)

## Contexto

La consigna pide en el Sprint 2 una interfaz usable, no el producto completo.
El Sprint 1 dejó el ciclo Scrum mínimo en API (HU-001 a HU-010). El frontend
sigue en el placeholder de TASK-003. No hay `GET /projects` para armar un home.

Este recorte es el **incremento del Sprint 2**. Planning Poker, esfuerzo,
defectos, métricas y PDF quedan para sprints siguientes.

## Decisión

Entran al Sprint 2 estos ítems, en este orden de dependencia:

| Ítem | Issue | Por qué entra |
|---|---|---|
| TASK-010 Alcance Sprint 2 | [#43](https://github.com/FaustinoDuran/seguimiento-medicion/issues/43) | Fija el recorte. |
| TASK-011 Shell frontend | [#44](https://github.com/FaustinoDuran/seguimiento-medicion/issues/44) | Rutas, layout y cliente HTTP. Sin reglas. |
| HU-023 Listar proyectos | [#45](https://github.com/FaustinoDuran/seguimiento-medicion/issues/45) | API que falta para el home. TDD en Go. |
| HU-024 UI proyectos | [#46](https://github.com/FaustinoDuran/seguimiento-medicion/issues/46) | Crear, listar y ver detalle desde la UI. |
| HU-025 UI integrantes | [#47](https://github.com/FaustinoDuran/seguimiento-medicion/issues/47) | Alta y listado sobre un proyecto. |
| HU-026 UI backlog | [#48](https://github.com/FaustinoDuran/seguimiento-medicion/issues/48) | Crear y ver historias. |
| HU-011 Consultar sprints anteriores | [#20](https://github.com/FaustinoDuran/seguimiento-medicion/issues/20) | Listar sprints del proyecto; la UI de sprints lo necesita. |
| HU-027 UI sprints | [#49](https://github.com/FaustinoDuran/seguimiento-medicion/issues/49) | Crear, asignar, completar y cerrar desde la UI. |

**Sprint Goal / demo de la Review:** repetir la demo del Sprint 1 **desde la
UI**: un proyecto, al menos dos integrantes, dos historias, un Sprint con
goal, asignar, completar al menos una y cerrar.

HU-009 y HU-010 siguen en Sprint 1 hasta Done. HU-027 espera esas APIs.

### Fuera de este incremento

- HU-002 y HU-006 (modificar proyecto / actualizar historia): si sobra capacidad.
- HU-004 (consultar estado): el GET por id ya existe en HU-001; la UI de detalle
  queda en HU-024.
- HU-012 a HU-015 (Planning Poker).
- HU-016 (esfuerzo).
- HU-017 y HU-018 (defectos).
- HU-019 a HU-022 (métricas, dashboard, reporte, PDF).

No se reabren HUs Done del Sprint 1. Las pantallas son HUs nuevas.

### Cómo se construye una HU de UI

Issue → SDD (pantalla, llamadas a API, vacío y error) → BDD Gherkin → React.
Sin TDD en Go salvo que falte una regla: entonces primero la HU o TASK de
backend (como HU-023). React no calcula métricas ni Planning Poker.

## Consecuencias

- El Sprint 2 se planifica contra estas issues.
- TASK-011 no tiene SDD/BDD; las HUs de UI sí.
- Valor y estimación (SP) se acuerdan en el planning; no en este ADR.
- El dueño de cada ítem se confirma en `docs/sprints/sprint-2.md`.
