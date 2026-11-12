---
name: pre-commit
description: Checks TP Integrador process rules before implementing, committing, or opening a PR. Business rules only in Go, React presentation only, Issue to SDD to BDD to TDD, Go 1.24 and pgx, no GORM, no auth until a HU asks. Use when committing, pushing, opening a pull request, implementing a HU or TASK, or when the user mentions pre-commit, TDD, or reglas del TP.
---

# Pre-commit del TP

Leé `docs/guia-de-trabajo.md` y `.cursor/rules/` si hay duda. No adelantes producto.

## Antes de escribir código

1. Hay Issue `HU-XXX`, `TASK-XXX` o `BUG-XXX`. Si no, parar.
2. HU: SDD en `docs/sdd/` **antes** de tests y código. Luego `features/*.feature`.
3. TASK: no inventar SDD/BDD. No adelantar la tarea si no está asignada.
4. Un PR por HU. `App.tsx` de a un PR. No mezclar módulos ajenos.

## Dónde va cada cosa

- Reglas: `backend/internal/<feature>/` (service). Handler sin reglas.
- UI: `frontend/src/features/<feature>/`. Solo presenta números de la API.
- Prohibido en React: velocidad, desviación, porcentajes de historias, reglas de Planning Poker, cierre de sprint.
- Stack: Go 1.24.x, pgx v5.7.x, chi, PostgreSQL. Sin GORM. Sin auth hasta que una HU lo pida. No subir `go` del `go.mod` a 1.25+.
- Código y archivos en inglés. Specs, issues y Gherkin en español.

## Commits

TDD separado, issue `#N` (número de GitHub, no el HU):

- `test(#N): red - …`
- `feat(#N): green - …`
- `refactor(#N)` si hubo diseño
- TASK: `chore(#N): …` o `docs(#N): …`

No juntar RED y GREEN. PR con `Closes #N` (sin espacio). Otra persona Aprueba.

## Checklist rápido

- [ ] Issue correcta y asignada
- [ ] HU: SDD y BDD antes del GREEN
- [ ] Reglas solo en Go
- [ ] React no calcula
- [ ] Commits RED/GREEN separados
- [ ] Trazabilidad en el mismo PR (`docs/traceability.md`)
- [ ] Sin GORM, sin auth inventada, sin bump de Go

Si algo falla, no commits ni PR hasta corregirlo.
