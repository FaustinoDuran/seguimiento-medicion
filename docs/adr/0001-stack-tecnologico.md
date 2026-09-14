# ADR 0001 — Stack tecnológico

- Estado: aceptada
- Fecha: 2026-09-14
- Autores: Faustino, Valentino, Ignacio

## Contexto

La consigna exige Go para el núcleo y las reglas de negocio. El equipo no había programado en Go; sí conoce PostgreSQL, React, TypeScript y Tailwind. La UI puede ser web.

## Decisión

- Backend: Go 1.24, chi, pgx/v5.
- Base de datos: PostgreSQL 16 (Docker Compose).
- Frontend: TypeScript, React, Tailwind, Vite. Arranca en el Sprint 2.
- Tests: paquete `testing` de Go. BDD en Gherkin (`features/`) y Godog cuando se automatice el primer flujo.
- Sin GORM. Sin autenticación hasta que una HU o el profesor la pidan.

## Consecuencias

- Las métricas, el Planning Poker y las validaciones de negocio se implementan y testean en Go.
- React no calcula velocidad, desviación ni reglas de votación.
- PostgreSQL se configurará mediante una tarea técnica separada antes de la
  primera HU que requiera persistencia.
