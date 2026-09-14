# ADR 0002 — Arquitectura feature-based

- Estado: aceptada
- Fecha: 2026-09-14
- Autores: Faustino, Valentino, Ignacio

## Contexto

Hay que mostrar modularidad y trazabilidad. Un layout MVC plano mezcla features y dificulta asignar HUs.

## Decisión

Monolito modular. Cada feature es un paquete en `backend/internal/<feature>/` con `model.go`, `service.go`, `repository.go`, `repository_postgres.go`, `handler.go`, `routes.go` y tests al lado.

Infra transversal: `config`, `httpx`, `apperrors`, `platform/postgres`.

El frontend replica features en `frontend/src/features/`.

La documentación de trazabilidad vive en `docs/` y `features/`, no dentro del paquete Go.

## Consecuencias

- Un módulo no importa handlers ni SQL de otro.
- Las carpetas vacías se conservan con `.gitkeep`; el primer archivo Go de cada
  paquete se crea únicamente al implementar una HU o tarea aprobada.
