# Trazabilidad

Índice Historia → SDD → criterios → BDD → tests → código.

Completar una fila por cada HU mergeada. La demo final se apoya en esta tabla.

| HU | Issue | SDD | Criterios | BDD | Tests | Código | PR | Evidencia TDD |
|---|---|---|---|---|---|---|---|---|
| HU-001 | [#10](https://github.com/FaustinoDuran/seguimiento-medicion/issues/10) | [SDD](sdd/HU-001-crear-proyecto.md) | AC-001-01 a AC-001-04 | [feature](../features/HU-001-crear-proyecto.feature) | `backend/internal/projects/service_test.go` | `backend/internal/projects/service.go` | [#36](https://github.com/FaustinoDuran/seguimiento-medicion/pull/36) | `test(#10): red` → `feat(#10): green` |
| HU-003 | [#12](https://github.com/FaustinoDuran/seguimiento-medicion/issues/12) | [SDD](sdd/HU-003-registrar-integrantes.md) | AC-003-01 a AC-003-04 | [feature](../features/HU-003-registrar-integrantes.feature) | `backend/internal/projects/service_test.go` | `backend/internal/projects/service.go` | [#37](https://github.com/FaustinoDuran/seguimiento-medicion/pull/37) | `test(#12): red` → `feat(#12): green` |
| HU-005 | [#14](https://github.com/FaustinoDuran/seguimiento-medicion/issues/14) | [SDD](sdd/HU-005-crear-historia-backlog.md) | AC-005-01 a AC-005-09 | [feature](../features/HU-005-crear-historia-backlog.feature) | `backend/internal/backlog/{service,handler,repository_postgres}_test.go` | `backend/internal/backlog/` | [#38](https://github.com/FaustinoDuran/seguimiento-medicion/pull/38) | RED `2c96efa`, `47db0ca`, `7d31919`, `dc75d2c`, `7852b2e` → GREEN `0a918d5`, `c7b9bbc`, `2f30a63`, `b8d50f7`, `bb66d7d` → REFACTOR `89dc48b` |
| HU-007 | [#16](https://github.com/FaustinoDuran/seguimiento-medicion/issues/16) | [SDD](sdd/HU-007-crear-sprint.md) | AC-007-01 a AC-007-03 | [feature](../features/HU-007-crear-sprint.feature) | `backend/internal/sprints/service_test.go` | `backend/internal/sprints/service.go` | [#39](https://github.com/FaustinoDuran/seguimiento-medicion/pull/39) | `test(#16): red` → `feat(#16): green` |
| HU-008 | [#17](https://github.com/FaustinoDuran/seguimiento-medicion/issues/17) | [SDD](sdd/HU-008-asignar-historia-sprint.md) | AC-008-01 a AC-008-04 | [feature](../features/HU-008-asignar-historia-sprint.feature) | `backend/internal/sprints/{service,handler,repository_postgres}_test.go` | `backend/internal/sprints/` | [#42](https://github.com/FaustinoDuran/seguimiento-medicion/pull/42); en revisión | RED `972c7e0`, `133fda5`, `9bc136f`, `7a47d6b`, `6b38a02` → GREEN `f7202ac`, `b6f6423`, `a104c7f`, `f089cc3`, `6f5652a` |
| HU-009 | [#18](https://github.com/FaustinoDuran/seguimiento-medicion/issues/18) | [SDD](sdd/HU-009-completar-historia.md) | AC-009-01 a AC-009-07 | [feature](../features/HU-009-completar-historia.feature) | `backend/internal/sprints/{service,handler,repository_postgres}_test.go`; `backend/internal/backlog/repository_postgres_test.go` | `backend/internal/sprints/`; puerto y persistencia de `backlog`; migración `006` | Pendiente de apertura | RED `45b344a`, `3e21321`, `d1fa167`, `1003650`, `baadf55`, `d979ff3` → GREEN `86425f7`, `77b82b3`, `e7fc515`, `f940656`, `5d6b5df`, `c5e7a98` → REFACTOR `82bf223` |

HU-001, HU-003, HU-005 y HU-007 están mergeadas en `main`. HU-008 permanece en revisión: el PR [#42](https://github.com/FaustinoDuran/seguimiento-medicion/pull/42) está abierto.

HU-009 está implementada en `feat/HU-009-completar-historia`, pendiente de revisión y merge.

Las tareas técnicas del Sprint 0 no se agregan a esta matriz porque no recorren
SDD/BDD. Se registran como Issues `TASK-XXX` y en `docs/sprints/sprint-0.md`.
