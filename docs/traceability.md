# Trazabilidad

Índice Historia → SDD → criterios → BDD → tests → código.

Completar una fila por cada HU mergeada. La demo final se apoya en esta tabla.

| HU | Issue | SDD | Criterios | BDD | Tests | Código | PR | Evidencia TDD |
|---|---|---|---|---|---|---|---|---|
| HU-001 | [#10](https://github.com/FaustinoDuran/seguimiento-medicion/issues/10) | [SDD](sdd/HU-001-crear-proyecto.md) | AC-001-01 a AC-001-04 | [feature](../features/HU-001-crear-proyecto.feature) | `backend/internal/projects/service_test.go` | `backend/internal/projects/service.go` | [#36](https://github.com/FaustinoDuran/seguimiento-medicion/pull/36) | `test(#10): red` → `feat(#10): green` |
| HU-003 | [#12](https://github.com/FaustinoDuran/seguimiento-medicion/issues/12) | [SDD](sdd/HU-003-registrar-integrantes.md) | AC-003-01 a AC-003-04 | [feature](../features/HU-003-registrar-integrantes.feature) | `backend/internal/projects/service_test.go` | `backend/internal/projects/service.go` | [#37](https://github.com/FaustinoDuran/seguimiento-medicion/pull/37) | `test(#12): red` → `feat(#12): green` |
| HU-005 | [#14](https://github.com/FaustinoDuran/seguimiento-medicion/issues/14) | [SDD](sdd/HU-005-crear-historia-backlog.md) | AC-005-01 a AC-005-09 | [feature](../features/HU-005-crear-historia-backlog.feature) | `backend/internal/backlog/{service,handler,repository_postgres}_test.go` | `backend/internal/backlog/` | [Abrir PR](https://github.com/FaustinoDuran/seguimiento-medicion/compare/main...feat/HU-005-crear-historia-backlog?expand=1); pendiente | RED `2c96efa`, `47db0ca`, `7d31919`, `dc75d2c`, `7852b2e` → GREEN `0a918d5`, `c7b9bbc`, `2f30a63`, `b8d50f7`, `bb66d7d` → REFACTOR `89dc48b` |

HU-005 permanece pendiente de revisión y merge; su implementación en rama no implica estado Done.
La rama está publicada. La creación del PR mediante la integración devolvió
`403: Resource not accessible by integration`; falta abrirlo con permisos de escritura sobre pull requests.

Las tareas técnicas del Sprint 0 no se agregan a esta matriz porque no recorren
SDD/BDD. Se registran como Issues `TASK-XXX` y en `docs/sprints/sprint-0.md`.
