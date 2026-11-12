---
name: plan-de-accion
description: Classifies a GitHub ticket (HU, TASK, BUG, or other) and writes a concrete action plan from repo docs and code. Use when the user pastes an issue, names HU-XXX, TASK-XXX, BUG-XXX, a #N ticket, or asks for a plan de acción, qué hacer, or next steps on a card.
---

# Plan de acción

No codees hasta entregar el plan y, si es HU, hasta tener SDD (y BDD).

## 1. Clasificar el ticket

Leer título, labels y cuerpo (Issue o el texto que pasó la persona).

| Señal | Tipo |
|---|---|
| `[HU-…]`, label `historia` | HU |
| `[TASK-…]`, label `tarea técnica` | TASK |
| `[BUG-…]`, label defecto | BUG |
| refactor / docs sueltos | otro — pedir Issue TASK o HU antes de tocar producto |

Si no hay Issue: parar. No inventar HU para infra.

## 2. Leer el repo (en este orden)

1. `docs/guia-de-trabajo.md`, `AGENTS.md`, `.cursor/rules/proyecto.mdc`
2. Acta vigente: `docs/sprints/sprint-*.md` más reciente y su ADR
3. Issue `#N` (`gh issue view N`) y tablero (asignado, Sprint, estado)
4. Si es HU: `docs/sdd/HU-XXX-*.md`, `features/HU-XXX-*.feature`, fila en `docs/traceability.md`
5. Código del módulo: `backend/internal/<modulo>/` y, si hay UI, `frontend/src/features/<modulo>/`
6. Dependencias nombradas en la SDD o el ADR (qué tiene que estar en `main`)

Si falta SDD/BDD de una HU, el plan empieza por escribirlos. No tests ni implementación antes.

## 3. Entregar el plan (plantilla)

```markdown
## Tipo
HU | TASK | BUG | otro

## Ticket
HU-XXX / TASK-XXX / BUG-XXX · #N · dueño · estado del tablero

## Qué hay
- Issue / SDD / BDD / código / PR (sí/no + ruta)

## Qué falta
- eslabones rotos de la cadena

## Pasos
1. …
2. …

## Archivos
- paths concretos

## Commits
- `test(#N): red - …`
- `feat(#N): green - …`
(o `chore(#N)` / `docs(#N)` si es TASK)

## Qué no hacer
- no calcular métricas ni Planning Poker en React
- no GORM / auth / bump de Go
- no un PR con varias HU
- no SDD/BDD inventados en una TASK
```

## 4. Después del plan

HU: SDD → BDD → RED → GREEN → trazabilidad → PR `Closes #N`.
TASK: criterios de la Issue → PR. Sin SDD/BDD.
BUG: test que reproduce → arreglo → PR apuntando a la HU afectada.
